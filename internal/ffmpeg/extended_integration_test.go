package ffmpeg

import (
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

type realResolver map[string]string

func (r realResolver) Executable(pkg, name string) (string, error) {
	path := r[name]
	if path == "" {
		return "", &operation.Error{Code: operation.CodeDependencyMissing, Message: "missing " + pkg}
	}
	return path, nil
}

func TestExtendedMediaWithManagedTools(t *testing.T) {
	ffmpeg := os.Getenv("FINISHBIT_TEST_FFMPEG")
	ffprobe := os.Getenv("FINISHBIT_TEST_FFPROBE")
	if ffmpeg == "" || ffprobe == "" {
		t.Skip("set FINISHBIT_TEST_FFMPEG and FINISHBIT_TEST_FFPROBE to run real media acceptance tests")
	}
	ffmpeg, _ = filepath.Abs(ffmpeg)
	ffprobe, _ = filepath.Abs(ffprobe)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	dir := filepath.Join(t.TempDir(), "素材 ' [case]")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	input := filepath.Join(dir, "input.mp4")
	silent := filepath.Join(dir, "silence.wav")
	subtitle := filepath.Join(dir, "caption's.srt")
	picture := filepath.Join(dir, "logo.png")
	execute := func(t *testing.T, args ...string) {
		t.Helper()
		out, err := exec.CommandContext(ctx, ffmpeg, args...).CombinedOutput()
		if err != nil {
			t.Fatalf("FFmpeg fixture/verification failed: %v\n%s", err, out)
		}
	}
	execute(t, "-v", "error", "-y", "-f", "lavfi", "-i", "testsrc2=size=160x120:rate=10", "-f", "lavfi", "-i", "sine=frequency=880:sample_rate=48000", "-t", "2", "-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "aac", input)
	execute(t, "-v", "error", "-y", "-f", "lavfi", "-i", "aevalsrc=if(lt(t\\,1)\\,0\\,0.2*sin(2*PI*440*t)):s=48000:d=2", silent)
	if err := os.WriteFile(subtitle, []byte("1\n00:00:00,000 --> 00:00:01,500\nFinishBit test\n"), 0600); err != nil {
		t.Fatal(err)
	}
	file, err := os.Create(picture)
	if err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 32, 24))
	for y := 0; y < 24; y++ {
		for x := 0; x < 32; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}
	if err := png.Encode(file, img); err != nil {
		t.Fatal(err)
	}
	_ = file.Close()
	registry := operation.NewRegistry()
	if err := Register(registry, realResolver{"ffmpeg": ffmpeg, "ffprobe": ffprobe}); err != nil {
		t.Fatal(err)
	}
	probe := func(t *testing.T, path string) map[string]any {
		t.Helper()
		out, err := exec.CommandContext(ctx, ffprobe, "-v", "error", "-show_streams", "-show_format", "-of", "json", path).Output()
		if err != nil {
			t.Fatalf("probe %s: %v", path, err)
		}
		var result map[string]any
		if err := json.Unmarshal(out, &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	duration := func(t *testing.T, data map[string]any) float64 {
		t.Helper()
		format, ok := data["format"].(map[string]any)
		if !ok {
			t.Fatal("missing format")
		}
		text, _ := format["duration"].(string)
		d, err := strconv.ParseFloat(text, 64)
		if err != nil {
			t.Fatalf("bad duration %q", text)
		}
		return d
	}
	cases := []struct {
		id, extension string
		inputs        []string
		options       map[string]any
		check         func(*testing.T, operation.Result)
	}{
		{"media.info", "", nil, nil, func(t *testing.T, r operation.Result) {
			if len(r.Data["streams"].([]any)) != 2 {
				t.Fatal("expected video and audio streams")
			}
		}},
		{"video.convert", "mp4", nil, nil, nil},
		{"video.remux", "mkv", nil, nil, nil},
		{"video.resize", "mp4", nil, map[string]any{"width": 80, "height": 60}, nil},
		{"video.crop", "mp4", nil, map[string]any{"width": 80, "height": 60}, nil},
		{"video.rotate", "mp4", nil, nil, nil},
		{"video.flip", "mp4", nil, nil, nil},
		{"video.speed", "mp4", nil, map[string]any{"factor": "2"}, nil},
		{"video.framerate", "mp4", nil, map[string]any{"fps": 5}, nil},
		{"video.concat", "mp4", nil, map[string]any{"files": []string{input}, "width": 160, "height": 120}, nil},
		{"video.thumbnail", "png", nil, map[string]any{"time": "0.5", "width": 80}, nil},
		{"video.frames", "", nil, map[string]any{"count": 3, "interval": "0.5", "width": 80}, func(t *testing.T, r operation.Result) {
			if len(r.Outputs) != 3 {
				t.Fatalf("frames: %v", r.Outputs)
			}
		}},
		{"video.contact-sheet", "png", nil, map[string]any{"columns": 2, "rows": 2, "width": 80, "interval": "0.5"}, nil},
		{"video.gif", "gif", nil, map[string]any{"fps": 5, "width": 80, "duration": "1"}, nil},
		{"video.from-images", "mp4", []string{picture}, map[string]any{"files": []string{picture}, "width": 160, "height": 120, "seconds": "0.5"}, nil},
		{"video.watermark", "mp4", []string{input, picture}, nil, nil},
		{"video.mute", "mp4", nil, nil, nil},
		{"video.replace-audio", "mp4", []string{input, silent}, nil, nil},
		{"video.add-music", "mp4", []string{input, silent}, nil, nil},
		{"subtitle.add", "mp4", []string{input, subtitle}, nil, nil},
		{"subtitle.extract", "srt", []string{filepath.Join(dir, "subtitle.add.mp4")}, nil, func(t *testing.T, r operation.Result) {
			data, err := os.ReadFile(r.Outputs[0])
			if err != nil || !strings.Contains(string(data), "FinishBit test") {
				t.Fatalf("subtitle: %s %v", data, err)
			}
		}},
		{"subtitle.burn", "mp4", []string{input, subtitle}, nil, nil},
		{"audio.convert", "mp3", nil, nil, nil},
		{"audio.trim", "wav", nil, map[string]any{"duration": "0.5"}, nil},
		{"audio.concat", "wav", nil, map[string]any{"files": []string{input}}, nil},
		{"audio.mix", "wav", []string{input, silent}, nil, nil},
		{"audio.volume", "wav", nil, map[string]any{"gain": "0.5"}, nil},
		{"audio.normalize", "wav", nil, nil, nil},
		{"audio.denoise", "wav", nil, nil, nil},
		{"audio.silence-detect", "", []string{silent}, nil, func(t *testing.T, r operation.Result) {
			frames, ok := r.Data["frames"].([]any)
			if !ok {
				t.Fatal("missing silence frames")
			}
			found := false
			for _, f := range frames {
				tags, _ := f.(map[string]any)["tags"].(map[string]any)
				if _, ok := tags["lavfi.silence_end"]; ok {
					found = true
				}
			}
			if !found {
				t.Fatal("expected silence end")
			}
		}},
		{"audio.silence-remove", "wav", []string{silent}, nil, nil},
		{"audio.fade", "wav", nil, nil, nil},
		{"audio.waveform", "png", nil, map[string]any{"width": 320, "height": 80}, nil},
	}
	for _, test := range cases {
		t.Run(test.id, func(t *testing.T) {
			cap, ok := registry.Get(test.id)
			if !ok {
				t.Fatal("missing operation")
			}
			inputs := test.inputs
			if inputs == nil {
				inputs = []string{input}
			}
			options := test.options
			if options == nil {
				options = map[string]any{}
			}
			if test.extension != "" {
				options["output"] = filepath.Join(dir, test.id+"."+test.extension)
			} else if test.id == "video.frames" {
				options["output"] = filepath.Join(dir, "frames")
			}
			result, err := cap.Runner.Run(ctx, operation.Request{Inputs: inputs, Options: options})
			if err != nil {
				t.Fatalf("%v: %#v", err, operation.AsError(err).Details)
			}
			if test.check != nil {
				test.check(t, result)
			}
			for _, path := range result.Outputs {
				info, err := os.Stat(path)
				if err != nil || info.Size() == 0 {
					t.Fatalf("empty output: %s %v", path, err)
				}
				if test.extension != "srt" {
					data := probe(t, path)
					streams := data["streams"].([]any)
					if len(streams) == 0 {
						t.Fatal("output has no stream")
					}
					stream := streams[0].(map[string]any)
					switch test.id {
					case "video.resize", "video.crop":
						if stream["width"] != float64(80) || stream["height"] != float64(60) {
							t.Fatalf("wrong dimensions: %v", stream)
						}
					case "video.rotate":
						if stream["width"] != float64(120) || stream["height"] != float64(160) {
							t.Fatal("rotation did not swap dimensions")
						}
					case "video.mute":
						for _, s := range streams {
							if s.(map[string]any)["codec_type"] == "audio" {
								t.Fatal("audio not removed")
							}
						}
					case "video.speed":
						if d := duration(t, data); d < 0.9 || d > 1.3 {
							t.Fatalf("speed duration: %g", d)
						}
					case "video.concat", "audio.concat":
						if d := duration(t, data); d < 3.8 || d > 4.3 {
							t.Fatalf("concat duration: %g", d)
						}
					case "audio.trim":
						if d := duration(t, data); d < 0.45 || d > 0.55 {
							t.Fatalf("trim duration: %g", d)
						}
					case "audio.silence-remove":
						if d := duration(t, data); d < 0.8 || d > 1.3 {
							t.Fatalf("silence removal duration: %g", d)
						}
					case "video.contact-sheet":
						if stream["width"] != float64(160) || stream["height"] != float64(120) {
							t.Fatal("incorrect tile grid")
						}
					}
					execute(t, "-v", "error", "-i", path, "-f", "null", "-")
				}
			}
		})
	}
	t.Run("protect-output-and-validate", func(t *testing.T) {
		cap, _ := registry.Get("video.resize")
		if _, err := cap.Runner.Run(ctx, operation.Request{Inputs: []string{silent}, Options: map[string]any{"output": filepath.Join(dir, "not-video.mp4")}}); err == nil {
			t.Fatal("video operation accepted audio-only input")
		}
		output := filepath.Join(dir, "keep.mp4")
		sentinel := []byte("original")
		_ = os.WriteFile(output, sentinel, 0600)
		for _, opts := range []map[string]any{{"output": output}, {"output": output, "width": 3, "overwrite": true}, {"output": output, "width": "NaN", "overwrite": true}} {
			if _, err := cap.Runner.Run(ctx, operation.Request{Inputs: []string{input}, Options: opts}); err == nil {
				t.Fatal("invalid operation succeeded")
			}
			got, _ := os.ReadFile(output)
			if string(got) != string(sentinel) {
				t.Fatal("destination damaged")
			}
		}
		broken := filepath.Join(dir, "broken.mp4")
		_ = os.WriteFile(broken, []byte("invalid media"), 0600)
		if _, err := cap.Runner.Run(ctx, operation.Request{Inputs: []string{broken}, Options: map[string]any{"output": output, "overwrite": true}}); err == nil {
			t.Fatal("corrupt media succeeded")
		}
		got, _ := os.ReadFile(output)
		if string(got) != string(sentinel) {
			t.Fatal("failed tool damaged output")
		}
		cancelled, cancel := context.WithCancel(ctx)
		cancel()
		if _, err := cap.Runner.Run(cancelled, operation.Request{Inputs: []string{input}, Options: map[string]any{"output": output, "overwrite": true}}); err == nil {
			t.Fatal("cancelled operation succeeded")
		}
	})
}
