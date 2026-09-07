package toolrun

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

type limitedWriter struct {
	writer    io.Writer
	remaining int64
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	if int64(len(p)) > w.remaining {
		return 0, fmt.Errorf("output exceeds configured limit")
	}
	n, e := w.writer.Write(p)
	w.remaining -= int64(n)
	return n, e
}

// RunToFile streams bounded stdout into a provider-owned temporary file.
// The provider is responsible for staging and publishing the result.
func RunToFile(ctx context.Context, resolver Resolver, pkg, name, dir, path string, limit int64, args ...string) error {
	binary, e := resolver.Executable(pkg, name)
	if e != nil {
		return e
	}
	file, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Dir = dir
	diagnostic := &capture{limit: 1 << 20}
	cmd.Stderr = diagnostic
	cmd.Stdout = &limitedWriter{file, limit}
	runErr := cmd.Run()
	closeErr := file.Close()
	if runErr != nil {
		return &operation.Error{Code: operation.CodeExecutionFailed, Message: pkg + " operation failed", Details: map[string]any{"diagnostic": strings.TrimSpace(string(diagnostic.data))}, Err: runErr}
	}
	return closeErr
}

// DirectoryTree stages a nested regular-file tree and reserves a new destination.
// It returns the root directory; no symbolic links or special files are published.
func DirectoryTree(ctx context.Context, path string, generate func(string) error) ([]string, error) {
	if path == "" {
		return nil, Invalid("output directory is required")
	}
	root, e := filepath.Abs(path)
	if e != nil {
		return nil, e
	}
	if _, e = os.Lstat(root); !os.IsNotExist(e) {
		return nil, Invalid("output directory must not already exist")
	}
	if e = os.MkdirAll(filepath.Dir(root), 0755); e != nil {
		return nil, e
	}
	temp, e := os.MkdirTemp(filepath.Dir(root), ".finishbit-output-")
	if e != nil {
		return nil, e
	}
	defer os.RemoveAll(temp)
	if e = generate(temp); e != nil {
		return nil, e
	}
	if e = filepath.WalkDir(temp, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if e := ctx.Err(); e != nil {
			return e
		}
		if !d.IsDir() && !d.Type().IsRegular() {
			return Invalid("unexpected non-file output")
		}
		return nil
	}); e != nil {
		return nil, e
	}
	if e = os.Mkdir(root, 0755); e != nil {
		return nil, Invalid("cannot reserve output directory")
	}
	success := false
	defer func() {
		if !success {
			_ = os.RemoveAll(root)
		}
	}()
	entries, e := os.ReadDir(temp)
	if e != nil {
		return nil, e
	}
	for _, entry := range entries {
		if e = ctx.Err(); e != nil {
			return nil, e
		}
		if e = os.Rename(filepath.Join(temp, entry.Name()), filepath.Join(root, entry.Name())); e != nil {
			return nil, e
		}
	}
	success = true
	return []string{root}, nil
}
