package pdf

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/fu9zhou/finish-bit/internal/toolrun"
	"github.com/fu9zhou/finish-bit/pkg/operation"
	"github.com/fu9zhou/finish-bit/pkg/packagemanager"
)

func fixturePDF(t *testing.T, path string) {
	t.Helper()
	stream1 := "BT /F1 18 Tf 40 220 Td (FinishBit page one) Tj ET"
	stream2 := "BT /F1 18 Tf 40 220 Td (FinishBit page two) Tj ET"
	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R /AcroForm 9 0 R >>",
		"<< /Type /Pages /Kids [4 0 R 6 0 R] /Count 2 >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 300 300] /Resources << /Font << /F1 3 0 R >> >> /Contents 5 0 R /Annots [8 0 R] >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(stream1), stream1),
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 300 300] /Resources << /Font << /F1 3 0 R >> >> /Contents 7 0 R >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(stream2), stream2),
		"<< /Type /Annot /Subtype /Widget /FT /Tx /T (customer) /V (Alice) /DV (Default) /Rect [40 40 200 70] /P 4 0 R /F 4 /DA (/Helv 12 Tf 0 g) >>",
		"<< /Fields [8 0 R] /NeedAppearances true /DA (/Helv 12 Tf 0 g) /DR << /Font << /Helv 3 0 R >> >> >>",
	}
	var buffer bytes.Buffer
	buffer.WriteString("%PDF-1.7\n")
	offsets := []int{0}
	for i, obj := range objects {
		offsets = append(offsets, buffer.Len())
		fmt.Fprintf(&buffer, "%d 0 obj\n%s\nendobj\n", i+1, obj)
	}
	xref := buffer.Len()
	fmt.Fprintf(&buffer, "xref\n0 %d\n0000000000 65535 f \n", len(objects)+1)
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&buffer, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&buffer, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xref)
	if err := os.WriteFile(path, buffer.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestPDFWithManagedTools(t *testing.T) {
	home := os.Getenv("FINISHBIT_TEST_HOME")
	if home == "" {
		t.Skip("set FINISHBIT_TEST_HOME to the managed runtime root for real PDF tests")
	}
	packages, err := packagemanager.BuiltinRegistry()
	if err != nil {
		t.Fatal(err)
	}
	manager := packagemanager.New(home, packages)
	for pkg, exe := range map[string]string{"pdfcpu": "pdfcpu", "poppler": "pdftotext"} {
		if _, err := manager.Executable(pkg, exe); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	registry := operation.NewRegistry()
	if err := Register(registry, manager); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(t.TempDir(), "PDF 素材 ' [case]")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	input := filepath.Join(dir, "original.pdf")
	fixturePDF(t, input)
	picture := filepath.Join(dir, "image.png")
	f, err := os.Create(picture)
	if err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 60, 40))
	for y := 0; y < 40; y++ {
		for x := 0; x < 60; x++ {
			img.Set(x, y, color.RGBA{200, 10, 30, 255})
		}
	}
	_ = png.Encode(f, img)
	_ = f.Close()
	attachment := filepath.Join(dir, "notes.txt")
	_ = os.WriteFile(attachment, []byte("attachment content"), 0600)
	run := func(t *testing.T, id string, inputs []string, options map[string]any) operation.Result {
		t.Helper()
		cap, ok := registry.Get(id)
		if !ok {
			t.Fatal("operation missing:", id)
		}
		r, err := cap.Runner.Run(ctx, operation.Request{Inputs: inputs, Options: options})
		if err != nil {
			t.Fatalf("%s: %v %#v", id, err, operation.AsError(err).Details)
		}
		return r
	}
	pageCount := func(t *testing.T, path string) int {
		t.Helper()
		raw, err := toolrun.Run(ctx, manager, "poppler", "pdfinfo", "", path)
		if err != nil {
			t.Fatal(err)
		}
		match := regexp.MustCompile(`(?m)^Pages:\s+(\d+)`).FindSubmatch(raw)
		if len(match) != 2 {
			t.Fatalf("no page count: %s", raw)
		}
		count, _ := strconv.Atoi(string(match[1]))
		return count
	}
	output := func(id string) string { return filepath.Join(dir, id+".pdf") }
	tests := []struct {
		id     string
		input  string
		second string
		opts   map[string]any
		count  int
	}{
		{"pdf.info", "", "", nil, 0}, {"pdf.validate", "", "", nil, 0},
		{"pdf.merge", "", "", map[string]any{"files": []string{input}}, 4},
		{"pdf.split", "", "", map[string]any{"output": filepath.Join(dir, "split"), "span": 1}, 0},
		{"pdf.select-pages", "", "", map[string]any{"pages": "2,1,2"}, 3},
		{"pdf.remove-pages", "", "", map[string]any{"pages": "2"}, 1},
		{"pdf.insert-pages", "", "", map[string]any{"pages": "1"}, 3},
		{"pdf.rotate", "", "", nil, 2}, {"pdf.crop", "", "", nil, 2},
		{"pdf.watermark", "", "", map[string]any{"text": "TEST WATERMARK"}, 2},
		{"pdf.stamp", "", "", map[string]any{"text": "APPROVED"}, 2},
		{"pdf.remove-watermark", output("pdf.watermark"), "", nil, 2},
		{"pdf.remove-stamp", output("pdf.stamp"), "", nil, 2},
		{"pdf.from-images", picture, "", map[string]any{"files": []string{picture}}, 2},
		{"pdf.optimize", "", "", nil, 2},
		{"pdf.encrypt", "", "", map[string]any{"owner-password": "owner-123", "user-password": "reader-123"}, 0},
		{"pdf.decrypt", output("pdf.encrypt"), "", map[string]any{"password": "reader-123"}, 2},
		{"pdf.attach", "", "", map[string]any{"files": []string{attachment}}, 2},
		{"pdf.attachments", output("pdf.attach"), "", nil, 0},
		{"pdf.extract-attachments", output("pdf.attach"), "", map[string]any{"output": filepath.Join(dir, "attachments")}, 0},
		{"pdf.remove-attachments", output("pdf.attach"), "", nil, 2},
		{"pdf.form-fields", "", "", nil, 0},
		{"pdf.form-export", "", "", map[string]any{"output": filepath.Join(dir, "form.json")}, 0},
		{"pdf.form-fill", "", filepath.Join(dir, "filled-data.json"), nil, 2},
		{"pdf.form-multifill", "", filepath.Join(dir, "multi-data.json"), map[string]any{"output": filepath.Join(dir, "filled-forms")}, 0},
		{"pdf.form-reset", output("pdf.form-fill"), "", nil, 2},
		{"pdf.form-lock", "", "", nil, 2}, {"pdf.form-unlock", output("pdf.form-lock"), "", nil, 2},
		{"pdf.bookmarks-export", output("pdf.merge"), "", map[string]any{"output": filepath.Join(dir, "bookmarks.json")}, 0},
		{"pdf.bookmarks-import", "", filepath.Join(dir, "bookmarks.json"), nil, 2},
		{"pdf.bookmarks-remove", output("pdf.merge"), "", nil, 4},
		{"pdf.keywords-add", "", "", map[string]any{"values": []string{"finishbit", "acceptance"}}, 2},
		{"pdf.keywords", output("pdf.keywords-add"), "", nil, 0},
		{"pdf.keywords-remove", output("pdf.keywords-add"), "", nil, 2},
		{"pdf.nup", "", "", map[string]any{"count": 2}, 1},
		{"pdf.extract-text", "", "", nil, 0}, {"pdf.text-boxes", "", "", nil, 0},
		{"pdf.render", "", "", map[string]any{"output": filepath.Join(dir, "render"), "first": 1, "last": 2, "dpi": 72}, 0},
		{"pdf.extract-images", output("pdf.from-images"), "", map[string]any{"output": filepath.Join(dir, "images"), "first": 1, "last": 2}, 0},
		{"pdf.fonts", "", "", nil, 0},
		{"pdf.to-html", "", "", map[string]any{"output": filepath.Join(dir, "html"), "first": 1, "last": 2}, 0},
		{"pdf.to-ps", "", "", map[string]any{"output": filepath.Join(dir, "print.ps")}, 0},
	}
	for _, test := range tests {
		t.Run(test.id, func(t *testing.T) {
			in := test.input
			if in == "" {
				in = input
			}
			inputs := []string{in}
			if test.second != "" {
				inputs = append(inputs, test.second)
			}
			opts := test.opts
			if opts == nil {
				opts = map[string]any{}
			}
			cap, _ := registry.Get(test.id)
			for _, param := range cap.Definition.Options {
				if param.Name == "output" && param.Required {
					if _, ok := opts["output"]; !ok {
						opts["output"] = output(test.id)
					}
				}
			}
			result := run(t, test.id, inputs, opts)
			for _, path := range result.Outputs {
				info, err := os.Stat(path)
				if err != nil || info.Size() == 0 {
					t.Fatalf("missing/empty output %s: %v", path, err)
				}
			}
			if test.count > 0 {
				if count := pageCount(t, result.Outputs[0]); count != test.count {
					t.Fatalf("pages=%d want=%d", count, test.count)
				}
				run(t, "pdf.validate", []string{result.Outputs[0]}, nil)
			}
			switch test.id {
			case "pdf.split":
				if len(result.Outputs) != 2 {
					t.Fatalf("split outputs=%v", result.Outputs)
				}
				for _, path := range result.Outputs {
					if pageCount(t, path) != 1 {
						t.Fatal("split page count")
					}
				}
			case "pdf.info":
				if len(result.Data) == 0 {
					t.Fatal("empty info")
				}
			case "pdf.attachments":
				if !strings.Contains(fmt.Sprint(result.Data), "notes.txt") {
					t.Fatal("attachment not listed")
				}
			case "pdf.extract-attachments":
				if len(result.Outputs) != 1 {
					t.Fatal("expected one attachment")
				}
				data, _ := os.ReadFile(result.Outputs[0])
				if string(data) != "attachment content" {
					t.Fatal("attachment content changed")
				}
			case "pdf.form-fields":
				if !strings.Contains(fmt.Sprint(result.Data), "Alice") {
					t.Fatalf("form data: %v", result.Data)
				}
			case "pdf.form-export":
				raw, _ := os.ReadFile(result.Outputs[0])
				if !bytes.Contains(raw, []byte("Alice")) {
					t.Fatal("missing original field value")
				}
				raw = bytes.ReplaceAll(raw, []byte(`"Alice"`), []byte(`"Bob"`))
				if err := os.WriteFile(filepath.Join(dir, "filled-data.json"), raw, 0600); err != nil {
					t.Fatal(err)
				}
				var multi map[string]any
				if err := json.Unmarshal(raw, &multi); err != nil {
					t.Fatal(err)
				}
				forms := multi["forms"].([]any)
				multi["forms"] = append(forms, forms[0])
				encoded, _ := json.Marshal(multi)
				if err := os.WriteFile(filepath.Join(dir, "multi-data.json"), encoded, 0600); err != nil {
					t.Fatal(err)
				}
			case "pdf.form-multifill":
				if len(result.Outputs) != 2 {
					t.Fatalf("expected two filled forms: %v", result.Outputs)
				}
				for _, path := range result.Outputs {
					if pageCount(t, path) != 2 {
						t.Fatal("bad filled form page count")
					}
					fields := run(t, "pdf.form-fields", []string{path}, nil)
					if !strings.Contains(fmt.Sprint(fields.Data), "Bob") {
						t.Fatal("multifill value missing")
					}
				}
			case "pdf.form-fill":
				fields := run(t, "pdf.form-fields", result.Outputs, nil)
				if !strings.Contains(fmt.Sprint(fields.Data), "Bob") {
					t.Fatal("form was not filled")
				}
			case "pdf.bookmarks-export":
				raw, _ := os.ReadFile(result.Outputs[0])
				var value any
				if json.Unmarshal(raw, &value) != nil {
					t.Fatal("invalid bookmarks JSON")
				} // Restrict imported bookmarks to the original two pages.
				raw = bytes.ReplaceAll(raw, []byte(`"page": 3`), []byte(`"page": 2`))
				raw = bytes.ReplaceAll(raw, []byte(`"pageNr": 3`), []byte(`"pageNr": 2`))
				_ = os.WriteFile(result.Outputs[0], raw, 0600)
			case "pdf.keywords":
				if !strings.Contains(fmt.Sprint(result.Data), "acceptance") {
					t.Fatal("keyword missing")
				}
			case "pdf.extract-text":
				text, _ := result.Data["text"].(string)
				if !strings.Contains(text, "FinishBit page one") || !strings.Contains(text, "FinishBit page two") {
					t.Fatalf("extracted text: %q", text)
				}
			case "pdf.text-boxes":
				pages, ok := result.Data["pages"].([]textPage)
				if !ok || len(pages) != 2 || len(pages[0].Words) == 0 {
					t.Fatalf("word coordinates: %v", result.Data)
				}
				if pages[0].Words[0].XMax <= pages[0].Words[0].XMin {
					t.Fatal("invalid word coordinates")
				}
			case "pdf.render":
				if len(result.Outputs) != 2 {
					t.Fatal("rendered page count")
				}
				f, err := os.Open(result.Outputs[0])
				if err != nil {
					t.Fatal(err)
				}
				config, err := png.DecodeConfig(f)
				_ = f.Close()
				if err != nil || config.Width != 300 || config.Height != 300 {
					t.Fatalf("render dimensions: %v %v", config, err)
				}
			case "pdf.extract-images":
				if len(result.Outputs) < 1 {
					t.Fatal("no extracted images")
				}
			case "pdf.fonts":
				if !strings.Contains(fmt.Sprint(result.Data), "Helvetica") {
					t.Fatal("font missing")
				}
			}
		})
	}
	t.Run("image-and-pdf-stamps", func(t *testing.T) {
		for _, mode := range []string{"image", "pdf"} {
			asset := picture
			if mode == "pdf" {
				asset = input
			}
			path := filepath.Join(dir, "stamp-"+mode+".pdf")
			result := run(t, "pdf.stamp", []string{input}, map[string]any{"mode": mode, "asset": asset, "output": path})
			if pageCount(t, result.Outputs[0]) != 2 {
				t.Fatal("stamp changed page count")
			}
			run(t, "pdf.validate", result.Outputs, nil)
		}
	})
	t.Run("failure-preserves-files", func(t *testing.T) {
		target := filepath.Join(dir, "preserve.pdf")
		_ = os.WriteFile(target, []byte("original"), 0600)
		cap, _ := registry.Get("pdf.rotate")
		if _, err := cap.Runner.Run(ctx, operation.Request{Inputs: []string{input}, Options: map[string]any{"output": target, "angle": "45", "overwrite": true}}); err == nil {
			t.Fatal("invalid angle succeeded")
		}
		raw, _ := os.ReadFile(target)
		if string(raw) != "original" {
			t.Fatal("original changed")
		}
		corrupt := filepath.Join(dir, "corrupt.pdf")
		_ = os.WriteFile(corrupt, []byte("not PDF"), 0600)
		if _, err := cap.Runner.Run(ctx, operation.Request{Inputs: []string{corrupt}, Options: map[string]any{"output": target, "overwrite": true}}); err == nil {
			t.Fatal("corrupt input succeeded")
		}
		raw, _ = os.ReadFile(target)
		if string(raw) != "original" {
			t.Fatal("failure damaged output")
		}
	})
}
