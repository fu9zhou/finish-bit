// Package pdf exposes local PDF editing and reading through managed providers.
package pdf

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/fu9zhou/finish-bit/internal/toolrun"
	"github.com/fu9zhou/finish-bit/pkg/operation"
)

type Provider struct{ resolver toolrun.Resolver }
type spec struct {
	id, summary, alias string
	options            []operation.Parameter
	inputs             int
	mode               string
}

func opt(name, description, value string) operation.Parameter {
	return toolrun.Option(name, operation.TypeString, description, value)
}
func number(name, description string, value int) operation.Parameter {
	return toolrun.Option(name, operation.TypeInteger, description, value)
}
func list(name, description string) operation.Parameter {
	return toolrun.Option(name, operation.TypeStrings, description, nil)
}
func pageOption() operation.Parameter {
	return opt("pages", "One-based pages/ranges, odd/even; empty means all", "")
}
func stampOptions() []operation.Parameter {
	return []operation.Parameter{pageOption(), opt("mode", "text, image or pdf", "text"), opt("text", "Text for text mode", "FinishBit"), opt("asset", "Local image/PDF for image/pdf mode", ""), number("size", "Text font size in points", 24), number("opacity", "Opacity percentage", 60), opt("position", "tl, tc, tr, l, c, r, bl, bc or br", "c")}
}
func catalog() []spec {
	return []spec{
		{"pdf.info", "Inspect PDF structure and page information", "PDF 信息", nil, 1, "inspect"},
		{"pdf.validate", "Validate PDF structure in strict or relaxed mode", "校验 PDF", []operation.Parameter{opt("mode", "strict or relaxed", "relaxed")}, 1, "inspect"},
		{"pdf.merge", "Merge local PDFs in explicit order", "合并 PDF", []operation.Parameter{list("files", "Additional PDF files in order")}, 1, "file"},
		{"pdf.split", "Split PDF into groups of pages", "拆分 PDF", []operation.Parameter{number("span", "Pages per output file", 1)}, 1, "directory"},
		{"pdf.select-pages", "Extract or reorder a specified page sequence", "提取重排 PDF 页面", []operation.Parameter{pageOption()}, 1, "file"},
		{"pdf.remove-pages", "Remove selected PDF pages", "删除 PDF 页面", []operation.Parameter{pageOption()}, 1, "file"},
		{"pdf.insert-pages", "Insert blank pages before or after selected pages", "插入 PDF 空白页", []operation.Parameter{pageOption(), opt("position", "before or after", "before")}, 1, "file"},
		{"pdf.rotate", "Rotate selected PDF pages clockwise", "旋转 PDF", []operation.Parameter{pageOption(), opt("angle", "90, 180 or 270", "90")}, 1, "file"},
		{"pdf.crop", "Set page crop boxes using independent margins in points", "裁剪 PDF", []operation.Parameter{pageOption(), number("margin", "Default margin on all sides in PDF points", 10), number("top", "Top margin; -1 uses margin", -1), number("right", "Right margin; -1 uses margin", -1), number("bottom", "Bottom margin; -1 uses margin", -1), number("left", "Left margin; -1 uses margin", -1)}, 1, "file"},
		{"pdf.watermark", "Add a text, image or PDF watermark below page content", "PDF 水印", stampOptions(), 1, "file"},
		{"pdf.stamp", "Add a text, image or PDF stamp above page content", "PDF 盖章", stampOptions(), 1, "file"},
		{"pdf.remove-watermark", "Remove recognized pdfcpu watermarks", "移除 PDF 水印", []operation.Parameter{pageOption()}, 1, "file"},
		{"pdf.remove-stamp", "Remove recognized pdfcpu stamps", "移除 PDF 印章", []operation.Parameter{pageOption()}, 1, "file"},
		{"pdf.from-images", "Create PDF pages from ordered image files", "图片转 PDF", []operation.Parameter{list("files", "Additional image files in order"), number("dpi", "Image DPI used to size full-image pages", 72)}, 1, "file"},
		{"pdf.resize-pages", "Resize PDF pages to a named paper size", "修改PDF页面尺寸", []operation.Parameter{pageOption(), opt("paper", "A3, A4, A5, Letter or Legal; optional P/L orientation suffix", "A4")}, 1, "file"},
		{"pdf.metadata-set", "Set PDF title, author, subject and creator fields", "修改PDF元数据", []operation.Parameter{opt("title", "Document title", ""), opt("author", "Document author", ""), opt("subject", "Document subject", ""), opt("creator", "Content creator", "")}, 1, "file"},
		{"pdf.page-numbers", "Stamp page numbers and total page count", "PDF加页码", []operation.Parameter{pageOption(), opt("text", "ASCII template; %p current page, %P total; %p3 adds offset 3", "%p / %P"), number("size", "Font size in points", 12), opt("position", "tl, tc, tr, l, c, r, bl, bc or br", "bc")}, 1, "file"},
		{"pdf.sign-image", "Stamp a supplied signature image on selected PDF pages", "PDF签名图片", []operation.Parameter{pageOption(), toolrun.Param("signature", "Local signature PNG/JPEG", true), number("scale-percent", "Relative image scale percentage", 25), number("x", "Horizontal offset in points", 0), number("y", "Vertical offset in points", 0), opt("position", "tl, tc, tr, l, c, r, bl, bc or br", "br")}, 1, "file"},
		{"pdf.optimize", "Remove redundant PDF resources", "优化 PDF", nil, 1, "file"},
		{"pdf.encrypt", "Encrypt a PDF with AES-256 and password permissions", "PDF 加密", []operation.Parameter{toolrun.Param("owner-password", "Owner password", true), toolrun.Param("user-password", "User password", true), opt("permissions", "none, print or all", "none")}, 1, "file"},
		{"pdf.decrypt", "Decrypt a PDF using a supplied password", "PDF 解密", []operation.Parameter{toolrun.Param("password", "User or owner password", true)}, 1, "file"},
		{"pdf.attachments", "List embedded PDF attachments", "PDF 附件列表", nil, 1, "inspect"},
		{"pdf.attach", "Add local attachments to a copy of a PDF", "添加 PDF 附件", []operation.Parameter{list("files", "Local files to attach")}, 1, "file"},
		{"pdf.extract-attachments", "Extract embedded attachments into a new directory", "提取 PDF 附件", nil, 1, "directory"},
		{"pdf.remove-attachments", "Remove selected or all embedded attachments", "删除 PDF 附件", []operation.Parameter{list("names", "Attachment names; empty means all")}, 1, "file"},
		{"pdf.form-fields", "Read structured form fields and values", "读取 PDF 表单", nil, 1, "json"},
		{"pdf.form-export", "Export form data in pdfcpu JSON format", "导出 PDF 表单", nil, 1, "json-file"},
		{"pdf.form-fill", "Fill a PDF form from pdfcpu JSON data", "填写 PDF 表单", nil, 2, "file"},
		{"pdf.form-multifill", "Fill multiple form instances from JSON or CSV", "批量填写 PDF 表单", nil, 2, "directory"},
		{"pdf.form-reset", "Reset selected or all PDF form fields", "重置 PDF 表单", []operation.Parameter{list("fields", "Field names or IDs; empty means all")}, 1, "file"},
		{"pdf.form-lock", "Make selected or all form fields read-only", "锁定 PDF 表单", []operation.Parameter{list("fields", "Field names or IDs; empty means all")}, 1, "file"},
		{"pdf.form-unlock", "Unlock selected or all form fields", "解锁 PDF 表单", []operation.Parameter{list("fields", "Field names or IDs; empty means all")}, 1, "file"},
		{"pdf.bookmarks-export", "Export PDF bookmarks as JSON", "导出 PDF 书签", nil, 1, "json-file"},
		{"pdf.bookmarks-import", "Replace PDF bookmarks from pdfcpu JSON", "导入 PDF 书签", nil, 2, "file"},
		{"pdf.bookmarks-remove", "Remove PDF bookmarks", "删除 PDF 书签", nil, 1, "file"},
		{"pdf.keywords", "List PDF keywords", "PDF 关键词", nil, 1, "inspect"},
		{"pdf.keywords-add", "Add PDF keywords", "添加 PDF 关键词", []operation.Parameter{list("values", "Keywords to add")}, 1, "file"},
		{"pdf.keywords-remove", "Remove selected or all PDF keywords", "删除 PDF 关键词", []operation.Parameter{list("values", "Keywords; empty means all")}, 1, "file"},
		{"pdf.nup", "Arrange multiple PDF pages on each output sheet", "PDF 多页合一", []operation.Parameter{number("count", "Pages per sheet: 2, 3, 4, 8, 9, 12, 16", 4)}, 1, "file"},
	}
}

func Register(registry *operation.Registry, resolver toolrun.Resolver) error {
	p := &Provider{resolver: resolver}
	for _, item := range catalog() {
		def := operation.Definition{ID: item.id, Summary: item.summary, Description: item.summary + ". Uses managed pdfcpu offline with a disabled user config; modifies only staged copies. See docs/pdf.md.", Aliases: []string{item.alias}, Tags: []string{"pdf", "document"}, Options: append([]operation.Parameter{}, item.options...), Requirements: []operation.Requirement{{Package: "pdfcpu"}}, Source: "pdfcpu"}
		for i := 0; i < item.inputs; i++ {
			name := "input"
			if i == 1 {
				name = "data"
			}
			def.Inputs = append(def.Inputs, toolrun.Param(name, "Local input file", true))
		}
		switch item.mode {
		case "directory":
			def.Options = append(def.Options, toolrun.Param("output", "New output directory", true))
		case "file", "json-file":
			def.Options = append(def.Options, toolrun.OutputOptions()...)
		}
		if err := registry.Register(operation.Capability{Definition: def, Runner: operation.Func(func(ctx context.Context, r operation.Request) (operation.Result, error) {
			if err := operation.ValidateRequest(def, r); err != nil {
				return operation.Result{}, err
			}
			return p.run(ctx, item, r)
		})}); err != nil {
			return err
		}
	}
	return p.registerReading(registry)
}

var pagePattern = regexp.MustCompile(`^(?:[1-9][0-9]*(?:-[1-9][0-9]*)?|odd|even)(?:,(?:[1-9][0-9]*(?:-[1-9][0-9]*)?|odd|even))*$`)

func pageFlags(v *toolrun.Values, required bool) []string {
	pages := v.String("pages", "")
	if pages == "" {
		v.Check(!required, "pages is required")
		return nil
	}
	v.Check(len(pages) <= 4096 && pagePattern.MatchString(pages), "pages must contain positive numbers, ranges, odd or even")
	return []string{"--pages", pages}
}
func (p *Provider) command(ctx context.Context, args ...string) ([]byte, error) {
	return toolrun.Run(ctx, p.resolver, "pdfcpu", "pdfcpu", "", append([]string{"--conf", "disable", "--offline", "--force"}, args...)...)
}

func (p *Provider) run(ctx context.Context, item spec, r operation.Request) (operation.Result, error) {
	v := &toolrun.Values{Request: r}
	paths := []string{}
	for _, path := range r.Inputs {
		paths = append(paths, v.File(path))
	}
	for _, path := range v.Strings("files") {
		paths = append(paths, v.File(path))
	}
	v.Check(len(paths) <= 128, "at most 128 input files are supported")
	output := v.String("output", "")
	overwrite := v.Bool("overwrite", false)
	pages := pageFlags(v, item.id == "pdf.select-pages" || item.id == "pdf.remove-pages" || item.id == "pdf.insert-pages")
	if v.Err != nil {
		return operation.Result{}, v.Err
	}
	input := paths[0]
	result := operation.Result{Operation: item.id}
	if item.mode == "inspect" {
		args := []string{}
		switch item.id {
		case "pdf.info":
			args = []string{"info", "--json", input}
		case "pdf.validate":
			args = []string{"validate", "--mode", v.Enum("mode", "relaxed", "relaxed", "strict"), input}
		case "pdf.attachments":
			args = []string{"attachments", "list", input}
		case "pdf.keywords":
			args = []string{"keywords", "list", input}
		}
		if v.Err != nil {
			return result, v.Err
		}
		raw, err := p.command(ctx, args...)
		if err != nil {
			return result, err
		}
		if item.id == "pdf.info" {
			if err := json.Unmarshal(raw, &result.Data); err != nil {
				return result, fmt.Errorf("decode PDF info: %w", err)
			}
		} else {
			result.Data = map[string]any{"report": strings.TrimSpace(string(raw))}
			if item.id == "pdf.validate" {
				result.Data["valid"] = true
			}
		}
		return result, nil
	}
	if item.mode == "json" {
		dir, err := os.MkdirTemp("", "finishbit-pdf-form-")
		if err != nil {
			return result, err
		}
		defer os.RemoveAll(dir)
		path := filepath.Join(dir, "form.json")
		if _, err := p.command(ctx, "form", "export", input, path); err != nil {
			return result, err
		}
		raw, err := readSmall(path)
		if err != nil {
			return result, err
		}
		if err := json.Unmarshal(raw, &result.Data); err != nil {
			return result, err
		}
		return result, nil
	}
	if item.mode == "directory" {
		span := 1
		if item.id == "pdf.split" {
			span = v.Int("span", 1, 1, 10000)
		}
		if v.Err != nil {
			return result, v.Err
		}
		outputs, err := toolrun.Directory(ctx, output, func(dir string) error {
			args := []string{"attachments", "extract", input, dir}
			if item.id == "pdf.form-multifill" {
				args = []string{"form", "multifill", input, paths[1], dir}
			}
			if item.id == "pdf.split" {
				args = []string{"split", input, dir, strconv.Itoa(span)}
			}
			_, err := p.command(ctx, args...)
			return err
		})
		result.Outputs = outputs
		return result, err
	}
	extension := ".pdf"
	if item.mode == "json-file" {
		extension = ".json"
	}
	v.Check(strings.EqualFold(filepath.Ext(output), extension), "output extension must be "+extension)
	args := []string{}
	copyInput := false
	// The placeholder is replaced only in argument positions selected by the builder.
	const target = "{finishbit-output}"
	switch item.id {
	case "pdf.merge":
		v.Check(len(paths) >= 2, "at least one additional files entry is required")
		args = append([]string{"merge", "--bookmarks", target}, paths...)
	case "pdf.select-pages":
		args = append(append([]string{"collect"}, pages...), input, target)
	case "pdf.remove-pages":
		args = append(append([]string{"pages", "remove"}, pages...), input, target)
	case "pdf.insert-pages":
		args = append([]string{"pages", "insert", "--mode", v.Enum("position", "before", "before", "after")}, pages...)
		args = append(args, input, target)
	case "pdf.rotate":
		angle := v.Enum("angle", "90", "90", "180", "270")
		args = append(append([]string{"rotate"}, pages...), input, angle, target)
	case "pdf.crop":
		margin := v.Int("margin", 10, 0, 10000)
		margins := []string{}
		for _, side := range []string{"top", "right", "bottom", "left"} {
			n := v.Int(side, -1, -1, 10000)
			if n == -1 {
				n = margin
			}
			margins = append(margins, strconv.Itoa(n))
		}
		args = append(append([]string{"crop"}, pages...), strings.Join(margins, " ")+" abs", input, target)
	case "pdf.watermark", "pdf.stamp":
		command := "watermark"
		if item.id == "pdf.stamp" {
			command = "stamp"
		}
		mode := v.Enum("mode", "text", "text", "image", "pdf")
		text := v.String("text", "FinishBit")
		if mode == "text" {
			v.Check(text != "" && len(text) <= 4096, "text must contain 1 to 4096 bytes")
		} else {
			text = v.File(v.String("asset", ""))
		}
		size := v.Int("size", 24, 1, 500)
		opacity := v.Int("opacity", 60, 1, 100)
		position := v.Enum("position", "c", "tl", "tc", "tr", "l", "c", "r", "bl", "bc", "br")
		description := fmt.Sprintf("rotation:0,opacity:%.2f,position:%s", float64(opacity)/100, position)
		if mode == "text" {
			description += fmt.Sprintf(",font:Helvetica,points:%d,scale:1 abs", size)
		}
		args = append([]string{command, "add", "--mode", mode}, pages...)
		args = append(args, "--", text, description, input, target)
	case "pdf.remove-watermark", "pdf.remove-stamp":
		command := "watermark"
		if item.id == "pdf.remove-stamp" {
			command = "stamp"
		}
		args = append(append([]string{command, "remove"}, pages...), input, target)
	case "pdf.from-images":
		dpi := v.Int("dpi", 72, 36, 1200)
		args = append([]string{"import", fmt.Sprintf("pos:full,dpi:%d", dpi), target}, paths...)
	case "pdf.resize-pages":
		paper := v.Enum("paper", "A4", "A3", "A4", "A5", "Letter", "Legal", "A3P", "A4P", "A5P", "LetterP", "LegalP", "A3L", "A4L", "A5L", "LetterL", "LegalL")
		args = append(append([]string{"resize"}, pages...), "form:"+paper, input, target)
	case "pdf.metadata-set":
		args = []string{"properties", "add", input, target}
		count := 0
		for _, entry := range []struct{ key, name string }{{"title", "Title"}, {"author", "Author"}, {"subject", "Subject"}, {"creator", "Creator"}} {
			if _, present := r.Options[entry.key]; present {
				val := v.String(entry.key, "")
				v.Check(len(val) <= 4096 && !strings.ContainsRune(val, '\x00'), "metadata value exceeds 4096 bytes or contains NUL")
				args = append(args, entry.name+" = "+val)
				count++
			}
		}
		v.Check(count > 0, "provide at least one metadata field")
	case "pdf.page-numbers":
		text := v.String("text", "%p / %P")
		v.Check(len(text) > 0 && len(text) <= 256, "page-number template must be 1 to 256 bytes")
		for _, r := range text {
			v.Check(r >= 32 && r <= 126, "page-number template must be printable ASCII")
		}
		size := v.Int("size", 12, 1, 100)
		position := v.Enum("position", "bc", "tl", "tc", "tr", "l", "c", "r", "bl", "bc", "br")
		args = append(append([]string{"stamp", "add", "--mode", "text"}, pages...), "--", text, fmt.Sprintf("font:Helvetica,points:%d,scale:1 abs,rot:0,pos:%s", size, position), input, target)
	case "pdf.sign-image":
		signature := v.File(v.String("signature", ""))
		ext := strings.ToLower(filepath.Ext(signature))
		v.Check(ext == ".png" || ext == ".jpg" || ext == ".jpeg", "signature must be PNG/JPEG")
		scale := v.Int("scale-percent", 25, 1, 100)
		x, y := v.Int("x", 0, -10000, 10000), v.Int("y", 0, -10000, 10000)
		position := v.Enum("position", "br", "tl", "tc", "tr", "l", "c", "r", "bl", "bc", "br")
		args = append(append([]string{"stamp", "add", "--mode", "image"}, pages...), "--", signature, fmt.Sprintf("scale:%.2f rel,rot:0,pos:%s,off:%d %d", float64(scale)/100, position, x, y), input, target)
	case "pdf.optimize":
		args = []string{"optimize", input, target}
	case "pdf.encrypt":
		owner := v.String("owner-password", "")
		user := v.String("user-password", "")
		v.Check(owner != "" && user != "" && len(owner) <= 127 && len(user) <= 127, "passwords must contain 1 to 127 bytes")
		args = []string{"encrypt", "--mode", "aes", "--key", "256", "--opw", owner, "--upw", user, "--perm", v.Enum("permissions", "none", "none", "print", "all"), input, target}
	case "pdf.decrypt":
		password := v.String("password", "")
		v.Check(password != "", "password is required")
		args = []string{"decrypt", "--upw", password, input, target}
	case "pdf.attach":
		v.Check(len(paths) >= 2, "files must contain at least one attachment")
		copyInput = true
		args = append([]string{"attachments", "add", target}, paths[1:]...)
	case "pdf.remove-attachments":
		copyInput = true
		args = append([]string{"attachments", "remove", target, "--"}, v.Strings("names")...)
	case "pdf.form-export":
		args = []string{"form", "export", input, target}
	case "pdf.form-fill":
		args = []string{"form", "fill", input, paths[1], target}
	case "pdf.form-reset", "pdf.form-lock", "pdf.form-unlock":
		action := strings.TrimPrefix(item.id, "pdf.form-")
		args = append([]string{"form", action, input, target, "--"}, v.Strings("fields")...)
	case "pdf.bookmarks-export":
		args = []string{"bookmarks", "export", input, target}
	case "pdf.bookmarks-import":
		args = []string{"bookmarks", "import", "--replace", input, paths[1], target}
	case "pdf.bookmarks-remove":
		args = []string{"bookmarks", "remove", input, target}
	case "pdf.keywords-add":
		values := v.Strings("values")
		v.Check(len(values) > 0, "values must contain keywords")
		args = append([]string{"keywords", "add", input, target, "--"}, values...)
	case "pdf.keywords-remove":
		args = append([]string{"keywords", "remove", input, target, "--"}, v.Strings("values")...)
	case "pdf.nup":
		count := v.Int("count", 4, 2, 16)
		v.Check(count == 2 || count == 3 || count == 4 || count == 8 || count == 9 || count == 12 || count == 16, "unsupported pages-per-sheet count")
		args = []string{"nup", target, strconv.Itoa(count), input}
	default:
		return result, toolrun.Invalid("unsupported PDF operation")
	}
	if v.Err != nil {
		return result, v.Err
	}
	if err := toolrun.Artifact(ctx, output, overwrite, func(path string) error {
		if copyInput {
			if err := toolrun.CopyFile(input, path); err != nil {
				return err
			}
		}
		actual := append([]string{}, args...)
		for i, a := range actual {
			if a == target {
				actual[i] = path
			}
		}
		if item.id == "pdf.attach" {
			// pdfcpu expands attachment arguments as globs. A private one-file
			// directory preserves the original basename without interpreting it.
			actual = []string{"attachments", "add", path}
			for i, source := range paths[1:] {
				name := fmt.Sprintf("attachment-%03d", i)
				dir := filepath.Join(filepath.Dir(path), name)
				if err := os.Mkdir(dir, 0700); err != nil {
					return err
				}
				if err := toolrun.CopyFile(source, filepath.Join(dir, filepath.Base(source))); err != nil {
					return err
				}
				actual = append(actual, name+"/*")
			}
			_, err := toolrun.Run(ctx, p.resolver, "pdfcpu", "pdfcpu", filepath.Dir(path), append([]string{"--conf", "disable", "--offline", "--force"}, actual...)...)
			return err
		}
		_, err := p.command(ctx, actual...)
		return err
	}); err != nil {
		return result, err
	}
	result.Outputs = []string{output}
	return result, nil
}
func readSmall(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.Size() > 16<<20 {
		return nil, toolrun.Invalid("result exceeds 16 MiB")
	}
	return os.ReadFile(path)
}
