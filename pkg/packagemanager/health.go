package packagemanager

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

// HealthCheck distinguishes an optional, absent package from a damaged install.
type HealthCheck struct {
	Name      string
	Installed bool
	Err       error
}

// Health verifies installed bytes without executing the runtimes being checked.
func (m *Manager) Health() []HealthCheck {
	checks := make([]HealthCheck, 0, len(m.registry.Packages))
	for _, pkg := range m.registry.Packages {
		check := HealthCheck{Name: pkg.Name}
		directory := filepath.Join(m.root, "packages", pkg.Name, pkg.Version)
		if _, err := os.Lstat(directory); os.IsNotExist(err) {
			checks = append(checks, check)
			continue
		}
		check.Installed = true
		installed, err := m.Info(pkg.Name)
		if err == nil {
			err = m.checkInstalled(pkg, installed)
		}
		if err != nil {
			check.Err = &operation.Error{Code: operation.CodeIntegrityFailure, Message: fmt.Sprintf("package %q is damaged or unverifiable: %v; run fnsh pkg add %s", pkg.Name, err, pkg.Name), Suggestion: "fnsh pkg add " + pkg.Name, Err: err}
		}
		checks = append(checks, check)
	}
	return checks
}

func (m *Manager) checkInstalled(pkg Package, installed Installed) error {
	if installed.Name != pkg.Name || installed.Version != pkg.Version || installed.Platform != PlatformKey() {
		return fmt.Errorf("installed metadata does not match package/version/platform")
	}
	artifact, ok := pkg.Artifacts[installed.Platform]
	if !ok {
		return fmt.Errorf("unsupported installed platform")
	}
	if len(installed.Files) == 0 {
		return fmt.Errorf("legacy installation has no integrity manifest; reinstall from verified cache or source")
	}
	if len(installed.Executables) != len(artifact.Executables) {
		return fmt.Errorf("incomplete executable manifest")
	}
	for logical, entry := range artifact.Executables {
		expected := logical
		if strings.HasPrefix(installed.Platform, "win32-") && filepath.Ext(expected) == "" {
			expected += ".exe"
		}
		if artifact.Format != "raw" && artifact.Format != "gzip" {
			expected = "payload/" + entry
		}
		if installed.Executables[logical] != expected || installed.Files[expected] == "" {
			return fmt.Errorf("missing executable integrity record: %s", logical)
		}
	}
	for _, path := range artifact.Resources {
		if installed.Files["payload/"+path] == "" {
			return fmt.Errorf("missing resource integrity record: %s", path)
		}
	}
	for _, resource := range artifact.Downloads {
		if !strings.EqualFold(installed.Files["payload/"+resource.Path], resource.SHA256) {
			return fmt.Errorf("missing or invalid resource integrity record: %s", resource.Path)
		}
	}
	actual, err := fileDigests(filepath.Join(m.root, "packages", pkg.Name, pkg.Version))
	if err != nil {
		return err
	}
	for path, expected := range installed.Files {
		if actual[path] == "" {
			return fmt.Errorf("missing installed file: %s", path)
		}
		if !strings.EqualFold(actual[path], expected) {
			return fmt.Errorf("installed file SHA-256 mismatch: %s", path)
		}
	}
	if len(actual) != len(installed.Files) {
		return fmt.Errorf("unexpected files in installed package")
	}
	return nil
}

// Files are a local corruption baseline derived only after archive verification.
// They do not authenticate against another user who can also change package.json.
func fileDigests(directory string) (map[string]string, error) {
	info, err := os.Lstat(directory)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("package directory is not a regular directory")
	}
	result := map[string]string{}
	var total int64
	err = filepath.WalkDir(directory, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symbolic link in package: %s", path)
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(directory, path)
		if err != nil {
			return err
		}
		if rel == "package.json" {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("non-regular package file: %s", rel)
		}
		total += info.Size()
		if total > maxPackageBytes {
			return fmt.Errorf("installed package exceeds size limit")
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		hash := sha256.New()
		_, err = io.Copy(hash, io.LimitReader(f, maxPackageBytes+1))
		closeErr := f.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		result[filepath.ToSlash(rel)] = hex.EncodeToString(hash.Sum(nil))
		return nil
	})
	return result, err
}
