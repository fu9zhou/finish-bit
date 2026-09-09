package packagemanager

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

func TestReinstallRestoresDamageAndUpgradesLegacyManifestFromCache(t *testing.T) {
	content := []byte("verified runtime bytes")
	hash := sha256.Sum256(content)
	var requests atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { requests.Add(1); _, _ = w.Write(content) }))
	defer server.Close()
	m := New(t.TempDir(), Registry{Schema: 1, Packages: []Package{{Name: "tool", Version: "1", Artifacts: map[string]Artifact{PlatformKey(): {URL: server.URL, SHA256: hex.EncodeToString(hash[:]), Format: "raw", Executables: map[string]string{"tool": "tool"}}}}}})
	m.client = server.Client()
	if _, err := m.Install(context.Background(), "tool"); err != nil {
		t.Fatal(err)
	}
	path, err := m.Executable("tool", "tool")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("damaged"), 0600); err != nil {
		t.Fatal(err)
	}
	if m.Health()[0].Err == nil {
		t.Fatal("damage not detected")
	}
	if _, err := m.Install(context.Background(), "tool"); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(path); string(got) != string(content) {
		t.Fatal("cached reinstall did not restore bytes")
	}
	installed, err := m.Info("tool")
	if err != nil {
		t.Fatal(err)
	}
	installed.Files = nil
	data, _ := json.Marshal(installed)
	if err := os.WriteFile(filepath.Join(m.root, "packages", "tool", "1", "package.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	if m.Health()[0].Err == nil {
		t.Fatal("legacy installation misleadingly verified")
	}
	var stages []string
	ctx := WithProgress(context.Background(), func(event Progress) { stages = append(stages, event.Stage) })
	if _, err := m.Install(ctx, "tool"); err != nil {
		t.Fatal(err)
	}
	if m.Health()[0].Err != nil {
		t.Fatal(m.Health()[0].Err)
	}
	if requests.Load() != 1 {
		t.Fatalf("expected verified cache reuse; downloads=%d", requests.Load())
	}
	found := false
	for _, stage := range stages {
		if stage == "cache-hit" {
			found = true
		}
	}
	if !found {
		t.Fatal("no cache hit event")
	}
}
