package raster_test

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fu9zhou/finish-bit/pkg/app"
	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func TestImagesWithManagedTools(t *testing.T) {
	home := os.Getenv("FINISHBIT_TEST_HOME")
	if home == "" {
		t.Skip("set FINISHBIT_TEST_HOME for real image acceptance tests")
	}
	service, err := app.New(app.Config{Root: home})
	if err != nil {
		t.Fatal(err)
	}
	status, err := service.PackageInfo("imagemagick")
	if err != nil || status.Installed == nil {
		t.Fatal("managed ImageMagick is required", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	dir := filepath.Join(t.TempDir(), "图片 ' [case]")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	input := filepath.Join(dir, "source.png")
	second := filepath.Join(dir, "second.png")
	logo := filepath.Join(dir, "watermark.png")
	makePNG := func(path string, w, h int, other bool) {
		t.Helper()
		img := image.NewNRGBA(image.Rect(0, 0, w, h))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				c := color.NRGBA{uint8(x * 3), uint8(y * 4), 40, 255}
				if other {
					c = color.NRGBA{10, 200, 80, 255}
				}
				if x == 0 && y == 0 {
					c.A = 0
				}
				img.SetNRGBA(x, y, c)
			}
		}
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, img); err != nil {
			t.Fatal(err)
		}
		_ = f.Close()
	}
	makePNG(input, 80, 60, false)
	makePNG(second, 80, 60, true)
	makePNG(logo, 20, 10, true)
	font := os.Getenv("FINISHBIT_TEST_FONT")
	if font == "" {
		font = filepath.Join(os.Getenv("WINDIR"), "Fonts", "arial.ttf")
	}
	if _, err := os.Stat(font); err != nil {
		t.Fatal("set FINISHBIT_TEST_FONT to a local font", err)
	}
	// Hand-built EXIF orientation=6 verifies actual orientation handling, not just the existence of a command.
	oriented := filepath.Join(dir, "rotated.jpg")
	source, _ := os.Open(input)
	pixels, err := png.Decode(source)
	_ = source.Close()
	if err != nil {
		t.Fatal(err)
	}
	var jpegData bytes.Buffer
	_ = jpeg.Encode(&jpegData, pixels, nil)
	exif := []byte{'E', 'x', 'i', 'f', 0, 0, 'I', 'I', 42, 0, 8, 0, 0, 0, 1, 0, 0x12, 0x01, 3, 0, 1, 0, 0, 0, 6, 0, 0, 0, 0, 0, 0, 0}
	encoded := append([]byte{0xff, 0xd8, 0xff, 0xe1, 0, byte(len(exif) + 2)}, exif...)
	encoded = append(encoded, jpegData.Bytes()[2:]...)
	_ = os.WriteFile(oriented, encoded, 0600)
	run := func(t *testing.T, id string, inputs []string, opts map[string]any) operation.Result {
		t.Helper()
		result, err := service.Execute(ctx, id, operation.Request{Inputs: inputs, Options: opts})
		if err != nil {
			t.Fatalf("%s: %v %#v", id, err, operation.AsError(err).Details)
		}
		return result
	}
	readPNG := func(t *testing.T, path string) image.Image {
		t.Helper()
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		img, err := png.Decode(f)
		if err != nil {
			t.Fatal(err)
		}
		return img
	}
	tests := []struct {
		id     string
		inputs []string
		opts   map[string]any
		ext    string
		w, h   int
	}{
		{"image.formats", []string{}, nil, "", 0, 0},
		{"image.info", nil, map[string]any{"engine": "imagemagick"}, "", 80, 60},
		{"image.convert", nil, map[string]any{"engine": "imagemagick"}, "webp", 80, 60},
		{"image.resize", nil, map[string]any{"engine": "imagemagick", "width": 40}, "png", 40, 30},
		{"image.crop", nil, map[string]any{"width": 30, "height": 20, "x": 5, "y": 7}, "png", 30, 20},
		{"image.rotate", nil, nil, "png", 60, 80},
		{"image.flip", nil, nil, "png", 80, 60},
		{"image.orient", []string{oriented}, nil, "png", 60, 80},
		{"image.compress", nil, map[string]any{"quality": 60}, "jpeg", 80, 60},
		{"image.watermark", []string{input, logo}, nil, "png", 80, 60},
		{"image.annotate", nil, map[string]any{"font": font, "text": "FinishBit 100%", "size": 12, "x": 0, "y": 0}, "png", 80, 60},
		{"image.join", nil, map[string]any{"files": []string{second}}, "png", 160, 60},
		{"image.contact-sheet", nil, map[string]any{"files": []string{second, input, second}, "columns": 2, "width": 40, "height": 30, "gap": 0}, "png", 80, 60},
		{"image.canvas", nil, map[string]any{"width": 100, "height": 80}, "png", 100, 80},
		{"image.flatten", nil, nil, "png", 80, 60},
		{"image.transparent", []string{second}, map[string]any{"color": "#0ac850"}, "png", 80, 60},
		{"image.grayscale", nil, nil, "png", 80, 60},
		{"image.blur", nil, nil, "png", 80, 60},
		{"image.sharpen", nil, nil, "png", 80, 60},
		{"image.adjust", nil, map[string]any{"brightness": 120}, "png", 80, 60},
		{"image.difference", []string{input, input}, nil, "png", 80, 60},
		{"image.gif-create", nil, map[string]any{"files": []string{second}}, "gif", 80, 60},
		{"image.gif-split", []string{filepath.Join(dir, "image.gif-create.gif")}, map[string]any{"output": filepath.Join(dir, "frames")}, "", 0, 0},
		{"image.icon", nil, nil, "ico", 0, 0},
		{"image.strip-metadata", nil, nil, "png", 80, 60},
	}
	for _, test := range tests {
		t.Run(test.id, func(t *testing.T) {
			inputs := test.inputs
			if inputs == nil {
				inputs = []string{input}
			}
			opts := test.opts
			if opts == nil {
				opts = map[string]any{}
			}
			if test.ext != "" {
				opts["output"] = filepath.Join(dir, test.id+"."+test.ext)
			}
			result := run(t, test.id, inputs, opts)
			if test.w > 0 {
				if result.Data["width"] != test.w || result.Data["height"] != test.h {
					t.Fatalf("unexpected dimensions: %v", result.Data)
				}
			}
			for _, path := range result.Outputs {
				info, err := os.Stat(path)
				if err != nil || info.Size() == 0 {
					t.Fatalf("missing/empty output %s: %v", path, err)
				}
				run(t, "image.info", []string{path}, map[string]any{"engine": "imagemagick"})
			}
			switch test.id {
			case "image.formats":
				if !strings.Contains(result.Data["report"].(string), "WEBP") {
					t.Fatal("WebP support missing")
				}
			case "image.crop":
				got := readPNG(t, result.Outputs[0])
				original := readPNG(t, input)
				if color.NRGBAModel.Convert(got.At(0, 0)) != color.NRGBAModel.Convert(original.At(5, 7)) {
					t.Fatal("crop picked wrong pixels")
				}
			case "image.flatten":
				got := readPNG(t, result.Outputs[0])
				r, g, b, a := got.At(0, 0).RGBA()
				if a != 65535 || r != 65535 || g != 65535 || b != 65535 {
					t.Fatal("transparent pixel did not flatten onto white")
				}
			case "image.transparent":
				got := readPNG(t, result.Outputs[0])
				_, _, _, alpha := got.At(2, 2).RGBA()
				if alpha != 0 {
					t.Fatal("color was not made transparent")
				}
			case "image.grayscale":
				got := readPNG(t, result.Outputs[0])
				r, g, b, _ := got.At(10, 20).RGBA()
				if r != g || g != b {
					t.Fatal("pixel is not gray")
				}
			case "image.difference":
				got := readPNG(t, result.Outputs[0])
				r, g, b, _ := got.At(10, 20).RGBA()
				if r != 0 || g != 0 || b != 0 {
					t.Fatal("identical images have a nonzero difference")
				}
			case "image.gif-split":
				if len(result.Outputs) != 2 {
					t.Fatalf("GIF frame count: %v", result.Outputs)
				}
			case "image.icon":
				if result.Data["frames"].(int) < 5 {
					t.Fatal("icon must contain multiple sizes")
				}
			}
		})
	}
	t.Run("additional-formats-and-orientation", func(t *testing.T) {
		for _, format := range []string{"tiff", "bmp", "avif", "png"} {
			path := filepath.Join(dir, "converted."+format)
			run(t, "image.convert", []string{input}, map[string]any{"engine": "imagemagick", "output": path})
			read := run(t, "image.info", []string{path}, map[string]any{"engine": "imagemagick"})
			if read.Data["width"] != 80 {
				t.Fatal("format conversion changed width")
			}
		}
		result := run(t, "image.resize", []string{oriented}, map[string]any{"engine": "imagemagick", "width": 30, "output": filepath.Join(dir, "oriented-resized.png")})
		if result.Data["width"] != 30 || result.Data["height"] != 40 {
			t.Fatal("resize did not apply EXIF orientation")
		}
	})
	t.Run("failure-preserves-output", func(t *testing.T) {
		path := filepath.Join(dir, "preserve.png")
		_ = os.WriteFile(path, []byte("original"), 0600)
		for _, options := range []map[string]any{{"output": path}, {"output": path, "width": 1000, "overwrite": true}, {"output": path, "width": "not-a-number", "overwrite": true}} {
			if _, err := service.Execute(ctx, "image.crop", operation.Request{Inputs: []string{input}, Options: options}); err == nil {
				t.Fatal("invalid request succeeded")
			}
			raw, _ := os.ReadFile(path)
			if string(raw) != "original" {
				t.Fatal("output changed")
			}
		}
		cancelled, cancel := context.WithCancel(ctx)
		cancel()
		if _, err := service.Execute(cancelled, "image.blur", operation.Request{Inputs: []string{input}, Options: map[string]any{"output": path, "overwrite": true}}); err == nil {
			t.Fatal("cancelled operation succeeded")
		}
	})
}
