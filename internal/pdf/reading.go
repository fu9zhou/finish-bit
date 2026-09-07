package pdf

import (
	"context"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/fu9zhou/finish-bit/internal/toolrun"
	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func readingCatalog() []spec {
	rangeOptions := []operation.Parameter{number("first", "First page, one-based", 1), number("last", "Last page, one-based (0 means all for text)", 0)}
	textOptions := append(append([]operation.Parameter{}, rangeOptions...),
		toolrun.Option("layout", operation.TypeBoolean, "Preserve physical text layout", false),
		toolrun.Option("unwrap", operation.TypeBoolean, "Join visual line breaks within detected paragraphs", false),
	)
	return []spec{
		{"pdf.extract-text", "Extract existing text from PDF pages", "提取 PDF 文字", textOptions, 1, "text"},
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
		} else {
			layout := v.Bool("layout", false)
			unwrap := v.Bool("unwrap", false)
			v.Check(!(layout && unwrap), "layout and unwrap cannot be used together")
			if layout {
				args = append(args, "-layout")
			}
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
		if v.Bool("unwrap", false) {
			raw = []byte(unwrapExtractedText(string(raw)))
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

func unwrapExtractedText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	trailingNewline := strings.HasSuffix(text, "\n")
	pages := strings.Split(text, "\f")
	for i, page := range pages {
		lines := strings.Split(page, "\n")
		out := make([]string, 0, len(lines))
		block := make([]string, 0, 8)
		flush := func() {
			if len(block) == 0 {
				return
			}
			out = append(out, unwrapTextBlock(block)...)
			block = block[:0]
		}
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				flush()
				if len(out) > 0 && out[len(out)-1] != "" {
					out = append(out, "")
				}
				continue
			}
			block = append(block, line)
		}
		flush()
		out = joinContinuationBlocks(out)
		for len(out) > 0 && out[len(out)-1] == "" {
			out = out[:len(out)-1]
		}
		pages[i] = strings.Join(out, "\n")
	}
	result := strings.Join(pages, "\n\f\n")
	if trailingNewline && !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	return result
}

func joinContinuationBlocks(lines []string) []string {
	for i := 1; i+1 < len(lines); {
		if lines[i] != "" || !startsWithLowercase(lines[i+1]) || isPDFListLine(lines[i+1]) {
			i++
			continue
		}
		previous, next := lines[i-1], lines[i+1]
		if previous == "" || isPDFListLine(previous) || !canContinuePDFSentence(previous) {
			i++
			continue
		}
		separator := " "
		if strings.HasSuffix(previous, "-") {
			separator = ""
		}
		lines[i-1] = previous + separator + next
		lines = append(lines[:i], lines[i+2:]...)
	}
	return lines
}

func canContinuePDFSentence(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" {
		return false
	}
	for _, ending := range []string{".", "!", "?", ":", ";", "。", "！", "？", "：", "；"} {
		if strings.HasSuffix(line, ending) {
			return false
		}
	}
	return true
}

func startsWithLowercase(line string) bool {
	for _, r := range strings.TrimSpace(line) {
		if unicode.IsLetter(r) {
			return unicode.IsLower(r)
		}
		if unicode.IsDigit(r) {
			return false
		}
	}
	return false
}

func unwrapTextBlock(lines []string) []string {
	if len(lines) < 2 {
		return append([]string(nil), lines...)
	}
	lengths := make([]int, 0, len(lines))
	for _, line := range lines {
		if !isPDFListLine(line) && !isPDFHeading(line) {
			lengths = append(lengths, len([]rune(line)))
		}
	}
	sort.Ints(lengths)
	typicalLength := 0
	if len(lengths) > 0 {
		typicalLength = lengths[len(lengths)/2]
	}
	out := []string{lines[0]}
	for i := 1; i < len(lines); i++ {
		current := lines[i]
		previous := out[len(out)-1]
		physicalPrevious := lines[i-1]
		paragraphEnd := typicalLength >= 40 && len([]rune(physicalPrevious))*4 < typicalLength*3 && endsPDFSentence(physicalPrevious)
		if isPDFListLine(previous) || isPDFListLine(current) || isPDFHeading(previous) || isPDFHeading(current) || paragraphEnd {
			out = append(out, current)
			continue
		}
		if strings.HasSuffix(previous, "-") {
			out[len(out)-1] = previous + current
		} else {
			out[len(out)-1] = previous + " " + current
		}
	}
	return out
}

func isPDFListLine(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" {
		return false
	}
	for _, prefix := range []string{"• ", "- ", "* ", "· "} {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	first := strings.IndexByte(line, '.')
	if first < 1 || first > 3 || first+1 >= len(line) || line[first+1] != ' ' {
		return false
	}
	for _, r := range line[:first] {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func endsPDFSentence(line string) bool {
	line = strings.TrimRightFunc(strings.TrimSpace(line), func(r rune) bool {
		return strings.ContainsRune("\"')]}”’", r)
	})
	for _, ending := range []string{".", "!", "?", "。", "！", "？"} {
		if strings.HasSuffix(line, ending) {
			return true
		}
	}
	return false
}

func isPDFHeading(line string) bool {
	letters, upper := 0, 0
	for _, r := range line {
		if unicode.IsLetter(r) {
			letters++
			if unicode.IsUpper(r) {
				upper++
			}
		}
	}
	if letters > 0 && letters == upper {
		return true
	}
	if len([]rune(line)) > 80 || len(strings.Fields(line)) > 10 {
		return false
	}
	if strings.ContainsAny(line, ".!?;:。！？；：") {
		return false
	}
	contentWords, titleWords := 0, 0
	stopWords := map[string]bool{"a": true, "an": true, "and": true, "for": true, "in": true, "of": true, "on": true, "or": true, "the": true, "to": true, "with": true}
	for _, field := range strings.Fields(line) {
		word := strings.TrimFunc(field, func(r rune) bool { return !unicode.IsLetter(r) })
		if word == "" || stopWords[strings.ToLower(word)] {
			continue
		}
		contentWords++
		first, _ := utf8.DecodeRuneInString(word)
		if unicode.IsUpper(first) {
			titleWords++
		}
	}
	return contentWords > 0 && contentWords == titleWords
}
