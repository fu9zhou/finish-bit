package extension

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

type Manager struct{ root string }

const maxExtensionBytes int64 = 512 << 20

func New(root string) *Manager { return &Manager{root: root} }

func (m *Manager) Install(source string) (Manifest, error) {
	return m.InstallValidated(source, nil)
}

// InstallValidated stages and validates an extension before making it visible.
// The optional validator can enforce application-wide constraints such as
// Operation ID uniqueness without leaving an invalid extension installed.
func (m *Manager) InstallValidated(source string, validate func(Manifest) error) (Manifest, error) {
	if err := os.MkdirAll(m.root, 0o755); err != nil {
		return Manifest{}, err
	}
	staging, err := os.MkdirTemp(m.root, ".install-")
	if err != nil {
		return Manifest{}, err
	}
	defer os.RemoveAll(staging)
	info, err := os.Stat(source)
	if err != nil {
		return Manifest{}, fmt.Errorf("inspect extension source: %w", err)
	}
	if info.IsDir() {
		err = copyDirectory(source, staging)
	} else if strings.EqualFold(filepath.Ext(source), ".zip") {
		err = extractZip(source, staging)
	} else {
		err = fmt.Errorf("extension source must be a directory or .zip archive")
	}
	if err != nil {
		return Manifest{}, err
	}
	manifestPath := filepath.Join(staging, "finishbit-extension.json")
	manifest, err := ReadManifest(manifestPath)
	if err != nil {
		return Manifest{}, err
	}
	if validate != nil {
		if err := validate(manifest); err != nil {
			return Manifest{}, err
		}
	}
	executable, err := resolveExecutable(staging, manifest.Executable)
	if err != nil {
		return Manifest{}, err
	}
	if err := os.Chmod(executable, 0o755); err != nil {
		return Manifest{}, fmt.Errorf("mark extension executable: %w", err)
	}
	if runtime.GOOS == "windows" && filepath.Ext(executable) == "" {
		windowsExecutable := executable + ".exe"
		if err := os.Rename(executable, windowsExecutable); err != nil {
			return Manifest{}, fmt.Errorf("normalize Windows extension executable: %w", err)
		}
	}
	destination := filepath.Join(m.root, manifest.Name)
	if _, err := os.Stat(destination); err == nil {
		return Manifest{}, fmt.Errorf("extension %q is already installed; remove it before reinstalling", manifest.Name)
	}
	if err := os.Rename(staging, destination); err != nil {
		return Manifest{}, fmt.Errorf("activate extension: %w", err)
	}
	return ReadManifest(filepath.Join(destination, "finishbit-extension.json"))
}

func (m *Manager) Remove(name string) error {
	if !validName(name) {
		return fmt.Errorf("invalid extension name %q", name)
	}
	return removeDirectory(filepath.Join(m.root, name))
}

func (m *Manager) List() []Manifest {
	manifests, _ := m.Scan()
	return manifests
}

// Scan retains diagnostics for damaged installations instead of hiding them.
func (m *Manager) Scan() ([]Manifest, map[string]error) {
	issues := map[string]error{}
	entries, err := os.ReadDir(m.root)
	if err != nil {
		if !os.IsNotExist(err) {
			issues["directory"] = err
		}
		return []Manifest{}, issues
	}
	manifests := []Manifest{}
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			if manifest, err := ReadManifest(filepath.Join(m.root, entry.Name(), "finishbit-extension.json")); err == nil {
				if manifest.Name != entry.Name() {
					issues[entry.Name()] = fmt.Errorf("manifest name %q does not match installation directory", manifest.Name)
				} else {
					manifests = append(manifests, manifest)
				}
			} else {
				issues[entry.Name()] = err
			}
		}
	}
	sort.Slice(manifests, func(i, j int) bool { return manifests[i].Name < manifests[j].Name })
	return manifests, issues
}

// RegisterAvailable isolates invalid extensions so diagnostics and removal work.
func (m *Manager) RegisterAvailable(registry *operation.Registry) map[string]error {
	manifests, issues := m.Scan()
	for _, manifest := range manifests {
		seen := map[string]bool{}
		for _, def := range manifest.Operations {
			_, exists := registry.Get(def.ID)
			if exists || seen[def.ID] {
				issues[manifest.Name] = fmt.Errorf("operation %q already registered", def.ID)
				break
			}
			seen[def.ID] = true
		}
		if issues[manifest.Name] != nil {
			continue
		}
		executable, err := resolveExecutable(filepath.Join(m.root, manifest.Name), manifest.Executable)
		if err != nil {
			issues[manifest.Name] = err
			continue
		}
		for _, def := range manifest.Operations {
			def.Source = "extension:" + manifest.Name
			if err := registry.Register(operation.Capability{Definition: def, Runner: &processRunner{executable: executable, operationID: def.ID}}); err != nil {
				issues[manifest.Name] = err
				break
			}
		}
	}
	return issues
}

func (m *Manager) Info(name string) (Manifest, error) {
	if !validName(name) {
		return Manifest{}, fmt.Errorf("invalid extension name %q", name)
	}
	return ReadManifest(filepath.Join(m.root, name, "finishbit-extension.json"))
}

func (m *Manager) Register(registry *operation.Registry) error {
	for _, manifest := range m.List() {
		root := filepath.Join(m.root, manifest.Name)
		executable, err := resolveExecutable(root, manifest.Executable)
		if err != nil {
			return err
		}
		for _, definition := range manifest.Operations {
			definition.Source = "extension:" + manifest.Name
			runner := &processRunner{executable: executable, operationID: definition.ID}
			if err := registry.Register(operation.Capability{Definition: definition, Runner: runner}); err != nil {
				return fmt.Errorf("register extension %q: %w", manifest.Name, err)
			}
		}
	}
	return nil
}

type protocolRequest struct {
	Protocol  int               `json:"protocol"`
	Operation string            `json:"operation"`
	Request   operation.Request `json:"request"`
}
type protocolResponse struct {
	Result *operation.Result `json:"result,omitempty"`
	Error  *operation.Error  `json:"error,omitempty"`
}
type processRunner struct{ executable, operationID string }

type limitedWriter struct {
	builder   strings.Builder
	remaining int
}

func (w *limitedWriter) Write(data []byte) (int, error) {
	original := len(data)
	if len(data) > w.remaining {
		data = data[:w.remaining]
	}
	if len(data) > 0 {
		_, _ = w.builder.Write(data)
		w.remaining -= len(data)
	}
	return original, nil
}

func (w *limitedWriter) String() string { return w.builder.String() }

func (r *processRunner) Run(ctx context.Context, request operation.Request) (operation.Result, error) {
	command := exec.CommandContext(ctx, r.executable)
	stdin, err := command.StdinPipe()
	if err != nil {
		return operation.Result{}, err
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		return operation.Result{}, err
	}
	stderr := &limitedWriter{remaining: 1 << 20}
	command.Stderr = stderr
	if err := command.Start(); err != nil {
		return operation.Result{}, fmt.Errorf("start extension operation: %w", err)
	}
	encodeErr := json.NewEncoder(stdin).Encode(protocolRequest{Protocol: ProtocolVersion, Operation: r.operationID, Request: request})
	closeErr := stdin.Close()
	if encodeErr != nil {
		_ = command.Process.Kill()
		_ = command.Wait()
		return operation.Result{}, fmt.Errorf("send extension request: %w", encodeErr)
	}
	if closeErr != nil {
		_ = command.Process.Kill()
		_ = command.Wait()
		return operation.Result{}, fmt.Errorf("close extension request: %w", closeErr)
	}
	responseBytes, readErr := io.ReadAll(io.LimitReader(stdout, (16<<20)+1))
	if readErr != nil {
		_ = command.Process.Kill()
		_ = command.Wait()
		return operation.Result{}, fmt.Errorf("read extension response: %w", readErr)
	}
	if len(responseBytes) > 16<<20 {
		_ = command.Process.Kill()
		_ = command.Wait()
		return operation.Result{}, &operation.Error{Code: operation.CodeExecutionFailed, Message: "extension response exceeds 16 MiB"}
	}
	var response protocolResponse
	decoder := json.NewDecoder(bytes.NewReader(responseBytes))
	decodeErr := decoder.Decode(&response)
	if decodeErr == nil {
		var trailing any
		if err := decoder.Decode(&trailing); err != io.EOF {
			decodeErr = fmt.Errorf("extension returned trailing output")
		}
	}
	if decodeErr == nil && (response.Result == nil) == (response.Error == nil) {
		decodeErr = fmt.Errorf("extension response must contain exactly one of result or error")
	}
	waitErr := command.Wait()
	if decodeErr != nil {
		return operation.Result{}, &operation.Error{Code: operation.CodeExecutionFailed, Message: "extension returned an invalid response", Details: map[string]any{"stderr": stderr.String()}, Err: decodeErr}
	}
	if waitErr != nil {
		return operation.Result{}, &operation.Error{Code: operation.CodeExecutionFailed, Message: "extension process failed", Details: map[string]any{"stderr": stderr.String()}, Err: waitErr}
	}
	if response.Error != nil {
		response.Error.Operation = r.operationID
		return operation.Result{}, response.Error
	}
	response.Result.Operation = r.operationID
	return *response.Result, nil
}

func copyDirectory(source, destination string) error {
	var total int64
	return filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target, err := safeJoin(destination, relative)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("extension archives cannot contain symbolic links")
		}
		if info.Size() < 0 || info.Size() > maxExtensionBytes-total {
			return fmt.Errorf("extension directory exceeds 512 MiB size limit")
		}
		total += info.Size()
		input, err := os.Open(path)
		if err != nil {
			return err
		}
		output, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, info.Mode().Perm())
		if err != nil {
			input.Close()
			return err
		}
		written, copyErr := io.Copy(output, io.LimitReader(input, info.Size()+1))
		inputCloseErr := input.Close()
		closeErr := output.Close()
		if copyErr != nil {
			return copyErr
		}
		if written != info.Size() {
			return fmt.Errorf("extension file %q changed while it was being copied", relative)
		}
		if inputCloseErr != nil {
			return inputCloseErr
		}
		return closeErr
	})
}

func extractZip(path, destination string) error {
	archive, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("open extension archive: %w", err)
	}
	defer archive.Close()
	var total int64
	for _, file := range archive.File {
		if file.UncompressedSize64 > uint64(maxExtensionBytes-total) {
			return fmt.Errorf("extension archive exceeds 512 MiB extracted size limit")
		}
		total += int64(file.UncompressedSize64)
		target, err := safeJoin(destination, file.Name)
		if err != nil {
			return err
		}
		if file.FileInfo().Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("extension archives cannot contain symbolic links")
		}
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		input, err := file.Open()
		if err != nil {
			return err
		}
		output, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, file.Mode().Perm())
		if err != nil {
			input.Close()
			return err
		}
		written, copyErr := io.Copy(output, io.LimitReader(input, int64(file.UncompressedSize64)+1))
		input.Close()
		closeErr := output.Close()
		if copyErr != nil {
			return copyErr
		}
		if written != int64(file.UncompressedSize64) {
			return fmt.Errorf("extension archive entry %q has an invalid size", file.Name)
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}
