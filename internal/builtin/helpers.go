package builtin

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func readInput(value string) ([]byte, error) {
	if value == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("read stdin: %w", err)
		}
		return data, nil
	}
	info, err := os.Stat(value)
	if err == nil && !info.IsDir() {
		data, readErr := os.ReadFile(value)
		if readErr != nil {
			return nil, fmt.Errorf("read %q: %w", value, readErr)
		}
		return data, nil
	}
	return []byte(value), nil
}

func finishText(id, text, output string) (operation.Result, error) {
	result := operation.Result{Operation: id, Data: map[string]any{"text": text}}
	if output == "" {
		return result, nil
	}
	if parent := filepath.Dir(output); parent != "." {
		if err := os.MkdirAll(parent, 0o755); err != nil {
			return operation.Result{}, fmt.Errorf("create output directory: %w", err)
		}
	}
	if err := os.WriteFile(output, []byte(text), 0o644); err != nil {
		return operation.Result{}, fmt.Errorf("write %q: %w", output, err)
	}
	result.Outputs = []string{output}
	delete(result.Data, "text")
	return result, nil
}

func outputOption(request operation.Request) (string, error) {
	return operation.StringOption(request, "output", "")
}

func registerAll(registry *operation.Registry, capabilities ...operation.Capability) error {
	for _, capability := range capabilities {
		if err := registry.Register(capability); err != nil {
			return err
		}
	}
	return nil
}

func param(name, description string, required bool) operation.Parameter {
	return operation.Parameter{Name: name, Type: operation.TypeString, Description: description, Required: required}
}

func option(name string, valueType operation.ValueType, description string, defaultValue any) operation.Parameter {
	return operation.Parameter{Name: name, Type: valueType, Description: description, Default: defaultValue}
}
