package ffmpeg

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

type testResolver struct {
	path string
	err  error
}

func (r testResolver) Executable(packageName, logicalName string) (string, error) {
	if packageName != "ffmpeg" || logicalName != "ffmpeg" {
		return "", errors.New("unexpected executable request")
	}
	return r.path, r.err
}

func TestProviderRegistersMediaOperations(t *testing.T) {
	registry := operation.NewRegistry()
	if err := Register(registry, testResolver{err: errors.New("missing")}); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"video.trim", "video.compress", "audio.extract"} {
		if _, ok := registry.Get(id); !ok {
			t.Fatalf("operation %s was not registered", id)
		}
	}
}

func TestProviderValidatesRequestsBeforeExecution(t *testing.T) {
	provider := &Provider{resolver: testResolver{err: errors.New("missing")}}
	tests := []struct {
		name string
		run  func() error
	}{
		{"trim input", func() error { _, err := provider.trim(context.Background(), operation.Request{}); return err }},
		{"trim duration", func() error {
			_, err := provider.trim(context.Background(), operation.Request{Inputs: []string{"input.mp4"}, Options: map[string]any{"output": "out.mp4"}})
			return err
		}},
		{"trim output", func() error {
			_, err := provider.trim(context.Background(), operation.Request{Inputs: []string{"input.mp4"}, Options: map[string]any{"duration": "1s"}})
			return err
		}},
		{"compress target", func() error {
			_, err := provider.compress(context.Background(), operation.Request{Inputs: []string{"input.mp4"}, Options: map[string]any{"target-mb": 0, "output": "out.mp4"}})
			return err
		}},
		{"audio format", func() error {
			_, err := provider.extractAudio(context.Background(), operation.Request{Inputs: []string{"input.mp4"}, Options: map[string]any{"format": "ogg", "output": "out.ogg"}})
			return err
		}},
		{"dependency", func() error {
			_, err := provider.extractAudio(context.Background(), operation.Request{Inputs: []string{"input.mp4"}, Options: map[string]any{"output": "out.mp3"}})
			return err
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.run(); err == nil {
				t.Fatal("invalid request succeeded")
			}
		})
	}
}

func TestMediaOperationsWithFFmpeg(t *testing.T) {
	binary, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg is not installed on this test host")
	}
	directory := t.TempDir()
	input := filepath.Join(directory, "input.mp4")
	generate := exec.Command(binary, "-hide_banner", "-loglevel", "error", "-y", "-f", "lavfi", "-i", "testsrc=size=160x120:rate=15", "-f", "lavfi", "-i", "sine=frequency=880:sample_rate=44100", "-t", "2", "-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "aac", input)
	if output, err := generate.CombinedOutput(); err != nil {
		t.Skipf("local ffmpeg cannot generate the integration fixture: %v: %s", err, output)
	}
	provider := &Provider{resolver: testResolver{path: binary}}

	trimmed := filepath.Join(directory, "outputs", "trimmed.mp4")
	if _, err := provider.trim(context.Background(), operation.Request{Inputs: []string{input}, Options: map[string]any{"start": "0.25s", "duration": "1s", "output": trimmed}}); err != nil {
		t.Fatal(err)
	}
	compressed := filepath.Join(directory, "compressed.mp4")
	if _, err := provider.compress(context.Background(), operation.Request{Inputs: []string{input}, Options: map[string]any{"target-mb": 1, "audio-kbps": 64, "output": compressed}}); err != nil {
		t.Fatal(err)
	}
	audio := filepath.Join(directory, "audio.mp3")
	if _, err := provider.extractAudio(context.Background(), operation.Request{Inputs: []string{input}, Options: map[string]any{"format": "mp3", "output": audio}}); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{trimmed, compressed, audio} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("output %s: %v", path, err)
		}
		if info.Size() == 0 {
			t.Fatalf("output %s is empty", path)
		}
	}
}
