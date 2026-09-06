package builtin

import (
	"fmt"
	"github.com/fu9zhou/finish-bit/pkg/operation"
	"io"
	"os"
	"path/filepath"
)

// writeArtifact completes an output beside its destination before publishing it.
// Without overwrite, a hard link publishes exclusively without a check/write race.
func writeArtifact(path string, overwrite bool, write func(io.Writer) error) error {
	if path == "" {
		return &operation.Error{Code: operation.CodeInvalidInput, Message: "output path is required"}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".finishbit-output-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	if err := write(file); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if overwrite {
		err = os.Rename(file.Name(), path)
	} else {
		err = os.Link(file.Name(), path)
	}
	if os.IsExist(err) {
		return &operation.Error{Code: operation.CodeInvalidInput, Message: fmt.Sprintf("output %q already exists; use --overwrite to replace it", path)}
	}
	return err
}

func finishArtifactText(id, text string, request operation.Request) (operation.Result, error) {
	output, err := outputOption(request)
	if err != nil {
		return operation.Result{}, err
	}
	overwrite, err := operation.BoolOption(request, "overwrite", false)
	if err != nil {
		return operation.Result{}, err
	}
	if output == "" {
		return operation.Result{Operation: id, Data: map[string]any{"text": text}}, nil
	}
	if err := writeArtifact(output, overwrite, func(w io.Writer) error { _, err := io.WriteString(w, text); return err }); err != nil {
		return operation.Result{}, err
	}
	return operation.Result{Operation: id, Outputs: []string{output}}, nil
}
