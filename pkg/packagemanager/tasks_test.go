package packagemanager

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeFrozenTask(t *testing.T, w *taskWriter) {
	t.Helper()
	data, err := json.Marshal(w.task)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(w.directory, w.task.ID+".json"), data, 0600); err != nil {
		t.Fatal(err)
	}
}

func taskTestRegistry(url string) Registry {
	hash := sha256.Sum256([]byte("test executable"))
	return Registry{Schema: 1, Packages: []Package{{Name: "tool", Version: "1", Artifacts: map[string]Artifact{
		PlatformKey(): {URL: url, SHA256: hex.EncodeToString(hash[:]), Format: "raw", Executables: map[string]string{"tool": "tool"}},
	}}}}
}

func TestTaskProcessHelper(t *testing.T) {
	if os.Getenv("FINISHBIT_TASK_TEST_HELPER") != "1" {
		return
	}
	m := New(os.Getenv("FINISHBIT_TASK_TEST_ROOT"), taskTestRegistry(os.Getenv("FINISHBIT_TASK_TEST_URL")))
	m.client = &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}} // Local test server only.
	observed := false
	ctx := WithProgress(context.Background(), func(p Progress) { observed = true })
	if _, err := m.Install(ctx, "tool"); err != nil {
		t.Fatal(err)
	}
	if !observed {
		t.Fatal("existing CLI observer was lost")
	}
}

func waitTask(t *testing.T, m *Manager, predicate func(Task) bool) Task {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		tasks, err := m.Tasks()
		if err != nil {
			t.Fatal(err)
		}
		for _, task := range tasks {
			if predicate(task) {
				return task
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("task state did not arrive")
	return Task{}
}

func TestTaskProgressVisibleAcrossProcesses(t *testing.T) {
	release := make(chan struct{})
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "15")
		w.WriteHeader(200)
		w.(http.Flusher).Flush()
		select {
		case <-release:
			_, _ = w.Write([]byte("test executable"))
		case <-r.Context().Done():
		}
	}))
	defer server.Close()
	root := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestTaskProcessHelper$")
	cmd.Env = append(os.Environ(), "FINISHBIT_TASK_TEST_HELPER=1", "FINISHBIT_TASK_TEST_ROOT="+root, "FINISHBIT_TASK_TEST_URL="+server.URL)
	var output strings.Builder
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cmd.Process.Kill() }()
	reader := New(root, Registry{})
	active := waitTask(t, reader, func(task Task) bool { return task.Progress.Stage == "downloading" })
	if active.Origin != "cli" || active.PID == os.Getpid() || active.Progress.Total != 15 {
		t.Fatalf("bad cross-process snapshot: %+v", active)
	}
	competitor := New(root, taskTestRegistry(server.URL))
	if _, err := competitor.Install(context.Background(), "tool"); err == nil || !strings.Contains(err.Error(), "busy") {
		t.Fatalf("concurrent installation was not rejected: %v", err)
	}
	if err := competitor.Remove("tool"); err == nil || !strings.Contains(err.Error(), "busy") {
		t.Fatalf("removal during installation was not rejected: %v", err)
	}
	close(release)
	if err := cmd.Wait(); err != nil {
		t.Fatalf("helper: %v: %s", err, output.String())
	}
	// A newly opened UI reads the final state from disk after the writer exits.
	final := waitTask(t, New(root, Registry{}), func(task Task) bool { return task.ID == active.ID && task.Status == "done" })
	if final.Progress.Stage != "ready" {
		t.Fatalf("unexpected final progress: %+v", final)
	}
	unlock, err := competitor.lockPackage("tool")
	if err != nil {
		t.Fatalf("owner exit left package locked: %v", err)
	}
	unlock()
}

func TestBackgroundTaskSurvivesReaderRecreationAndCancelsWithOwner(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	defer server.Close()
	root := t.TempDir()
	m := New(root, taskTestRegistry(server.URL))
	m.client = server.Client()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	task, err := m.StartInstall(ctx, "tool", false)
	if err != nil {
		t.Fatal(err)
	}
	waitTask(t, New(root, Registry{}), func(s Task) bool {
		return s.ID == task.ID && s.Progress.Stage == "downloading" && s.Status == "running"
	})
	cancel()
	waitTask(t, New(root, Registry{}), func(s Task) bool { return s.ID == task.ID && s.Status == "cancelled" })
}

func TestStaleHeartbeatIsReconciledWithoutRewritingHistory(t *testing.T) {
	m := New(t.TempDir(), Registry{})
	w, err := m.newTask("tool", false, "cli")
	if err != nil {
		t.Fatal(err)
	}
	w.task.Updated = time.Now().Add(-time.Hour)
	// Write a frozen snapshot to model a killed process.
	writeFrozenTask(t, w)
	tasks, err := m.Tasks()
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].Status != "interrupted" {
		t.Fatalf("stale snapshot: %+v", tasks)
	}
	// A held OS lock proves package work is still active even if its heartbeat
	// cannot be written. Do not invite a second installation in this state.
	unlock, err := m.lockPackage("tool")
	if err != nil {
		t.Fatal(err)
	}
	tasks, err = m.Tasks()
	unlock()
	if err != nil || len(tasks) != 1 || tasks[0].Status != "recovering" {
		t.Fatalf("active owner lost during recovery: %+v, %v", tasks, err)
	}
	data, err := readTaskSnapshot(filepath.Join(w.directory, w.task.ID+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var persisted Task
	if err := json.Unmarshal(data, &persisted); err != nil || persisted.Status != "running" {
		t.Fatalf("reconciliation rewrote history: %+v, %v", persisted, err)
	}
	w.mu.Lock()
	_ = w.save()
	w.mu.Unlock()
	tasks, _ = m.Tasks()
	if tasks[0].Status != "running" {
		t.Fatal("reader modified writer state")
	}
}

func TestStaleTaskLockInspectionErrorRemainsUnknown(t *testing.T) {
	m := New(t.TempDir(), Registry{})
	w, err := m.newTask("tool", false, "web")
	if err != nil {
		t.Fatal(err)
	}
	w.task.Updated = time.Now().Add(-time.Hour)
	writeFrozenTask(t, w)
	// A filesystem failure is not evidence that the package is idle.
	if err := os.WriteFile(filepath.Join(m.root, "package-locks"), []byte("blocked"), 0600); err != nil {
		t.Fatal(err)
	}
	tasks, err := m.Tasks()
	if err != nil || len(tasks) != 1 || tasks[0].Status != "unknown" {
		t.Fatalf("inspection failure reported a stopped task: %+v, %v", tasks, err)
	}
}

func TestHistoricalInspectionSeparatesInstalledFilesAndOrphanDownloads(t *testing.T) {
	m := New(t.TempDir(), taskTestRegistry("https://example.invalid/tool"))
	w, err := m.newTask("tool", true, "web")
	if err != nil {
		t.Fatal(err)
	}
	w.task.Updated = time.Now().Add(-time.Hour)
	writeFrozenTask(t, w)
	dir := filepath.Join(m.root, "packages", "tool", ".install-legacy")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "download"), []byte("partial"), 0600); err != nil {
		t.Fatal(err)
	}
	inspection, err := m.InspectTask(w.task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Installation != "missing" || inspection.ResidualDownloads != 1 || inspection.ResidualBytes != 7 || inspection.CompletionConfirmed {
		t.Fatalf("wrong historical evidence: %+v", inspection)
	}
	unlock, err := m.lockPackage("tool")
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	inspection, err = m.InspectTask(w.task.ID)
	if err != nil || inspection.Status != "recovering" || inspection.Installation != "unchecked" {
		t.Fatalf("inspected files while package was mutating: %+v %v", inspection, err)
	}
}

func TestHistoricalInspectionConfirmsOnlyMatchingIntactInstallation(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("test executable"))
	}))
	defer server.Close()
	m := New(t.TempDir(), taskTestRegistry(server.URL))
	m.client = server.Client()
	installed, err := m.Install(context.Background(), "tool")
	if err != nil {
		t.Fatal(err)
	}
	tasks, err := m.Tasks()
	if err != nil || len(tasks) != 1 {
		t.Fatalf("tasks: %+v %v", tasks, err)
	}
	frozen := &taskWriter{directory: filepath.Join(m.root, "tasks"), task: tasks[0]}
	frozen.task.Status = "running"
	frozen.task.Updated = time.Now().Add(-time.Hour)
	writeFrozenTask(t, frozen)
	evidence, err := m.InspectTask(frozen.task.ID)
	if err != nil || evidence.Status != "done" || !evidence.CompletionConfirmed || evidence.Cache != "verified" {
		t.Fatalf("lost successful activation: %+v %v", evidence, err)
	}
	other, err := m.newTask("tool", true, "web")
	if err != nil {
		t.Fatal(err)
	}
	other.task.Updated = time.Now().Add(-time.Hour)
	writeFrozenTask(t, other)
	evidence, err = m.InspectTask(other.task.ID)
	if err != nil || evidence.CompletionConfirmed || evidence.Installation != "verified" {
		t.Fatalf("old installation incorrectly proves repair: %+v %v", evidence, err)
	}
	for _, executable := range installed.Executables {
		if err := os.WriteFile(filepath.Join(m.root, "packages", "tool", "1", executable), []byte("damaged"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	evidence, err = m.InspectTask(frozen.task.ID)
	if err != nil || evidence.CompletionConfirmed || evidence.Installation != "unverified" {
		t.Fatalf("damaged installation confirmed: %+v %v", evidence, err)
	}
}
