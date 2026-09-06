package packagemanager

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallVerifiesAndActivatesPackage(t *testing.T) {
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	_, _ = writer.Write([]byte("test executable"))
	_ = writer.Close()
	hash := sha256.Sum256(compressed.Bytes())
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) { _, _ = response.Write(compressed.Bytes()) }))
	defer server.Close()
	registry := Registry{Schema: 1, Packages: []Package{{Name: "tool", Version: "1.0.0", Artifacts: map[string]Artifact{PlatformKey(): {URL: server.URL, SHA256: hex.EncodeToString(hash[:]), Format: "gzip", Executables: map[string]string{"tool": "tool"}}}}}}
	manager := New(t.TempDir(), registry)
	manager.client = server.Client()
	installed, err := manager.Install(context.Background(), "tool")
	if err != nil {
		t.Fatal(err)
	}
	path, err := manager.Executable("tool", "tool")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "test executable" {
		t.Fatalf("content = %q", data)
	}
	if installed.Version != "1.0.0" {
		t.Fatalf("version = %q", installed.Version)
	}
	if filepath.Dir(path) == "" {
		t.Fatal("executable path has no parent")
	}
	if manager.Root() == "" {
		t.Fatal("manager root is empty")
	}
	status, err := manager.Status("tool")
	if err != nil || status.Installed == nil || !status.Supported {
		t.Fatalf("Status() = %#v, %v", status, err)
	}
	if values := manager.List(); len(values) != 1 || values[0].Name != "tool" {
		t.Fatalf("List() = %#v", values)
	}
	if err := manager.Remove("tool"); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Executable("tool", "tool"); err == nil {
		t.Fatal("removed package remained executable")
	}
}

func TestInstallRejectsBadChecksum(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) { _, _ = response.Write([]byte("bad")) }))
	defer server.Close()
	registry := Registry{Schema: 1, Packages: []Package{{Name: "tool", Version: "1", Artifacts: map[string]Artifact{PlatformKey(): {URL: server.URL, SHA256: strings.Repeat("0", 64), Format: "raw", Executables: map[string]string{"tool": "tool"}}}}}}
	manager := New(t.TempDir(), registry)
	manager.client = server.Client()
	if _, err := manager.Install(context.Background(), "tool"); err == nil {
		t.Fatal("bad checksum was accepted")
	}
}

func TestInstallFallsBackToVerifiedMirror(t *testing.T) {
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	_, _ = writer.Write([]byte("mirrored executable"))
	_ = writer.Close()
	hash := sha256.Sum256(compressed.Bytes())
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/primary" {
			response.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = response.Write(compressed.Bytes())
	}))
	defer server.Close()
	registry := Registry{Schema: 1, Packages: []Package{{Name: "tool", Version: "1", Artifacts: map[string]Artifact{PlatformKey(): {URL: server.URL + "/primary", Mirrors: []string{server.URL + "/mirror"}, SHA256: hex.EncodeToString(hash[:]), Format: "gzip", Executables: map[string]string{"tool": "tool"}}}}}}
	manager := New(t.TempDir(), registry)
	manager.client = server.Client()
	if _, err := manager.Install(context.Background(), "tool"); err != nil {
		t.Fatal(err)
	}
}

func TestRepairFailureKeepsCurrentPackageUsable(t *testing.T) {
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	_, _ = writer.Write([]byte("working executable"))
	_ = writer.Close()
	hash := sha256.Sum256(compressed.Bytes())
	fail := false
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		if fail {
			response.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = response.Write(compressed.Bytes())
	}))
	defer server.Close()
	registry := Registry{Schema: 1, Packages: []Package{{Name: "tool", Version: "1", Artifacts: map[string]Artifact{PlatformKey(): {URL: server.URL, SHA256: hex.EncodeToString(hash[:]), Format: "gzip", Executables: map[string]string{"tool": "tool"}}}}}}
	manager := New(t.TempDir(), registry)
	manager.client = server.Client()
	if _, err := manager.Install(context.Background(), "tool"); err != nil {
		t.Fatal(err)
	}

	fail = true
	if _, err := manager.Repair(context.Background(), "tool"); err == nil {
		t.Fatal("Repair() succeeded after the package source failed")
	}
	path, err := manager.Executable("tool", "tool")
	if err != nil {
		t.Fatalf("current package is no longer usable: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "working executable" {
		t.Fatalf("current executable changed to %q", data)
	}
}

func TestInstallRawPackageAndRepairIt(t *testing.T) {
	content := []byte("first executable")
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) { _, _ = response.Write(content) }))
	defer server.Close()
	makeArtifact := func() Artifact {
		hash := sha256.Sum256(content)
		return Artifact{URL: server.URL, SHA256: hex.EncodeToString(hash[:]), Format: "raw", Executables: map[string]string{"tool": "tool"}}
	}
	manager := New(t.TempDir(), Registry{Schema: 1, Packages: []Package{{Name: "tool", Version: "1", Artifacts: map[string]Artifact{PlatformKey(): makeArtifact()}}}})
	manager.client = server.Client()
	if _, err := manager.Install(context.Background(), "tool"); err != nil {
		t.Fatal(err)
	}
	content = []byte("repaired executable")
	manager.registry.Packages[0].Artifacts[PlatformKey()] = makeArtifact()
	if _, err := manager.Repair(context.Background(), "tool"); err != nil {
		t.Fatal(err)
	}
	path, err := manager.Executable("tool", "tool")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "repaired executable" {
		t.Fatalf("repaired executable = %q", data)
	}
}

func TestBuiltinRegistryIsValid(t *testing.T) {
	registry, err := BuiltinRegistry()
	if err != nil {
		t.Fatal(err)
	}
	if pkg, ok := registry.Find("ffmpeg"); !ok || pkg.Version == "" {
		t.Fatalf("ffmpeg package = %#v, %v", pkg, ok)
	}
}

func TestCopyBoundedRejectsOverflow(t *testing.T) {
	var output bytes.Buffer
	if err := copyBounded(&output, strings.NewReader("12345"), 4, "test data"); err == nil {
		t.Fatal("oversized data was silently truncated")
	}
}
