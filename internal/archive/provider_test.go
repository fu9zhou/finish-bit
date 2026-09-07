package archive

import (
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

type runtimeResolver string

func (r runtimeResolver) Executable(string, string) (string, error) { return string(r), nil }
func TestListingBoundaries(t *testing.T) {
	for _, name := range []string{"../escape", "/root", "C:\\x", "a/../b", "NUL.txt", "a:stream", "a/CON", "a\nSize = 0", "a.", "a//b", "@list", "a?b", "a\x00b"} {
		if _, e := safeName(name); e == nil {
			t.Errorf("accepted %q", name)
		}
	}
	for _, body := range []string{"Path = a\nSize = 1\nSize = 2", "Path = a\nSize = -1", "Path = a\nSize = 67108865", "Path = a\nSize = 0\n\nPath = A\nSize = 0", "Path = a\nSize = 0\n\nPath = a/b\nSize = 0", "Path = a\nSize = 0\nAttributes = A lrwxrwxrwx", "Path = a\nSize = 0\nHard Link = x"} {
		if _, e := parseListing([]byte(body)); e == nil {
			t.Errorf("accepted listing %q", body)
		}
	}
}
func TestReal7Zip(t *testing.T) {
	binary := os.Getenv("FINISHBIT_TEST_7ZIP")
	if binary == "" {
		t.Skip("set FINISHBIT_TEST_7ZIP to managed 7za executable")
	}
	registry := operation.NewRegistry()
	if e := Register(registry, runtimeResolver(binary)); e != nil {
		t.Fatal(e)
	}
	root := t.TempDir()
	write := func(name string, data []byte) string {
		path := filepath.Join(root, name)
		if e := os.MkdirAll(filepath.Dir(path), 0755); e != nil {
			t.Fatal(e)
		}
		if e := os.WriteFile(path, data, 0600); e != nil {
			t.Fatal(e)
		}
		return path
	}
	input := write("资料/first.txt", []byte("Hello archive 42\n"))
	write("资料/nested/第二.txt", []byte("中文内容\n"))
	if e := os.Mkdir(filepath.Join(root, "资料/empty"), 0755); e != nil {
		t.Fatal(e)
	}
	call := func(id string, inputs []string, opts map[string]any) operation.Result {
		t.Helper()
		c, ok := registry.Get(id)
		if !ok {
			t.Fatal(id)
		}
		r, e := c.Runner.Run(context.Background(), operation.Request{Inputs: inputs, Options: opts})
		if e != nil {
			t.Fatalf("%s: %v (%+v)", id, e, operation.AsError(e).Details)
		}
		return r
	}
	contents := func(path string) []byte {
		t.Helper()
		data, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		return data
	}
	call("archive.formats", nil, nil)
	for _, format := range writeFormats {
		t.Run(format, func(t *testing.T) {
			source := filepath.Dir(input)
			if format == "gzip" || format == "bzip2" || format == "xz" {
				source = input
			}
			archive := filepath.Join(root, "bundle."+format)
			call("archive.create", []string{source}, map[string]any{"format": format, "output": archive})
			listed := call("archive.list", []string{archive}, map[string]any{"format": format})
			if listed.Data["count"].(int) == 0 {
				t.Fatal("empty listing")
			}
			call("archive.test", []string{archive}, map[string]any{"format": format})
			destination := filepath.Join(root, "out-"+format)
			call("archive.extract", []string{archive}, map[string]any{"format": format, "output": destination})
			found := false
			if e := filepath.WalkDir(destination, func(path string, d os.DirEntry, e error) error {
				if e != nil {
					return e
				}
				if !d.IsDir() && string(contents(path)) == "Hello archive 42\n" {
					found = true
				}
				return nil
			}); e != nil {
				t.Fatal(e)
			}
			if !found {
				t.Fatal("round-trip lost file bytes")
			}
		})
	}
	original := filepath.Join(root, "bundle.zip")
	digest := sha256.Sum256(contents(original))
	selectedOut := filepath.Join(root, "selected")
	call("archive.extract", []string{original}, map[string]any{"entries": []string{"资料/nested"}, "output": selectedOut})
	if string(contents(filepath.Join(selectedOut, "资料/nested/第二.txt"))) != "中文内容\n" {
		t.Fatal("selected extraction lost Unicode bytes")
	}
	if _, e := os.Stat(filepath.Join(selectedOut, "资料/first.txt")); !os.IsNotExist(e) {
		t.Fatal("selection extracted unrelated file")
	}
	for _, kind := range []string{"zip", "7z"} {
		source := filepath.Join(root, "bundle."+kind)
		added := write("extra.txt", []byte("Added data"))
		updated := filepath.Join(root, "updated."+kind)
		call("archive.update", []string{source}, map[string]any{"files": []string{added}, "output": updated})
		listed := call("archive.list", []string{updated}, nil)
		found := false
		for _, item := range listed.Data["entries"].([]entry) {
			if item.Path == "extra.txt" {
				found = true
			}
		}
		if !found {
			t.Fatal("update omitted new file")
		}
		renamed := filepath.Join(root, "renamed."+kind)
		call("archive.rename", []string{updated}, map[string]any{"entry": "extra.txt", "name": "renamed/deep.txt", "output": renamed})
		removed := filepath.Join(root, "removed."+kind)
		call("archive.remove", []string{renamed}, map[string]any{"entries": []string{"资料"}, "output": removed})
		dest := filepath.Join(root, "mutated-"+kind)
		call("archive.extract", []string{removed}, map[string]any{"output": dest})
		if string(contents(filepath.Join(dest, "renamed/deep.txt"))) != "Added data" {
			t.Fatal("mutation changed file bytes")
		}
	}
	repacked := filepath.Join(root, "repacked.tar.xz")
	call("archive.repack", []string{original}, map[string]any{"to": "tar.xz", "output": repacked})
	call("archive.test", []string{repacked}, nil)
	t.Setenv("FINISHBIT_ARCHIVE_TEST_PASSWORD", "fixture-password-123")
	for _, kind := range []string{"zip", "7z"} {
		encrypted := filepath.Join(root, "encrypted."+kind)
		call("archive.create", []string{input}, map[string]any{"format": kind, "password-env": "FINISHBIT_ARCHIVE_TEST_PASSWORD", "output": encrypted})
		call("archive.test", []string{encrypted}, map[string]any{"password-env": "FINISHBIT_ARCHIVE_TEST_PASSWORD"})
		call("archive.extract", []string{encrypted}, map[string]any{"password-env": "FINISHBIT_ARCHIVE_TEST_PASSWORD", "output": filepath.Join(root, "decrypted-"+kind)})
		c, _ := registry.Get("archive.extract")
		badOut := filepath.Join(root, "wrong-"+kind)
		_, e := c.Runner.Run(context.Background(), operation.Request{Inputs: []string{encrypted}, Options: map[string]any{"output": badOut}})
		if e == nil {
			t.Fatal("encrypted archive accepted without password")
		}
		if _, e = os.Stat(badOut); !os.IsNotExist(e) {
			t.Fatal("failed extraction published destination")
		}
	}
	data := make([]byte, 3<<20)
	for i := range data {
		data[i] = byte(i * 13)
	}
	large := write("large.bin", data)
	for _, kind := range []string{"zip", "7z"} {
		volumes := call("archive.volumes", []string{large}, map[string]any{"format": kind, "level": 0, "volume-mib": 1, "output": filepath.Join(root, "volumes-"+kind)})
		if len(volumes.Outputs) < 3 {
			t.Fatal("volume split did not create multiple files")
		}
		first := volumes.Outputs[0]
		dest := filepath.Join(root, "volume-out-"+kind)
		call("archive.extract", []string{first}, map[string]any{"output": dest})
		if sha256.Sum256(contents(filepath.Join(dest, "large.bin"))) != sha256.Sum256(data) {
			t.Fatal("volume round-trip differs")
		}
	}
	if sha256.Sum256(contents(original)) != digest {
		t.Fatal("mutation changed original archive")
	}
	existing := write("existing.zip", []byte("keep"))
	c, _ := registry.Get("archive.create")
	_, e := c.Runner.Run(context.Background(), operation.Request{Inputs: []string{input}, Options: map[string]any{"output": existing}})
	if e == nil || string(contents(existing)) != "keep" {
		t.Fatal("overwrote existing file")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, e = c.Runner.Run(ctx, operation.Request{Inputs: []string{input}, Options: map[string]any{"output": existing, "overwrite": true}})
	if e == nil || string(contents(existing)) != "keep" {
		t.Fatal("cancelled archive changed destination")
	}
	// A real traversal archive must fail before any destination is published.
	bad := filepath.Join(root, "traversal.zip")
	f, e := os.Create(bad)
	if e != nil {
		t.Fatal(e)
	}
	z := zip.NewWriter(f)
	w, e := z.Create("../escaped.txt")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = w.Write([]byte("escape")); e != nil {
		t.Fatal(e)
	}
	if e = z.Close(); e != nil {
		t.Fatal(e)
	}
	if e = f.Close(); e != nil {
		t.Fatal(e)
	}
	extractor, _ := registry.Get("archive.extract")
	destination := filepath.Join(root, "unsafe-output")
	_, e = extractor.Runner.Run(context.Background(), operation.Request{Inputs: []string{bad}, Options: map[string]any{"output": destination}})
	if e == nil {
		t.Fatal("traversal archive accepted")
	}
	if _, e = os.Stat(destination); !os.IsNotExist(e) {
		t.Fatal("unsafe destination published")
	}
	if runtime.GOOS == "windows" {
		zone := []byte("[ZoneTransfer]\r\nZoneId=3\r\n")
		if e := os.WriteFile(original+":Zone.Identifier", zone, 0600); e != nil {
			t.Fatal(e)
		}
		marked := filepath.Join(root, "marked")
		call("archive.extract", []string{original}, map[string]any{"output": marked})
		if string(contents(filepath.Join(marked, "资料/first.txt")+":Zone.Identifier")) != string(zone) {
			t.Fatal("download origin mark was lost")
		}
		markCopy := filepath.Join(root, "marked-copy.7z")
		call("archive.repack", []string{original}, map[string]any{"to": "7z", "output": markCopy})
		if string(contents(markCopy+":Zone.Identifier")) != string(zone) {
			t.Fatal("repack lost origin mark")
		}
	}
	encrypted := filepath.Join(root, "encrypted.zip")
	unencrypted := filepath.Join(root, "unencrypted.tar")
	call("archive.repack", []string{encrypted}, map[string]any{"password-env": "FINISHBIT_ARCHIVE_TEST_PASSWORD", "drop-password": true, "to": "tar", "output": unencrypted})
	call("archive.test", []string{unencrypted}, nil)
	t.Setenv("FINISHBIT_ARCHIVE_WRONG_PASSWORD", "wrong")
	_, e = extractor.Runner.Run(context.Background(), operation.Request{Inputs: []string{encrypted}, Options: map[string]any{"password-env": "FINISHBIT_ARCHIVE_WRONG_PASSWORD", "output": filepath.Join(root, "wrong-password")}})
	if e == nil {
		t.Fatal("wrong password succeeded")
	}
	bomb := filepath.Join(root, "oversized.gz")
	bf, e := os.Create(bomb)
	if e != nil {
		t.Fatal(e)
	}
	gz := gzip.NewWriter(bf)
	chunk := make([]byte, 1<<20)
	for i := 0; i < 65; i++ {
		if _, e = gz.Write(chunk); e != nil {
			t.Fatal(e)
		}
	}
	if e = gz.Close(); e != nil {
		t.Fatal(e)
	}
	if e = bf.Close(); e != nil {
		t.Fatal(e)
	}
	bombOut := filepath.Join(root, "oversized-out")
	_, e = extractor.Runner.Run(context.Background(), operation.Request{Inputs: []string{bomb}, Options: map[string]any{"output": bombOut}})
	if e == nil {
		t.Fatal("expanded stream escaped limit")
	}
	if _, e = os.Stat(bombOut); !os.IsNotExist(e) {
		t.Fatal("oversized stream published output")
	}
	// Exercise zero-sized and >16 MiB streaming entries, beyond Run's capture limit.
	zero := write("zero.txt", nil)
	big := write("big.bin", []byte(strings.Repeat("Z", 17<<20)))
	bigZip := filepath.Join(root, "stream.zip")
	call("archive.create", []string{zero}, map[string]any{"files": []string{big}, "format": "zip", "output": bigZip})
	dest := filepath.Join(root, "stream-out")
	call("archive.extract", []string{bigZip}, map[string]any{"output": dest})
	if sha256.Sum256(contents(big)) != sha256.Sum256(contents(filepath.Join(dest, "big.bin"))) {
		t.Fatal("streamed file differs")
	}
}
