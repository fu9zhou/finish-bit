package builtin

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

const maxInputBytes int64 = 64 << 20

func readInput(value string) ([]byte, error) {
	if value == "-" {
		return readBounded(os.Stdin, "stdin")
	}
	info, err := os.Stat(value)
	if err == nil && !info.IsDir() {
		if info.Size() > maxInputBytes {
			return nil, inputTooLarge(value)
		}
		file, openErr := os.Open(value)
		if openErr != nil {
			return nil, fmt.Errorf("open %q: %w", value, openErr)
		}
		defer file.Close()
		return readBounded(file, value)
	}
	if int64(len(value)) > maxInputBytes {
		return nil, inputTooLarge("literal input")
	}
	return []byte(value), nil
}

func readBounded(reader io.Reader, source string) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(reader, maxInputBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", source, err)
	}
	if int64(len(data)) > maxInputBytes {
		return nil, inputTooLarge(source)
	}
	return data, nil
}

func inputTooLarge(source string) error {
	return &operation.Error{Code: operation.CodeInvalidInput, Message: fmt.Sprintf("%s exceeds the 64 MiB input limit", source)}
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
