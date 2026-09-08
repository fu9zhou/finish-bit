package app

import (
	"context"
	"encoding/base64"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"

	"github.com/fu9zhou/finish-bit/internal/toolrun"
	"github.com/fu9zhou/finish-bit/pkg/operation"
)

// registerWorkflows composes existing contracts in the application layer. Adapters
// cannot bypass validation or implement a second version of these use cases.
func registerWorkflows(reg *operation.Registry) error {
	type workflow struct {
		id, summary, alias string
		packages           []string
	}
	rows := []workflow{
		{"pdf.rasterize", "Create an image-only PDF from a selected page range", "转纯图PDF", []string{"poppler", "pdfcpu"}},
		{"pdf.compress-images", "Rebuild selected PDF pages with JPEG compression and report the size change", "扫描PDF图像压缩", []string{"poppler", "pdfcpu", "imagemagick"}},
		{"pdf.to-docx", "Extract PDF text into a reflowed Word document", "PDF转Word文字版", []string{"poppler", "pandoc"}},
		{"pdf.to-pptx", "Create a presentation with one rendered PDF page per slide", "PDF转PPT图片版", []string{"poppler", "pandoc"}},
		{"pdf.to-long-image", "Join rendered PDF pages into a vertical image", "PDF转长图", []string{"poppler", "imagemagick"}},
		{"pdf.ocr", "Rebuild selected PDF pages with a searchable OCR layer", "扫描PDF文字识别", []string{"poppler", "tesseract", "pdfcpu"}},
		{"pdf.ocr-text", "Recognize text in selected PDF page images", "扫描PDF提取文字", []string{"poppler", "tesseract"}},
		{"document.scan", "Recognize a local document image into a reflowed Word document", "扫描图片转Word", []string{"tesseract", "pandoc"}},
		{"document.compare", "Compare extracted document text without changing either input", "文档合同文本对比", []string{"pandoc"}},
	}
	for _, s := range rows {
		inputs := []operation.Parameter{toolrun.Param("input", "Local source file", true)}
		options := []operation.Parameter{}
		if strings.HasPrefix(s.id, "pdf.") {
			options = append(options, toolrun.Option("first", operation.TypeInteger, "First page", 1), toolrun.Option("last", operation.TypeInteger, "Last page (explicit, up to 50 pages)", 1), toolrun.Option("dpi", operation.TypeInteger, "Rendering DPI", 120))
		}
		if strings.Contains(s.id, "ocr") || s.id == "document.scan" {
			options = append(options, toolrun.Option("language", operation.TypeString, "eng, chi_sim or chi_sim+eng", "chi_sim+eng"))
		}
		if s.id == "document.compare" {
			inputs = append(inputs, toolrun.Param("second", "Second local document", true))
		}
		if s.id == "pdf.compress-images" {
			options = append(options, toolrun.Option("quality", operation.TypeInteger, "JPEG quality; rasterizes pages and removes original text/links/forms", 75))
		}
		options = append(options, toolrun.OutputOptions()...)
		requirements := []operation.Requirement{}
		for _, p := range s.packages {
			requirements = append(requirements, operation.Requirement{Package: p})
		}
		def := operation.Definition{ID: s.id, Summary: s.summary, Description: s.summary + ". Local application workflow; source files are preserved and the final artifact is published only after every step succeeds. Reflowed text and page-image slides are not editable layout reconstruction.", Aliases: []string{s.alias}, Tags: []string{strings.Split(s.id, ".")[0], "workflow", "offline"}, Inputs: inputs, Options: options, Requirements: requirements, Source: "app"}
		if e := reg.Register(operation.Capability{Definition: def, Runner: operation.Func(func(ctx context.Context, r operation.Request) (operation.Result, error) {
			if e := operation.ValidateRequest(def, r); e != nil {
				return operation.Result{}, e
			}
			return runWorkflow(ctx, reg, s.id, r)
		})}); e != nil {
			return e
		}
	}
	return nil
}
func invoke(ctx context.Context, reg *operation.Registry, id string, inputs []string, options map[string]any) (operation.Result, error) {
	cap, ok := reg.Get(id)
	if !ok {
		return operation.Result{}, fmt.Errorf("workflow dependency missing: %s", id)
	}
	r := operation.Request{Inputs: inputs, Options: options}
	if e := operation.ValidateRequest(cap.Definition, r); e != nil {
		return operation.Result{}, e
	}
	if e := ctx.Err(); e != nil {
		return operation.Result{}, e
	}
	return cap.Runner.Run(ctx, r)
}
func runWorkflow(ctx context.Context, reg *operation.Registry, id string, r operation.Request) (operation.Result, error) {
	v := &toolrun.Values{Request: r}
	out := v.String("output", "")
	overwrite := v.Bool("overwrite", false)
	first, last, dpi := 1, 1, 120
	language := "chi_sim+eng"
	if strings.HasPrefix(id, "pdf.") {
		first = v.Int("first", 1, 1, 100000)
		last = v.Int("last", 1, 1, 100000)
		dpi = v.Int("dpi", 120, 36, 300)
		v.Check(last >= first && last-first < 50, "select a forward range of at most 50 pages")
	}
	if strings.Contains(id, "ocr") || id == "document.scan" {
		language = v.Enum("language", "chi_sim+eng", "eng", "chi_sim", "chi_sim+eng")
	}
	if v.Err != nil {
		return operation.Result{}, v.Err
	}
	work, e := os.MkdirTemp("", "finishbit-workflow-")
	if e != nil {
		return operation.Result{}, e
	}
	defer os.RemoveAll(work)
	final := ""
	details := map[string]any{"workflow": id}
	toWord := func(text string) (string, error) {
		if len(text) > 4<<20 {
			return "", toolrun.Invalid("extracted text exceeds 4 MiB")
		}
		source := filepath.Join(work, "text.html")
		markup := "<!doctype html><html><head><meta charset=\"utf-8\"></head><body>"
		for _, paragraph := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n\n") {
			markup += "<p>" + strings.ReplaceAll(html.EscapeString(paragraph), "\n", "<br>") + "</p>"
		}
		markup += "</body></html>"
		if e := os.WriteFile(source, []byte(markup), 0600); e != nil {
			return "", e
		}
		target := filepath.Join(work, "result.docx")
		_, e := invoke(ctx, reg, "document.convert", []string{source}, map[string]any{"from": "html", "to": "docx", "output": target})
		return target, e
	}
	if id == "document.scan" {
		res, e := invoke(ctx, reg, "ocr.text", r.Inputs, map[string]any{"language": language})
		if e != nil {
			return operation.Result{}, e
		}
		text, _ := res.Data["text"].(string)
		if strings.TrimSpace(text) == "" {
			return operation.Result{}, toolrun.Invalid("OCR returned no text")
		}
		final, e = toWord(text)
		if e != nil {
			return operation.Result{}, e
		}
		details["content"] = "recognized text; reflowed layout"
	} else if id == "document.compare" {
		texts := []string{}
		for i, p := range r.Inputs {
			target := filepath.Join(work, fmt.Sprintf("text-%d.txt", i))
			op := "document.text"
			opts := map[string]any{"output": target}
			if strings.EqualFold(filepath.Ext(p), ".pdf") {
				op = "pdf.extract-text"
				opts["unwrap"] = true
			}
			if _, e := invoke(ctx, reg, op, []string{p}, opts); e != nil {
				return operation.Result{}, e
			}
			b, e := os.ReadFile(target)
			if e != nil {
				return operation.Result{}, e
			}
			texts = append(texts, string(b))
		}
		final = filepath.Join(work, "difference.json")
		if _, e := invoke(ctx, reg, "text.diff", texts, map[string]any{"output": final}); e != nil {
			return operation.Result{}, e
		}
		details["content"] = "extracted text differences; no legal interpretation or visual comparison"
	} else if id == "pdf.to-docx" {
		res, e := invoke(ctx, reg, "pdf.extract-text", r.Inputs, map[string]any{"first": first, "last": last, "unwrap": true})
		if e != nil {
			return operation.Result{}, e
		}
		text, _ := res.Data["text"].(string)
		if strings.TrimSpace(text) == "" {
			return operation.Result{}, toolrun.Invalid("PDF has no readable text; use OCR for scanned pages")
		}
		final, e = toWord(text)
		if e != nil {
			return operation.Result{}, e
		}
		details["content"] = "existing PDF text; reflowed layout"
	} else {
		rendered, e := invoke(ctx, reg, "pdf.render", r.Inputs, map[string]any{"first": first, "last": last, "dpi": dpi, "format": "png", "output": filepath.Join(work, "pages")})
		if e != nil {
			return operation.Result{}, e
		}
		pages := rendered.Outputs
		if len(pages) == 0 {
			return operation.Result{}, toolrun.Invalid("no PDF pages rendered")
		}
		details["pages"] = len(pages)
		details["dpi"] = dpi
		switch id {
		case "pdf.rasterize", "pdf.compress-images":
			if id == "pdf.compress-images" {
				quality := v.Int("quality", 75, 1, 100)
				if v.Err != nil {
					return operation.Result{}, v.Err
				}
				for i, p := range pages {
					jpeg := filepath.Join(work, fmt.Sprintf("compressed-%04d.jpg", i))
					if _, err := invoke(ctx, reg, "image.compress", []string{p}, map[string]any{"quality": quality, "output": jpeg}); err != nil {
						return operation.Result{}, err
					}
					pages[i] = jpeg
				}
				details["quality"] = quality
			}
			final = filepath.Join(work, "result.pdf")
			_, e = invoke(ctx, reg, "pdf.from-images", pages[:1], map[string]any{"files": pages[1:], "dpi": dpi, "output": final})
			details["content"] = "raster images; text, links, forms and signatures are not preserved"
		case "pdf.to-long-image":
			final = filepath.Join(work, "result.png")
			if len(pages) == 1 {
				e = toolrun.CopyFile(pages[0], final)
			} else {
				_, e = invoke(ctx, reg, "image.join", pages[:1], map[string]any{"files": pages[1:], "axis": "vertical", "output": final})
			}
		case "pdf.to-pptx":
			var md strings.Builder
			for _, p := range pages {
				b, err := os.ReadFile(p)
				if err != nil {
					return operation.Result{}, err
				}
				if md.Len()+base64.StdEncoding.EncodedLen(len(b)) > 24<<20 {
					return operation.Result{}, toolrun.Invalid("slide image payload exceeds 24 MiB; lower DPI or page count")
				}
				md.WriteString("#\n\n![](data:image/png;base64,")
				md.WriteString(base64.StdEncoding.EncodeToString(b))
				md.WriteString(")\n\n")
			}
			source := filepath.Join(work, "slides.md")
			if e = os.WriteFile(source, []byte(md.String()), 0600); e != nil {
				return operation.Result{}, e
			}
			final = filepath.Join(work, "result.pptx")
			_, e = invoke(ctx, reg, "document.convert", []string{source}, map[string]any{"from": "markdown", "to": "pptx", "output": final})
			details["content"] = "one page image per slide; text is not editable"
		case "pdf.ocr", "pdf.ocr-text":
			texts := []string{}
			searchable := []string{}
			textBytes := 0
			for i, p := range pages {
				if e := ctx.Err(); e != nil {
					return operation.Result{}, e
				}
				if id == "pdf.ocr" {
					target := filepath.Join(work, fmt.Sprintf("ocr-%04d.pdf", i))
					_, e = invoke(ctx, reg, "ocr.to-pdf", []string{p}, map[string]any{"language": language, "dpi": dpi, "output": target})
					searchable = append(searchable, target)
				} else {
					var res operation.Result
					res, e = invoke(ctx, reg, "ocr.text", []string{p}, map[string]any{"language": language, "dpi": dpi})
					text, _ := res.Data["text"].(string)
					textBytes += len(text)
					if textBytes > 16<<20 {
						return operation.Result{}, toolrun.Invalid("OCR text exceeds 16 MiB")
					}
					texts = append(texts, text)
				}
				if e != nil {
					return operation.Result{}, e
				}
			}
			if id == "pdf.ocr" {
				final = filepath.Join(work, "result.pdf")
				if len(searchable) == 1 {
					e = toolrun.CopyFile(searchable[0], final)
				} else {
					_, e = invoke(ctx, reg, "pdf.merge", searchable[:1], map[string]any{"files": searchable[1:], "output": final})
				}
				details["content"] = "rendered pages with recognized text; original forms/links/signatures not retained"
			} else {
				final = filepath.Join(work, "result.txt")
				e = os.WriteFile(final, []byte(strings.Join(texts, "\n\f\n")), 0600)
			}
		default:
			return operation.Result{}, toolrun.Invalid("unknown workflow")
		}
		if e != nil {
			return operation.Result{}, e
		}
	}
	if !strings.EqualFold(filepath.Ext(out), filepath.Ext(final)) {
		return operation.Result{}, toolrun.Invalid("output extension must be " + filepath.Ext(final))
	}
	if id == "pdf.compress-images" {
		before, err := os.Stat(r.Inputs[0])
		if err != nil {
			return operation.Result{}, err
		}
		after, err := os.Stat(final)
		if err != nil {
			return operation.Result{}, err
		}
		details["input_bytes"], details["output_bytes"], details["smaller"] = before.Size(), after.Size(), after.Size() < before.Size()
		details["size_comparison"] = "selected page range against whole input; no target-size guarantee"
	}
	e = toolrun.Artifact(ctx, out, overwrite, func(p string) error { return toolrun.CopyFile(final, p) })
	if e != nil {
		return operation.Result{}, e
	}
	return operation.Result{Operation: id, Outputs: []string{out}, Data: details}, nil
}
