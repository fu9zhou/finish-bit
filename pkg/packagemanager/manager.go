package packagemanager

import (
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

type Manager struct {
	root     string
	registry Registry
	client   *http.Client
}

type Installed struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	License     string            `json:"license"`
	Source      string            `json:"source"`
	Platform    string            `json:"platform"`
	Installed   time.Time         `json:"installed"`
	Executables map[string]string `json:"executables"`
}

type Status struct {
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	Description string     `json:"description"`
	License     string     `json:"license"`
	Source      string     `json:"source"`
	Platform    string     `json:"platform"`
	Supported   bool       `json:"supported"`
	Installed   *Installed `json:"installed,omitempty"`
}

func DefaultRoot() (string, error) {
	if override := os.Getenv("FINISHBIT_HOME"); override != "" {
		return filepath.Abs(override)
	}
	config, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locate user config directory: %w", err)
	}
	return filepath.Join(config, "finishbit"), nil
}

func New(root string, registry Registry) *Manager {
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ForceAttemptHTTP2:     false,
		DisableCompression:    true,
		ResponseHeaderTimeout: 30 * time.Second,
		IdleConnTimeout:       90 * time.Second,
	}
	return &Manager{root: root, registry: registry, client: &http.Client{Timeout: 30 * time.Minute, Transport: transport}}
}

func (m *Manager) Root() string { return m.root }

func PlatformKey() string {
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x64"
	}
	platform := runtime.GOOS
	if platform == "windows" {
		platform = "win32"
	}
	return platform + "-" + arch
}

func (m *Manager) Install(ctx context.Context, name string) (Installed, error) {
	pkg, ok := m.registry.Find(name)
	if !ok {
		return Installed{}, &operation.Error{Code: operation.CodeInvalidInput, Message: fmt.Sprintf("unknown package %q", name)}
	}
	platform := PlatformKey()
	artifact, ok := pkg.Artifacts[platform]
	if !ok {
		return Installed{}, &operation.Error{Code: operation.CodeUnsupportedPlatform, Message: fmt.Sprintf("package %q does not support %s", name, platform)}
	}
	if err := validateArtifact(artifact); err != nil {
		return Installed{}, err
	}
	if installed, err := m.Info(name); err == nil && installed.Version == pkg.Version && installed.Platform == platform {
		return installed, nil
	}
	if err := os.MkdirAll(filepath.Join(m.root, "packages", name), 0o755); err != nil {
		return Installed{}, fmt.Errorf("create package directory: %w", err)
	}
	staging, err := os.MkdirTemp(filepath.Join(m.root, "packages", name), ".install-")
	if err != nil {
		return Installed{}, fmt.Errorf("create package staging directory: %w", err)
	}
	defer os.RemoveAll(staging)
	archivePath := filepath.Join(staging, "download")
	if err := m.download(ctx, artifact, archivePath); err != nil {
		return Installed{}, err
	}
	if err := verifyFile(archivePath, artifact.SHA256); err != nil {
		return Installed{}, err
	}
	executables := map[string]string{}
	for logical := range artifact.Executables {
		destinationName := logical
		if runtime.GOOS == "windows" && filepath.Ext(destinationName) == "" {
			destinationName += ".exe"
		}
		destination := filepath.Join(staging, destinationName)
		switch artifact.Format {
		case "gzip":
			if len(artifact.Executables) != 1 {
				return Installed{}, fmt.Errorf("gzip artifacts must contain exactly one executable")
			}
			if err := gunzip(archivePath, destination); err != nil {
				return Installed{}, err
			}
		case "raw":
			if err := copyFile(archivePath, destination); err != nil {
				return Installed{}, err
			}
		default:
			return Installed{}, fmt.Errorf("unsupported artifact format %q", artifact.Format)
		}
		if err := os.Chmod(destination, 0o755); err != nil {
			return Installed{}, fmt.Errorf("mark executable: %w", err)
		}
		executables[logical] = destinationName
	}
	installed := Installed{Name: name, Version: pkg.Version, License: pkg.License, Source: pkg.Source, Platform: platform, Installed: time.Now().UTC(), Executables: executables}
	metadata, err := json.MarshalIndent(installed, "", "  ")
	if err != nil {
		return Installed{}, err
	}
	if err := os.WriteFile(filepath.Join(staging, "package.json"), append(metadata, '\n'), 0o644); err != nil {
		return Installed{}, fmt.Errorf("write package metadata: %w", err)
	}
	final := filepath.Join(m.root, "packages", name, pkg.Version)
	if err := os.RemoveAll(final); err != nil {
		return Installed{}, fmt.Errorf("replace old package: %w", err)
	}
	if err := os.Rename(staging, final); err != nil {
		return Installed{}, fmt.Errorf("activate package: %w", err)
	}
	return installed, nil
}

func validateArtifact(artifact Artifact) error {
	for _, source := range append([]string{artifact.URL}, artifact.Mirrors...) {
		parsed, err := url.Parse(source)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return fmt.Errorf("package artifact sources must use absolute HTTPS URLs")
		}
	}
	if len(artifact.SHA256) != 64 {
		return fmt.Errorf("package artifact has an invalid SHA-256")
	}
	if len(artifact.Executables) == 0 {
		return fmt.Errorf("package artifact has no executables")
	}
	return nil
}

func (m *Manager) download(ctx context.Context, artifact Artifact, destination string) error {
	var failures []error
	sources := append([]string{artifact.URL}, artifact.Mirrors...)
	for index, source := range sources {
		_ = os.Remove(destination)
		sourceContext := ctx
		cancel := func() {}
		if index < len(sources)-1 {
			sourceContext, cancel = context.WithTimeout(ctx, 45*time.Second)
		}
		err := m.downloadOnce(sourceContext, source, destination)
		cancel()
		if err == nil {
			return nil
		} else {
			failures = append(failures, err)
		}
	}
	return fmt.Errorf("all package download sources failed: %w", errors.Join(failures...))
}

func (m *Manager) downloadOnce(ctx context.Context, source, destination string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", "FinishBit/0.1 (+https://github.com/fu9zhou/finish-bit)")
	request.Header.Set("Accept", "application/octet-stream")
	response, err := m.client.Do(request)
	if err != nil {
		return fmt.Errorf("download %s: %w", source, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: server returned %s", source, response.Status)
	}
	file, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("create download: %w", err)
	}
	defer file.Close()
	if _, err := io.Copy(file, io.LimitReader(response.Body, 1<<30)); err != nil {
		return fmt.Errorf("save package: %w", err)
	}
	return file.Close()
}

func verifyFile(path, expected string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return err
	}
	actual := hex.EncodeToString(hasher.Sum(nil))
	if !strings.EqualFold(actual, expected) {
		return &operation.Error{Code: operation.CodeIntegrityFailure, Message: "package SHA-256 verification failed", Details: map[string]any{"expected": expected, "actual": actual}}
	}
	return nil
}

func gunzip(source, destination string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	reader, err := gzip.NewReader(input)
	if err != nil {
		return fmt.Errorf("open gzip package: %w", err)
	}
	defer reader.Close()
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	defer output.Close()
	if _, err := io.Copy(output, io.LimitReader(reader, 1<<30)); err != nil {
		return fmt.Errorf("extract gzip package: %w", err)
	}
	return output.Close()
}

func copyFile(source, destination string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	defer output.Close()
	_, err = io.Copy(output, input)
	return err
}

func (m *Manager) Info(name string) (Installed, error) {
	pkg, ok := m.registry.Find(name)
	if !ok {
		return Installed{}, os.ErrNotExist
	}
	data, err := os.ReadFile(filepath.Join(m.root, "packages", name, pkg.Version, "package.json"))
	if err != nil {
		return Installed{}, err
	}
	var installed Installed
	if err := json.Unmarshal(data, &installed); err != nil {
		return Installed{}, fmt.Errorf("decode installed package metadata: %w", err)
	}
	return installed, nil
}

func (m *Manager) Status(name string) (Status, error) {
	pkg, ok := m.registry.Find(name)
	if !ok {
		return Status{}, &operation.Error{Code: operation.CodeInvalidInput, Message: fmt.Sprintf("unknown package %q", name)}
	}
	_, supported := pkg.Artifacts[PlatformKey()]
	status := Status{Name: pkg.Name, Version: pkg.Version, Description: pkg.Description, License: pkg.License, Source: pkg.Source, Platform: PlatformKey(), Supported: supported}
	if installed, err := m.Info(name); err == nil {
		status.Installed = &installed
	}
	return status, nil
}

func (m *Manager) List() []Installed {
	installed := []Installed{}
	for _, pkg := range m.registry.Packages {
		if info, err := m.Info(pkg.Name); err == nil {
			installed = append(installed, info)
		}
	}
	sort.Slice(installed, func(i, j int) bool { return installed[i].Name < installed[j].Name })
	return installed
}

func (m *Manager) Executable(packageName, logicalName string) (string, error) {
	installed, err := m.Info(packageName)
	if err != nil {
		return "", &operation.Error{Code: operation.CodeDependencyMissing, Message: fmt.Sprintf("operation requires package %q", packageName), Suggestion: fmt.Sprintf("fnsh pkg add %s", packageName), Err: err}
	}
	name, ok := installed.Executables[logicalName]
	if !ok {
		return "", fmt.Errorf("package %q does not provide executable %q", packageName, logicalName)
	}
	path := filepath.Join(m.root, "packages", packageName, installed.Version, name)
	if _, err := os.Stat(path); err != nil {
		return "", &operation.Error{Code: operation.CodeDependencyMissing, Message: fmt.Sprintf("package %q is damaged", packageName), Suggestion: fmt.Sprintf("fnsh pkg repair %s", packageName), Err: err}
	}
	return path, nil
}

func (m *Manager) Remove(name string) error {
	if _, ok := m.registry.Find(name); !ok {
		return &operation.Error{Code: operation.CodeInvalidInput, Message: fmt.Sprintf("unknown package %q", name)}
	}
	return os.RemoveAll(filepath.Join(m.root, "packages", name))
}

func (m *Manager) Repair(ctx context.Context, name string) (Installed, error) {
	if err := m.Remove(name); err != nil {
		return Installed{}, err
	}
	return m.Install(ctx, name)
}
