package pdf

import (
	"context"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fu9zhou/finish-bit/internal/toolrun"
	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func readingCatalog() []spec {
	rangeOptions := []operation.Parameter{number("first", "First page, one-based", 1), number("last", "Last page, one-based (0 means all for text)", 0)}
	return []spec{
		{"pdf.extract-text", "Extract existing text from PDF pages", "提取 PDF 文字", append(append([]operation.Parameter{}, rangeOptions...), toolrun.Option("layout", operation.TypeBoolean, "Preserve physical text layout", false)), 1, "text"},
		{"pdf.text-boxes", "Extract words and their page coordinates", "PDF 文字坐标", rangeOptions, 1, "inspect"},
		{"pdf.render", "Render selected PDF pages to PNG or JPEG", "PDF 转图片", []operation.Parameter{number("first", "First page", 1), number("last", "Last page", 1), number("dpi", "Rendering resolution, 36 to 300", 120), opt("format", "png or jpeg", "png")}, 1, "directory"},
		{"pdf.extract-images", "Extract images embedded in selected PDF pages", "提取 PDF 图片", []operation.Parameter{number("first", "First page", 1), number("last", "Last page", 1)}, 1, "directory"},
		{"pdf.fonts", "List PDF fonts and embedding information", "PDF 字体列表", nil, 1, "inspect"},
		{"pdf.to-html", "Convert selected PDF pages to an HTML bundle", "PDF 转 HTML", []operation.Parameter{number("first", "First page", 1), number("last", "Last page", 1)}, 1, "directory"},
		{"pdf.to-ps", "Convert selected PDF pages to PostScript", "PDF 转 PostScript", rangeOptions, 1, "file"},
	}
}
func (p *Provider) registerReading(registry *operation.Registry) error {
	for _, item := range readingCatalog() {
		def := operation.Definition{ID: item.id, Summary: item.summary, Description: item.summary + ". Local, managed Poppler; existing text only, no OCR. See docs/pdf.md.", Aliases: []string{item.alias}, Tags: []string{"pdf", "document"}, Inputs: []operation.Parameter{toolrun.Param("input", "Local PDF file", true)}, Options: append([]operation.Parameter{}, item.options...), Requirements: []operation.Requirement{{Package: "poppler"}}, Source: "poppler"}
		switch item.mode {
		case "directory":
			def.Options = append(def.Options, toolrun.Param("output", "New output directory", true))
		case "file":
			def.Options = append(def.Options, toolrun.OutputOptions()...)
		case "text":
			def.Options = append(def.Options, opt("output", "Optional UTF-8 destination file", ""), toolrun.Option("overwrite", operation.TypeBoolean, "Replace existing destination", false))
		}
		if err := registry.Register(operation.Capability{Definition: def, Runner: operation.Func(func(ctx context.Context, r operation.Request) (operation.Result, error) {
			if err := operation.ValidateRequest(def, r); err != nil {
				return operation.Result{}, err
			}
			return p.read(ctx, item, r)
		})}); err != nil {
			return err
		}
	}
	return nil
}

type word struct {
	Text string  `xml:",chardata" json:"text"`
	XMin float64 `xml:"xMin,attr" json:"x_min"`
	YMin float64 `xml:"yMin,attr" json:"y_min"`
	XMax float64 `xml:"xMax,attr" json:"x_max"`
	YMax float64 `xml:"yMax,attr" json:"y_max"`
}
type textPage struct {
	Number int     `json:"number"`
	Width  float64 `xml:"width,attr" json:"width"`
	Height float64 `xml:"height,attr" json:"height"`
	Words  []word  `xml:"word" json:"words"`
}

func (p *Provider) read(ctx context.Context, item spec, r operation.Request) (operation.Result, error) {
	v := &toolrun.Values{Request: r}
	input := v.File(r.Inputs[0])
	output := v.String("output", "")
	overwrite := v.Bool("overwrite", false)
	first, last := 1, 0
	flags := []string{}
	if item.id != "pdf.fonts" {
		first = v.Int("first", 1, 1, 100000)
		fallback := 0
		if item.mode == "directory" {
			fallback = 1
		}
		last = v.Int("last", fallback, 0, 100000)
		v.Check(last == 0 || last >= first, "last page must not precede first")
		flags = []string{"-f", strconv.Itoa(first)}
		if last > 0 {
			flags = append(flags, "-l", strconv.Itoa(last))
		}
		if item.mode == "directory" {
			v.Check(last > 0 && last-first < 200, "directory output supports at most 200 explicitly selected pages")
		}
	}
	result := operation.Result{Operation: item.id}
	if v.Err != nil {
		return result, v.Err
	}
	run := func(name string, args ...string) ([]byte, error) {
		return toolrun.Run(ctx, p.resolver, "poppler", name, "", args...)
	}
	switch item.id {
	case "pdf.extract-text", "pdf.text-boxes":
		args := append([]string{"-enc", "UTF-8"}, flags...)
		if item.id == "pdf.text-boxes" {
			args = append(args, "-bbox")
		} else if v.Bool("layout", false) {
			args = append(args, "-layout")
		}
		if v.Err != nil {
			return result, v.Err
		}
		raw, err := run("pdftotext", append(args, input, "-")...)
		if err != nil {
			return result, err
		}
		if item.id == "pdf.text-boxes" {
			var document struct {
				Pages []textPage `xml:"body>doc>page"`
			}
			if err := xml.Unmarshal(raw, &document); err != nil {
				return result, fmt.Errorf("decode word coordinates: %w", err)
			}
			for i := range document.Pages {
				document.Pages[i].Number = first + i
			}
			result.Data = map[string]any{"pages": document.Pages, "unit": "PDF points", "origin": "top-left"}
			return result, nil
		}
		result.Data = map[string]any{"text": string(raw)}
		if output != "" {
			if err := toolrun.Artifact(ctx, output, overwrite, func(path string) error { return os.WriteFile(path, raw, 0600) }); err != nil {
				return result, err
			}
			result.Outputs = []string{output}
			result.Data = map[string]any{"bytes": len(raw)}
		}
		return result, nil
	case "pdf.fonts":
		raw, err := run("pdffonts", input)
		if err != nil {
			return result, err
		}
		result.Data = map[string]any{"report": string(raw)}
		return result, nil
	case "pdf.to-ps":
		v.Check(strings.EqualFold(filepath.Ext(output), ".ps"), "output extension must be .ps")
		if v.Err != nil {
			return result, v.Err
		}
		err := toolrun.Artifact(ctx, output, overwrite, func(path string) error { _, err := run("pdftops", append(flags, input, path)...); return err })
		result.Outputs = []string{output}
		return result, err
	default:
		name := ""
		args := append([]string{}, flags...)
		switch item.id {
		case "pdf.render":
			name = "pdftoppm"
			format := v.Enum("format", "png", "png", "jpeg")
			dpi := v.Int("dpi", 120, 36, 300)
			args = append(args, "-r", strconv.Itoa(dpi), "-"+format)
		case "pdf.extract-images":
			name = "pdfimages"
			args = append(args, "-all")
		case "pdf.to-html":
			name = "pdftohtml"
			args = append(args, "-c", "-s", "-enc", "UTF-8")
		}
		if v.Err != nil {
			return result, v.Err
		}
		outputs, err := toolrun.Directory(ctx, output, func(dir string) error {
			// Some Windows Poppler utilities use narrow output-path APIs.
			// Set a Unicode-aware process working directory and use a relative prefix.
			_, err := toolrun.Run(ctx, p.resolver, "poppler", name, dir, append(args, input, "page")...)
			return err
		})
		result.Outputs = outputs
		return result, err
	}
}
