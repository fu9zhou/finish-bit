package packagemanager

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/ulikunitz/xz"
)

func TestArchiveInstallPreservesRuntimeLayout(t *testing.T) {
	var data bytes.Buffer
	archive := zip.NewWriter(&data)
	for name, content := range map[string]string{"release/bin/tool.exe": "executable", "release/bin/support.dll": "library", "release/LICENSE": "license"} {
		w, _ := archive.Create(name)
		_, _ = w.Write([]byte(content))
	}
	_ = archive.Close()
	hash := sha256.Sum256(data.Bytes())
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(data.Bytes()) }))
	defer server.Close()
	manager := New(t.TempDir(), Registry{Schema: 1, Packages: []Package{{Name: "tool", Version: "1", Artifacts: map[string]Artifact{PlatformKey(): {URL: server.URL, Format: "zip", SHA256: hex.EncodeToString(hash[:]), Executables: map[string]string{"tool": "release/bin/tool.exe"}}}}}})
	manager.client = server.Client()
	if _, err := manager.Install(context.Background(), "tool"); err != nil {
		t.Fatal(err)
	}
	path, err := manager.Executable("tool", "tool")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(path), "support.dll")); err != nil {
		t.Fatal("runtime DLL missing:", err)
	}
}

func TestArchiveRejectsUnsafeEntries(t *testing.T) {
	for _, name := range []string{"../escape", "/absolute", "C:/absolute", "ok/../../escape", "ok\\..\\escape", "file:stream", "NUL.txt", "ok/CON", "bad. ", "."} {
		t.Run(name, func(t *testing.T) {
			var data bytes.Buffer
			archive := zip.NewWriter(&data)
			w, _ := archive.Create(name)
			_, _ = w.Write([]byte("bad"))
			_ = archive.Close()
			root := t.TempDir()
			source := filepath.Join(root, "bad.zip")
			_ = os.WriteFile(source, data.Bytes(), 0600)
			if err := extractArchive(context.Background(), source, filepath.Join(root, "out"), "zip"); err == nil {
				t.Fatal("unsafe archive accepted")
			}
		})
	}
	t.Run("symlink", func(t *testing.T) {
		var data bytes.Buffer
		archive := zip.NewWriter(&data)
		header := &zip.FileHeader{Name: "link"}
		header.SetMode(os.ModeSymlink | 0777)
		w, _ := archive.CreateHeader(header)
		_, _ = w.Write([]byte("../target"))
		_ = archive.Close()
		root := t.TempDir()
		source := filepath.Join(root, "link.zip")
		_ = os.WriteFile(source, data.Bytes(), 0600)
		if err := extractArchive(context.Background(), source, filepath.Join(root, "out"), "zip"); err == nil {
			t.Fatal("symlink accepted")
		}
	})
}

func TestTarXZExtraction(t *testing.T) {
	var data bytes.Buffer
	x, err := xz.NewWriter(&data)
	if err != nil {
		t.Fatal(err)
	}
	archive := tar.NewWriter(x)
	_ = archive.WriteHeader(&tar.Header{Name: "tool/bin", Mode: 0755, Size: 4, Typeflag: tar.TypeReg})
	_, _ = archive.Write([]byte("tool"))
	_ = archive.Close()
	_ = x.Close()
	root := t.TempDir()
	source := filepath.Join(root, "tool.tar.xz")
	_ = os.WriteFile(source, data.Bytes(), 0600)
	out := filepath.Join(root, "out")
	if err := extractArchive(context.Background(), source, out, "tar.xz"); err != nil {
		t.Fatal(err)
	}
	actual, _ := os.ReadFile(filepath.Join(out, "tool", "bin"))
	if string(actual) != "tool" {
		t.Fatal("extraction mismatch")
	}
}

func TestSelectedArchiveResources(t *testing.T) {
	var data bytes.Buffer
	archive := zip.NewWriter(&data)
	for name, body := range map[string]string{"bin/tool.exe": "program", "bin/support.dll": "dependency", "LICENSE": "license", "alternate.exe": "unused variant"} {
		w, e := archive.Create(name)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = w.Write([]byte(body)); e != nil {
			t.Fatal(e)
		}
	}
	if e := archive.Close(); e != nil {
		t.Fatal(e)
	}
	root := t.TempDir()
	source := filepath.Join(root, "runtime.zip")
	if e := os.WriteFile(source, data.Bytes(), 0600); e != nil {
		t.Fatal(e)
	}
	target := filepath.Join(root, "selected")
	if e := extractArchive(context.Background(), source, target, "zip", []string{"bin/tool.exe", "bin/support.dll", "LICENSE"}); e != nil {
		t.Fatal(e)
	}
	if _, e := os.Stat(filepath.Join(target, "alternate.exe")); !os.IsNotExist(e) {
		t.Fatal("alternate build extracted", e)
	}
	for _, name := range []string{"bin/tool.exe", "bin/support.dll", "LICENSE"} {
		if _, e := os.Stat(filepath.Join(target, name)); e != nil {
			t.Fatal("missing selected resource", name, e)
		}
	}
	if e := extractArchive(context.Background(), source, filepath.Join(root, "missing"), "zip", []string{"absent.dll"}); e == nil {
		t.Fatal("missing resource accepted")
	}
}
