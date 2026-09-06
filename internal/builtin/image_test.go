package builtin

import (
	"bytes"
	"context"
	"encoding/binary"
	"github.com/fu9zhou/finish-bit/pkg/operation"
	"hash/crc32"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func imageFixture(t *testing.T) string {
	t.Helper()
	im := image.NewNRGBA(image.Rect(0, 0, 4, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 4; x++ {
			im.SetNRGBA(x, y, color.NRGBA{R: 255, A: 255})
		}
	}
	path := filepath.Join(t.TempDir(), "input.png")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(file, im); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}
func decodeFixture(t *testing.T, path string) image.Image {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	im, _, err := image.Decode(file)
	if err != nil {
		t.Fatal(err)
	}
	return im
}

func TestImageOperations(t *testing.T) {
	input := imageFixture(t)
	info := runTestOperation(t, "image.info", operation.Request{Inputs: []string{input}})
	if info.Data["width"] != 4 || info.Data["height"] != 2 || info.Data["format"] != "png" {
		t.Fatal(info)
	}
	output := filepath.Join(t.TempDir(), "nested", "small.png")
	runTestOperation(t, "image.resize", operation.Request{Inputs: []string{input}, Options: map[string]any{"width": 2, "height": 2, "output": output}})
	actual := decodeFixture(t, output)
	if actual.Bounds().Dx() != 2 || actual.Bounds().Dy() != 1 {
		t.Fatal(actual.Bounds())
	}
	r, g, b, a := actual.At(0, 0).RGBA()
	if r != 65535 || g != 0 || b != 0 || a != 65535 {
		t.Fatalf("wrong pixel %d %d %d %d", r, g, b, a)
	}
	larger := filepath.Join(t.TempDir(), "larger.png")
	runTestOperation(t, "image.resize", operation.Request{Inputs: []string{input}, Options: map[string]any{"width": 8, "output": larger}})
	if decodeFixture(t, larger).Bounds().Dx() != 4 {
		t.Fatal("unexpected upscale")
	}
	runTestOperation(t, "image.resize", operation.Request{Inputs: []string{input}, Options: map[string]any{"width": 8, "upscale": true, "output": larger, "overwrite": true}})
	if decodeFixture(t, larger).Bounds().Dx() != 8 {
		t.Fatal("upscale failed")
	}
	jpegPath := filepath.Join(t.TempDir(), "converted.jpg")
	runTestOperation(t, "image.convert", operation.Request{Inputs: []string{input}, Options: map[string]any{"output": jpegPath}})
	if decodeFixture(t, jpegPath).Bounds().Dx() != 4 {
		t.Fatal("JPEG conversion failed")
	}
}

func TestImageTransparencyAndSampling(t *testing.T) {
	source := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	source.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	source.SetNRGBA(1, 0, color.NRGBA{B: 255, A: 255})
	scaled, err := resizeBilinear(context.Background(), source, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	r, _, b, _ := scaled.At(0, 0).RGBA()
	if r < 32000 || r > 33500 || b < 32000 || b > 33500 {
		t.Fatalf("bad interpolation: %d %d", r, b)
	}
	input := filepath.Join(t.TempDir(), "transparent.png")
	var data bytes.Buffer
	png.Encode(&data, image.NewNRGBA(image.Rect(0, 0, 2, 2)))
	os.WriteFile(input, data.Bytes(), 0600)
	output := filepath.Join(t.TempDir(), "white.jpg")
	runTestOperation(t, "image.convert", operation.Request{Inputs: []string{input}, Options: map[string]any{"output": output}})
	r, g, b, _ := decodeFixture(t, output).At(0, 0).RGBA()
	if r < 65000 || g < 65000 || b < 65000 {
		t.Fatal("transparency did not become white")
	}
}

func TestImageRejectsInvalidAndPreservesDestination(t *testing.T) {
	input := imageFixture(t)
	output := filepath.Join(t.TempDir(), "out.png")
	os.WriteFile(output, []byte("keep"), 0600)
	cap, _ := testRegistry(t).Get("image.resize")
	for _, options := range []map[string]any{{"output": output}, {"output": output, "width": -1}, {"output": output, "width": 2}, {"output": output, "width": 2, "quality": 101}, {"output": output, "width": 40000}} {
		if _, err := cap.Runner.Run(context.Background(), operation.Request{Inputs: []string{input}, Options: options}); err == nil {
			t.Fatal("invalid request succeeded")
		}
		data, _ := os.ReadFile(output)
		if string(data) != "keep" {
			t.Fatal("existing output changed")
		}
	}
	data, _ := os.ReadFile(input)
	binary.BigEndian.PutUint32(data[16:20], 40000)
	binary.BigEndian.PutUint32(data[29:33], crc32.ChecksumIEEE(data[12:29]))
	oversized := filepath.Join(t.TempDir(), "oversized.png")
	os.WriteFile(oversized, data, 0600)
	if _, err := runImageInfo(context.Background(), operation.Request{Inputs: []string{oversized}}); err == nil {
		t.Fatal("oversized decoded dimensions accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := runImageConvert(ctx, operation.Request{Inputs: []string{input}, Options: map[string]any{"output": output, "overwrite": true}}); err == nil {
		t.Fatal("canceled image request succeeded")
	}
}

func TestArtifactFailurePreservesDestination(t *testing.T) {
	path := filepath.Join(t.TempDir(), "output")
	os.WriteFile(path, []byte("original"), 0600)
	err := writeArtifact(path, true, func(w io.Writer) error { w.Write([]byte("partial")); return io.ErrUnexpectedEOF })
	if err == nil {
		t.Fatal("write unexpectedly succeeded")
	}
	data, _ := os.ReadFile(path)
	if string(data) != "original" {
		t.Fatal("original truncated")
	}
}
