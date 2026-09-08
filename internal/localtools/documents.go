package localtools

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/fu9zhou/finish-bit/internal/toolrun"
	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func documentSpecs() []spec {
	return []spec{
		{id: "worksheet.handwriting", summary: "Generate an offline printable Chinese handwriting worksheet", alias: "字帖生成", inputs: in("text"), options: []operation.Parameter{integer("columns", "Cells per row", 12), integer("repeat", "Rows per character group", 3), str("grid", "square, cross or rice", "rice"), str("title", "Worksheet title", "练字字帖")}, run: func(ctx context.Context, v *toolrun.Values, a []string) (map[string]any, error) {
			cols := v.Int("columns", 12, 4, 12)
			repeat := v.Int("repeat", 3, 1, 10)
			grid := v.Enum("grid", "rice", "square", "cross", "rice")
			title := v.String("title", "练字字帖")
			runes := []rune(strings.Join(strings.Fields(a[0]), ""))
			if len(runes) == 0 || len(runes) > 300 {
				return nil, invalid("provide 1 to 300 characters")
			}
			if v.Err != nil {
				return nil, v.Err
			}
			var b strings.Builder
			b.WriteString("<!doctype html><html lang=\"zh-CN\"><head><meta charset=\"utf-8\"><title>" + escapeHTML(title) + "</title><style>@page{size:A4;margin:15mm}body{font-family:serif}h1{font-size:20px}.row{display:flex;break-inside:avoid;margin-bottom:3mm}.cell{position:relative;box-sizing:border-box;border:1px solid #b57b7b;width:14mm;height:14mm;text-align:center;line-height:14mm;font-size:10mm}.trace{color:#ccc}.cross:before{content:'';position:absolute;inset:0;background:linear-gradient(transparent 49%,#ddd 50%,transparent 51%),linear-gradient(90deg,transparent 49%,#ddd 50%,transparent 51%)}.rice:after{content:'';position:absolute;inset:0;background:linear-gradient(45deg,transparent 49%,#eee 50%,transparent 51%),linear-gradient(-45deg,transparent 49%,#eee 50%,transparent 51%)}</style></head><body><h1>" + escapeHTML(title) + "</h1>")
			for start := 0; start < len(runes); start += cols {
				for row := 0; row < repeat; row++ {
					b.WriteString("<div class=\"row\">")
					for col := 0; col < cols; col++ {
						class := "cell"
						if grid != "square" {
							class += " cross"
						}
						if grid == "rice" {
							class += " rice"
						}
						if row > 0 {
							class += " trace"
						}
						text := ""
						if start+col < len(runes) {
							text = string(runes[start+col])
						}
						b.WriteString("<div class=\"" + class + "\">" + escapeHTML(text) + "</div>")
					}
					b.WriteString("</div>")
				}
			}
			b.WriteString("</body></html>")
			return value(b.String()), nil
		}},
		{id: "text.led", summary: "Generate an offline fullscreen scrolling text page", alias: "手持弹幕LED", inputs: in("text"), options: []operation.Parameter{integer("seconds", "Seconds per scroll", 10), str("color", "Text #RRGGBB", "#00FF66"), integer("size", "Font size in viewport-height percent", 30)}, run: func(ctx context.Context, v *toolrun.Values, a []string) (map[string]any, error) {
			seconds := v.Int("seconds", 10, 1, 300)
			size := v.Int("size", 30, 1, 80)
			color := v.String("color", "#00FF66")
			if _, e := parseRGB(color); e != nil {
				return nil, e
			}
			if len(a[0]) > 4096 {
				return nil, invalid("LED text exceeds 4096 bytes")
			}
			if v.Err != nil {
				return nil, v.Err
			}
			return value(fmt.Sprintf("<!doctype html><html><head><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width\"><title>LED</title><style>body{margin:0;background:#000;color:%s;overflow:hidden;display:flex;align-items:center;height:100vh}div{font:bold %dvh sans-serif;white-space:nowrap;animation:scroll %ds linear infinite}body:hover div{animation-play-state:paused}@keyframes scroll{from{transform:translateX(100vw)}to{transform:translateX(-100%%)}}@media(prefers-reduced-motion:reduce){div{animation:none;white-space:normal;font-size:10vh}}</style></head><body><div>%s</div></body></html>", color, size, seconds, escapeHTML(a[0]))), nil
		}},
		{id: "markdown.mindmap", summary: "Generate an offline collapsible outline from Markdown headings and lists", alias: "便捷思维导图", inputs: in("markdown"), run: func(ctx context.Context, v *toolrun.Values, a []string) (map[string]any, error) {
			if len(a[0]) > 1<<20 {
				return nil, invalid("outline input exceeds 1 MiB")
			}
			type node struct {
				text     string
				level    int
				children []*node
			}
			root := &node{level: -1}
			stack := []*node{root}
			count := 0
			for _, line := range strings.Split(a[0], "\n") {
				trim := strings.TrimSpace(line)
				if trim == "" {
					continue
				}
				level := 0
				label := ""
				if strings.HasPrefix(trim, "#") {
					marks := len(trim) - len(strings.TrimLeft(trim, "#"))
					if marks <= 6 && len(trim) > marks && trim[marks] == ' ' {
						level = marks - 1
						label = strings.TrimSpace(trim[marks:])
					}
				} else if strings.HasPrefix(trim, "- ") || strings.HasPrefix(trim, "* ") {
					level = 6 + (len(line)-len(strings.TrimLeft(line, " \t")))/2
					label = trim[2:]
				}
				if label == "" {
					continue
				}
				if level > 64 {
					return nil, invalid("outline depth exceeds 64")
				}
				count++
				if count > 2000 {
					return nil, invalid("outline exceeds 2000 nodes")
				}
				for len(stack) > 1 && stack[len(stack)-1].level >= level {
					stack = stack[:len(stack)-1]
				}
				n := &node{text: label, level: level}
				parent := stack[len(stack)-1]
				parent.children = append(parent.children, n)
				stack = append(stack, n)
			}
			if count == 0 {
				return nil, invalid("no Markdown headings or bullet items found")
			}
			var b strings.Builder
			b.WriteString("<!doctype html><html><head><meta charset=\"utf-8\"><title>Outline</title><style>body{font:18px/1.6 system-ui;margin:2em}details{margin:.4em 0 .4em 1em;border-left:1px solid #ccc;padding-left:1em}summary{cursor:pointer}p{margin:.25em 1em}</style></head><body>")
			var render func(*node)
			render = func(n *node) {
				for _, c := range n.children {
					if len(c.children) > 0 {
						b.WriteString("<details open><summary>" + escapeHTML(c.text) + "</summary>")
						render(c)
						b.WriteString("</details>")
					} else {
						b.WriteString("<p>" + escapeHTML(c.text) + "</p>")
					}
				}
			}
			render(root)
			b.WriteString("</body></html>")
			return value(b.String()), nil
		}},
	}
}

func registerDocumentCompression(reg *operation.Registry) error {
	def := operation.Definition{ID: "document.compress", Summary: "Recompress OOXML ZIP entries and reduce embedded JPEG/PNG bytes", Description: "Local DOCX/PPTX/XLSX containers, at most 32 MiB compressed and 128 MiB expanded. Preserves member names and relationships, changes only smaller re-encodings, rejects package signatures. No Office rendering or service required.", Aliases: []string{"文档瘦身"}, Tags: []string{"document", "offline", "compression"}, Source: "core", Inputs: []operation.Parameter{toolrun.Param("input", "Local .docx/.pptx/.xlsx", true)}, Options: append([]operation.Parameter{integer("jpeg-quality", "JPEG re-encoding quality", 80)}, toolrun.OutputOptions()...)}
	return reg.Register(operation.Capability{Definition: def, Runner: operation.Func(func(ctx context.Context, r operation.Request) (operation.Result, error) {
		if e := operation.ValidateRequest(def, r); e != nil {
			return operation.Result{}, e
		}
		v := &toolrun.Values{Request: r}
		quality := v.Int("jpeg-quality", 80, 1, 100)
		input := v.File(r.Inputs[0])
		out := v.String("output", "")
		overwrite := v.Bool("overwrite", false)
		if v.Err != nil {
			return operation.Result{}, v.Err
		}
		ext := strings.ToLower(filepath.Ext(input))
		if (ext != ".docx" && ext != ".pptx" && ext != ".xlsx") || !strings.EqualFold(filepath.Ext(out), ext) {
			return operation.Result{}, invalid("matching .docx/.pptx/.xlsx extensions required")
		}
		info, e := os.Stat(input)
		if e != nil {
			return operation.Result{}, e
		}
		if info.Size() > 32<<20 {
			return operation.Result{}, invalid("OOXML input exceeds 32 MiB")
		}
		z, e := zip.OpenReader(input)
		if e != nil {
			return operation.Result{}, invalid("invalid OOXML ZIP")
		}
		defer z.Close()
		if len(z.File) > 10000 {
			return operation.Result{}, invalid("too many OOXML entries")
		}
		var total uint64
		seen := map[string]bool{}
		contentTypes := false
		for _, f := range z.File {
			n := f.Name
			if strings.HasPrefix(n, "/") || strings.Contains(n, "\\") || path.Clean(n) != strings.TrimSuffix(n, "/") || strings.HasPrefix(n, "../") || strings.Contains(n, ":") || !f.Mode().IsRegular() && !f.FileInfo().IsDir() {
				return operation.Result{}, invalid("unsafe OOXML member")
			}
			if seen[n] {
				return operation.Result{}, invalid("duplicate OOXML member")
			}
			seen[n] = true
			if n == "[Content_Types].xml" {
				contentTypes = true
			}
			if strings.HasPrefix(strings.ToLower(n), "_xmlsignatures/") {
				return operation.Result{}, invalid("package is digitally signed; recompression would invalidate signatures")
			}
			if f.UncompressedSize64 > 128<<20 || total+f.UncompressedSize64 > 128<<20 {
				return operation.Result{}, invalid("expanded OOXML exceeds 128 MiB")
			}
			total += f.UncompressedSize64
		}
		if !contentTypes {
			return operation.Result{}, invalid("missing OOXML content types")
		}
		changed := 0
		var outputBytes int64
		e = toolrun.Artifact(ctx, out, overwrite, func(target string) error {
			file, e := os.Create(target)
			if e != nil {
				return e
			}
			zw := zip.NewWriter(file)
			for _, member := range z.File {
				if e := ctx.Err(); e != nil {
					_ = zw.Close()
					_ = file.Close()
					return e
				}
				header := member.FileHeader
				header.Method = zip.Deflate
				writer, e := zw.CreateHeader(&header)
				if e != nil {
					_ = zw.Close()
					_ = file.Close()
					return e
				}
				reader, e := member.Open()
				if e != nil {
					_ = zw.Close()
					_ = file.Close()
					return e
				}
				data, e := io.ReadAll(io.LimitReader(reader, int64(member.UncompressedSize64)+1))
				_ = reader.Close()
				if e != nil || uint64(len(data)) != member.UncompressedSize64 {
					_ = zw.Close()
					_ = file.Close()
					return invalid("OOXML member size mismatch")
				}
				lower := strings.ToLower(member.Name)
				// Preserve metadata-bearing media: re-encoding could remove orientation,
				// color profiles or resolution information used by the Office renderer.
				hasMetadata := bytes.Contains(data, []byte("Exif")) || bytes.Contains(data, []byte("ICC_PROFILE")) || bytes.Contains(data, []byte("iCCP")) || bytes.Contains(data, []byte("pHYs")) || bytes.Contains(data, []byte("eXIf"))
				if !hasMetadata && strings.Contains(lower, "/media/") && (strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg") || strings.HasSuffix(lower, ".png")) {
					cfg, format, e := image.DecodeConfig(bytes.NewReader(data))
					if e == nil && cfg.Width > 0 && cfg.Height > 0 && cfg.Width <= 25000000/cfg.Height {
						img, _, e := image.Decode(bytes.NewReader(data))
						if e == nil {
							var encoded bytes.Buffer
							if format == "jpeg" {
								e = jpeg.Encode(&encoded, img, &jpeg.Options{Quality: quality})
							} else if format == "png" {
								encoder := png.Encoder{CompressionLevel: png.BestCompression}
								e = encoder.Encode(&encoded, img)
							}
							if e == nil && encoded.Len() > 0 && encoded.Len() < len(data) {
								data = encoded.Bytes()
								changed++
							}
						}
					}
				}
				if _, e = writer.Write(data); e != nil {
					_ = zw.Close()
					_ = file.Close()
					return e
				}
			}
			e = zw.Close()
			ce := file.Close()
			if e != nil {
				return e
			}
			if ce != nil {
				return ce
			}
			stat, e := os.Stat(target)
			if e != nil {
				return e
			}
			if stat.Size() >= info.Size() {
				if e = os.Remove(target); e != nil {
					return e
				}
				if e = toolrun.CopyFile(input, target); e != nil {
					return e
				}
				changed = 0
				outputBytes = info.Size()
			} else {
				outputBytes = stat.Size()
			}
			return nil
		})
		return operation.Result{Operation: def.ID, Outputs: []string{out}, Data: map[string]any{"input_bytes": info.Size(), "output_bytes": outputBytes, "saved_bytes": info.Size() - outputBytes, "images_reencoded": changed, "smaller": outputBytes < info.Size()}}, e
	})})
}
