package packagemanager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fu9zhou/finish-bit/internal/toolrun"
)

// Prefer an existing managed reader. The raw upstream bootstrap requires no
// extractor of its own, so cold installation never depends on a 7z package.
func (m *Manager) sevenzipExtractor(ctx context.Context, installing string) (string, error) {
	if installing == "7zip-bootstrap" {
		return "", fmt.Errorf("7zip-bootstrap must not require archive extraction")
	}
	for _, name := range []string{"7zip-full", "7zip", "7zip-bootstrap"} {
		if name == installing {
			continue
		}
		pkg, ok := m.registry.Find(name)
		if !ok {
			continue
		}
		installed, err := m.Info(name)
		if err != nil || installed.Version != pkg.Version || installed.Platform != PlatformKey() {
			continue
		}
		if _, err := m.Executable(name, "7zip"); err == nil {
			return name, nil
		}
	}
	if _, err := m.Install(ctx, "7zip-bootstrap"); err != nil {
		return "", fmt.Errorf("install 7-Zip bootstrap: %w", err)
	}
	return "7zip-bootstrap", nil
}

type sevenzipEntry struct {
	name string
	size int64
	dir  bool
}

// Only 7-Zip reads the archive; FinishBit owns all destination paths and writes.
// A separate exact-name stdout stream per file preserves the size limit during
// extraction, including solid archives. No untrusted archive path reaches a
// tool-controlled filesystem write.
func (m *Manager) extract7z(ctx context.Context, installing, source, root string, files []string) error {
	extractor, err := m.sevenzipExtractor(ctx, installing)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	source, err = filepath.Abs(source)
	if err != nil {
		return err
	}
	listing, err := toolrun.Run(ctx, m, extractor, "7zip", "", "l", "-t7z", "-slt", "-ba", "-sccUTF-8", "-pfinishbit", "--", source)
	if err != nil {
		return err
	}
	entries, err := sevenzipEntries(string(listing), root)
	if err != nil {
		return err
	}
	selected := map[string]bool{}
	for _, name := range files {
		if _, err := archiveTarget(root, name); err != nil {
			return err
		}
		selected[strings.ReplaceAll(name, "\\", "/")] = false
	}
	// Check required resources before starting any output process.
	for _, entry := range entries {
		if !entry.dir {
			if _, ok := selected[entry.name]; ok {
				selected[entry.name] = true
			}
		}
	}
	for name, found := range selected {
		if !found {
			return fmt.Errorf("archive missing required resource %q", name)
		}
	}
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		if len(selected) > 0 && !selected[entry.name] {
			continue
		}
		target, err := archiveTarget(root, entry.name)
		if err != nil {
			return err
		}
		if entry.dir {
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		// -spd and -ssc make the include literal and case-sensitive; -r- avoids
		// recursive matches. -i! prevents leading @/- in a name becoming syntax.
		if err := toolrun.RunToFile(ctx, m, extractor, "7zip", "", target, entry.size,
			"x", "-t7z", "-so", "-spd", "-ssc", "-r-", "-y", "-pfinishbit", "-bb0", "-bd", "-i!"+entry.name, "--", source); err != nil {
			return err
		}
		info, err := os.Stat(target)
		if err != nil {
			return err
		}
		if info.Size() != entry.size {
			return fmt.Errorf("archive entry size mismatch: %s", entry.name)
		}
	}
	return os.MkdirAll(root, 0755)
}

func sevenzipEntries(listing, root string) ([]sevenzipEntry, error) {
	listing = strings.ReplaceAll(listing, "\r\n", "\n")
	entries := []sevenzipEntry{}
	seen := map[string]bool{}
	var total int64
	for _, block := range strings.Split(strings.Trim(listing, "\n"), "\n\n") {
		if block == "" {
			continue
		}
		if len(entries) >= 50000 {
			return nil, fmt.Errorf("archive exceeds 50000 entries")
		}
		fields := map[string]string{}
		for _, line := range strings.Split(block, "\n") {
			pair := strings.SplitN(line, " = ", 2)
			if len(pair) != 2 || pair[0] == "" || strings.ContainsAny(line, "\r\x00") {
				return nil, fmt.Errorf("invalid 7-Zip listing")
			}
			if _, ok := fields[pair[0]]; ok {
				return nil, fmt.Errorf("duplicate 7-Zip listing field")
			}
			fields[pair[0]] = pair[1]
		}
		name := strings.ReplaceAll(fields["Path"], "\\", "/")
		if _, err := archiveTarget(root, name); err != nil {
			return nil, err
		}
		attributes := fields["Attributes"]
		if fields["Symbolic Link"] != "" || fields["Hard Link"] != "" || strings.ContainsAny(attributes, "Ll") {
			return nil, fmt.Errorf("archive links are not supported: %s", name)
		}
		for _, mode := range strings.Fields(attributes) {
			if len(mode) == 10 && mode[0] != '-' && mode[0] != 'd' {
				return nil, fmt.Errorf("archive special files are not supported: %s", name)
			}
		}
		if fields["Encrypted"] != "-" {
			return nil, fmt.Errorf("encrypted or unrecognized 7-Zip entry: %s", name)
		}
		size, err := strconv.ParseInt(fields["Size"], 10, 64)
		if err != nil || size < 0 || size > maxPackageBytes-total {
			return nil, fmt.Errorf("invalid archive size or expanded payload exceeds 1 GiB")
		}
		dir := fields["Folder"] == "+" || strings.Contains(attributes, "D") || strings.Contains(attributes, " dr")
		if dir && size != 0 {
			return nil, fmt.Errorf("nonempty archive directory")
		}
		name = strings.TrimSuffix(name, "/")
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			return nil, fmt.Errorf("duplicate archive entry %q", name)
		}
		seen[key] = dir
		total += size
		entries = append(entries, sevenzipEntry{name, size, dir})
	}
	for _, entry := range entries {
		parts := strings.Split(strings.ToLower(entry.name), "/")
		for i := 1; i < len(parts); i++ {
			if dir, exists := seen[strings.Join(parts[:i], "/")]; exists && !dir {
				return nil, fmt.Errorf("archive file conflicts with parent directory: %s", entry.name)
			}
		}
	}
	return entries, nil
}
