// Package document provides local, sandboxed document workflows using managed Pandoc.
package document

import (
	"archive/zip"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/fu9zhou/finish-bit/internal/toolrun"
	"github.com/fu9zhou/finish-bit/pkg/operation"
)

var readers = strings.Fields("markdown gfm commonmark html docx odt epub rst latex org textile docbook jats ipynb typst json")
var writers = strings.Fields("markdown gfm commonmark html html5 docx odt epub epub3 rst latex org textile docbook jats ipynb typst json plain rtf pptx asciidoc man revealjs beamer")
var bibliographyFormats = strings.Fields("bibtex biblatex csljson ris")

type spec struct {
	id, summary, alias string
	input, output      bool
	options            []operation.Parameter
}

func str(n, d, v string) operation.Parameter { return toolrun.Option(n, operation.TypeString, d, v) }
func num(n, d string, v int) operation.Parameter {
	return toolrun.Option(n, operation.TypeInteger, d, v)
}
func flag(n, d string, v bool) operation.Parameter {
	return toolrun.Option(n, operation.TypeBoolean, d, v)
}
func names(n, d string, required bool) operation.Parameter {
	return operation.Parameter{Name: n, Description: d, Type: operation.TypeStrings, Required: required}
}
func conversion() []operation.Parameter {
	return []operation.Parameter{
		str("to", "Output format; document.formats lists supported formats", "gfm"),
		flag("standalone", "Produce a complete document", true), flag("toc", "Include table of contents where supported", false),
		num("toc-depth", "Maximum table of contents heading depth", 3), flag("number-sections", "Number sections where supported", false),
		num("shift-headings", "Shift heading levels by this amount", 0),
		str("wrap", "auto, none or preserve", "auto"), num("columns", "Text wrapping width", 80), str("eol", "lf or crlf", "lf"),
		str("track-changes", "DOCX revisions: accept, reject or all", "accept"),
		str("highlight-style", "pygments, tango, espresso, zenburn, kate, monochrome, breezedark or haddock", "pygments"),
		str("title", "Override document title", ""), str("author", "Override author", ""), str("lang", "Override language tag", ""), str("date", "Override date", ""),
		str("reference", "Local reference DOCX, ODT or PPTX matching output format", ""),
		str("bibliography", "Local BibTeX, BibLaTeX, CSL JSON or RIS file; enables citeproc", ""), str("csl", "Local citation style (requires bibliography)", ""),
	}
}
func catalog() []spec {
	return []spec{
		{"document.formats", "List supported document and bibliography formats", "文档格式清单", false, false, nil},
		{"document.convert", "Convert structured documents and presentation formats", "转换文档格式", true, true, conversion()},
		{"document.merge", "Merge documents of one input format in order", "合并文档", true, true, append(conversion(), names("files", "Additional local documents in order (up to 15)", true))},
		{"document.text", "Extract readable plain text from a document", "提取文档文本", true, true, []operation.Parameter{str("wrap", "auto, none or preserve", "none"), num("columns", "Text wrapping width", 80)}},
		{"document.inspect", "Inspect document metadata, headings, links and structure counts", "查看文档结构", true, false, nil},
		{"document.media", "Extract embedded office or EPUB media into numbered files", "提取文档内嵌资源", true, false, []operation.Parameter{toolrun.Param("output", "New directory for embedded resources", true)}},
		{"document.split", "Split top-level document blocks at headings into text documents", "按标题拆分文档", true, false, []operation.Parameter{num("level", "Split at headings at or above this level", 1), str("to", "gfm, markdown, html, plain, rst or latex", "gfm"), str("wrap", "auto, none or preserve", "none"), num("columns", "Text wrapping width", 80), toolrun.Param("output", "New directory for numbered sections", true)}},
		{"document.template", "Export a built-in document template", "导出文档模板", false, true, []operation.Parameter{str("to", "html, latex, revealjs, beamer, rst, man, typst or rtf", "html")}},
		{"document.reference", "Export a default office reference document for styling", "导出文档样式参考", false, true, []operation.Parameter{str("to", "docx, odt or pptx", "docx")}},
		{"document.bibliography", "Convert bibliography records between citation formats", "转换参考文献格式", true, true, []operation.Parameter{str("to", "bibtex, biblatex or csljson", "csljson")}},
	}
}

func Register(registry *operation.Registry, resolver toolrun.Resolver) error {
	for _, s := range catalog() {
		def := operation.Definition{ID: s.id, Summary: s.summary, Description: s.summary + ". Managed Pandoc 3.11; local files, sandboxed IO, 3 minute timeout and staged outputs. See docs/documents.md.", Aliases: []string{s.alias}, Tags: []string{"document", "office", "markup"}, Source: "pandoc", Requirements: []operation.Requirement{{Package: "pandoc"}}, Options: append([]operation.Parameter{}, s.options...)}
		if s.input {
			def.Inputs = []operation.Parameter{toolrun.Param("input", "Local document (maximum 32 MiB)", true)}
			def.Options = append(def.Options, str("from", "Input format; auto infers from extension", "auto"))
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

func format(path string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	if x, ok := map[string]string{"md": "markdown", "markdown": "markdown", "htm": "html", "html": "html", "tex": "latex", "bib": "biblatex", "json": "json", "txt": "markdown", "xml": "docbook"}[ext]; ok {
		return x
	}
	return ext
}
func local(path string) (string, error) {
	p, e := toolrun.LocalFile(path)
	if e != nil {
		return "", e
	}
	info, e := os.Stat(p)
	if e != nil {
		return "", e
	}
	if info.Size() > 32<<20 {
		return "", toolrun.Invalid("document exceeds 32 MiB")
	}
	if slices.Contains([]string{".docx", ".odt", ".epub", ".pptx"}, strings.ToLower(filepath.Ext(p))) {
		z, err := zip.OpenReader(p)
		if err != nil {
			return "", toolrun.Invalid("invalid office/EPUB container")
		}
		defer z.Close()
		var size uint64
		if len(z.File) > 10000 {
			return "", toolrun.Invalid("document contains too many entries")
		}
		for _, f := range z.File {
			name := strings.ReplaceAll(f.Name, "\\", "/")
			if strings.HasPrefix(name, "/") || strings.Contains(name, ":") || slices.Contains(strings.Split(name, "/"), "..") || f.Mode()&os.ModeSymlink != 0 {
				return "", toolrun.Invalid("unsafe document container entry")
			}
			if f.UncompressedSize64 > 128<<20-size {
				return "", toolrun.Invalid("expanded document exceeds 128 MiB")
			}
			size += f.UncompressedSize64
		}
	}
	return p, nil
}
func invoke(ctx context.Context, resolver toolrun.Resolver, dir string, args ...string) ([]byte, error) {
	env := []string{}
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if !strings.EqualFold(key, "GHCRTS") && !strings.HasPrefix(strings.ToUpper(key), "PANDOC_") {
			env = append(env, entry)
		}
	}
	args = append([]string{"+RTS", "-M512M", "-RTS", "--sandbox", "--fail-if-warnings"}, args...)
	return toolrun.RunWithEnv(ctx, resolver, "pandoc", "pandoc", dir, env, args...)
}
func run(ctx context.Context, resolver toolrun.Resolver, s spec, r operation.Request) (operation.Result, error) {
	result := operation.Result{}
	v := toolrun.Values{Request: r}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	if s.id == "document.formats" {
		return operation.Result{Data: map[string]any{"readers": readers, "writers": writers, "bibliographyReaders": bibliographyFormats, "bibliographyWriters": []string{"bibtex", "biblatex", "csljson"}}}, nil
	}
	dir, e := os.MkdirTemp("", "finishbit-document-")
	if e != nil {
		return result, e
	}
	defer os.RemoveAll(dir)
	from := ""
	inputs := []string{}
	if s.input {
		from = v.String("from", "auto")
		if from == "auto" {
			from = format(r.Inputs[0])
			if s.id == "document.bibliography" && from == "json" {
				from = "csljson"
			}
		}
		allowed := readers
		if s.id == "document.bibliography" {
			allowed = bibliographyFormats
		}
		v.Check(slices.Contains(allowed, from), "unsupported input format")
		paths := append([]string{}, r.Inputs...)
		if s.id == "document.merge" {
			more := v.Strings("files")
			v.Check(len(more) > 0 && len(more) <= 15, "files must contain 1 to 15 documents")
			paths = append(paths, more...)
		}
		if v.Err != nil {
			return result, v.Err
		}
		var total int64
		for i, path := range paths {
			p, err := local(path)
			if err != nil {
				return result, err
			}
			info, err := os.Stat(p)
			if err != nil {
				return result, err
			}
			total += info.Size()
			if total > 64<<20 {
				return result, toolrun.Invalid("combined documents exceed 64 MiB")
			}
			dest := filepath.Join(dir, fmt.Sprintf("input-%02d%s", i, filepath.Ext(p)))
			if err := toolrun.CopyFile(p, dest); err != nil {
				return result, err
			}
			inputs = append(inputs, dest)
		}
	}
	args := []string{}
	if s.input {
		args = append(args, "--from="+from)
	}
	switch s.id {
	case "document.media":
		v.Check(slices.Contains([]string{"docx", "odt", "epub"}, from), "media extraction requires docx, odt or epub input")
		out := v.String("output", "")
		if v.Err != nil {
			return result, v.Err
		}
		media := filepath.Join(dir, "media")
		_, err := invoke(ctx, resolver, dir, append(args, append([]string{"--to=json", "--extract-media=" + media}, inputs...)...)...)
		if err != nil {
			return result, err
		}
		mapping := []any{}
		result.Outputs, err = toolrun.Directory(ctx, out, func(target string) error {
			var total int64
			count := 0
			return filepath.WalkDir(media, func(path string, entry fs.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if entry.IsDir() {
					return nil
				}
				if !entry.Type().IsRegular() {
					return toolrun.Invalid("unexpected non-file media")
				}
				info, e := entry.Info()
				if e != nil {
					return e
				}
				total += info.Size()
				count++
				if total > 128<<20 || count > 10000 {
					return toolrun.Invalid("embedded media exceeds limits")
				}
				name := fmt.Sprintf("resource-%04d%s", count, filepath.Ext(path))
				relative, e := filepath.Rel(media, path)
				if e != nil {
					return e
				}
				mapping = append(mapping, map[string]any{"source": filepath.ToSlash(relative), "file": name})
				return toolrun.CopyFile(path, filepath.Join(target, name))
			})
		})
		result.Data = map[string]any{"resources": mapping}
		return result, err
	case "document.inspect", "document.split":
		data, err := invoke(ctx, resolver, dir, append(args, append([]string{"--to=json"}, inputs...)...)...)
		if err != nil {
			return result, err
		}
		var ast map[string]any
		if err := json.Unmarshal(data, &ast); err != nil {
			return result, err
		}
		if s.id == "document.inspect" {
			return operation.Result{Data: inspect(ast)}, nil
		}
		level := v.Int("level", 1, 1, 6)
		to := v.Enum("to", "gfm", "gfm", "markdown", "html", "plain", "rst", "latex")
		wrap := v.Enum("wrap", "none", "none", "auto", "preserve")
		columns := v.Int("columns", 80, 20, 240)
		out := v.String("output", "")
		if v.Err != nil {
			return result, v.Err
		}
		blocks, ok := ast["blocks"].([]any)
		if !ok {
			return result, toolrun.Invalid("document has no block list")
		}
		groups := [][]any{}
		current := []any{}
		for _, block := range blocks {
			node, _ := block.(map[string]any)
			if node["t"] == "Header" {
				parts, _ := node["c"].([]any)
				if len(parts) > 0 {
					n, _ := parts[0].(float64)
					if n <= float64(level) && len(current) > 0 {
						groups = append(groups, current)
						current = nil
					}
				}
			}
			current = append(current, block)
		}
		if len(current) > 0 {
			groups = append(groups, current)
		}
		if len(groups) > 200 {
			return result, toolrun.Invalid("split exceeds 200 sections")
		}
		result.Outputs, err = toolrun.Directory(ctx, out, func(target string) error {
			for i, group := range groups {
				ast["blocks"] = group
				b, err := json.Marshal(ast)
				if err != nil {
					return err
				}
				input := filepath.Join(dir, "section.json")
				if err = os.WriteFile(input, b, 0600); err != nil {
					return err
				}
				_, err = invoke(ctx, resolver, dir, "--from=json", "--to="+to, "--wrap="+wrap, "--columns="+strconv.Itoa(columns), "--output="+filepath.Join(target, fmt.Sprintf("section-%03d.%s", i+1, extension(to))), input)
				if err != nil {
					return err
				}
			}
			return nil
		})
		return result, err
	case "document.template":
		args = append(args, "--print-default-template="+v.Enum("to", "html", "html", "latex", "revealjs", "beamer", "rst", "man", "typst", "rtf"))
	case "document.reference":
		args = append(args, "--print-default-data-file=reference."+v.Enum("to", "docx", "docx", "odt", "pptx"))
	case "document.bibliography":
		args = append(args, "--to="+v.Enum("to", "csljson", "bibtex", "biblatex", "csljson"))
	case "document.text":
		args = append(args, "--to=plain", "--wrap="+v.Enum("wrap", "none", "none", "auto", "preserve"), "--columns="+strconv.Itoa(v.Int("columns", 80, 20, 240)))
	default:
		to := v.Enum("to", "gfm", writers...)
		args = append(args, "--to="+to, "--wrap="+v.Enum("wrap", "auto", "auto", "none", "preserve"), "--columns="+strconv.Itoa(v.Int("columns", 80, 20, 240)), "--eol="+v.Enum("eol", "lf", "lf", "crlf"), "--track-changes="+v.Enum("track-changes", "accept", "accept", "reject", "all"), "--shift-heading-level-by="+strconv.Itoa(v.Int("shift-headings", 0, -5, 5)), "--toc-depth="+strconv.Itoa(v.Int("toc-depth", 3, 1, 6)), "--highlight-style="+v.Enum("highlight-style", "pygments", "pygments", "tango", "espresso", "zenburn", "kate", "monochrome", "breezedark", "haddock"))
		for _, option := range []string{"standalone", "toc", "number-sections"} {
			if v.Bool(option, option == "standalone") {
				args = append(args, "--"+option)
			}
		}
		for _, key := range []string{"title", "author", "lang", "date"} {
			if value := v.String(key, ""); value != "" {
				args = append(args, "--metadata="+key+":"+value)
			}
		}
		// A fallback title avoids Pandoc's missing-title warning in standalone HTML.
		args = append(args, "--variable=pagetitle:"+strings.TrimSuffix(filepath.Base(r.Inputs[0]), filepath.Ext(r.Inputs[0])))
		if s.id == "document.merge" {
			args = append(args, "--file-scope")
		}
		for _, key := range []string{"reference", "bibliography", "csl"} {
			if path := v.String(key, ""); path != "" {
				p, err := local(path)
				if err != nil {
					return result, err
				}
				ext := strings.ToLower(filepath.Ext(p))
				if key == "reference" {
					v.Check(slices.Contains([]string{"docx", "odt", "pptx"}, to) && ext == "."+to, "reference format must match office output")
				}
				if key == "bibliography" {
					v.Check(slices.Contains([]string{".bib", ".bibtex", ".biblatex", ".json", ".ris"}, ext), "unsupported bibliography extension")
					// Built-in citeproc has IO outside the reader/writer sandbox. Reject
					// document-supplied resource metadata before enabling it.
					for _, input := range inputs {
						data, err := invoke(ctx, resolver, dir, "--from="+from, "--to=json", input)
						if err != nil {
							return result, err
						}
						var ast map[string]any
						if err = json.Unmarshal(data, &ast); err != nil {
							return result, err
						}
						meta, _ := ast["meta"].(map[string]any)
						for _, name := range []string{"bibliography", "csl", "citation-abbreviations"} {
							if _, present := meta[name]; present {
								return result, toolrun.Invalid("citation resource metadata must be removed from input; use explicit local bibliography/csl options")
							}
						}
					}
				}
				if key == "csl" {
					v.Check(v.String("bibliography", "") != "", "csl requires bibliography")
					v.Check(ext == ".csl", "citation style must be a .csl file")
					data, err := os.ReadFile(p)
					if err != nil {
						return result, err
					}
					if err = checkStyle(data); err != nil {
						return result, err
					}
				}
				dest := filepath.Join(dir, key+ext)
				if err := toolrun.CopyFile(p, dest); err != nil {
					return result, err
				}
				option := key
				if key == "reference" {
					option = "reference-doc"
				}
				args = append(args, "--"+option+"="+dest)
				if key == "bibliography" {
					args = append(args, "--citeproc")
				}
			}
		}
	}
	out := v.String("output", "")
	v.Check(!strings.EqualFold(filepath.Ext(out), ".pdf"), "PDF output requires a separate typesetting engine and is not supported")
	overwrite := v.Bool("overwrite", false)
	if v.Err != nil {
		return result, v.Err
	}
	args = append(args, inputs...)
	e = toolrun.Artifact(ctx, out, overwrite, func(target string) error {
		if s.id == "document.template" || s.id == "document.reference" {
			data, err := invoke(ctx, resolver, dir, args...)
			if err != nil {
				return err
			}
			return os.WriteFile(target, data, 0600)
		}
		// Explicit --to prevents a .pdf destination from selecting an external PDF engine.
		_, err := invoke(ctx, resolver, dir, append(args, "--output="+target)...)
		return err
	})
	if e == nil {
		result.Outputs = []string{out}
	}
	return result, e
}

// Dependent CSL styles can fetch their parent outside Pandoc's sandbox.
func checkStyle(data []byte) error {
	d := xml.NewDecoder(strings.NewReader(string(data)))
	for {
		token, e := d.Token()
		if e == io.EOF {
			return nil
		}
		if e != nil {
			return toolrun.Invalid("invalid CSL XML")
		}
		if element, ok := token.(xml.StartElement); ok && element.Name.Local == "link" {
			for _, a := range element.Attr {
				if a.Name.Local == "rel" && a.Value == "independent-parent" {
					return toolrun.Invalid("CSL must be a self-contained independent style")
				}
			}
		}
	}
}
func extension(to string) string {
	if v, ok := map[string]string{"gfm": "md", "markdown": "md", "plain": "txt", "latex": "tex"}[to]; ok {
		return v
	}
	return to
}
func inlineText(value any) string {
	var b strings.Builder
	var walk func(any)
	walk = func(v any) {
		switch x := v.(type) {
		case []any:
			for _, child := range x {
				walk(child)
			}
		case map[string]any:
			switch x["t"] {
			case "Str":
				s, _ := x["c"].(string)
				b.WriteString(s)
			case "Space", "SoftBreak", "LineBreak":
				b.WriteByte(' ')
			default:
				walk(x["c"])
			}
		}
	}
	walk(value)
	return b.String()
}
func inspect(ast map[string]any) map[string]any {
	counts := map[string]int{}
	headings := []any{}
	links := []any{}
	images := []any{}
	var walk func(any)
	walk = func(v any) {
		switch x := v.(type) {
		case []any:
			for _, child := range x {
				walk(child)
			}
		case map[string]any:
			kind, _ := x["t"].(string)
			if kind != "" {
				counts[kind]++
			}
			parts, _ := x["c"].([]any)
			if kind == "Header" && len(parts) == 3 {
				headings = append(headings, map[string]any{"level": parts[0], "text": inlineText(parts[2])})
			}
			if (kind == "Link" || kind == "Image") && len(parts) == 3 {
				target, _ := parts[2].([]any)
				if len(target) > 0 {
					entry := map[string]any{"text": inlineText(parts[1]), "target": target[0]}
					if kind == "Link" {
						links = append(links, entry)
					} else {
						images = append(images, entry)
					}
				}
			}
			for _, child := range x {
				walk(child)
			}
		}
	}
	walk(ast["blocks"])
	return map[string]any{"metadata": ast["meta"], "headings": headings, "links": links, "images": images, "counts": counts}
}
