package packagemanager

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Task is a durable snapshot shared by processes using the same package root.
// Stale snapshots are reconciled against the package lock. Unknown is reserved
// for cases where the lock cannot be inspected reliably.
type Task struct {
	ID       string    `json:"id"`
	Package  string    `json:"package"`
	Action   string    `json:"action"`
	Origin   string    `json:"origin"`
	PID      int       `json:"pid"`
	Status   string    `json:"status"`
	Started  time.Time `json:"started"`
	Updated  time.Time `json:"updated"`
	Progress Progress  `json:"progress"`
	Error    string    `json:"error,omitempty"`
	Checked  time.Time `json:"checked,omitempty"`
}

const taskLease = 30 * time.Second

type taskOriginKey struct{}
type taskIDKey struct{}

func (m *Manager) lockPackage(name string) (func(), error) {
	directory := filepath.Join(m.root, "package-locks")
	if err := os.MkdirAll(directory, 0700); err != nil {
		return nil, err
	}
	hash := sha256.Sum256([]byte(name))
	unlock, err := lockPackageFile(filepath.Join(directory, hex.EncodeToString(hash[:])+".lock"))
	if err != nil {
		return nil, fmt.Errorf("package %q is busy or cannot be locked: %w", name, err)
	}
	return unlock, nil
}

type taskWriter struct {
	mu        sync.Mutex
	task      Task
	directory string
}

func (w *taskWriter) save() error {
	w.task.Updated = time.Now().UTC()
	data, err := json.Marshal(w.task)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(w.directory, ".snapshot-")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	_, err = f.Write(data)
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	// On Windows a reader without FILE_SHARE_DELETE makes replacement fail
	// with access denied, even though opening and reading the file succeeds.
	// Retry briefly, including sharing/lock violations from other readers.
	for attempt := 0; ; attempt++ {
		err = os.Rename(f.Name(), filepath.Join(w.directory, w.task.ID+".json"))
		if err == nil || runtime.GOOS != "windows" || attempt == 5 ||
			(!errors.Is(err, syscall.Errno(5)) && !errors.Is(err, syscall.Errno(32)) && !errors.Is(err, syscall.Errno(33))) {
			return err
		}
		time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond)
	}
}

func (m *Manager) newTask(name string, force bool, origin string) (*taskWriter, error) {
	directory := filepath.Join(m.root, "tasks")
	if err := os.MkdirAll(directory, 0700); err != nil {
		return nil, err
	}
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return nil, err
	}
	action := "install"
	if force {
		action = "repair"
	}
	w := &taskWriter{directory: directory, task: Task{
		ID: hex.EncodeToString(id[:]), Package: name, Action: action, Origin: origin,
		PID: os.Getpid(), Status: "running", Started: time.Now().UTC(),
		Progress: Progress{Package: name, Stage: "starting"},
	}}
	return w, w.save()
}

func (m *Manager) runTask(ctx context.Context, name string, force bool, w *taskWriter) (result Installed, err error) {
	done, stopped := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(stopped)
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				w.mu.Lock()
				_ = w.save()
				w.mu.Unlock()
			}
		}
	}()
	defer func() {
		close(done)
		<-stopped
		w.mu.Lock()
		defer w.mu.Unlock()
		w.task.Status = "done"
		if err != nil {
			w.task.Status = "failed"
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				w.task.Status = "cancelled"
			}
			w.task.Error = err.Error()
		}
		if saveErr := w.save(); saveErr != nil {
			err = errors.Join(err, fmt.Errorf("save final package task snapshot: %w", saveErr))
		}
	}()
	unlock, err := m.lockPackage(name)
	if err != nil {
		return Installed{}, err
	}
	defer unlock()
	ctx = context.WithValue(ctx, taskOriginKey{}, w.task.Origin)
	ctx = context.WithValue(ctx, taskIDKey{}, w.task.ID)
	observer, _ := ctx.Value(progressKey{}).(func(Progress))
	ctx = WithProgress(ctx, func(p Progress) {
		w.mu.Lock()
		w.task.Progress = p
		_ = w.save()
		w.mu.Unlock()
		if observer != nil {
			observer(p)
		}
	})
	return m.install(ctx, name, force)
}

func (m *Manager) trackedInstall(ctx context.Context, name string, force bool) (Installed, error) {
	origin, _ := ctx.Value(taskOriginKey{}).(string)
	if origin == "" {
		origin = "cli"
	}
	w, err := m.newTask(name, force, origin)
	if err != nil {
		return Installed{}, err
	}
	return m.runTask(ctx, name, force, w)
}

// StartInstall starts a web-owned task. Its lifetime is the supplied service
// context, not an HTTP request. The snapshot exists before this call returns.
func (m *Manager) StartInstall(ctx context.Context, name string, force bool) (Task, error) {
	w, err := m.newTask(name, force, "web")
	if err != nil {
		return Task{}, err
	}
	task := w.task
	go func() { _, _ = m.runTask(ctx, name, force, w) }()
	return task, nil
}

// Tasks returns recent snapshots, including all active tasks. Readers never
// modify another process's state; a stale heartbeat is a presentation state.
func (m *Manager) Tasks() ([]Task, error) {
	entries, err := os.ReadDir(filepath.Join(m.root, "tasks"))
	if os.IsNotExist(err) {
		return []Task{}, nil
	}
	if err != nil {
		return nil, err
	}
	tasks := []Task{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := readTaskSnapshot(filepath.Join(m.root, "tasks", entry.Name()))
		if err != nil {
			return nil, err
		}
		var task Task
		if json.Unmarshal(data, &task) != nil {
			continue
		}
		if task.Status == "running" && time.Since(task.Updated) > taskLease {
			task = m.reconcileTask(task)
		}
		task.Checked = time.Now().UTC()
		tasks = append(tasks, task)
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Started.After(tasks[j].Started) })
	result := []Task{}
	for _, task := range tasks {
		if len(result) < 100 || task.Status == "running" || task.Status == "recovering" || task.Status == "unknown" {
			result = append(result, task)
		}
	}
	return result, nil
}

// An OS lock survives a stalled writer but is released when its process exits.
// Probe without waiting; never modify persisted history or infer completion
// from an older installation that may have existed before a repair began.
func (m *Manager) reconcileTask(task Task) Task {
	unlock, err := m.lockPackage(task.Package)
	if err != nil {
		task.Status = "unknown"
		if packageLockBusy(err) {
			task.Status = "recovering"
		}
		return task
	}
	defer unlock()
	// The writer may have completed between the initial read and lock probe.
	data, err := readTaskSnapshot(filepath.Join(m.root, "tasks", task.ID+".json"))
	if err != nil {
		task.Status = "unknown"
		return task
	}
	var latest Task
	if json.Unmarshal(data, &latest) != nil || latest.ID != task.ID || latest.Package != task.Package {
		task.Status = "unknown"
		return task
	}
	if latest.Status != "running" || time.Since(latest.Updated) <= taskLease {
		return latest
	}
	task.Status = "interrupted"
	return task
}

// Windows can briefly deny reads while a snapshot is replaced or scanned.
// Retry only sharing/lock violations; persistent and unrelated errors surface.
func readTaskSnapshot(path string) ([]byte, error) {
	for attempt := 0; ; attempt++ {
		data, err := os.ReadFile(path)
		if err == nil || runtime.GOOS != "windows" || attempt == 5 ||
			(!errors.Is(err, syscall.Errno(32)) && !errors.Is(err, syscall.Errno(33))) {
			return data, err
		}
		time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond)
	}
}
