package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func TestInstallExtensionRejectsOperationConflictWithoutBreakingNextStart(t *testing.T) {
	root := t.TempDir()
	source := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "runner"), []byte("runner"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{
  "schema": 1,
  "name": "conflicting-extension",
  "version": "1.0.0",
  "executable": "runner",
  "operations": [{"id":"json.format","summary":"Conflict","source":"extension"}]
}`
	if err := os.WriteFile(filepath.Join(source, "finishbit-extension.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	application, err := New(Config{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := application.InstallExtension(source); err == nil || !strings.Contains(err.Error(), "already registered") || operation.AsError(err).Code != operation.CodeInvalidInput {
		t.Fatalf("InstallExtension() error = %v, want operation conflict", err)
	}
	if _, err := os.Stat(filepath.Join(root, "extensions", "conflicting-extension")); !os.IsNotExist(err) {
		t.Fatalf("conflicting extension remains installed: %v", err)
	}
	if _, err := New(Config{Root: root}); err != nil {
		t.Fatalf("next application start failed: %v", err)
	}
}

func TestApplicationCoreWorkflow(t *testing.T) {
	application, err := New(Config{Root: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if len(application.Capabilities()) < 10 {
		t.Fatalf("capabilities = %d", len(application.Capabilities()))
	}
	if matches := application.Search("format json", 3); len(matches) == 0 || matches[0].ID != "json.format" {
		t.Fatalf("search matches = %#v", matches)
	}
	definition, err := application.Describe("json.format")
	if err != nil || definition.ID != "json.format" {
		t.Fatalf("Describe() = %#v, %v", definition, err)
	}
	result, err := application.Execute(context.Background(), "json.minify", operationRequest(`{"ok": true}`))
	if err != nil || result.Operation != "json.minify" || result.Data["text"] != `{"ok":true}` {
		t.Fatalf("Execute() = %#v, %v", result, err)
	}
	if _, err := application.Describe("missing.operation"); err == nil {
		t.Fatal("missing operation was described")
	}
	if _, err := application.Execute(context.Background(), "missing.operation", operationRequest("value")); err == nil {
		t.Fatal("missing operation was executed")
	}
	checks := application.Doctor()
	if len(checks) < 3 {
		t.Fatalf("doctor checks = %#v", checks)
	}
	status, err := application.PackageInfo("ffmpeg")
	if err != nil || status.Name != "ffmpeg" || !status.Supported {
		t.Fatalf("PackageInfo() = %#v, %v", status, err)
	}
	if packages := application.Packages(); len(packages) != 0 {
		t.Fatalf("Packages() = %#v", packages)
	}
	if err := application.RemovePackage("ffmpeg"); err != nil {
		t.Fatal(err)
	}
}

func TestApplicationExecuteValidatesOperationContract(t *testing.T) {
	application, err := New(Config{Root: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	called := false
	definition := operation.Definition{
		ID:      "test.contract",
		Summary: "Validate a request",
		Source:  "test",
		Inputs:  []operation.Parameter{{Name: "input", Type: operation.TypeString, Required: true}},
		Options: []operation.Parameter{{Name: "count", Type: operation.TypeInteger, Required: true}},
	}
	if err := application.registry.Register(operation.Capability{Definition: definition, Runner: operation.Func(func(_ context.Context, _ operation.Request) (operation.Result, error) {
		called = true
		return operation.Result{}, nil
	})}); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		request operation.Request
	}{
		{name: "missing input", request: operation.Request{Options: map[string]any{"count": 1}}},
		{name: "excess input", request: operation.Request{Inputs: []string{"value", "extra"}, Options: map[string]any{"count": 1}}},
		{name: "unknown option", request: operation.Request{Inputs: []string{"value"}, Options: map[string]any{"count": 1, "unknown": true}}},
		{name: "missing required option", request: operation.Request{Inputs: []string{"value"}}},
		{name: "wrong option type", request: operation.Request{Inputs: []string{"value"}, Options: map[string]any{"count": true}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			called = false
			_, err := application.Execute(context.Background(), definition.ID, test.request)
			if err == nil || operation.AsError(err).Code != operation.CodeInvalidInput {
				t.Fatalf("Execute() error = %v, want invalid_input", err)
			}
			if called {
				t.Fatal("runner was called before request validation")
			}
		})
	}
}

func operationRequest(input string) operation.Request {
	return operation.Request{Inputs: []string{input}, Options: map[string]any{}}
}

func TestApplicationExtensionLifecycle(t *testing.T) {
	root := t.TempDir()
	source := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "runner"), []byte("runner"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := map[string]any{
		"schema": 1, "name": "lifecycle", "version": "1.0.0", "executable": "runner",
		"operations": []map[string]any{{"id": "lifecycle.echo", "summary": "Echo", "source": "extension"}},
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "finishbit-extension.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	application, err := New(Config{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	installed, err := application.InstallExtension(source)
	if err != nil || installed.Name != "lifecycle" {
		t.Fatalf("InstallExtension() = %#v, %v", installed, err)
	}
	if values := application.Extensions(); len(values) != 1 || values[0].Name != "lifecycle" {
		t.Fatalf("Extensions() = %#v", values)
	}
	if info, err := application.ExtensionInfo("lifecycle"); err != nil || info.Version != "1.0.0" {
		t.Fatalf("ExtensionInfo() = %#v, %v", info, err)
	}
	if err := application.RemoveExtension("lifecycle"); err != nil {
		t.Fatal(err)
	}
	if values := application.Extensions(); len(values) != 0 {
		t.Fatalf("extensions after removal = %#v", values)
	}
}
