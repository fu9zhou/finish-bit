package extension

import (
	"os"
	"path/filepath"
	"testing"
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

func TestRemoveRejectsPathSegments(t *testing.T) {
	manager := New(t.TempDir())
	for _, name := range []string{".", "..", "../other", "bad/name"} {
		if err := manager.Remove(name); err == nil {
			t.Fatalf("Remove(%q) succeeded", name)
		}
	}
}
