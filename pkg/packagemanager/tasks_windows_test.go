//go:build windows

package packagemanager

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestTaskSnapshotSaveSurvivesShortWindowsSharingLock(t *testing.T) {
	m := New(t.TempDir(), Registry{})
	w, err := m.newTask("tool", false, "cli")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(w.directory, w.task.ID+".json")
	file, err := os.Open(path) // Go readers allow reads/writes, but not deletion.
	if err != nil {
		t.Fatal(err)
	}
	released := make(chan struct{})
	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = file.Close()
		close(released)
	}()
	defer func() { <-released }()
	w.task.Status = "done"
	if err := w.save(); err != nil {
		t.Fatal(err)
	}
	data, err := readTaskSnapshot(path)
	if err != nil {
		t.Fatal(err)
	}
	var task Task
	if err := json.Unmarshal(data, &task); err != nil || task.Status != "done" {
		t.Fatalf("terminal snapshot was lost: %+v %v", task, err)
	}
}

func TestInstallReportsTerminalSnapshotSharingFailure(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("test executable"))
	}))
	defer server.Close()
	m := New(t.TempDir(), taskTestRegistry(server.URL))
	m.client = server.Client()
	var reader *os.File
	ctx := WithProgress(context.Background(), func(p Progress) {
		if p.Stage != "ready" {
			return
		}
		tasks, err := m.Tasks()
		if err != nil || len(tasks) != 1 {
			t.Fatalf("tasks: %+v %v", tasks, err)
		}
		reader, err = os.Open(filepath.Join(m.root, "tasks", tasks[0].ID+".json"))
		if err != nil {
			t.Fatal(err)
		}
	})
	defer func() {
		if reader != nil {
			_ = reader.Close()
		}
	}()
	installed, err := m.Install(ctx, "tool")
	if err == nil || !strings.Contains(err.Error(), "save final package task snapshot") {
		t.Fatalf("final persistence failure was hidden: %v", err)
	}
	if installed.Name != "tool" {
		t.Fatalf("successful installation result was lost: %+v", installed)
	}
}

func TestTaskSnapshotPersistentSharingLockIsReported(t *testing.T) {
	path := filepath.Join(t.TempDir(), "snapshot.json")
	if err := os.WriteFile(path, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	name, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := syscall.CreateFile(name, syscall.GENERIC_READ, 0, nil, syscall.OPEN_EXISTING, syscall.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer syscall.CloseHandle(handle)
	if _, err := readTaskSnapshot(path); !errors.Is(err, syscall.Errno(32)) {
		t.Fatalf("persistent lock was hidden: %v", err)
	}
	if _, err := readTaskSnapshot(filepath.Join(t.TempDir(), "missing")); !os.IsNotExist(err) {
		t.Fatalf("missing file was hidden: %v", err)
	}
}

func TestTasksReadSurvivesShortWindowsSharingLock(t *testing.T) {
	m := New(t.TempDir(), Registry{})
	w, err := m.newTask("tool", false, "cli")
	if err != nil {
		t.Fatal(err)
	}
	name, err := syscall.UTF16PtrFromString(filepath.Join(w.directory, w.task.ID+".json"))
	if err != nil {
		t.Fatal(err)
	}
	handle, err := syscall.CreateFile(name, syscall.GENERIC_READ, 0, nil, syscall.OPEN_EXISTING, syscall.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		t.Fatal(err)
	}
	released := make(chan struct{})
	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = syscall.CloseHandle(handle)
		close(released)
	}()
	defer func() { <-released }()
	tasks, err := m.Tasks()
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].ID != w.task.ID {
		t.Fatalf("lost task: %+v", tasks)
	}
}
