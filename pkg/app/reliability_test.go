package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtensionRegistryRefreshAndRecovery(t *testing.T) {
	root := t.TempDir()
	source := t.TempDir()
	os.WriteFile(filepath.Join(source, "runner.exe"), []byte("fixture"), 0755)
	manifest := `{"schema":1,"name":"recovery","version":"1","executable":"runner.exe","operations":[{"id":"recovery.echo","summary":"Echo","source":"extension"}]}`
	os.WriteFile(filepath.Join(source, "finishbit-extension.json"), []byte(manifest), 0644)
	a, err := New(Config{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = a.InstallExtension(source); err != nil {
		t.Fatal(err)
	}
	if _, err = a.Describe("recovery.echo"); err != nil {
		t.Fatal(err)
	}
	if err = a.RemoveExtension("recovery"); err != nil {
		t.Fatal(err)
	}
	if _, err = a.Describe("recovery.echo"); err == nil {
		t.Fatal("removed operation remains registered")
	}
	dir := filepath.Join(root, "extensions", "broken")
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "finishbit-extension.json"), []byte("broken"), 0644)
	a, err = New(Config{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, check := range a.Doctor() {
		if check.Name == "extension:broken" && !check.OK {
			found = true
		}
	}
	if !found {
		t.Fatal("doctor hid broken extension")
	}
	if err = a.RemoveExtension("broken"); err != nil {
		t.Fatal(err)
	}
	// A valid manifest with a conflicting ID must not disable recovery commands.
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "runner.exe"), []byte("fixture"), 0755)
	os.WriteFile(filepath.Join(dir, "finishbit-extension.json"), []byte(`{"schema":1,"name":"broken","version":"1","executable":"runner.exe","operations":[{"id":"json.format","summary":"Conflict","source":"extension"}]}`), 0644)
	a, err = New(Config{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	found = false
	for _, check := range a.Doctor() {
		if check.Name == "extension:broken" && !check.OK {
			found = true
		}
	}
	if !found {
		t.Fatal("doctor hid conflicting extension")
	}
	if _, err = a.Describe("json.format"); err != nil {
		t.Fatal(err)
	}
	if err = a.RemoveExtension("broken"); err != nil {
		t.Fatal(err)
	}
}
