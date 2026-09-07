package document

import (
	"archive/zip"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

type runtimeResolver string

func (r runtimeResolver) Executable(string, string) (string, error) { return string(r), nil }
func TestDocumentBoundaries(t *testing.T) {
	root := t.TempDir()
	input := filepath.Join(root, "input.md")
	if e := os.WriteFile(input, []byte("# Test"), 0600); e != nil {
		t.Fatal(e)
	}
	registry := operation.NewRegistry()
	if e := Register(registry, runtimeResolver("missing-pandoc")); e != nil {
		t.Fatal(e)
	}
	for _, options := range []map[string]any{{"to": "pdf"}, {"to": "custom.lua"}, {"from": "https://example.org"}, {"columns": 0}, {"track-changes": "wrong"}, {"output": filepath.Join(root, "output.pdf")}} {
		if options["output"] == nil {
			options["output"] = filepath.Join(root, "out")
		}
		c, _ := registry.Get("document.convert")
		_, e := c.Runner.Run(context.Background(), operation.Request{Inputs: []string{input}, Options: options})
		if e == nil || operation.AsError(e).Code != operation.CodeInvalidInput {
			t.Errorf("%v: %v", options, e)
		}
	}
	bad := filepath.Join(root, "bad.docx")
	f, e := os.Create(bad)
	if e != nil {
		t.Fatal(e)
	}
	w := zip.NewWriter(f)
	part, e := w.Create("../escape")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = part.Write([]byte("x")); e != nil {
		t.Fatal(e)
	}
	if e = w.Close(); e != nil {
		t.Fatal(e)
	}
	if e = f.Close(); e != nil {
		t.Fatal(e)
	}
	if _, e = local(bad); e == nil {
		t.Fatal("accepted traversal container")
	}
}
func TestRealPandoc(t *testing.T) {
	binary := os.Getenv("FINISHBIT_TEST_PANDOC")
	if binary == "" {
		t.Skip("set FINISHBIT_TEST_PANDOC to managed Pandoc executable")
	}
	root := t.TempDir()
	registry := operation.NewRegistry()
	if e := Register(registry, runtimeResolver(binary)); e != nil {
		t.Fatal(e)
	}
	write := func(name, body string) string {
		p := filepath.Join(root, name)
		if e := os.WriteFile(p, []byte(body), 0600); e != nil {
			t.Fatal(e)
		}
		return p
	}
	input := write("sample.md", "---\ntitle: Example\nauthor: Tester\n---\n# First\n\nHello **world**. [Example](https://example.org).\n\n| Item | Value |\n|---|---|\n| A | 42 |\n\n# Second\n\nFinal paragraph.\n")
	call := func(id string, inputs []string, opts map[string]any) operation.Result {
		t.Helper()
		c, ok := registry.Get(id)
		if !ok {
			t.Fatal(id)
		}
		r, e := c.Runner.Run(context.Background(), operation.Request{Inputs: inputs, Options: opts})
		if e != nil {
			t.Fatalf("%s %v: %v (%+v)", id, opts, e, operation.AsError(e))
		}
		return r
	}
	read := func(path string) string {
		b, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		return string(b)
	}
	call("document.formats", nil, nil)
	inspection := call("document.inspect", []string{input}, nil)
	if len(inspection.Data["headings"].([]any)) != 2 || len(inspection.Data["links"].([]any)) != 1 || inspection.Data["counts"].(map[string]int)["Table"] != 1 {
		t.Fatalf("inspection: %+v", inspection)
	}
	for _, to := range writers {
		t.Run(to, func(t *testing.T) {
			output := filepath.Join(root, "converted."+to)
			call("document.convert", []string{input}, map[string]any{"to": to, "output": output})
			info, e := os.Stat(output)
			if e != nil || info.Size() == 0 {
				t.Fatalf("no output: %v", e)
			}
			if to == "docx" || to == "odt" || to == "epub" || to == "html" || to == "json" {
				plain := output + ".txt"
				call("document.text", []string{output}, map[string]any{"from": to, "output": plain})
				if text := read(plain); !strings.Contains(text, "Hello world") || !strings.Contains(text, "Final paragraph") {
					t.Fatalf("lost content: %s", text)
				}
			}
		})
	}
	for _, to := range []string{"docx", "odt", "pptx"} {
		ref := filepath.Join(root, "reference."+to)
		call("document.reference", nil, map[string]any{"to": to, "output": ref})
		call("document.convert", []string{input}, map[string]any{"to": to, "reference": ref, "output": filepath.Join(root, "styled."+to)})
	}
	for _, to := range []string{"html", "latex", "revealjs", "beamer", "rst", "man", "typst", "rtf"} {
		out := filepath.Join(root, "template."+to)
		call("document.template", nil, map[string]any{"to": to, "output": out})
		if len(read(out)) == 0 {
			t.Fatal("empty template")
		}
	}
	second := write("second.md", "# Third\n\nUnique ending.\n")
	merged := filepath.Join(root, "merged.md")
	call("document.merge", []string{input}, map[string]any{"files": []string{second}, "output": merged})
	if !strings.Contains(read(merged), "Unique ending") {
		t.Fatal("merge lost second input")
	}
	sections := call("document.split", []string{input}, map[string]any{"output": filepath.Join(root, "sections")})
	if len(sections.Outputs) != 2 || !strings.Contains(read(sections.Outputs[1]), "Final paragraph") {
		t.Fatal("split lost sections")
	}
	bib := write("references.bib", "@book{doe, title={Testing Books}, author={Doe, Jane}, year={2024}, publisher={Example}}")
	for _, to := range []string{"bibtex", "biblatex", "csljson"} {
		out := filepath.Join(root, "references."+to)
		call("document.bibliography", []string{bib}, map[string]any{"to": to, "output": out})
		if !strings.Contains(strings.ToLower(read(out)), "testing books") {
			t.Fatal("bibliography lost title")
		}
		if to == "csljson" {
			var records []map[string]any
			if e := json.Unmarshal([]byte(read(out)), &records); e != nil || len(records) != 1 {
				t.Fatal("invalid CSL JSON")
			}
		}
	}
	cited := write("cited.md", "# Citations\n\nAccording to @doe.\n")
	out := filepath.Join(root, "cited.html")
	call("document.convert", []string{cited}, map[string]any{"to": "html", "bibliography": bib, "output": out})
	if !strings.Contains(strings.ToLower(read(out)), "testing books") {
		t.Fatal("citeproc did not render bibliography")
	}
	existing := write("existing.md", "keep-me")
	c, _ := registry.Get("document.convert")
	_, e := c.Runner.Run(context.Background(), operation.Request{Inputs: []string{input}, Options: map[string]any{"output": existing}})
	if e == nil || read(existing) != "keep-me" {
		t.Fatal("overwrote existing output")
	}
	bad := write("broken.md", "![missing](private.png)")
	_, e = c.Runner.Run(context.Background(), operation.Request{Inputs: []string{bad}, Options: map[string]any{"to": "docx", "output": existing, "overwrite": true}})
	if e == nil || read(existing) != "keep-me" {
		t.Fatal("failed conversion damaged existing file")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, e = c.Runner.Run(ctx, operation.Request{Inputs: []string{input}, Options: map[string]any{"output": existing, "overwrite": true}})
	if e == nil || read(existing) != "keep-me" {
		t.Fatal("cancelled conversion changed target")
	}
	var fetched atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fetched.Add(1); w.Write([]byte("PRIVATE-SENTINEL")) }))
	defer server.Close()
	hostile := write("remote.html", "<iframe src=\""+server.URL+"\"></iframe>")
	htmlCap, _ := registry.Get("document.text")
	_, _ = htmlCap.Runner.Run(context.Background(), operation.Request{Inputs: []string{hostile}, Options: map[string]any{"output": filepath.Join(root, "remote.txt")}})
	hostile = write("remote-citation.md", "---\ncsl: "+server.URL+"/remote.csl\n---\nAccording to @doe.\n")
	_, _ = c.Runner.Run(context.Background(), operation.Request{Inputs: []string{hostile}, Options: map[string]any{"bibliography": bib, "output": filepath.Join(root, "remote-cited.md")}})
	if fetched.Load() != 0 {
		t.Fatal("sandbox made an unexpected network request")
	}
	// Build a DOCX with an embedded image via Pandoc's sandbox-supported data URI.
	picture := write("picture.md", "# Image\n\n![pixel](data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+aDZkAAAAASUVORK5CYII=)\n")
	docx := filepath.Join(root, "picture.docx")
	call("document.convert", []string{picture}, map[string]any{"to": "docx", "output": docx})
	media := call("document.media", []string{docx}, map[string]any{"output": filepath.Join(root, "media")})
	if len(media.Outputs) != 1 || !strings.HasPrefix(read(media.Outputs[0]), "\x89PNG") {
		t.Fatal("embedded image missing")
	}
}
