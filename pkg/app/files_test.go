package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBrowseFilesReturnsActualPathsWithoutCreatingOutputs(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "中文 directory")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(root, "hello world.txt")
	if err := os.WriteFile(file, []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	a := &App{}
	listing, err := a.BrowseFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	if listing.Path != root || listing.Parent != filepath.Dir(root) || len(listing.Entries) != 2 {
		t.Fatalf("unexpected listing: %+v", listing)
	}
	if !listing.Entries[0].Directory || listing.Entries[0].Path != dir || listing.Entries[1].Path != file {
		t.Fatalf("unexpected entries: %+v", listing.Entries)
	}
	empty, err := a.BrowseFiles(dir)
	if err != nil || len(empty.Entries) != 0 {
		t.Fatalf("empty directory: %+v %v", empty, err)
	}
	if _, err = a.BrowseFiles(file); err == nil {
		t.Fatal("file accepted as directory")
	}
	missing := filepath.Join(root, "not-created")
	if _, err = a.BrowseFiles(missing); err == nil {
		t.Fatal("missing directory accepted")
	}
	if _, err = os.Stat(missing); !os.IsNotExist(err) {
		t.Fatal("browse created a path")
	}
}
