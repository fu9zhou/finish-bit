package app

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func TestLocalWorkflowsWithManagedTools(t *testing.T) {
	root := os.Getenv("FINISHBIT_TEST_LOCAL_HOME")
	if root == "" {
		t.Skip("set FINISHBIT_TEST_LOCAL_HOME for real offline OCR/PDF/document acceptance")
	}
	a, e := New(Config{Root: root})
	if e != nil {
		t.Fatal(e)
	}
	ctx := context.Background()
	dir := t.TempDir()
	run := func(id string, inputs []string, opts map[string]any) operation.Result {
		t.Helper()
		r, e := a.Execute(ctx, id, operation.Request{Inputs: inputs, Options: opts})
		if e != nil {
			t.Fatalf("%s: %v (%+v)", id, e, operation.AsError(e).Details)
		}
		return r
	}
	font := os.Getenv("FINISHBIT_TEST_FONT")
	if font == "" {
		font = filepath.Join(os.Getenv("WINDIR"), "Fonts", "arial.ttf")
	}
	canvas := filepath.Join(dir, "canvas.png")
	run("image.solid", nil, map[string]any{"width": 1200, "height": 240, "output": canvas})
	image := filepath.Join(dir, "printed.png")
	run("image.annotate", []string{canvas}, map[string]any{"text": "FINISHBIT LOCAL 2026", "font": font, "size": 64, "x": 40, "y": 40, "color": "#000000", "output": image})
	recognized := run("ocr.text", []string{image}, map[string]any{"language": "eng", "layout": 6})
	if !strings.Contains(recognized.Data["text"].(string), "FINISHBIT LOCAL 2026") {
		t.Fatal(recognized.Data)
	}
	words := run("ocr.words", []string{image}, map[string]any{"language": "eng", "layout": 6})
	if len(words.Data["words"].([]map[string]any)) < 3 {
		t.Fatal(words.Data)
	}
	for _, word := range words.Data["words"].([]map[string]any) {
		left, top := word["left"].(int), word["top"].(int)
		width, height := word["width"].(int), word["height"].(int)
		confidence := word["confidence"].(float64)
		if left < 0 || top < 0 || width <= 0 || height <= 0 || left+width > 1200 || top+height > 240 || confidence < 0 || confidence > 100 {
			t.Fatalf("invalid OCR word geometry/confidence: %v", word)
		}
	}
	scanned := filepath.Join(dir, "scanned.pdf")
	run("ocr.to-pdf", []string{image}, map[string]any{"language": "eng", "layout": 6, "output": scanned})
	text := run("pdf.extract-text", []string{scanned}, nil)
	if !strings.Contains(text.Data["text"].(string), "FINISHBIT") {
		t.Fatal(text)
	}
	hocr := filepath.Join(dir, "ocr.html")
	run("ocr.to-html", []string{image}, map[string]any{"language": "eng", "output": hocr})
	b, _ := os.ReadFile(hocr)
	if !strings.Contains(string(b), "ocrx_word") {
		t.Fatal("missing hOCR")
	}
	chineseFont := filepath.Join(os.Getenv("WINDIR"), "Fonts", "msyh.ttc")
	if _, e := os.Stat(chineseFont); e == nil {
		cn := filepath.Join(dir, "中文.png")
		run("image.annotate", []string{canvas}, map[string]any{"text": "你好世界 测试文字", "font": chineseFont, "size": 64, "x": 40, "y": 40, "color": "#000000", "output": cn})
		r := run("ocr.text", []string{cn}, map[string]any{"language": "chi_sim", "layout": 6})
		if !strings.Contains(strings.ReplaceAll(r.Data["text"].(string), " ", ""), "你好世界") {
			t.Fatal(r.Data)
		}
		cnPDF := filepath.Join(dir, "中文识别.pdf")
		run("ocr.to-pdf", []string{cn}, map[string]any{"language": "chi_sim", "layout": 6, "output": cnPDF})
		cnText := run("pdf.extract-text", []string{cnPDF}, nil)
		// PDF extraction may put a line break between adjacent Chinese glyphs.
		if !strings.Contains(strings.Join(strings.Fields(cnText.Data["text"].(string)), ""), "你好世界") {
			t.Fatal("Chinese PDF text layer lost recognized text", cnText.Data)
		}
	}
	two := filepath.Join(dir, "two.pdf")
	run("pdf.merge", []string{scanned}, map[string]any{"files": []string{scanned}, "output": two})
	for _, id := range []string{"pdf.rasterize", "pdf.compress-images", "pdf.ocr", "pdf.ocr-text", "pdf.to-docx", "pdf.to-pptx", "pdf.to-long-image"} {
		t.Run(id, func(t *testing.T) {
			ext := ".pdf"
			switch id {
			case "pdf.ocr-text":
				ext = ".txt"
			case "pdf.to-docx":
				ext = ".docx"
			case "pdf.to-pptx":
				ext = ".pptx"
			case "pdf.to-long-image":
				ext = ".png"
			}
			out := filepath.Join(dir, strings.ReplaceAll(id, ".", "-")+ext)
			opts := map[string]any{"first": 1, "last": 2, "dpi": 120, "output": out}
			if strings.Contains(id, "ocr") {
				opts["language"] = "eng"
			}
			r, e := a.Execute(ctx, id, operation.Request{Inputs: []string{two}, Options: opts})
			if e != nil {
				t.Fatalf("%s: %v (%v)", id, e, operation.AsError(e).Details)
			}
			if len(r.Outputs) != 1 {
				t.Fatal(r)
			}
			stat, e := os.Stat(out)
			if e != nil || stat.Size() == 0 {
				t.Fatal("empty artifact", e)
			}
			if ext == ".pdf" {
				run("pdf.validate", []string{out}, nil)
			}
			if id == "pdf.to-pptx" {
				z, e := zip.OpenReader(out)
				if e != nil {
					t.Fatal(e)
				}
				defer z.Close()
				slides := 0
				for _, f := range z.File {
					if strings.HasPrefix(f.Name, "ppt/slides/slide") && strings.HasSuffix(f.Name, ".xml") {
						slides++
					}
				}
				if slides != 2 {
					t.Fatalf("slide count %d", slides)
				}
			}
		})
	}
	docx := filepath.Join(dir, "scan.docx")
	run("document.scan", []string{image}, map[string]any{"language": "eng", "output": docx})
	plain := filepath.Join(dir, "scan.txt")
	run("document.text", []string{docx}, map[string]any{"output": plain})
	b, _ = os.ReadFile(plain)
	if !strings.Contains(string(b), "FINISHBIT") {
		t.Fatal(string(b))
	}
	comparison := filepath.Join(dir, "comparison.json")
	run("document.compare", []string{docx, docx}, map[string]any{"output": comparison})
	b, _ = os.ReadFile(comparison)
	if !strings.Contains(string(b), `"equal":true`) {
		t.Fatal(string(b))
	}
	numbered := filepath.Join(dir, "numbered.pdf")
	run("pdf.page-numbers", []string{two}, map[string]any{"text": "Page %p of %P", "output": numbered})
	r := run("pdf.extract-text", []string{numbered}, nil)
	if !strings.Contains(r.Data["text"].(string), "Page 2 of 2") {
		t.Fatal(r.Data)
	}
	meta := filepath.Join(dir, "meta.pdf")
	run("pdf.metadata-set", []string{two}, map[string]any{"title": "FinishBit Test", "author": "Offline", "output": meta})
	run("pdf.validate", []string{meta}, nil)
	resized := filepath.Join(dir, "a4.pdf")
	run("pdf.resize-pages", []string{two}, map[string]any{"paper": "A4P", "output": resized})
	run("pdf.validate", []string{resized}, nil)
	cropped := filepath.Join(dir, "cropped.pdf")
	run("pdf.crop", []string{two}, map[string]any{"top": 2, "right": 3, "bottom": 4, "left": 5, "output": cropped})
	run("pdf.validate", []string{cropped}, nil)
	signed := filepath.Join(dir, "signed.pdf")
	run("pdf.sign-image", []string{two}, map[string]any{"signature": image, "pages": "2", "output": signed})
	run("pdf.validate", []string{signed}, nil)
	filtered := filepath.Join(dir, "oil.png")
	run("image.filter", []string{image}, map[string]any{"output": filtered, "preset": "oil"})
	compressed := filepath.Join(dir, "compressed.docx")
	run("document.compress", []string{docx}, map[string]any{"output": compressed})
	run("document.text", []string{compressed}, map[string]any{"output": filepath.Join(dir, "compressed.txt")})
}
