package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
)

type FileEntry struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Directory bool   `json:"directory"`
}

type DirectoryListing struct {
	Path      string      `json:"path"`
	Parent    string      `json:"parent"`
	Roots     []FileEntry `json:"roots"`
	Entries   []FileEntry `json:"entries"`
	Truncated bool        `json:"truncated"`
}

// BrowseFiles lists local names for a picker; it never uploads, opens or creates files.
// Paths are absolute so selecting a file does not depend on the adapter's cwd.
func (a *App) BrowseFiles(path string) (DirectoryListing, error) {
	if path == "" {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return DirectoryListing{}, err
		}
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return DirectoryListing{}, err
	}
	f, err := os.Open(absolute)
	if err != nil {
		return DirectoryListing{}, fmt.Errorf("open directory: %w", err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return DirectoryListing{}, err
	}
	if !info.IsDir() {
		return DirectoryListing{}, fmt.Errorf("path is not a directory")
	}
	items, err := f.ReadDir(5001)
	if err != nil && err != io.EOF {
		return DirectoryListing{}, err
	}
	listing := DirectoryListing{Path: absolute, Parent: filepath.Dir(absolute), Roots: []FileEntry{}, Entries: []FileEntry{}, Truncated: len(items) > 5000}
	if len(items) > 5000 {
		items = items[:5000]
	}
	for _, item := range items {
		isDir := item.IsDir()
		entryPath := filepath.Join(absolute, item.Name())
		if item.Type()&os.ModeSymlink != 0 {
			target, e := os.Stat(entryPath)
			if e != nil {
				continue
			}
			isDir = target.IsDir()
			if !isDir && !target.Mode().IsRegular() {
				continue
			}
		} else if !isDir && !item.Type().IsRegular() {
			continue
		}
		listing.Entries = append(listing.Entries, FileEntry{Name: item.Name(), Path: entryPath, Directory: isDir})
	}
	sort.Slice(listing.Entries, func(i, j int) bool {
		x, y := listing.Entries[i], listing.Entries[j]
		if x.Directory != y.Directory {
			return x.Directory
		}
		return x.Name < y.Name
	})
	if home, e := os.UserHomeDir(); e == nil {
		listing.Roots = append(listing.Roots, FileEntry{Name: "Home", Path: home, Directory: true})
	}
	if runtime.GOOS == "windows" {
		for drive := 'A'; drive <= 'Z'; drive++ {
			root := fmt.Sprintf("%c:\\", drive)
			if info, e := os.Stat(root); e == nil && info.IsDir() {
				listing.Roots = append(listing.Roots, FileEntry{Name: root, Path: root, Directory: true})
			}
		}
	} else {
		listing.Roots = append(listing.Roots, FileEntry{Name: "/", Path: "/", Directory: true})
	}
	return listing, nil
}
