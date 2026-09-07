// Package toolrun supplies bounded process execution and staged artifacts to providers.
package toolrun

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

type Resolver interface {
	Executable(string, string) (string, error)
}

type capture struct {
	sync.Mutex
	data      []byte
	limit     int
	truncated bool
}

func (b *capture) Write(p []byte) (int, error) {
	b.Lock()
	defer b.Unlock()
	n := len(p)
	if len(p) > b.limit-len(b.data) {
		p = p[:b.limit-len(b.data)]
		b.truncated = true
	}
	b.data = append(b.data, p...)
	return n, nil
}

// Run separates data from diagnostics and never returns silently truncated data.
func Run(ctx context.Context, resolver Resolver, pkg, name, dir string, args ...string) ([]byte, error) {
	return RunWithEnv(ctx, resolver, pkg, name, dir, nil, args...)
}

// RunWithEnv permits providers to isolate configuration from caller environment variables.
// A nil environment keeps the process defaults used by Run.
func RunWithEnv(ctx context.Context, resolver Resolver, pkg, name, dir string, environment []string, args ...string) ([]byte, error) {
	binary, err := resolver.Executable(pkg, name)
	if err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Dir = dir
	cmd.Env = environment
	out := &capture{limit: 16 << 20}
	diagnostic := &capture{limit: 1 << 20}
	cmd.Stdout = out
	cmd.Stderr = diagnostic
	if err = cmd.Run(); err != nil {
		message := strings.TrimSpace(string(diagnostic.data))
		if diagnostic.truncated {
			message += "\n[output truncated]"
		}
		return nil, &operation.Error{Code: operation.CodeExecutionFailed, Message: pkg + " operation failed", Details: map[string]any{"diagnostic": message}, Err: err}
	}
	if out.truncated {
		return nil, &operation.Error{Code: operation.CodeExecutionFailed, Message: pkg + " result exceeds 16 MiB"}
	}
	return out.data, nil
}

func Invalid(message string) error {
	return &operation.Error{Code: operation.CodeInvalidInput, Message: message}
}

func LocalFile(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", Invalid("invalid file path")
	}
	info, err := os.Stat(absolute)
	if err != nil || !info.Mode().IsRegular() {
		return "", Invalid("input must be an existing regular file: " + path)
	}
	return absolute, nil
}

// Artifact publishes a completed file; failed tools cannot damage the destination.
func Artifact(ctx context.Context, path string, overwrite bool, generate func(string) error) error {
	if path == "" {
		return Invalid("output path is required")
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if !overwrite {
		if _, err := os.Lstat(path); err == nil {
			return Invalid("output already exists; use --overwrite")
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	dir, err := os.MkdirTemp(filepath.Dir(path), ".finishbit-output-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	temporary := filepath.Join(dir, "result"+filepath.Ext(path))
	if err := generate(temporary); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	info, err := os.Stat(temporary)
	if err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("provider did not produce the expected output file")
	}
	if overwrite {
		return os.Rename(temporary, path)
	}
	if err := os.Link(temporary, path); err != nil {
		if os.IsExist(err) {
			return Invalid("output already exists; use --overwrite")
		}
		return err
	}
	return nil
}

// Directory publishes an entire multi-file result into a new directory.
func Directory(ctx context.Context, path string, generate func(string) error) ([]string, error) {
	if path == "" {
		return nil, Invalid("output directory is required")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if _, err := os.Lstat(absolute); err == nil {
		return nil, Invalid("output directory must not already exist")
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(absolute), 0755); err != nil {
		return nil, err
	}
	temporary, err := os.MkdirTemp(filepath.Dir(absolute), ".finishbit-output-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(temporary)
	if err := generate(temporary); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(temporary)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, Invalid("operation produced no files")
	}
	// Reserve the destination exclusively. Publish individual files, removing the reservation on failure.
	if err := os.Mkdir(absolute, 0755); err != nil {
		return nil, Invalid("cannot create output directory: " + err.Error())
	}
	success := false
	defer func() {
		if !success {
			_ = os.RemoveAll(absolute)
		}
	}()
	outputs := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			return nil, fmt.Errorf("unexpected non-file provider output")
		}
		destination := filepath.Join(absolute, entry.Name())
		if err := os.Rename(filepath.Join(temporary, entry.Name()), destination); err != nil {
			return nil, err
		}
		outputs = append(outputs, destination)
	}
	success = true
	return outputs, nil
}

func CopyFile(from, to string) error {
	r, err := os.Open(from)
	if err != nil {
		return err
	}
	defer r.Close()
	w, err := os.OpenFile(to, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(w, r)
	closeErr := w.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func Param(name, description string, required bool) operation.Parameter {
	return operation.Parameter{Name: name, Type: operation.TypeString, Description: description, Required: required}
}
func Option(name string, kind operation.ValueType, description string, value any) operation.Parameter {
	return operation.Parameter{Name: name, Type: kind, Description: description, Default: value}
}
func OutputOptions() []operation.Parameter {
	return []operation.Parameter{Param("output", "Destination file", true), Option("overwrite", operation.TypeBoolean, "Replace an existing destination", false)}
}

// Values collects validation errors so builders never ignore malformed parameters.
type Values struct {
	Request operation.Request
	Err     error
}

func (v *Values) fail(err error) {
	if v.Err == nil {
		v.Err = err
	}
}
func (v *Values) String(name, fallback string) string {
	s, e := operation.StringOption(v.Request, name, fallback)
	v.fail(e)
	return s
}
func (v *Values) Int(name string, fallback, low, high int) int {
	n, e := operation.IntOption(v.Request, name, fallback)
	v.fail(e)
	if n < low || n > high {
		v.fail(Invalid(fmt.Sprintf("%s must be between %d and %d", name, low, high)))
	}
	return n
}
func (v *Values) Bool(name string, fallback bool) bool {
	b, e := operation.BoolOption(v.Request, name, fallback)
	v.fail(e)
	return b
}
func (v *Values) Enum(name, fallback string, choices ...string) string {
	s := v.String(name, fallback)
	for _, c := range choices {
		if c == s {
			return s
		}
	}
	v.fail(Invalid(name + " must be one of " + strings.Join(choices, ", ")))
	return s
}
func (v *Values) File(path string) string { s, e := LocalFile(path); v.fail(e); return s }
func (v *Values) Strings(name string) []string {
	raw, ok := v.Request.Options[name]
	if !ok {
		return nil
	}
	switch x := raw.(type) {
	case []string:
		return x
	case []any:
		out := []string{}
		for _, item := range x {
			s, ok := item.(string)
			if !ok {
				v.fail(Invalid(name + " must contain strings"))
				return nil
			}
			out = append(out, s)
		}
		return out
	default:
		v.fail(Invalid(name + " must be a string array"))
		return nil
	}
}
func (v *Values) Check(ok bool, message string) {
	if !ok {
		v.fail(Invalid(message))
	}
}
