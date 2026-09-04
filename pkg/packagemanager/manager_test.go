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
