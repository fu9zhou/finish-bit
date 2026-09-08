package packagemanager

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fu9zhou/finish-bit/internal/toolrun"
)

func TestSevenzipListingValidation(t *testing.T) {
	row := func(name, size, attributes string) string {
		return fmt.Sprintf("Path = %s\nSize = %s\nAttributes = %s\nEncrypted = -\nBlock = \n\n", name, size, attributes)
	}
	good := row("bin", "0", "D") + row("bin\\tool.exe", "100", "A") + row("empty.txt", "0", "A")
	entries, err := sevenzipEntries(strings.ReplaceAll(good, "\n", "\r\n"), t.TempDir())
	if err != nil || len(entries) != 3 || entries[1].name != "bin/tool.exe" || !entries[0].dir {
		t.Fatalf("%+v %v", entries, err)
	}
	for name, bad := range map[string]string{
		"traversal":        row("../escape", "1", "A"),
		"reserved":         row("NUL.txt", "1", "A"),
		"absolute":         row("C:/outside", "1", "A"),
		"ads":              row("file:stream", "1", "A"),
		"link":             row("link", "1", "A_ lrwxrwxrwx"),
		"hardlink":         row("link", "1", "A") + "Hard Link = other\n",
		"special":          row("pipe", "1", "A_ prw-r--r--"),
		"duplicate":        good + row("BIN/TOOL.EXE", "1", "A"),
		"parent conflict":  row("bin/tool", "1", "A") + row("BIN", "1", "A"),
		"duplicate field":  "Path = x\nPath = y\nSize = 0\nEncrypted = -\n",
		"newline in name":  "Path = bad\nname\nSize = 0\nEncrypted = -\n",
		"negative size":    row("x", "-1", "A"),
		"oversize":         row("x", "1073741825", "A"),
		"aggregate limit":  row("x", "1073741824", "A") + row("y", "1", "A"),
		"nonempty folder":  row("bin", "1", "D"),
		"encrypted":        strings.Replace(good, "Encrypted = -", "Encrypted = +", 1),
		"unknown encoding": "Path = x\nSize = 1\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := sevenzipEntries(bad, t.TempDir()); err == nil {
				t.Fatal("unsafe listing accepted")
			}
		})
	}
	var many strings.Builder
	for i := 0; i <= 50000; i++ {
		many.WriteString(row(fmt.Sprint(i), "0", "A"))
	}
	if _, err := sevenzipEntries(many.String(), t.TempDir()); err == nil {
		t.Fatal("entry count limit ignored")
	}
}

func TestSevenzipReusesInstalledManagedReader(t *testing.T) {
	r, err := BuiltinRegistry()
	if err != nil {
		t.Fatal(err)
	}
	m := New(t.TempDir(), r)
	pkg, _ := r.Find("7zip-full")
	dir := filepath.Join(m.root, "packages", pkg.Name, pkg.Version)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "reader"), []byte("placeholder"), 0600); err != nil {
		t.Fatal(err)
	}
	metadata, err := json.Marshal(Installed{Name: pkg.Name, Version: pkg.Version, Platform: PlatformKey(), Executables: map[string]string{"7zip": "reader"}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "package.json"), metadata, 0600); err != nil {
		t.Fatal(err)
	}
	got, err := m.sevenzipExtractor(context.Background(), "imagemagick")
	if err != nil || got != "7zip-full" {
		t.Fatalf("did not reuse installed reader: %q %v", got, err)
	}
	if _, err := m.sevenzipExtractor(context.Background(), "7zip-bootstrap"); err == nil {
		t.Fatal("bootstrap extraction cycle accepted")
	}
	bootstrap, ok := r.Find("7zip-bootstrap")
	if !ok {
		t.Fatal("missing bootstrap")
	}
	for _, platform := range []string{"win32-x64", "win32-arm64"} {
		if bootstrap.Artifacts[platform].Format != "raw" {
			t.Fatalf("%s bootstrap is not independent of archive extraction", platform)
		}
	}
}

// Uses an already installed official 7-Zip; never downloads during unit tests.
func TestSevenzipManagedStreams(t *testing.T) {
	home := os.Getenv("FINISHBIT_TEST_7ZIP_HOME")
	if home == "" {
		t.Skip("set FINISHBIT_TEST_7ZIP_HOME to exercise the managed 7-Zip process")
	}
	r, err := BuiltinRegistry()
	if err != nil {
		t.Fatal(err)
	}
	m := New(home, r)
	ctx := context.Background()
	reader, err := m.sevenzipExtractor(ctx, "fixture")
	if err != nil {
		t.Fatal(err)
	}
	work := t.TempDir()
	in := filepath.Join(work, "in")
	want := map[string]string{"same.txt": "root", "nested/same.txt": "nested", "@list.txt": "literal list name", "-switch.txt": "literal switch name", "中文 空格.txt": "中文😀", "empty.txt": ""}
	for name, content := range want {
		path := filepath.Join(in, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(in, "empty-folder"), 0755); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(work, "fixture.7z")
	if _, err := toolrun.Run(ctx, m, reader, "7zip", in, "a", "-t7z", "-ms=on", source, "."); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(work, "out")
	if err := m.extract7z(ctx, "fixture", source, out, nil); err != nil {
		t.Fatal(err)
	}
	for name, content := range want {
		actual, err := os.ReadFile(filepath.Join(out, filepath.FromSlash(name)))
		if err != nil || string(actual) != content {
			t.Fatalf("%s: %q %v", name, actual, err)
		}
	}
	selected := filepath.Join(work, "selected")
	if err := m.extract7z(ctx, "fixture", source, selected, []string{"nested/same.txt"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(selected, "same.txt")); !os.IsNotExist(err) {
		t.Fatal("unselected file extracted", err)
	}
	if err := m.extract7z(ctx, "fixture", source, filepath.Join(work, "missing"), []string{"missing.txt"}); err == nil {
		t.Fatal("missing resource accepted")
	}
	if err := m.extract7z(ctx, "fixture", source, out, nil); err == nil {
		t.Fatal("existing output overwritten")
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if err := m.extract7z(cancelled, "fixture", source, filepath.Join(work, "cancelled"), nil); err == nil {
		t.Fatal("cancelled extraction accepted")
	}
}
