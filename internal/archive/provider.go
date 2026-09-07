// Package archive wraps managed 7-Zip with staged, bounded local archive workflows.
package archive

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/fu9zhou/finish-bit/internal/toolrun"
	"github.com/fu9zhou/finish-bit/pkg/operation"
)

var writeFormats = strings.Fields("7z zip tar gzip bzip2 xz tar.gz tar.xz")
var readFormats = append(append([]string{}, writeFormats...), "cab", "lzma", "zstd")

type spec struct {
	id, summary, alias string
	input, output      bool
	options            []operation.Parameter
}

func str(n, d, v string) operation.Parameter { return toolrun.Option(n, operation.TypeString, d, v) }
func num(n, d string, v int) operation.Parameter {
	return toolrun.Option(n, operation.TypeInteger, d, v)
}
func names(n, d string, required bool) operation.Parameter {
	return operation.Parameter{Name: n, Description: d, Type: operation.TypeStrings, Required: required}
}
func creation() []operation.Parameter {
	return []operation.Parameter{names("files", "Additional local files or directories", false), str("format", "7z, zip, tar, gzip, bzip2, xz, tar.gz or tar.xz", "7z"), num("level", "Compression level 0 to 9", 5)}
}
func catalog() []spec {
	return []spec{
		{"archive.formats", "List archive read/write and encryption formats", "压缩格式清单", false, false, nil},
		{"archive.list", "List validated archive entries, sizes and encryption", "查看压缩包目录", true, false, nil},
		{"archive.test", "Verify archive integrity and entry boundaries", "校验压缩包完整性", true, false, nil},
		{"archive.create", "Create an archive from local files and directories", "创建压缩包", true, true, creation()},
		{"archive.extract", "Extract all or exact selected entries into a new directory", "解压文件", true, false, []operation.Parameter{names("entries", "Exact file or directory paths; empty means all", false), toolrun.Param("output", "New output directory", true)}},
		{"archive.update", "Add or replace files in a copy of a ZIP or 7z archive", "更新压缩包文件", true, true, []operation.Parameter{names("files", "Local files/directories; names match archive roots", true), num("level", "Compression level 0 to 9", 5)}},
		{"archive.remove", "Remove exact entries from a copy of a ZIP or 7z archive", "删除压缩包成员", true, true, []operation.Parameter{names("entries", "Exact file/directory paths to remove", true)}},
		{"archive.rename", "Rename one archive entry in a copy of a ZIP or 7z archive", "重命名压缩包成员", true, true, []operation.Parameter{toolrun.Param("entry", "Exact existing file path", true), toolrun.Param("name", "New safe relative file path", true)}},
		{"archive.repack", "Recompress archive content into another format", "转换压缩包格式", true, true, []operation.Parameter{str("to", "7z, zip, tar, tar.gz or tar.xz", "zip"), num("level", "Compression level 0 to 9", 5), str("new-password-env", "Output password variable; empty reuses input password", ""), toolrun.Option("drop-password", operation.TypeBoolean, "Explicitly remove encryption from output", false)}},
		{"archive.volumes", "Create numbered ZIP or 7z volumes in a new directory", "分卷压缩", true, false, append(creation(), num("volume-mib", "Maximum volume size in MiB", 10), toolrun.Param("output", "New directory for archive.7z.001 or archive.zip.001 volumes", true))},
	}
}
func Register(registry *operation.Registry, resolver toolrun.Resolver) error {
	for _, s := range catalog() {
		def := operation.Definition{ID: s.id, Summary: s.summary, Description: s.summary + ". Managed 7-Zip Extra 26.03. Bounded local files; validated paths and staged publication. See docs/archives.md.", Aliases: []string{s.alias}, Tags: []string{"archive", "compression", "files"}, Source: "7zip", Requirements: []operation.Requirement{{Package: "7zip"}}, Options: append([]operation.Parameter{}, s.options...)}
		if s.input {
			def.Inputs = []operation.Parameter{toolrun.Param("input", "Local archive, or file/directory for creation", true)}
			def.Options = append(def.Options, str("password-env", "Environment variable containing a ZIP/7z password; empty means no password", ""))
			if s.id != "archive.create" && s.id != "archive.volumes" {
				def.Options = append(def.Options, str("format", "Input format; auto infers from extension", "auto"))
			}
		}
		if s.output {
			def.Options = append(def.Options, toolrun.OutputOptions()...)
		}
		if e := registry.Register(operation.Capability{Definition: def, Runner: operation.Func(func(ctx context.Context, r operation.Request) (operation.Result, error) {
			if e := operation.ValidateRequest(def, r); e != nil {
				return operation.Result{}, e
			}
			return run(ctx, resolver, s, r)
		})}); e != nil {
			return e
		}
	}
	return nil
}
func safeName(name string) (string, error) {
	name = strings.ReplaceAll(name, "\\", "/")
	name = strings.TrimSuffix(name, "/")
	if name == "" || strings.HasPrefix(name, "/") {
		return "", toolrun.Invalid("invalid archive path")
	}
	for _, part := range strings.Split(name, "/") {
		if part == "" || part == "." || part == ".." || strings.TrimRight(part, " .") != part || strings.ContainsAny(part, ":*?\"<>|@") || strings.IndexFunc(part, unicode.IsControl) >= 0 {
			return "", toolrun.Invalid("unsafe archive path: " + name)
		}
		base := strings.ToUpper(strings.SplitN(part, ".", 2)[0])
		if slices.Contains([]string{"CON", "PRN", "AUX", "NUL"}, base) || len(base) == 4 && (strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT")) && base[3] >= '0' && base[3] <= '9' {
			return "", toolrun.Invalid("reserved archive path: " + name)
		}
	}
	return name, nil
}

type entry struct {
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	Directory bool   `json:"directory"`
	Encrypted bool   `json:"encrypted"`
}

func parseListing(data []byte) ([]entry, error) {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	entries := []entry{}
	seen := map[string]bool{}
	var total int64
	for _, block := range strings.Split(strings.Trim(text, "\n"), "\n\n") {
		if strings.TrimSpace(block) == "" {
			continue
		}
		values := map[string]string{}
		for _, line := range strings.Split(block, "\n") {
			key, value, ok := strings.Cut(line, " = ")
			if !ok {
				return nil, toolrun.Invalid("ambiguous archive listing")
			}
			if _, exists := values[key]; exists {
				return nil, toolrun.Invalid("duplicate archive listing field")
			}
			values[key] = value
		}
		path, e := safeName(values["Path"])
		if e != nil {
			return nil, e
		}
		if seen[strings.ToLower(path)] {
			return nil, toolrun.Invalid("case-colliding or duplicate archive entries")
		}
		seen[strings.ToLower(path)] = true
		size, e := strconv.ParseInt(values["Size"], 10, 64)
		if e != nil || size < 0 || size > 64<<20 {
			return nil, toolrun.Invalid("entry exceeds 64 MiB or has unknown size")
		}
		total += size
		if total > 256<<20 || len(entries) >= 10000 {
			return nil, toolrun.Invalid("archive exceeds 256 MiB or 10000 entries")
		}
		attr := values["Attributes"]
		for _, field := range strings.Fields(attr) {
			if len(field) >= 10 && strings.ContainsAny(field[:1], "lbcp s") {
				return nil, toolrun.Invalid("archive links and special files are unsupported")
			}
		}
		for _, key := range []string{"Symbolic Link", "Hard Link", "Reparse Point", "Alternate Stream"} {
			if value := values[key]; value != "" && value != "-" {
				return nil, toolrun.Invalid("archive links and streams are unsupported")
			}
		}
		entries = append(entries, entry{path, size, values["Folder"] == "+" || strings.HasPrefix(attr, "D"), values["Encrypted"] == "+"})
	}
	// Reject file/directory conflicts independently of listing order.
	files := map[string]bool{}
	for _, item := range entries {
		if !item.Directory {
			files[strings.ToLower(item.Path)] = true
		}
	}
	for _, item := range entries {
		parts := strings.Split(item.Path, "/")
		for i := 1; i < len(parts); i++ {
			if files[strings.ToLower(strings.Join(parts[:i], "/"))] {
				return nil, toolrun.Invalid("archive file/directory collision")
			}
		}
	}
	return entries, nil
}
func password(v *toolrun.Values, key, fallback string) string {
	name := v.String(key, "")
	if name == "" {
		return fallback
	}
	value, ok := os.LookupEnv(name)
	v.Check(ok && value != "" && len(value) <= 1024 && !strings.ContainsAny(value, "\x00\r\n"), "password environment variable must contain 1 to 1024 characters without line breaks")
	return value
}
func baseArgs(command, pass string) []string {
	if pass == "" {
		pass = "-"
	}
	return []string{command, "-bd", "-bb0", "-y", "-sccUTF-8", "-scsUTF-8", "-spd", "-p" + pass}
}
func invoke(ctx context.Context, resolver toolrun.Resolver, dir string, args ...string) ([]byte, error) {
	return toolrun.Run(ctx, resolver, "7zip", "7zip", dir, args...)
}
func list(ctx context.Context, resolver toolrun.Resolver, dir, path, kind, pass string) ([]entry, error) {
	data, e := invoke(ctx, resolver, dir, append(baseArgs("l", pass), "-slt", "-ba", "-t"+kind, "--", path)...)
	if e != nil {
		return nil, e
	}
	entries, e := parseListing(data)
	if e != nil {
		return nil, e
	}
	for _, entry := range entries {
		if entry.Encrypted && pass == "" {
			return nil, toolrun.Invalid("encrypted archive requires password-env")
		}
	}
	return entries, nil
}
func infer(path string) string {
	s := strings.ToLower(path)
	s = strings.TrimSuffix(s, ".001")
	for _, compound := range []string{"tar.gz", "tar.xz"} {
		if strings.HasSuffix(s, "."+compound) {
			return compound
		}
	}
	ext := strings.TrimPrefix(filepath.Ext(s), ".")
	if x, ok := map[string]string{"gz": "gzip", "tgz": "tar.gz", "bz2": "bzip2", "txz": "tar.xz", "zst": "zstd"}[ext]; ok {
		return x
	}
	return ext
}
func copyInput(path, target string, max int64) error {
	info, e := os.Lstat(path)
	if e != nil || !info.Mode().IsRegular() {
		return toolrun.Invalid("input must be a regular file without links")
	}
	if info.Size() > max {
		return toolrun.Invalid("input file exceeds size limit")
	}
	r, e := os.Open(path)
	if e != nil {
		return e
	}
	defer r.Close()
	w, e := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	n, copyErr := io.Copy(w, io.LimitReader(r, max+1))
	closeErr := w.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	if n > max {
		return toolrun.Invalid("input grew beyond size limit")
	}
	return nil
}
func snapshot(path, dir string) (string, error) {
	absolute, e := filepath.Abs(path)
	if e != nil {
		return "", e
	}
	target := filepath.Join(dir, "source"+filepath.Ext(absolute))
	if strings.HasSuffix(absolute, ".001") {
		prefix := strings.TrimSuffix(absolute, ".001")
		target = filepath.Join(dir, "joined"+filepath.Ext(prefix))
		joined, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return "", err
		}
		defer joined.Close()
		var total int64
		for i := 1; i <= 1000; i++ {
			part := fmt.Sprintf("%s.%03d", prefix, i)
			info, err := os.Lstat(part)
			if os.IsNotExist(err) && i > 1 {
				break
			}
			if err != nil {
				return "", err
			}
			if !info.Mode().IsRegular() {
				return "", toolrun.Invalid("volume must be a regular file")
			}
			if info.Size() > 256<<20-total {
				return "", toolrun.Invalid("volumes exceed 256 MiB")
			}
			reader, err := os.Open(part)
			if err != nil {
				return "", err
			}
			n, copyErr := io.Copy(joined, io.LimitReader(reader, (256<<20)-total+1))
			closeErr := reader.Close()
			total += n
			if copyErr != nil {
				return "", copyErr
			}
			if closeErr != nil {
				return "", closeErr
			}
			if total > 256<<20 {
				return "", toolrun.Invalid("volumes exceed 256 MiB")
			}
			if i == 1000 {
				return "", toolrun.Invalid("too many volumes")
			}
		}
		if err = joined.Close(); err != nil {
			return "", err
		}
	} else if e = copyInput(absolute, target, 256<<20); e != nil {
		return "", e
	}
	return target, nil
}
func collect(paths []string, dir string) ([]string, error) {
	if len(paths) == 0 || len(paths) > 64 {
		return nil, toolrun.Invalid("provide 1 to 64 file/directory inputs")
	}
	names := []string{}
	seen := map[string]bool{}
	allNames := map[string]bool{}
	count := 0
	var total int64
	for _, input := range paths {
		absolute, e := filepath.Abs(input)
		if e != nil {
			return nil, e
		}
		base, e := safeName(filepath.Base(absolute))
		if e != nil {
			return nil, e
		}
		if seen[strings.ToLower(base)] {
			return nil, toolrun.Invalid("duplicate input base names")
		}
		seen[strings.ToLower(base)] = true
		names = append(names, base)
		e = filepath.WalkDir(absolute, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			rel, e := filepath.Rel(absolute, path)
			if e != nil {
				return e
			}
			name := base
			if rel != "." {
				name += "/" + filepath.ToSlash(rel)
			}
			name, e = safeName(name)
			if e != nil {
				return e
			}
			if allNames[strings.ToLower(name)] {
				return toolrun.Invalid("case-colliding input paths")
			}
			allNames[strings.ToLower(name)] = true
			count++
			if count > 10000 {
				return toolrun.Invalid("input exceeds 10000 entries")
			}
			target := filepath.Join(dir, filepath.FromSlash(name))
			if d.IsDir() {
				return os.MkdirAll(target, 0755)
			}
			if !d.Type().IsRegular() {
				return toolrun.Invalid("input links and special files are unsupported")
			}
			info, e := d.Info()
			if e != nil {
				return e
			}
			total += info.Size()
			if info.Size() > 64<<20 || total > 256<<20 {
				return toolrun.Invalid("input exceeds 64 MiB/file or 256 MiB total")
			}
			if e = os.MkdirAll(filepath.Dir(target), 0755); e != nil {
				return e
			}
			return copyInput(path, target, 64<<20)
		})
		if e != nil {
			return nil, e
		}
	}
	return names, nil
}
func selected(entries []entry, wanted []string) ([]entry, error) {
	if len(wanted) == 0 {
		return entries, nil
	}
	out := []entry{}
	matches := map[string]bool{}
	for _, name := range wanted {
		safe, e := safeName(name)
		if e != nil {
			return nil, e
		}
		exists := false
		for _, item := range entries {
			if item.Path == safe {
				exists = true
				break
			}
		}
		if !exists {
			return nil, toolrun.Invalid("selected entry does not exist: " + safe)
		}
		matches[safe] = true
	}
	for _, item := range entries {
		for name := range matches {
			if item.Path == name || strings.HasPrefix(item.Path, name+"/") {
				out = append(out, item)
				break
			}
		}
	}
	return out, nil
}
func extract(ctx context.Context, resolver toolrun.Resolver, dir, path, kind, pass, target string, entries []entry, zone []byte) error {
	for _, item := range entries {
		if e := ctx.Err(); e != nil {
			return e
		}
		dest := filepath.Join(target, filepath.FromSlash(item.Path))
		if item.Directory {
			if e := os.MkdirAll(dest, 0755); e != nil {
				return e
			}
			continue
		}
		if e := os.MkdirAll(filepath.Dir(dest), 0755); e != nil {
			return e
		}
		args := append(baseArgs("e", pass), "-so", "-t"+kind, "-i!"+item.Path, "--", path)
		var extractionError error
		if kind == "stream" {
			extractionError = toolrun.CopyFile(path, dest)
		} else {
			extractionError = toolrun.RunToFile(ctx, resolver, "7zip", "7zip", dir, dest, item.Size, args...)
		}
		if e := extractionError; e != nil {
			return e
		}
		info, e := os.Stat(dest)
		if e != nil || info.Size() != item.Size {
			return toolrun.Invalid("extracted size differs from archive directory")
		}
		if len(zone) > 0 {
			if e = os.WriteFile(dest+":Zone.Identifier", zone, 0600); e != nil {
				return e
			}
		}
	}
	return nil
}
func markOrigin(path string, zone []byte) error {
	if len(zone) == 0 {
		return nil
	}
	return os.WriteFile(path+":Zone.Identifier", zone, 0600)
}
func create(ctx context.Context, resolver toolrun.Resolver, dir, target, kind, pass string, level int, files []string, volume int) error {
	args := baseArgs("a", pass)
	if pass == "" {
		args = baseArgs("a", "")
		args = args[:len(args)-1]
	}
	if strings.HasPrefix(kind, "tar.") {
		tar := filepath.Join(filepath.Dir(dir), "intermediate.tar")
		if e := create(ctx, resolver, dir, tar, "tar", "", 0, files, 0); e != nil {
			return e
		}
		kind = map[string]string{"tar.gz": "gzip", "tar.xz": "xz"}[kind]
		files = []string{tar}
	}
	args = append(args, "-t"+kind, "-mx="+strconv.Itoa(level))
	if pass != "" {
		if kind == "zip" {
			args = append(args, "-mem=AES256")
		} else if kind == "7z" {
			args = append(args, "-mhe=on")
		} else {
			return toolrun.Invalid("encryption requires zip or 7z format")
		}
	}
	if volume > 0 {
		args = append(args, "-v"+strconv.Itoa(volume)+"m")
	}
	// Explicit relative paths cannot be interpreted as switches or @listfiles.
	for i, name := range files {
		if !filepath.IsAbs(name) {
			files[i] = "." + string(filepath.Separator) + filepath.FromSlash(name)
		}
	}
	args = append(args, "--", target)
	args = append(args, files...)
	_, e := invoke(ctx, resolver, dir, args...)
	return e
}
func run(ctx context.Context, resolver toolrun.Resolver, s spec, r operation.Request) (operation.Result, error) {
	result := operation.Result{}
	v := toolrun.Values{Request: r}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	if s.id == "archive.formats" {
		return operation.Result{Data: map[string]any{"read": readFormats, "write": writeFormats, "encrypt": []string{"7z", "zip"}, "mutate": []string{"7z", "zip"}, "volumes": []string{"7z", "zip"}}}, nil
	}
	pass := password(&v, "password-env", "")
	out := v.String("output", "")
	overwrite := v.Bool("overwrite", false)
	if v.Err != nil {
		return result, v.Err
	}
	dir, e := os.MkdirTemp("", "finishbit-archive-")
	if e != nil {
		return result, e
	}
	defer os.RemoveAll(dir)
	work := filepath.Join(dir, "files")
	if e = os.Mkdir(work, 0755); e != nil {
		return result, e
	}
	if s.id == "archive.create" || s.id == "archive.volumes" {
		kind := v.Enum("format", "7z", writeFormats...)
		level := v.Int("level", 5, 0, 9)
		volume := 0
		if s.id == "archive.volumes" {
			volume = v.Int("volume-mib", 10, 1, 256)
			v.Check(kind == "7z" || kind == "zip", "volumes require zip or 7z")
		}
		if v.Err != nil {
			return result, v.Err
		}
		files, err := collect(append(append([]string{}, r.Inputs...), v.Strings("files")...), work)
		if err != nil {
			return result, err
		}
		if slices.Contains([]string{"gzip", "bzip2", "xz"}, kind) {
			info, err := os.Stat(filepath.Join(work, files[0]))
			if err != nil || len(files) != 1 || !info.Mode().IsRegular() {
				return result, toolrun.Invalid("stream compression requires exactly one regular file; use tar.gz or tar.xz for directories")
			}
		}
		if volume > 0 {
			result.Outputs, e = toolrun.Directory(ctx, out, func(target string) error {
				return create(ctx, resolver, work, filepath.Join(target, "archive."+kind), kind, pass, level, files, volume)
			})
			return result, e
		}
		e = toolrun.Artifact(ctx, out, overwrite, func(target string) error { return create(ctx, resolver, work, target, kind, pass, level, files, 0) })
		if e == nil {
			result.Outputs = []string{out}
		}
		return result, e
	}
	kind := v.String("format", "auto")
	if kind == "auto" {
		kind = infer(r.Inputs[0])
	}
	v.Check(slices.Contains(readFormats, kind), "unsupported archive format")
	if v.Err != nil {
		return result, v.Err
	}
	path, e := snapshot(r.Inputs[0], dir)
	if e != nil {
		return result, e
	}
	zone := []byte(nil)
	if runtime.GOOS == "windows" {
		stream, err := os.Open(r.Inputs[0] + ":Zone.Identifier")
		if err == nil {
			data, readErr := io.ReadAll(io.LimitReader(stream, 4097))
			closeErr := stream.Close()
			if readErr != nil {
				return result, readErr
			}
			if closeErr != nil {
				return result, closeErr
			}
			if len(data) > 4096 {
				return result, toolrun.Invalid("archive origin metadata exceeds 4 KiB")
			}
			zone = data
		} else if !os.IsNotExist(err) {
			return result, err
		}
	}
	originalKind := kind
	if strings.HasPrefix(kind, "tar.") {
		outer := map[string]string{"tar.gz": "gzip", "tar.xz": "xz"}[kind]
		tar := filepath.Join(dir, "content.tar")
		args := append(baseArgs("e", pass), "-so", "-t"+outer, "--", path)
		if e = toolrun.RunToFile(ctx, resolver, "7zip", "7zip", dir, tar, 256<<20, args...); e != nil {
			return result, e
		}
		path = tar
		kind = "tar"
	}
	var entries []entry
	if slices.Contains([]string{"gzip", "bzip2", "xz", "lzma", "zstd"}, kind) {
		payload := filepath.Join(dir, "stream.bin")
		args := append(baseArgs("e", pass), "-so", "-t"+kind, "--", path)
		if e = toolrun.RunToFile(ctx, resolver, "7zip", "7zip", dir, payload, 64<<20, args...); e != nil {
			return result, e
		}
		info, err := os.Stat(payload)
		if err != nil {
			return result, err
		}
		name, err := safeName(strings.TrimSuffix(filepath.Base(r.Inputs[0]), filepath.Ext(r.Inputs[0])))
		if err != nil {
			return result, err
		}
		entries = []entry{{Path: name, Size: info.Size()}}
		path = payload
		kind = "stream"
	} else {
		entries, e = list(ctx, resolver, dir, path, kind, pass)
	}
	if e != nil {
		return result, e
	}
	switch s.id {
	case "archive.list":
		return operation.Result{Data: map[string]any{"entries": entries, "count": len(entries), "format": originalKind}}, nil
	case "archive.test":
		if kind != "stream" {
			_, e = invoke(ctx, resolver, dir, append(baseArgs("t", pass), "-t"+kind, "--", path)...)
		}
		if e == nil {
			result.Data = map[string]any{"valid": true, "entries": len(entries)}
		}
		return result, e
	case "archive.extract":
		chosen, err := selected(entries, v.Strings("entries"))
		if err != nil {
			return result, err
		}
		result.Outputs, e = toolrun.DirectoryTree(ctx, out, func(target string) error { return extract(ctx, resolver, dir, path, kind, pass, target, chosen, zone) })
		return result, e
	case "archive.repack":
		to := v.Enum("to", "zip", "7z", "zip", "tar", "tar.gz", "tar.xz")
		level := v.Int("level", 5, 0, 9)
		outputPass := password(&v, "new-password-env", pass)
		if v.Bool("drop-password", false) {
			v.Check(v.String("new-password-env", "") == "", "drop-password cannot be combined with new-password-env")
			outputPass = ""
		}
		if v.Err != nil {
			return result, v.Err
		}
		if e = extract(ctx, resolver, dir, path, kind, pass, work, entries, nil); e != nil {
			return result, e
		}
		children, e := os.ReadDir(work)
		if e != nil {
			return result, e
		}
		files := []string{}
		for _, child := range children {
			files = append(files, child.Name())
		}
		if len(files) == 0 {
			return result, toolrun.Invalid("cannot repack an empty archive")
		}
		e = toolrun.Artifact(ctx, out, overwrite, func(target string) error {
			if err := create(ctx, resolver, work, target, to, outputPass, level, files, 0); err != nil {
				return err
			}
			return markOrigin(target, zone)
		})
		if e == nil {
			result.Outputs = []string{out}
		}
		return result, e
	default:
		if kind != "7z" && kind != "zip" {
			return result, toolrun.Invalid("archive mutation requires zip or 7z")
		}
		if strings.HasSuffix(r.Inputs[0], ".001") {
			return result, toolrun.Invalid("repack volumes into one archive before mutation")
		}
		command := ""
		extra := []string{}
		switch s.id {
		case "archive.update":
			command = "a"
			level := v.Int("level", 5, 0, 9)
			if v.Err != nil {
				return result, v.Err
			}
			files, err := collect(v.Strings("files"), work)
			if err != nil {
				return result, err
			}
			extra = append(extra, "-mx="+strconv.Itoa(level))
			for _, name := range files {
				extra = append(extra, "-i!"+name)
			}
		case "archive.remove":
			command = "d"
			wanted := v.Strings("entries")
			if len(wanted) == 0 {
				return result, toolrun.Invalid("entries must not be empty")
			}
			chosen, err := selected(entries, wanted)
			if err != nil {
				return result, err
			}
			for _, item := range chosen {
				extra = append(extra, "-i!"+item.Path)
			}
		case "archive.rename":
			command = "rn"
			old, err := safeName(v.String("entry", ""))
			if err != nil {
				return result, err
			}
			name, err := safeName(v.String("name", ""))
			if err != nil {
				return result, err
			}
			found := false
			for _, item := range entries {
				if item.Path == old && !item.Directory {
					found = true
				}
				if strings.EqualFold(item.Path, name) || strings.HasPrefix(strings.ToLower(item.Path), strings.ToLower(name)+"/") || !item.Directory && strings.HasPrefix(strings.ToLower(name), strings.ToLower(item.Path)+"/") {
					return result, toolrun.Invalid("rename destination conflicts with an entry")
				}
			}
			if !found {
				return result, toolrun.Invalid("rename requires an existing file entry")
			}
			extra = []string{old, name}
		}
		e = toolrun.Artifact(ctx, out, overwrite, func(target string) error {
			if err := toolrun.CopyFile(path, target); err != nil {
				return err
			}
			args := baseArgs(command, pass)
			if pass == "" {
				args = args[:len(args)-1]
			}
			args = append(args, "-t"+kind)
			if command == "a" && pass != "" {
				if kind == "zip" {
					args = append(args, "-mem=AES256")
				} else {
					args = append(args, "-mhe=on")
				}
			}
			if command != "rn" {
				args = append(args, extra...)
			}
			args = append(args, "--", target)
			if command == "rn" {
				args = append(args, extra...)
			}
			if _, err := invoke(ctx, resolver, work, args...); err != nil {
				return err
			}
			_, err := list(ctx, resolver, dir, target, kind, pass)
			if err != nil {
				return err
			}
			return markOrigin(target, zone)
		})
		if e == nil {
			result.Outputs = []string{out}
		}
		return result, e
	}
}
