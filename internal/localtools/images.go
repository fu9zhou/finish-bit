package localtools

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fu9zhou/finish-bit/internal/toolrun"
	"github.com/fu9zhou/finish-bit/pkg/operation"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

type imageSpec struct {
	id, summary, alias string
	inputs             []operation.Parameter
	options            []operation.Parameter
	output             bool
}

func registerImages(reg *operation.Registry) error {
	rows := []imageSpec{
		{"qrcode.generate", "Generate a colored QR code and verify it can be read back", "二维码生成 美化", in("text"), []operation.Parameter{integer("size", "Minimum square pixel size", 512), str("level", "L, M, Q or H error correction", "M"), str("foreground", "Opaque #RRGGBB modules", "#000000"), str("background", "Opaque #RRGGBB background", "#FFFFFF")}, true},
		{"qrcode.decode", "Decode a QR code from a local PNG/JPEG/GIF image", "二维码扫描", in("input"), nil, false},
		{"qrcode.contact", "Generate a vCard QR code from contact fields", "二维码名片", in("name"), []operation.Parameter{str("phone", "Phone number", ""), str("email", "Email address", ""), str("organization", "Organization", ""), str("url", "Contact URL", ""), integer("size", "Minimum QR width", 512)}, true},
		{"image.pixelate", "Pixelate an image or selected rectangle", "图片像素化", in("input"), []operation.Parameter{integer("block", "Square block size", 12), integer("x", "Rectangle left", 0), integer("y", "Rectangle top", 0), integer("width", "Rectangle width; 0 to edge", 0), integer("height", "Rectangle height; 0 to edge", 0)}, true},
		{"image.ascii", "Convert an image to a grayscale character drawing", "图片转字符画", in("input"), []operation.Parameter{integer("columns", "Text columns", 100), str("characters", "Dark to light character ramp", "@%#*+=-:. "), boolean("invert", "Invert character ramp", false)}, false},
		{"image.split-grid", "Split an image into ordered grid tiles", "九宫格切图", in("input"), []operation.Parameter{integer("columns", "Grid columns", 3), integer("rows", "Grid rows", 3)}, true},
		{"image.palette", "Find approximate dominant colors using a bounded histogram", "图片配色提取", in("input"), []operation.Parameter{integer("count", "Palette entries", 8)}, false},
		{"image.color-at", "Read a pixel color from an image", "图片取色", in("input"), []operation.Parameter{integer("x", "Pixel column", 0), integer("y", "Pixel row", 0)}, false},
		{"image.solid", "Create a solid PNG canvas", "纯色图片生成", nil, []operation.Parameter{integer("width", "Pixel width", 512), integer("height", "Pixel height", 512), str("color", "Opaque #RRGGBB", "#FFFFFF")}, true},
		{"image.id-photo", "Center crop and resize a supplied portrait to a document photo preset", "证件照尺寸生成", in("input"), []operation.Parameter{str("preset", "one-inch, two-inch, passport or custom", "one-inch"), integer("width", "Custom width", 295), integer("height", "Custom height", 413), str("background", "Background for existing transparency", "#FFFFFF")}, true},
		{"image.hide", "Store a UTF-8 message in lossless PNG pixels with a checksum", "图片隐写", in("input", "message"), nil, true},
		{"image.reveal", "Recover a FinishBit message from PNG pixels", "图片隐写解码", in("input"), nil, false},
		{"image.minimum-bytes", "Increase PNG file size with a valid private ancillary chunk", "增加图片文件大小", in("input"), []operation.Parameter{integer("bytes", "Minimum PNG file bytes", 102400)}, true},
		{"image.target-size", "Encode JPEG at the highest tested quality meeting a byte budget", "图片目标大小压缩", in("input"), []operation.Parameter{integer("bytes", "Maximum output bytes", 102400), integer("min-quality", "Minimum JPEG quality", 10), integer("max-quality", "Maximum JPEG quality", 95)}, true},
	}
	for _, s := range rows {
		for i := range s.inputs {
			if s.inputs[i].Name == "input" {
				s.inputs[i].Description = "Local PNG/JPEG/GIF image file"
			} else {
				s.inputs[i].Description = "Literal text value"
			}
		}
		opts := append([]operation.Parameter{}, s.options...)
		opts = append(opts, boolean("overwrite", "Replace existing file", false))
		if s.output {
			description := "Destination PNG file"
			if s.id == "image.target-size" {
				description = "Destination JPEG file"
			}
			if s.id == "image.split-grid" {
				description = "New destination directory for PNG tiles"
			}
			opts = append(opts, toolrun.Param("output", description, true))
		} else {
			opts = append(opts, str("output", "Optional JSON/text output file", ""))
		}
		def := operation.Definition{ID: s.id, Summary: s.summary, Description: s.summary + ". Offline; local PNG/JPEG/GIF inputs up to 32 MiB and 25 million pixels. No automatic face detection or semantic image analysis.", Aliases: []string{s.alias}, Tags: []string{strings.Split(s.id, ".")[0], "offline"}, Inputs: s.inputs, Options: opts, Source: "core"}
		if e := reg.Register(operation.Capability{Definition: def, Runner: operation.Func(func(ctx context.Context, r operation.Request) (operation.Result, error) {
			if e := operation.ValidateRequest(def, r); e != nil {
				return operation.Result{}, e
			}
			v := &toolrun.Values{Request: r}
			if e := ctx.Err(); e != nil {
				return operation.Result{}, e
			}
			result, e := runLocalImage(ctx, s.id, v)
			if v.Err != nil {
				return operation.Result{}, v.Err
			}
			result.Operation = s.id
			if e != nil {
				return operation.Result{}, e
			}
			if e = ctx.Err(); e != nil {
				return operation.Result{}, e
			}
			return result, nil
		})}); e != nil {
			return e
		}
	}
	return nil
}
func loadImage(path string) (image.Image, error) {
	p, e := toolrun.LocalFile(path)
	if e != nil {
		return nil, e
	}
	f, e := os.Open(p)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, (32<<20)+1))
	if e != nil {
		return nil, e
	}
	if len(b) > 32<<20 {
		return nil, invalid("image exceeds 32 MiB")
	}
	cfg, format, e := image.DecodeConfig(bytes.NewReader(b))
	if e != nil {
		return nil, invalid("invalid image")
	}
	if format != "png" && format != "jpeg" && format != "gif" {
		return nil, invalid("supported inputs are PNG/JPEG/GIF")
	}
	if cfg.Width < 1 || cfg.Height < 1 || cfg.Width > 32768 || cfg.Height > 32768 || cfg.Width > 25000000/cfg.Height {
		return nil, invalid("image exceeds dimension/pixel limits")
	}
	img, _, e := image.Decode(bytes.NewReader(b))
	return img, e
}
func rgbaCopy(img image.Image) *image.NRGBA {
	out := image.NewNRGBA(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))
	draw.Draw(out, out.Bounds(), img, img.Bounds().Min, draw.Src)
	return out
}
func parseRGB(s string) (color.NRGBA, error) {
	if len(s) != 7 || s[0] != '#' {
		return color.NRGBA{}, invalid("color must be #RRGGBB")
	}
	var r, g, b uint8
	if _, e := fmt.Sscanf(s, "#%02x%02x%02x", &r, &g, &b); e != nil {
		return color.NRGBA{}, invalid("invalid RGB color")
	}
	return color.NRGBA{R: r, G: g, B: b, A: 255}, nil
}
func savePNG(ctx context.Context, v *toolrun.Values, img image.Image) (operation.Result, error) {
	out := v.String("output", "")
	overwrite := v.Bool("overwrite", false)
	if v.Err != nil {
		return operation.Result{}, v.Err
	}
	if strings.ToLower(filepath.Ext(out)) != ".png" {
		return operation.Result{}, invalid("output must use .png extension")
	}
	e := toolrun.Artifact(ctx, out, overwrite, func(p string) error {
		f, e := os.Create(p)
		if e != nil {
			return e
		}
		err := png.Encode(f, img)
		closeErr := f.Close()
		if err != nil {
			return err
		}
		return closeErr
	})
	return operation.Result{Outputs: []string{out}, Data: map[string]any{"width": img.Bounds().Dx(), "height": img.Bounds().Dy()}}, e
}
func imageData(ctx context.Context, v *toolrun.Values, data map[string]any) (operation.Result, error) {
	out := v.String("output", "")
	overwrite := v.Bool("overwrite", false)
	if v.Err != nil {
		return operation.Result{}, v.Err
	}
	if out == "" {
		return operation.Result{Data: data}, nil
	}
	payload, e := jsonBytes(data)
	if e != nil {
		return operation.Result{}, e
	}
	e = toolrun.Artifact(ctx, out, overwrite, func(p string) error { return os.WriteFile(p, payload, 0600) })
	return operation.Result{Outputs: []string{out}}, e
}
func runLocalImage(ctx context.Context, id string, v *toolrun.Values) (operation.Result, error) {
	a := v.Request.Inputs
	if id == "qrcode.generate" || id == "qrcode.contact" {
		content := a[0]
		size := v.Int("size", 512, 64, 4096)
		level := "M"
		fg, bg := color.NRGBA{A: 255}, color.NRGBA{255, 255, 255, 255}
		if id == "qrcode.contact" {
			esc := func(s string) string {
				return strings.NewReplacer("\\", "\\\\", "\r\n", "\\n", "\n", "\\n", "\r", "\\n", ";", "\\;", ",", "\\,").Replace(s)
			}
			content = "BEGIN:VCARD\r\nVERSION:3.0\r\nFN:" + esc(content) + "\r\nN:;" + esc(content) + ";;;\r\n"
			for _, field := range []struct{ k, t string }{{"phone", "TEL"}, {"email", "EMAIL"}, {"organization", "ORG"}, {"url", "URL"}} {
				if x := v.String(field.k, ""); x != "" {
					content += field.t + ":" + esc(x) + "\r\n"
				}
			}
			content += "END:VCARD\r\n"
		} else {
			level = v.Enum("level", "M", "L", "M", "Q", "H")
			var e error
			fg, e = parseRGB(v.String("foreground", "#000000"))
			if e != nil {
				return operation.Result{}, e
			}
			bg, e = parseRGB(v.String("background", "#FFFFFF"))
			if e != nil {
				return operation.Result{}, e
			}
		}
		if len(content) > 2000 {
			return operation.Result{}, invalid("QR content exceeds 2000 UTF-8 bytes")
		}
		if v.Err != nil {
			return operation.Result{}, v.Err
		}
		matrix, e := qrcode.NewQRCodeWriter().Encode(content, gozxing.BarcodeFormat_QR_CODE, size, size, map[gozxing.EncodeHintType]interface{}{gozxing.EncodeHintType_ERROR_CORRECTION: level, gozxing.EncodeHintType_MARGIN: 4, gozxing.EncodeHintType_CHARACTER_SET: "UTF-8"})
		if e != nil {
			return operation.Result{}, invalid("QR encoding failed: " + e.Error())
		}
		img := image.NewNRGBA(matrix.Bounds())
		for y := 0; y < img.Bounds().Dy(); y++ {
			for x := 0; x < img.Bounds().Dx(); x++ {
				c := bg
				if matrix.Get(x, y) {
					c = fg
				}
				img.SetNRGBA(x, y, c)
			}
		}
		decoded, e := decodeQR(img)
		if e != nil || decoded != content {
			return operation.Result{}, invalid("QR colors/size failed read-back verification")
		}
		return savePNG(ctx, v, img)
	}
	if id == "image.solid" {
		w, h := v.Int("width", 512, 1, 8192), v.Int("height", 512, 1, 8192)
		if v.Err != nil {
			return operation.Result{}, v.Err
		}
		if w > 25000000/h {
			return operation.Result{}, invalid("canvas exceeds 25 million pixels")
		}
		c, e := parseRGB(v.String("color", "#FFFFFF"))
		if e != nil {
			return operation.Result{}, e
		}
		img := image.NewNRGBA(image.Rect(0, 0, w, h))
		draw.Draw(img, img.Bounds(), image.NewUniform(c), image.Point{}, draw.Src)
		return savePNG(ctx, v, img)
	}
	img, e := loadImage(a[0])
	if e != nil {
		return operation.Result{}, e
	}
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	switch id {
	case "qrcode.decode":
		text, e := decodeQR(img)
		if e != nil {
			return operation.Result{}, invalid("no readable QR code: " + e.Error())
		}
		return imageData(ctx, v, value(text))
	case "image.pixelate":
		block := v.Int("block", 12, 1, 1024)
		x, y := v.Int("x", 0, 0, w-1), v.Int("y", 0, 0, h-1)
		rw, rh := v.Int("width", 0, 0, w), v.Int("height", 0, 0, h)
		if rw == 0 {
			rw = w - x
		}
		if rh == 0 {
			rh = h - y
		}
		if v.Err != nil {
			return operation.Result{}, v.Err
		}
		if x+rw > w || y+rh > h {
			return operation.Result{}, invalid("rectangle exceeds image")
		}
		out := rgbaCopy(img)
		for top := y; top < y+rh; top += block {
			if e := ctx.Err(); e != nil {
				return operation.Result{}, e
			}
			for left := x; left < x+rw; left += block {
				right, bottom := min(left+block, x+rw), min(top+block, y+rh)
				var r, g, b, alpha uint64
				for yy := top; yy < bottom; yy++ {
					for xx := left; xx < right; xx++ {
						c := out.NRGBAAt(xx, yy)
						r += uint64(c.R)
						g += uint64(c.G)
						b += uint64(c.B)
						alpha += uint64(c.A)
					}
				}
				n := uint64((right - left) * (bottom - top))
				c := color.NRGBA{uint8(r / n), uint8(g / n), uint8(b / n), uint8(alpha / n)}
				draw.Draw(out, image.Rect(left, top, right, bottom), image.NewUniform(c), image.Point{}, draw.Src)
			}
		}
		return savePNG(ctx, v, out)
	case "image.ascii":
		columns := v.Int("columns", 100, 1, 500)
		ramp := []rune(v.String("characters", "@%#*+=-:. "))
		invert := v.Bool("invert", false)
		if len(ramp) < 2 || len(ramp) > 256 {
			return operation.Result{}, invalid("character ramp needs 2 to 256 characters")
		}
		if v.Err != nil {
			return operation.Result{}, v.Err
		}
		rows := max(1, int(math.Round(float64(h)*float64(columns)/float64(w)*0.5)))
		if rows > 2000 {
			return operation.Result{}, invalid("character drawing exceeds 2000 rows")
		}
		var b strings.Builder
		for y := 0; y < rows; y++ {
			for x := 0; x < columns; x++ {
				r, g, blue, _ := img.At(x*w/columns, y*h/rows).RGBA()
				luma := (299*uint64(r) + 587*uint64(g) + 114*uint64(blue)) / 1000
				idx := int(luma * uint64(len(ramp)-1) / 65535)
				if invert {
					idx = len(ramp) - 1 - idx
				}
				b.WriteRune(ramp[idx])
			}
			b.WriteByte('\n')
		}
		return imageData(ctx, v, value(b.String()))
	case "image.split-grid":
		cols, rows := v.Int("columns", 3, 1, 32), v.Int("rows", 3, 1, 32)
		if v.Err != nil {
			return operation.Result{}, v.Err
		}
		if cols > w || rows > h || cols*rows > 256 {
			return operation.Result{}, invalid("grid must fit image and have at most 256 tiles")
		}
		out := v.String("output", "")
		files, e := toolrun.Directory(ctx, out, func(dir string) error {
			for y := 0; y < rows; y++ {
				for x := 0; x < cols; x++ {
					if e := ctx.Err(); e != nil {
						return e
					}
					rect := image.Rect(x*w/cols, y*h/rows, (x+1)*w/cols, (y+1)*h/rows)
					tile := image.NewNRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
					draw.Draw(tile, tile.Bounds(), img, rect.Min, draw.Src)
					f, e := os.Create(filepath.Join(dir, fmt.Sprintf("tile-%03d.png", y*cols+x+1)))
					if e != nil {
						return e
					}
					err := png.Encode(f, tile)
					closeErr := f.Close()
					if err != nil {
						return err
					}
					if closeErr != nil {
						return closeErr
					}
				}
			}
			return nil
		})
		return operation.Result{Outputs: files, Data: map[string]any{"columns": cols, "rows": rows, "order": "row-major", "edge_policy": "integer boundaries cover all pixels"}}, e
	case "image.palette":
		count := v.Int("count", 8, 1, 64)
		if v.Err != nil {
			return operation.Result{}, v.Err
		}
		hist := map[uint32]int{}
		step := max(1, int(math.Sqrt(float64(w*h)/100000)))
		total := 0
		for y := 0; y < h; y += step {
			for x := 0; x < w; x += step {
				c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
				if c.A < 128 {
					continue
				}
				key := uint32(c.R>>4)<<8 | uint32(c.G>>4)<<4 | uint32(c.B>>4)
				hist[key]++
				total++
			}
		}
		keys := []uint32{}
		for k := range hist {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			if hist[keys[i]] == hist[keys[j]] {
				return keys[i] < keys[j]
			}
			return hist[keys[i]] > hist[keys[j]]
		})
		out := []map[string]any{}
		for _, k := range keys[:min(count, len(keys))] {
			out = append(out, map[string]any{"hex": fmt.Sprintf("#%02X%02X%02X", (k>>8&15)*17, (k>>4&15)*17, (k&15)*17), "fraction": float64(hist[k]) / float64(total)})
		}
		return imageData(ctx, v, map[string]any{"colors": out, "sampled_pixels": total, "quantization_bits": 4})
	case "image.color-at":
		x, y := v.Int("x", 0, 0, w-1), v.Int("y", 0, 0, h-1)
		if v.Err != nil {
			return operation.Result{}, v.Err
		}
		c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
		return imageData(ctx, v, map[string]any{"hex": fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B), "rgba": []int{int(c.R), int(c.G), int(c.B), int(c.A)}, "x": x, "y": y})
	case "image.id-photo":
		preset := v.Enum("preset", "one-inch", "one-inch", "two-inch", "passport", "custom")
		tw, th := 295, 413
		switch preset {
		case "two-inch":
			tw, th = 413, 579
		case "passport":
			tw, th = 390, 567
		case "custom":
			tw, th = v.Int("width", 295, 1, 4096), v.Int("height", 413, 1, 4096)
		}
		bg, e := parseRGB(v.String("background", "#FFFFFF"))
		if e != nil {
			return operation.Result{}, e
		}
		if v.Err != nil {
			return operation.Result{}, v.Err
		}
		scale := math.Max(float64(tw)/float64(w), float64(th)/float64(h))
		sw, sh := float64(tw)/scale, float64(th)/scale
		ox, oy := (float64(w)-sw)/2, (float64(h)-sh)/2
		out := image.NewNRGBA(image.Rect(0, 0, tw, th))
		draw.Draw(out, out.Bounds(), image.NewUniform(bg), image.Point{}, draw.Src)
		for y := 0; y < th; y++ {
			for x := 0; x < tw; x++ {
				sx, sy := min(w-1, int(ox+(float64(x)+0.5)/scale)), min(h-1, int(oy+(float64(y)+0.5)/scale))
				c := color.NRGBAModel.Convert(img.At(sx, sy)).(color.NRGBA)
				alpha := uint32(c.A)
				out.SetNRGBA(x, y, color.NRGBA{uint8((uint32(c.R)*alpha + uint32(bg.R)*(255-alpha)) / 255), uint8((uint32(c.G)*alpha + uint32(bg.G)*(255-alpha)) / 255), uint8((uint32(c.B)*alpha + uint32(bg.B)*(255-alpha)) / 255), 255})
			}
		}
		result, e := savePNG(ctx, v, out)
		if e == nil {
			result.Data["crop"] = "center crop; no face detection or background removal"
			result.Data["nominal_dpi"] = 300
		}
		return result, e
	case "image.hide", "image.reveal":
		out := rgbaCopy(img)
		capacity := w * h * 3 / 8
		getByte := func(index int) byte {
			var b byte
			for bit := 0; bit < 8; bit++ {
				slot := index*8 + bit
				p := slot / 3
				channel := slot % 3
				offset := (p/w)*out.Stride + (p%w)*4 + channel
				b = (b << 1) | (out.Pix[offset] & 1)
			}
			return b
		}
		if id == "image.reveal" {
			if capacity < 13 {
				return operation.Result{}, invalid("image too small for payload")
			}
			header := make([]byte, 13)
			for i := range header {
				header[i] = getByte(i)
			}
			if string(header[:5]) != "FNSH1" {
				return operation.Result{}, invalid("no FinishBit payload")
			}
			n := int(binary.BigEndian.Uint32(header[5:9]))
			if n > 1<<20 || n > capacity-13 {
				return operation.Result{}, invalid("invalid payload length")
			}
			b := make([]byte, n)
			for i := range b {
				b[i] = getByte(i + 13)
			}
			if crc32.ChecksumIEEE(b) != binary.BigEndian.Uint32(header[9:13]) {
				return operation.Result{}, invalid("payload checksum failed")
			}
			return imageData(ctx, v, byteResult(b))
		}
		message := []byte(a[1])
		if len(message) > 1<<20 || len(message) > capacity-13 {
			return operation.Result{}, invalid("message exceeds image capacity or 1 MiB")
		}
		payload := make([]byte, 13+len(message))
		copy(payload, "FNSH1")
		binary.BigEndian.PutUint32(payload[5:9], uint32(len(message)))
		binary.BigEndian.PutUint32(payload[9:13], crc32.ChecksumIEEE(message))
		copy(payload[13:], message)
		for i, b := range payload {
			for bit := 0; bit < 8; bit++ {
				slot := i*8 + bit
				p := slot / 3
				offset := (p/w)*out.Stride + (p%w)*4 + slot%3
				out.Pix[offset] = (out.Pix[offset] & 254) | ((b >> (7 - bit)) & 1)
			}
		}
		return savePNG(ctx, v, out)
	case "image.minimum-bytes":
		target := v.Int("bytes", 102400, 1, 32<<20)
		output := v.String("output", "")
		overwrite := v.Bool("overwrite", false)
		if v.Err != nil {
			return operation.Result{}, v.Err
		}
		if strings.ToLower(filepath.Ext(output)) != ".png" {
			return operation.Result{}, invalid("output must be PNG")
		}
		var buf bytes.Buffer
		if e := png.Encode(&buf, img); e != nil {
			return operation.Result{}, e
		}
		encoded := buf.Bytes()
		if target > len(encoded) {
			payload := make([]byte, max(0, target-len(encoded)-12))
			chunk := make([]byte, 12+len(payload))
			binary.BigEndian.PutUint32(chunk, uint32(len(payload)))
			copy(chunk[4:], "npAD")
			copy(chunk[8:], payload)
			binary.BigEndian.PutUint32(chunk[len(chunk)-4:], crc32.ChecksumIEEE(chunk[4:len(chunk)-4]))
			combined := append([]byte{}, encoded[:len(encoded)-12]...)
			combined = append(combined, chunk...)
			combined = append(combined, encoded[len(encoded)-12:]...)
			encoded = combined
		}
		e := toolrun.Artifact(ctx, output, overwrite, func(p string) error { return os.WriteFile(p, encoded, 0600) })
		return operation.Result{Outputs: []string{output}, Data: map[string]any{"bytes": len(encoded), "pixel_details_added": false}}, e
	case "image.target-size":
		target := v.Int("bytes", 102400, 128, 32<<20)
		low, high := v.Int("min-quality", 10, 1, 100), v.Int("max-quality", 95, 1, 100)
		output := v.String("output", "")
		overwrite := v.Bool("overwrite", false)
		if v.Err != nil {
			return operation.Result{}, v.Err
		}
		if low > high {
			return operation.Result{}, invalid("min-quality exceeds max-quality")
		}
		ext := strings.ToLower(filepath.Ext(output))
		if ext != ".jpg" && ext != ".jpeg" {
			return operation.Result{}, invalid("output must be JPEG")
		}
		flat := image.NewRGBA(image.Rect(0, 0, w, h))
		draw.Draw(flat, flat.Bounds(), image.White, image.Point{}, draw.Src)
		draw.Draw(flat, flat.Bounds(), img, img.Bounds().Min, draw.Over)
		var best []byte
		chosen := 0
		for q := high; q >= low; q-- {
			if e := ctx.Err(); e != nil {
				return operation.Result{}, e
			}
			var b bytes.Buffer
			if e := jpeg.Encode(&b, flat, &jpeg.Options{Quality: q}); e != nil {
				return operation.Result{}, e
			}
			if b.Len() <= target {
				best = b.Bytes()
				chosen = q
				break
			}
		}
		if best == nil {
			return operation.Result{}, invalid("target cannot be reached within quality range; resize input or increase budget")
		}
		e := toolrun.Artifact(ctx, output, overwrite, func(p string) error { return os.WriteFile(p, best, 0600) })
		return operation.Result{Outputs: []string{output}, Data: map[string]any{"bytes": len(best), "quality": chosen, "width": w, "height": h}}, e
	}
	return operation.Result{}, invalid("unknown image operation")
}
func decodeQR(img image.Image) (string, error) {
	bitmap, e := gozxing.NewBinaryBitmapFromImage(img)
	if e != nil {
		return "", e
	}
	result, e := qrcode.NewQRCodeReader().Decode(bitmap, map[gozxing.DecodeHintType]interface{}{gozxing.DecodeHintType_TRY_HARDER: true})
	if e != nil {
		return "", e
	}
	return result.GetText(), nil
}
