package packagemanager

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fu9zhou/finish-bit/internal/toolrun"
)

// extractNSIS reads a pinned installer as an archive. It never executes the installer.
func (m *Manager) extractNSIS(ctx context.Context, extractor, source, root string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	listing, err := toolrun.Run(ctx, m, extractor, "7zip", "", "l", "-slt", "-ba", "-sccUTF-8", source)
	if err != nil {
		return err
	}
	expected, err := nsisEntries(string(listing), root)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(root, 0755); err != nil {
		return err
	}
	// NSIS may synthesize an uninstaller with unknown declared size; it is not
	// needed by a private, unpacked runtime and is explicitly excluded below.
	if _, err = toolrun.Run(ctx, m, extractor, "7zip", "", "x", "-y", "-bb0", "-bd", "-x!tesseract-uninstall.exe", "-o"+root, source); err != nil {
		return err
	}
	seen := map[string]bool{}
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("unexpected link or special NSIS entry")
		}
		relative, e := filepath.Rel(root, path)
		if e != nil {
			return e
		}
		key := strings.ReplaceAll(relative, "\\", "/")
		size, ok := expected[key]
		if !ok {
			return fmt.Errorf("unlisted extracted NSIS entry: %s", key)
		}
		info, e := d.Info()
		if e != nil {
			return e
		}
		if info.Size() != size {
			return fmt.Errorf("NSIS entry size mismatch: %s", key)
		}
		seen[key] = true
		return nil
	})
	if err != nil {
		return err
	}
	if len(seen) != len(expected) {
		return fmt.Errorf("NSIS extraction missing entries")
	}
	return nil
}
func nsisEntries(listing, root string) (map[string]int64, error) {
	listing = strings.ReplaceAll(listing, "\r\n", "\n")
	expected := map[string]int64{}
	folded := map[string]bool{}
	var total int64
	for _, block := range strings.Split(strings.TrimSpace(listing), "\n\n") {
		fields := map[string]string{}
		for _, line := range strings.Split(block, "\n") {
			pair := strings.SplitN(line, " = ", 2)
			if len(pair) == 2 {
				fields[pair[0]] = pair[1]
			}
		}
		name := fields["Path"]
		if name == "" {
			return nil, fmt.Errorf("invalid NSIS listing")
		}
		if _, e := archiveTarget(root, name); e != nil {
			return nil, e
		}
		if fields["Symbolic Link"] != "" || fields["Hard Link"] != "" || strings.Contains(fields["Attributes"], "L") {
			return nil, fmt.Errorf("NSIS links not supported")
		}
		if fields["Folder"] == "+" || strings.HasPrefix(fields["Attributes"], "D") {
			continue
		}
		if name == "tesseract-uninstall.exe" && fields["Size"] == "" {
			continue
		}
		size, e := strconv.ParseInt(fields["Size"], 10, 64)
		if e != nil || size < 0 || size > maxPackageBytes-total {
			return nil, fmt.Errorf("invalid NSIS size or expanded payload exceeds 1 GiB")
		}
		key := strings.ReplaceAll(name, "\\", "/")
		if folded[strings.ToLower(key)] {
			return nil, fmt.Errorf("duplicate NSIS entry")
		}
		folded[strings.ToLower(key)] = true
		expected[key] = size
		total += size
		if len(expected) > 50000 {
			return nil, fmt.Errorf("too many NSIS entries")
		}
	}
	if len(expected) == 0 {
		return nil, fmt.Errorf("empty NSIS archive")
	}
	return expected, nil
}
func checkPayloadSize(root string) error {
	var total int64
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("payload contains link or special file")
		}
		info, e := d.Info()
		if e != nil {
			return e
		}
		if info.Size() > maxPackageBytes-total {
			return fmt.Errorf("combined runtime and supplementary resources exceed 1 GiB")
		}
		total += info.Size()
		return nil
	})
}
