// Package extension installs and runs language-neutral FinishBit extensions.
package extension

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

const ProtocolVersion = 1

type Manifest struct {
	Schema     int                    `json:"schema"`
	Name       string                 `json:"name"`
	Version    string                 `json:"version"`
	Executable string                 `json:"executable"`
	Operations []operation.Definition `json:"operations"`
}

func ReadManifest(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("read extension manifest: %w", err)
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	var manifest Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("decode extension manifest: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Manifest{}, fmt.Errorf("extension manifest contains trailing content")
	}
	if err := manifest.Validate(filepath.Dir(path)); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func (m Manifest) Validate(root string) error {
	if m.Schema != ProtocolVersion {
		return fmt.Errorf("unsupported extension schema %d", m.Schema)
	}
	if !validName(m.Name) {
		return fmt.Errorf("extension name %q is invalid", m.Name)
	}
	if m.Version == "" {
		return fmt.Errorf("extension %q has no version", m.Name)
	}
	if m.Executable == "" || filepath.IsAbs(m.Executable) {
		return fmt.Errorf("extension %q has an invalid executable", m.Name)
	}
	executable, err := resolveExecutable(root, m.Executable)
	if err != nil {
		return err
	}
	if info, err := os.Stat(executable); err != nil || info.IsDir() {
		return fmt.Errorf("extension executable %q does not exist", m.Executable)
	}
	if len(m.Operations) == 0 {
		return fmt.Errorf("extension %q has no operations", m.Name)
	}
	for index := range m.Operations {
		m.Operations[index].Source = "extension:" + m.Name
		if err := operation.ValidateDefinition(m.Operations[index]); err != nil {
			return err
		}
	}
	return nil
}

func validName(name string) bool {
	return name != "" && name != "." && name != ".." && !strings.HasPrefix(name, "-") && !strings.HasSuffix(name, "-") && !strings.Contains(name, "--") && strings.Trim(name, "abcdefghijklmnopqrstuvwxyz0123456789-") == ""
}

func resolveExecutable(root, name string) (string, error) {
	executable, err := safeJoin(root, name)
	if err != nil {
		return "", err
	}
	if info, statErr := os.Stat(executable); statErr == nil && !info.IsDir() {
		return executable, nil
	}
	if runtime.GOOS == "windows" && filepath.Ext(executable) == "" {
		withExtension := executable + ".exe"
		if info, statErr := os.Stat(withExtension); statErr == nil && !info.IsDir() {
			return withExtension, nil
		}
	}
	return executable, nil
}

func safeJoin(root, name string) (string, error) {
	cleanRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	candidate, err := filepath.Abs(filepath.Join(cleanRoot, name))
	if err != nil {
		return "", err
	}
	if candidate != cleanRoot && !strings.HasPrefix(candidate, cleanRoot+string(os.PathSeparator)) {
		return "", fmt.Errorf("path %q escapes the extension directory", name)
	}
	return candidate, nil
}
