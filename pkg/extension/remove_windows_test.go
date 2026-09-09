//go:build windows

package extension

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestRemoveWaitsForLockedWindowsFile(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "example")
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "runner.exe")
	if err := os.WriteFile(path, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	name, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := syscall.CreateFile(name, syscall.GENERIC_READ, syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE, nil, syscall.OPEN_EXISTING, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() { defer close(done); time.Sleep(180 * time.Millisecond); _ = syscall.CloseHandle(handle) }()
	err = New(root).Remove("example")
	<-done
	if err != nil {
		t.Fatalf("short-lived lock prevented removal: %v", err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("directory still exists: %v", err)
	}
}
