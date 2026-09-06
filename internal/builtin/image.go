package builtin

import (
	"bytes"
	"context"
	"fmt"
	"github.com/fu9zhou/finish-bit/pkg/operation"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
)

const maxImagePixels = 25_000_000
const maxImageDimension = 32768

func registerImage(registry *operation.Registry) error {
	input := []operation.Parameter{param("input", "PNG or JPEG file path", true)}
	output := []operation.Parameter{param("output", "Destination PNG or JPEG file", true), option("format", operation.TypeString, "png or jpeg; default inferred from output extension", ""), option("quality", operation.TypeInteger, "JPEG quality from 1 to 100", 90), option("overwrite", operation.TypeBoolean, "Replace an existing destination", false)}
	return registerAll(registry,
		operation.Capability{Definition: operation.Definition{ID: "image.info", Summary: "Inspect PNG or JPEG dimensions and format", Description: "Read encoded pixel dimensions without applying EXIF orientation.", Aliases: []string{"图片信息", "image size"}, Tags: []string{"image", "metadata"}, Inputs: input, Source: "core"}, Runner: operation.Func(runImageInfo)},
		operation.Capability{Definition: operation.Definition{ID: "image.convert", Summary: "Convert between PNG and JPEG", Description: "Encode image pixels as PNG or JPEG; JPEG uses a white background for transparency. Metadata is not copied and EXIF orientation is not applied.", Aliases: []string{"图片格式转换", "convert image"}, Tags: []string{"image", "convert"}, Inputs: input, Options: output, Source: "core"}, Runner: operation.Func(runImageConvert)},
		operation.Capability{Definition: operation.Definition{ID: "image.resize", Summary: "Resize an image while preserving aspect ratio", Description: "Fit within width and/or height using bilinear sampling. Metadata is not copied and EXIF orientation is not applied.", Aliases: []string{"缩小图片", "图片缩放", "resize image"}, Tags: []string{"image", "resize"}, Inputs: input, Options: append([]operation.Parameter{option("width", operation.TypeInteger, "Maximum output width; 0 means unconstrained", 0), option("height", operation.TypeInteger, "Maximum output height; 0 means unconstrained", 0), option("upscale", operation.TypeBoolean, "Allow enlarging smaller images", false)}, output...), Source: "core"}, Runner: operation.Func(runImageResize)},
	)
}

func imageError(message string) error {
	return &operation.Error{Code: operation.CodeInvalidInput, Message: message}
}
func validImageSize(width, height int) bool {
	return width > 0 && height > 0 && width <= maxImageDimension && height <= maxImageDimension && width <= maxImagePixels/height
}

func imageInput(request operation.Request) ([]byte, image.Config, string, error) {
	if err := operation.RequireInputs(request, 1); err != nil {
		return nil, image.Config{}, "", err
	}
	file, err := os.Open(request.Inputs[0])
	if err != nil {
		return nil, image.Config{}, "", imageError("cannot open image file: " + err.Error())
	}
	defer file.Close()
	data, err := readBounded(file, "image")
	if err != nil {
		return nil, image.Config{}, "", err
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, image.Config{}, "", imageError("input must be a valid PNG or JPEG image")
	}
	if format != "png" && format != "jpeg" {
		return nil, image.Config{}, "", imageError("supported image formats are png and jpeg")
	}
	if !validImageSize(config.Width, config.Height) {
		return nil, image.Config{}, "", imageError("image exceeds 25 million pixels or 32768 pixels per dimension")
	}
	return data, config, format, nil
}

func runImageInfo(_ context.Context, request operation.Request) (operation.Result, error) {
	data, config, format, err := imageInput(request)
	if err != nil {
		return operation.Result{}, err
	}
	return operation.Result{Operation: "image.info", Data: map[string]any{"width": config.Width, "height": config.Height, "format": format, "bytes": len(data)}}, nil
}

func runImageConvert(ctx context.Context, request operation.Request) (operation.Result, error) {
	return transformImage(ctx, request, false)
}
func runImageResize(ctx context.Context, request operation.Request) (operation.Result, error) {
	return transformImage(ctx, request, true)
}

func transformImage(ctx context.Context, request operation.Request, resize bool) (operation.Result, error) {
	output, err := outputOption(request)
	if err != nil {
		return operation.Result{}, err
	}
	if output == "" {
		return operation.Result{}, imageError("output path is required")
	}
	format, err := operation.StringOption(request, "format", "")
	if err != nil {
		return operation.Result{}, err
	}
	if format == "" {
		format = strings.TrimPrefix(strings.ToLower(filepath.Ext(output)), ".")
	}
	if format == "jpg" {
		format = "jpeg"
	}
	if format != "png" && format != "jpeg" {
		return operation.Result{}, imageError("output format must be png or jpeg")
	}
	quality, err := operation.IntOption(request, "quality", 90)
	if err != nil {
		return operation.Result{}, err
	}
	if quality < 1 || quality > 100 {
		return operation.Result{}, imageError("quality must be between 1 and 100")
	}
	overwrite, err := operation.BoolOption(request, "overwrite", false)
	if err != nil {
		return operation.Result{}, err
	}
	data, config, _, err := imageInput(request)
	if err != nil {
		return operation.Result{}, err
	}
	width, height := config.Width, config.Height
	if resize {
		w, err := operation.IntOption(request, "width", 0)
		if err != nil {
			return operation.Result{}, err
		}
		h, err := operation.IntOption(request, "height", 0)
		if err != nil {
			return operation.Result{}, err
		}
		if w < 0 || h < 0 || w > maxImageDimension || h > maxImageDimension || (w == 0 && h == 0) {
			return operation.Result{}, imageError("provide width and/or height between 1 and 32768")
		}
		scale := math.Inf(1)
		if w > 0 {
			scale = float64(w) / float64(width)
		}
		if h > 0 {
			scale = math.Min(scale, float64(h)/float64(height))
		}
		upscale, err := operation.BoolOption(request, "upscale", false)
		if err != nil {
			return operation.Result{}, err
		}
		if !upscale {
			scale = math.Min(1, scale)
		}
		width = max(1, int(math.Floor(float64(width)*scale)))
		height = max(1, int(math.Floor(float64(height)*scale)))
		if !validImageSize(width, height) {
			return operation.Result{}, imageError("output exceeds the image size limit")
		}
	}
	if err := ctx.Err(); err != nil {
		return operation.Result{}, err
	}
	source, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return operation.Result{}, imageError("cannot decode image pixels: " + err.Error())
	}
	if width != config.Width || height != config.Height {
		source, err = resizeBilinear(ctx, source, width, height)
		if err != nil {
			return operation.Result{}, err
		}
	}
	if err := ctx.Err(); err != nil {
		return operation.Result{}, err
	}
	err = writeArtifact(output, overwrite, func(writer io.Writer) error {
		if format == "png" {
			return png.Encode(writer, source)
		}
		opaque := image.NewRGBA(source.Bounds())
		draw.Draw(opaque, opaque.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
		draw.Draw(opaque, opaque.Bounds(), source, source.Bounds().Min, draw.Over)
		return jpeg.Encode(writer, opaque, &jpeg.Options{Quality: quality})
	})
	if err != nil {
		return operation.Result{}, fmt.Errorf("write image: %w", err)
	}
	id := "image.convert"
	if resize {
		id = "image.resize"
	}
	return operation.Result{Operation: id, Outputs: []string{output}, Data: map[string]any{"width": width, "height": height, "format": format}}, nil
}

func resizeBilinear(ctx context.Context, source image.Image, width, height int) (image.Image, error) {
	bounds := source.Bounds()
	target := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		sy := math.Max(0, (float64(y)+0.5)*float64(bounds.Dy())/float64(height)-0.5)
		y0 := int(sy)
		y1 := min(y0+1, bounds.Dy()-1)
		fy := sy - float64(y0)
		for x := 0; x < width; x++ {
			sx := math.Max(0, (float64(x)+0.5)*float64(bounds.Dx())/float64(width)-0.5)
			x0 := int(sx)
			x1 := min(x0+1, bounds.Dx()-1)
			fx := sx - float64(x0)
			r0, g0, b0, a0 := source.At(bounds.Min.X+x0, bounds.Min.Y+y0).RGBA()
			r1, g1, b1, a1 := source.At(bounds.Min.X+x1, bounds.Min.Y+y0).RGBA()
			r2, g2, b2, a2 := source.At(bounds.Min.X+x0, bounds.Min.Y+y1).RGBA()
			r3, g3, b3, a3 := source.At(bounds.Min.X+x1, bounds.Min.Y+y1).RGBA()
			mix := func(a, b, c, d uint32) uint8 {
				return uint8(math.Round(((float64(a)*(1-fx)+float64(b)*fx)*(1-fy) + (float64(c)*(1-fx)+float64(d)*fx)*fy) / 257))
			}
			target.SetRGBA(x, y, color.RGBA{R: mix(r0, r1, r2, r3), G: mix(g0, g1, g2, g3), B: mix(b0, b1, b2, b3), A: mix(a0, a1, a2, a3)})
		}
	}
	return target, nil
}
