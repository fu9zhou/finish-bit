// Package raster provides optional ImageMagick operations for raster assets.
package raster

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/fu9zhou/finish-bit/internal/toolrun"
	"github.com/fu9zhou/finish-bit/pkg/operation"
)

type spec struct {
	id, summary, alias string
	options            []operation.Parameter
	inputs             int
	mode               string
}

func s(name, description, value string) operation.Parameter {
	return toolrun.Option(name, operation.TypeString, description, value)
}
func n(name, description string, value int) operation.Parameter {
	return toolrun.Option(name, operation.TypeInteger, description, value)
}
func files() operation.Parameter {
	return toolrun.Option("files", operation.TypeStrings, "Additional local images in order", nil)
}
func catalog() []spec {
	return []spec{
		{"image.formats", "List raster formats available in the installed runtime", "图片格式支持", nil, 0, "inspect"},
		{"image.crop", "Crop an image rectangle", "图片裁剪", []operation.Parameter{n("width", "Rectangle width", 100), n("height", "Rectangle height", 100), n("x", "Left coordinate", 0), n("y", "Top coordinate", 0)}, 1, "file"},
		{"image.rotate", "Rotate an image on a transparent canvas", "图片旋转", []operation.Parameter{n("angle", "Clockwise angle from -360 to 360", 90)}, 1, "file"},
		{"image.flip", "Flip an image horizontally or vertically", "图片翻转", []operation.Parameter{s("axis", "horizontal or vertical", "horizontal")}, 1, "file"},
		{"image.orient", "Apply EXIF orientation to image pixels", "照片方向校正", nil, 1, "file"},
		{"image.compress", "Strip metadata and encode with chosen quality", "图片压缩", []operation.Parameter{n("quality", "Encoding quality from 1 to 100", 80)}, 1, "file"},
		{"image.watermark", "Overlay a local watermark image", "图片水印", []operation.Parameter{n("x", "Left position", 10), n("y", "Top position", 10)}, 2, "file"},
		{"image.annotate", "Draw text using an explicitly supplied local font", "图片文字标注", []operation.Parameter{toolrun.Param("text", "Text to draw", true), toolrun.Param("font", "Local TTF or OTF font", true), n("size", "Font point size", 24), n("x", "Left offset", 10), n("y", "Top offset", 10), s("color", "Text color: #RRGGBB or #RRGGBBAA", "#ffffff")}, 1, "file"},
		{"image.join", "Join images horizontally or vertically", "图片拼接", []operation.Parameter{files(), s("axis", "horizontal or vertical", "horizontal")}, 1, "file"},
		{"image.contact-sheet", "Build an image grid with fitted thumbnails", "图片联系表", []operation.Parameter{files(), n("columns", "Grid columns", 4), n("width", "Tile width", 200), n("height", "Tile height", 150), n("gap", "Gap between tiles", 4)}, 1, "file"},
		{"image.canvas", "Place an image on a larger or smaller canvas", "扩展图片画布", []operation.Parameter{n("width", "Canvas width", 640), n("height", "Canvas height", 480), s("background", "#RRGGBB, #RRGGBBAA or transparent", "transparent")}, 1, "file"},
		{"image.flatten", "Flatten image transparency against a solid color", "去除图片透明度", []operation.Parameter{s("background", "Solid #RRGGBB color", "#ffffff")}, 1, "file"},
		{"image.transparent", "Replace a selected color with transparency", "图片颜色透明化", []operation.Parameter{s("color", "Color to replace, #RRGGBB", "#ffffff"), n("tolerance", "Color matching tolerance percent", 0)}, 1, "file"},
		{"image.grayscale", "Convert image pixels to grayscale", "图片灰度", nil, 1, "file"},
		{"image.blur", "Apply Gaussian blur", "图片模糊", []operation.Parameter{n("sigma", "Blur sigma in pixels", 2)}, 1, "file"},
		{"image.sharpen", "Sharpen image edges", "图片锐化", []operation.Parameter{n("sigma", "Sharpen sigma in pixels", 1)}, 1, "file"},
		{"image.adjust", "Adjust brightness, saturation and hue", "图片调色", []operation.Parameter{n("brightness", "Brightness percent", 100), n("saturation", "Saturation percent", 100), n("hue", "Hue percent: 100 unchanged", 100)}, 1, "file"},
		{"image.difference", "Render the pixel difference between equal-sized images", "图片差异图", nil, 2, "file"},
		{"image.gif-create", "Create an animated GIF from ordered images", "生成 GIF 动图", []operation.Parameter{files(), n("delay", "Frame delay in centiseconds", 20), n("loops", "Loop count; 0 is infinite", 0)}, 1, "file"},
		{"image.gif-split", "Coalesce and export GIF frames as PNG", "GIF 拆帧", nil, 1, "directory"},
		{"image.icon", "Generate a multi-resolution ICO icon", "生成图标", nil, 1, "file"},
		{"image.strip-metadata", "Remove profiles and comments while re-encoding pixels", "清除图片元数据", nil, 1, "file"},
	}
}

func Register(registry *operation.Registry, resolver toolrun.Resolver) error {
	for _, item := range catalog() {
		def := operation.Definition{ID: item.id, Summary: item.summary, Description: item.summary + ". Managed ImageMagick; local raster inputs, bounded processing and staged output. See docs/image-enhancements.md.", Aliases: []string{item.alias}, Tags: []string{"image", "raster"}, Options: append([]operation.Parameter{}, item.options...), Requirements: []operation.Requirement{{Package: "imagemagick"}}, Source: "imagemagick"}
		for i := 0; i < item.inputs; i++ {
			name := "input"
			if i > 0 {
				name = "second"
			}
			def.Inputs = append(def.Inputs, toolrun.Param(name, "Local raster image", true))
		}
		if item.mode == "file" {
			def.Options = append(def.Options, toolrun.OutputOptions()...)
		} else if item.mode == "directory" {
			def.Options = append(def.Options, toolrun.Param("output", "New output directory", true))
		}
		if err := registry.Register(operation.Capability{Definition: def, Runner: operation.Func(func(ctx context.Context, r operation.Request) (operation.Result, error) {
			if err := operation.ValidateRequest(def, r); err != nil {
				return operation.Result{}, err
			}
			return Run(ctx, resolver, item.id, r)
		})}); err != nil {
			return err
		}
	}
	return nil
}

var colorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}(?:[0-9a-fA-F]{2})?$`)

func colorValue(v *toolrun.Values, name, fallback string, transparent bool) string {
	value := v.String(name, fallback)
	v.Check(colorPattern.MatchString(value) || transparent && value == "transparent", name+" must be a hex color")
	if value == "transparent" {
		value = "none"
	}
	return value
}
func formatOf(path string) string {
	format := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	switch format {
	case "jpg":
		return "jpeg"
	case "tif":
		return "tiff"
	case "heif":
		return "heic"
	}
	return format
}
func allowed(format string) bool {
	switch format {
	case "png", "jpeg", "webp", "gif", "tiff", "bmp", "ico", "avif", "heic":
		return true
	}
	return false
}
func limits() []string {
	return []string{"-limit", "thread", "2", "-limit", "memory", "256MiB", "-limit", "map", "512MiB", "-limit", "disk", "1GiB", "-limit", "time", "120", "-limit", "list-length", "500"}
}

// Run serves both new Operations and the optional engine branch of the original image APIs.
func Run(ctx context.Context, resolver toolrun.Resolver, id string, r operation.Request) (operation.Result, error) {
	result := operation.Result{Operation: id}
	if resolver == nil {
		return result, &operation.Error{Code: operation.CodeDependencyMissing, Message: "operation requires ImageMagick", Suggestion: "fnsh pkg add imagemagick"}
	}
	if id == "image.formats" {
		raw, err := toolrun.Run(ctx, resolver, "imagemagick", "magick", "", "identify", "-list", "format")
		if err != nil {
			return result, err
		}
		result.Data = map[string]any{"report": string(raw), "supported_input_extensions": []string{"png", "jpg", "jpeg", "webp", "gif", "tif", "tiff", "bmp", "ico", "avif", "heic", "heif"}}
		return result, nil
	}
	v := &toolrun.Values{Request: r}
	inputs := append([]string{}, r.Inputs...)
	inputs = append(inputs, v.Strings("files")...)
	v.Check(len(inputs) > 0 && len(inputs) <= 100, "provide 1 to 100 local images")
	if v.Err != nil {
		return result, v.Err
	}
	work, err := os.MkdirTemp("", "finishbit-image-")
	if err != nil {
		return result, err
	}
	defer os.RemoveAll(work)
	paths := []string{}
	var totalBytes int64
	var inputBytes int64
	for i, input := range inputs {
		path := v.File(input)
		format := formatOf(input)
		v.Check(allowed(format), "unsupported input extension: "+format)
		if v.Err != nil {
			return result, v.Err
		}
		stat, err := os.Stat(path)
		if err != nil {
			return result, err
		}
		totalBytes += stat.Size()
		if i == 0 {
			inputBytes = stat.Size()
		}
		if stat.Size() > 64<<20 || totalBytes > 512<<20 {
			return result, toolrun.Invalid("images exceed 64 MiB per file or 512 MiB combined")
		}
		name := fmt.Sprintf("input-%03d.%s", i, format)
		if err := toolrun.CopyFile(path, filepath.Join(work, name)); err != nil {
			return result, err
		}
		paths = append(paths, format+":"+name)
	}
	// Raster-only coder prefixes prevent filenames or disguised extensions from selecting URL/document delegates.
	run := func(sub string, args ...string) ([]byte, error) {
		command := []string{}
		if sub != "" {
			command = append(command, sub)
		}
		command = append(command, limits()...)
		command = append(command, args...)
		return toolrun.Run(ctx, resolver, "imagemagick", "magick", work, command...)
	}
	info := func(path string) (map[string]any, error) {
		raw, err := run("identify", "-ping", "-format", "%w\t%h\t%m\t%z\n", path)
		if err != nil {
			return nil, err
		}
		frames := strings.Split(strings.TrimSpace(string(raw)), "\n")
		v.Check(len(frames) <= 500, "image has more than 500 frames")
		var first map[string]any
		var pixels int64
		for _, frame := range frames {
			fields := strings.Fields(frame)
			if len(fields) != 4 {
				return nil, fmt.Errorf("invalid image information")
			}
			width, e1 := strconv.Atoi(fields[0])
			height, e2 := strconv.Atoi(fields[1])
			depth, e3 := strconv.Atoi(fields[3])
			if e1 != nil || e2 != nil || e3 != nil || width < 1 || height < 1 {
				return nil, fmt.Errorf("invalid image dimensions")
			}
			pixels += int64(width) * int64(height)
			if width > 32768 || height > 32768 || pixels > 100_000_000 {
				return nil, toolrun.Invalid("image exceeds 100 million total pixels or 32768 pixels per dimension")
			}
			if first == nil {
				first = map[string]any{"width": width, "height": height, "format": strings.ToLower(fields[2]), "depth": depth, "frames": len(frames)}
			}
		}
		if v.Err != nil {
			return nil, v.Err
		}
		return first, nil
	}
	infos := []map[string]any{}
	for _, path := range paths {
		data, err := info(path)
		if err != nil {
			return result, err
		}
		infos = append(infos, data)
	}
	if id == "image.info" {
		result.Data = infos[0]
		result.Data["bytes"] = inputBytes
		return result, nil
	}
	output := v.String("output", "")
	overwrite := v.Bool("overwrite", false)
	format := formatOf(output)
	if id == "image.convert" || id == "image.resize" {
		if f := v.String("format", ""); f != "" {
			format = formatOf("file." + f)
		}
	}
	if id != "image.gif-split" {
		v.Check(allowed(format), "unsupported output extension/format: "+format)
	}
	args := []string{paths[0] + "[0]"}
	sub := ""
	directory := false
	switch id {
	case "image.convert", "image.resize":
		args = append(args, "-auto-orient")
		if id == "image.resize" {
			w := v.Int("width", 0, 0, 32768)
			h := v.Int("height", 0, 0, 32768)
			v.Check(w > 0 || h > 0, "provide width or height")
			geometry := ""
			if w > 0 {
				geometry = strconv.Itoa(w)
			}
			geometry += "x"
			if h > 0 {
				geometry += strconv.Itoa(h)
			}
			if !v.Bool("upscale", false) {
				geometry += ">"
			}
			args = append(args, "-resize", geometry)
		}
		args = append(args, "-quality", strconv.Itoa(v.Int("quality", 90, 1, 100)))
	case "image.crop":
		w := v.Int("width", 100, 1, 32768)
		h := v.Int("height", 100, 1, 32768)
		x := v.Int("x", 0, 0, 32768)
		y := v.Int("y", 0, 0, 32768)
		v.Check(x+w <= infos[0]["width"].(int) && y+h <= infos[0]["height"].(int), "crop rectangle exceeds input bounds")
		args = append(args, "-crop", fmt.Sprintf("%dx%d+%d+%d", w, h, x, y), "+repage")
	case "image.rotate":
		args = append(args, "-background", "none", "-rotate", strconv.Itoa(v.Int("angle", 90, -360, 360)), "+repage")
	case "image.flip":
		axis := v.Enum("axis", "horizontal", "horizontal", "vertical")
		flag := "-flop"
		if axis == "vertical" {
			flag = "-flip"
		}
		args = append(args, flag)
	case "image.orient":
		args = append(args, "-auto-orient")
	case "image.compress":
		args = append(args, "-strip", "-quality", strconv.Itoa(v.Int("quality", 80, 1, 100)))
	case "image.watermark":
		v.Check(len(paths) == 2, "watermark needs two images")
		if v.Err != nil {
			return result, v.Err
		}
		args = append(args, paths[1]+"[0]", "-geometry", fmt.Sprintf("+%d+%d", v.Int("x", 10, 0, 32768), v.Int("y", 10, 0, 32768)), "-compose", "over", "-composite")
	case "image.annotate":
		text := v.String("text", "")
		v.Check(text != "" && len(text) <= 16384, "text must contain 1 to 16384 bytes")
		font := v.File(v.String("font", ""))
		if v.Err != nil {
			return result, v.Err
		}
		if err := toolrun.CopyFile(font, filepath.Join(work, "font.ttf")); err != nil {
			return result, err
		}
		if err := os.WriteFile(filepath.Join(work, "annotation.txt"), []byte(strings.ReplaceAll(text, "%", "%%")), 0600); err != nil {
			return result, err
		}
		args = append(args, "-font", "font.ttf", "-pointsize", strconv.Itoa(v.Int("size", 24, 1, 1000)), "-fill", colorValue(v, "color", "#ffffff", false), "-gravity", "NorthWest", "-annotate", fmt.Sprintf("+%d+%d", v.Int("x", 10, 0, 32768), v.Int("y", 10, 0, 32768)), "@annotation.txt")
	case "image.join":
		v.Check(len(paths) > 1, "join requires additional files")
		args = nil
		for _, path := range paths {
			args = append(args, path+"[0]")
		}
		axis := v.Enum("axis", "horizontal", "horizontal", "vertical")
		flag := "+append"
		if axis == "vertical" {
			flag = "-append"
		}
		args = append(args, flag)
	case "image.contact-sheet":
		sub = "montage"
		args = nil
		for _, path := range paths {
			args = append(args, path+"[0]")
		}
		cols := v.Int("columns", 4, 1, 20)
		w := v.Int("width", 200, 1, 2000)
		h := v.Int("height", 150, 1, 2000)
		gap := v.Int("gap", 4, 0, 100)
		v.Check(int64(cols*(w+2*gap))*int64(((len(paths)+cols-1)/cols)*(h+2*gap)) <= 25_000_000, "contact sheet exceeds 25 million pixels")
		args = append(args, "-tile", fmt.Sprintf("%dx", cols), "-geometry", fmt.Sprintf("%dx%d+%d+%d", w, h, gap, gap), "-background", "none")
	case "image.canvas":
		w := v.Int("width", 640, 1, 32768)
		h := v.Int("height", 480, 1, 32768)
		v.Check(int64(w)*int64(h) <= 25_000_000, "canvas exceeds 25 million pixels")
		args = append(args, "-background", colorValue(v, "background", "transparent", true), "-gravity", "center", "-extent", fmt.Sprintf("%dx%d", w, h))
	case "image.flatten":
		background := colorValue(v, "background", "#ffffff", false)
		v.Check(len(background) == 7, "background must be an opaque #RRGGBB color")
		args = append(args, "-background", background, "-alpha", "remove", "-alpha", "off")
	case "image.transparent":
		args = append(args, "-alpha", "on", "-fuzz", strconv.Itoa(v.Int("tolerance", 0, 0, 100))+"%", "-transparent", colorValue(v, "color", "#ffffff", false))
	case "image.grayscale":
		args = append(args, "-colorspace", "Gray")
	case "image.blur":
		args = append(args, "-blur", "0x"+strconv.Itoa(v.Int("sigma", 2, 1, 100)))
	case "image.sharpen":
		args = append(args, "-sharpen", "0x"+strconv.Itoa(v.Int("sigma", 1, 1, 20)))
	case "image.adjust":
		args = append(args, "-modulate", fmt.Sprintf("%d,%d,%d", v.Int("brightness", 100, 0, 300), v.Int("saturation", 100, 0, 300), v.Int("hue", 100, 0, 200)))
	case "image.difference":
		v.Check(len(paths) == 2, "difference requires two images")
		if v.Err != nil {
			return result, v.Err
		}
		v.Check(infos[0]["width"] == infos[1]["width"] && infos[0]["height"] == infos[1]["height"], "images must have equal dimensions")
		args = append(args, paths[1]+"[0]", "-compose", "difference", "-composite")
	case "image.gif-create":
		v.Check(format == "gif", "output must be GIF")
		v.Check(len(paths) > 1, "provide at least two images")
		args = []string{"-delay", strconv.Itoa(v.Int("delay", 20, 1, 6000))}
		for _, path := range paths {
			args = append(args, path+"[0]")
		}
		args = append(args, "-loop", strconv.Itoa(v.Int("loops", 0, 0, 10000)), "-layers", "Optimize")
	case "image.gif-split":
		v.Check(formatOf(inputs[0]) == "gif", "input must be GIF")
		directory = true
		args = []string{paths[0], "-coalesce", "+repage"}
	case "image.icon":
		v.Check(format == "ico", "output must be ICO")
		args = append(args, "-background", "none", "-resize", "256x256", "-gravity", "center", "-extent", "256x256", "-define", "icon:auto-resize=256,128,64,48,32,16")
	case "image.strip-metadata":
		args = append(args, "-strip")
	default:
		return result, toolrun.Invalid("unsupported image operation")
	}
	if v.Err != nil {
		return result, v.Err
	}
	if directory {
		outputs, err := toolrun.Directory(ctx, output, func(dir string) error {
			if _, err := run(sub, append(args, "png:frame-%05d.png")...); err != nil {
				return err
			}
			matches, err := filepath.Glob(filepath.Join(work, "frame-*.png"))
			if err != nil {
				return err
			}
			for _, path := range matches {
				if err := toolrun.CopyFile(path, filepath.Join(dir, filepath.Base(path))); err != nil {
					return err
				}
			}
			return nil
		})
		result.Outputs = outputs
		return result, err
	}
	if format == "jpeg" {
		args = append(args, "-background", "white", "-alpha", "remove", "-alpha", "off")
	}
	generated := "result." + format
	args = append(args, format+":"+generated)
	if err := toolrun.Artifact(ctx, output, overwrite, func(path string) error {
		if _, err := run(sub, args...); err != nil {
			return err
		}
		data, err := info(format + ":" + generated)
		if err != nil {
			return err
		}
		result.Data = data
		return toolrun.CopyFile(filepath.Join(work, generated), path)
	}); err != nil {
		return result, err
	}
	result.Outputs = []string{output}
	return result, nil
}
