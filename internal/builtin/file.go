package builtin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func registerFile(registry *operation.Registry) error {
	input := []operation.Parameter{param("path", "File or directory path", true)}
	return registerAll(registry,
		operation.Capability{Definition: operation.Definition{ID: "file.info", Summary: "Inspect file metadata", Description: "Return path, type, size, permissions, modification time, and MIME type.", Aliases: []string{"stat file", "文件信息"}, Tags: []string{"file", "metadata"}, Inputs: input, Source: "core"}, Runner: operation.Func(runFileInfo)},
		operation.Capability{Definition: operation.Definition{ID: "file.checksum", Summary: "Calculate a file SHA-256 checksum", Description: "Stream a file and return its SHA-256 digest.", Aliases: []string{"sha256 file", "文件校验"}, Tags: []string{"file", "checksum", "sha256"}, Inputs: input, Source: "core"}, Runner: operation.Func(runFileChecksum)},
	)
}

func runFileInfo(_ context.Context, request operation.Request) (operation.Result, error) {
	if err := operation.RequireInputs(request, 1); err != nil {
		return operation.Result{}, err
	}
	absolute, err := filepath.Abs(request.Inputs[0])
	if err != nil {
		return operation.Result{}, fmt.Errorf("resolve path: %w", err)
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return operation.Result{}, &operation.Error{Code: operation.CodeInvalidInput, Message: fmt.Sprintf("cannot inspect %q", request.Inputs[0]), Err: err}
	}
	mimeType := ""
	if !info.IsDir() {
		mimeType = mime.TypeByExtension(filepath.Ext(info.Name()))
	}
	return operation.Result{Operation: "file.info", Data: map[string]any{"path": absolute, "name": info.Name(), "size": info.Size(), "directory": info.IsDir(), "permissions": info.Mode().Perm().String(), "modified": info.ModTime(), "mime": mimeType}}, nil
}

func runFileChecksum(_ context.Context, request operation.Request) (operation.Result, error) {
	if err := operation.RequireInputs(request, 1); err != nil {
		return operation.Result{}, err
	}
	file, err := os.Open(request.Inputs[0])
	if err != nil {
		return operation.Result{}, &operation.Error{Code: operation.CodeInvalidInput, Message: fmt.Sprintf("cannot open %q", request.Inputs[0]), Err: err}
	}
	defer file.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return operation.Result{}, fmt.Errorf("hash file: %w", err)
	}
	return operation.Result{Operation: "file.checksum", Data: map[string]any{"algorithm": "sha256", "digest": hex.EncodeToString(hasher.Sum(nil)), "path": request.Inputs[0]}}, nil
}
