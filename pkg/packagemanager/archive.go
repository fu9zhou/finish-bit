package packagemanager

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ulikunitz/xz"
)

func archiveTarget(root, name string) (string, error) {
	name = strings.ReplaceAll(name, "\\", "/")
	name = strings.TrimSuffix(name, "/")
	if name == "" || strings.HasPrefix(name, "/") {
		return "", fmt.Errorf("invalid archive path %q", name)
	}
	for _, part := range strings.Split(name, "/") {
		if part == "" || part == "." || part == ".." || strings.ContainsAny(part, ":\x00") || strings.TrimRight(part, " .") != part {
			return "", fmt.Errorf("unsafe archive path %q", name)
		}
		base := strings.ToUpper(strings.SplitN(part, ".", 2)[0])
		if base == "CON" || base == "PRN" || base == "AUX" || base == "NUL" || len(base) == 4 && (strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT")) && base[3] >= '0' && base[3] <= '9' {
			return "", fmt.Errorf("reserved archive path %q", name)
		}
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	target := filepath.Join(absolute, filepath.FromSlash(name))
	relative, err := filepath.Rel(absolute, target)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("archive path escapes root")
	}
	return target, nil
}

type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (r contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.r.Read(p)
}

func extractArchive(ctx context.Context, source, root, format string, files ...[]string) error {
	if err := os.MkdirAll(root, 0755); err != nil {
		return err
	}
	var total int64
	entries := 0
	seen := map[string]bool{}
	selected := map[string]bool{}
	if len(files) > 0 {
		for _, name := range files[0] {
			if _, err := archiveTarget(root, name); err != nil {
				return err
			}
			selected[strings.ReplaceAll(name, "\\", "/")] = false
		}
	}
	extract := func(name string, mode os.FileMode, size int64, open func() (io.ReadCloser, error)) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		entries++
		if entries > 50000 {
			return fmt.Errorf("archive exceeds 50000 entries")
		}
		target, err := archiveTarget(root, name)
		if err != nil {
			return err
		}
		if mode.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		if !mode.IsRegular() {
			return fmt.Errorf("archive links and special files are not supported: %s", name)
		}
		if len(selected) > 0 {
			key := strings.ReplaceAll(name, "\\", "/")
			if _, ok := selected[key]; !ok {
				return nil
			}
			selected[key] = true
		}
		key := strings.ToLower(target)
		if seen[key] {
			return fmt.Errorf("duplicate archive entry %q", name)
		}
		seen[key] = true
		if size < 0 || size > maxPackageBytes-total {
			return fmt.Errorf("archive exceeds 1 GiB expanded size")
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		r, err := open()
		if err != nil {
			return err
		}
		defer r.Close()
		w, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		written, copyErr := io.Copy(w, io.LimitReader(contextReader{ctx, r}, size+1))
		closeErr := w.Close()
		total += written
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		if written != size {
			return fmt.Errorf("archive entry size mismatch: %s", name)
		}
		return nil
	}
	switch format {
	case "zip":
		archive, err := zip.OpenReader(source)
		if err != nil {
			return err
		}
		defer archive.Close()
		for _, f := range archive.File {
			if f.UncompressedSize64 > uint64(maxPackageBytes) {
				return fmt.Errorf("archive entry too large")
			}
			if err := extract(f.Name, f.Mode(), int64(f.UncompressedSize64), f.Open); err != nil {
				return err
			}
		}
	case "tar.gz", "tar.xz":
		file, err := os.Open(source)
		if err != nil {
			return err
		}
		defer file.Close()
		var reader io.Reader
		if format == "tar.gz" {
			gz, err := gzip.NewReader(file)
			if err != nil {
				return err
			}
			defer gz.Close()
			reader = gz
		} else {
			x, err := xz.NewReader(file)
			if err != nil {
				return err
			}
			reader = x
		}
		archive := tar.NewReader(reader)
		for {
			header, err := archive.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			if err := extract(header.Name, header.FileInfo().Mode(), header.Size, func() (io.ReadCloser, error) { return io.NopCloser(archive), nil }); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unsupported archive %q", format)
	}
	for name, found := range selected {
		if !found {
			return fmt.Errorf("archive missing required resource %q", name)
		}
	}
	return nil
}
