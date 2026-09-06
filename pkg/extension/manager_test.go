package extension

import (
	"archive/zip"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func TestInstallDirectory(t *testing.T) {
	source := t.TempDir()
	executable := filepath.Join(source, "runner")
	if err := os.WriteFile(executable, []byte("runner"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{
  "schema": 1,
  "name": "example",
  "version": "1.0.0",
  "executable": "runner",
  "operations": [{"id":"example.echo","summary":"Echo input","description":"Echo an input value","source":"extension"}]
}`
	if err := os.WriteFile(filepath.Join(source, "finishbit-extension.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	manager := New(t.TempDir())
	installed, err := manager.Install(source)
	if err != nil {
		t.Fatal(err)
	}
	if installed.Name != "example" {
		t.Fatalf("name = %q", installed.Name)
	}
	if len(manager.List()) != 1 {
		t.Fatal("installed extension was not listed")
	}
}

func TestInstallZip(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "extension.zip")
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	archive := zip.NewWriter(file)
	entries := map[string]string{
		"finishbit-extension.json": `{"schema":1,"name":"zipped","version":"1","executable":"runner","operations":[{"id":"zipped.run","summary":"Run","source":"extension"}]}`,
		"runner":                   "runner",
	}
	for name, content := range entries {
		entry, err := archive.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	manager := New(t.TempDir())
	manifest, err := manager.Install(archivePath)
	if err != nil || manifest.Name != "zipped" {
		t.Fatalf("Install(zip) = %#v, %v", manifest, err)
	}
}

func TestRegisteredExtensionExecutesProtocol(t *testing.T) {
	source := t.TempDir()
	program := `package main
import ("encoding/json"; "fmt"; "os")
func main() {
  var request struct { Request struct { Inputs []string ` + "`json:\"inputs\"`" + ` } ` + "`json:\"request\"`" + ` }
  if err := json.NewDecoder(os.Stdin).Decode(&request); err != nil { os.Exit(2) }
  switch request.Request.Inputs[0] {
  case "protocol-error": json.NewEncoder(os.Stdout).Encode(map[string]any{"error": map[string]any{"code":"invalid_input", "message":"rejected"}}); return
  case "trailing": fmt.Fprint(os.Stdout, "{}{}"); return
  case "exit": os.Exit(3)
	case "empty": json.NewEncoder(os.Stdout).Encode(map[string]any{}); return
	case "both": json.NewEncoder(os.Stdout).Encode(map[string]any{"result": map[string]any{}, "error": map[string]any{"code":"invalid_input", "message":"rejected"}}); return
  }
  json.NewEncoder(os.Stdout).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"text": request.Request.Inputs[0]}}})
}`
	programPath := filepath.Join(source, "main.go")
	if err := os.WriteFile(programPath, []byte(program), 0o644); err != nil {
		t.Fatal(err)
	}
	executableName := "runner"
	if runtime.GOOS == "windows" {
		executableName += ".exe"
	}
	executable := filepath.Join(source, executableName)
	command := exec.Command("go", "build", "-o", executable, programPath)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build extension runner: %v: %s", err, output)
	}
	manifest := `{"schema":1,"name":"protocol","version":"1","executable":"runner","operations":[{"id":"protocol.echo","summary":"Echo","source":"extension","inputs":[{"name":"input","type":"string","description":"Value","required":true}]}]}`
	if err := os.WriteFile(filepath.Join(source, "finishbit-extension.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	manager := New(t.TempDir())
	if _, err := manager.Install(source); err != nil {
		t.Fatal(err)
	}
	registry := operation.NewRegistry()
	if err := manager.Register(registry); err != nil {
		t.Fatal(err)
	}
	capability, ok := registry.Get("protocol.echo")
	if !ok {
		t.Fatal("registered extension operation not found")
	}
	result, err := capability.Runner.Run(context.Background(), operation.Request{Inputs: []string{"hello"}, Options: map[string]any{}})
	if err != nil || result.Operation != "protocol.echo" || result.Data["text"] != "hello" {
		t.Fatalf("extension result = %#v, %v", result, err)
	}
	for _, input := range []string{"protocol-error", "trailing", "exit", "empty", "both"} {
		if _, err := capability.Runner.Run(context.Background(), operation.Request{Inputs: []string{input}, Options: map[string]any{}}); err == nil {
			t.Fatalf("extension input %q did not return an error", input)
		}
	}
}

func TestManifestRejectsEscapingExecutable(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(root, "..", "runner")
	if err := os.WriteFile(outside, []byte("runner"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{"schema":1,"name":"bad","version":"1","executable":"../runner","operations":[{"id":"bad.run","summary":"Bad","source":"extension"}]}`
	if err := os.WriteFile(filepath.Join(root, "finishbit-extension.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadManifest(filepath.Join(root, "finishbit-extension.json")); err == nil {
		t.Fatal("escaping executable was accepted")
	}
}

func TestManifestRequiresHostAssignedExtensionSource(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "runner"), []byte("runner"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, operationJSON := range []string{
		`{"id":"bad.run","summary":"Bad"}`,
		`{"id":"bad.run","summary":"Bad","source":"core"}`,
	} {
		manifest := `{"schema":1,"name":"bad","version":"1","executable":"runner","operations":[` + operationJSON + `]}`
		if err := os.WriteFile(filepath.Join(root, "finishbit-extension.json"), []byte(manifest), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := ReadManifest(filepath.Join(root, "finishbit-extension.json")); err == nil {
			t.Fatalf("invalid operation source was accepted: %s", operationJSON)
		}
	}
}

func TestRemoveRejectsPathSegments(t *testing.T) {
	manager := New(t.TempDir())
	for _, name := range []string{".", "..", "../other", "bad/name"} {
		if err := manager.Remove(name); err == nil {
			t.Fatalf("Remove(%q) succeeded", name)
		}
	}
}

func TestInstallDirectoryRejectsFileOverSizeLimit(t *testing.T) {
	source := t.TempDir()
	executable := filepath.Join(source, "runner")
	if err := os.WriteFile(executable, []byte("runner"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{"schema":1,"name":"oversized","version":"1","executable":"runner","operations":[{"id":"example.run","summary":"Run","source":"extension"}]}`
	if err := os.WriteFile(filepath.Join(source, "finishbit-extension.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	large := filepath.Join(source, "large.bin")
	file, err := os.Create(large)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Truncate((512 << 20) + 1); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	manager := New(t.TempDir())
	if _, err := manager.Install(source); err == nil {
		t.Fatal("oversized directory extension was accepted")
	}
}
