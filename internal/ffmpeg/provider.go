// Package ffmpeg exposes task-oriented operations backed by the managed FFmpeg package.
package ffmpeg

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

type Resolver interface {
	Executable(packageName, logicalName string) (string, error)
}

type Provider struct{ resolver Resolver }

const maxDiagnosticBytes = 1 << 20

const truncationMarker = "\n[output truncated]"

type limitedOutput struct {
	mu        sync.Mutex
	builder   strings.Builder
	remaining int
	truncated bool
}

func newLimitedOutput(limit int) *limitedOutput {
	return &limitedOutput{remaining: limit - len(truncationMarker)}
}

func (w *limitedOutput) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	original := len(data)
	if len(data) > w.remaining {
		data = data[:w.remaining]
		w.truncated = true
	}
	if len(data) > 0 {
		_, _ = w.builder.Write(data)
		w.remaining -= len(data)
	}
	return original, nil
}

func (w *limitedOutput) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	output := w.builder.String()
	if w.truncated {
		output += truncationMarker
	}
	return output
}

func Register(registry *operation.Registry, resolver Resolver) error {
	provider := &Provider{resolver: resolver}
	for _, capability := range []operation.Capability{
		{Definition: operation.Definition{ID: "video.trim", Summary: "Trim a video by start time and duration", Description: "Create a precisely trimmed video using managed FFmpeg.", Aliases: []string{"cut video", "clip video", "视频裁剪", "视频截取"}, Tags: []string{"video", "trim", "ffmpeg"}, Inputs: []operation.Parameter{{Name: "input", Type: operation.TypeString, Description: "Input media file", Required: true}}, Options: []operation.Parameter{{Name: "start", Type: operation.TypeString, Description: "Start time such as 10s or 00:00:10", Default: "0s"}, {Name: "duration", Type: operation.TypeString, Description: "Output duration", Required: true}, {Name: "output", Type: operation.TypeString, Description: "Output file", Required: true}}, Requirements: []operation.Requirement{{Package: "ffmpeg"}}, Source: "ffmpeg"}, Runner: operation.Func(provider.trim)},
		{Definition: operation.Definition{ID: "video.compress", Summary: "Compress a video toward a target file size", Description: "Use two-pass H.264 encoding to approach a target size in MiB.", Aliases: []string{"shrink video", "reduce video size", "压缩视频"}, Tags: []string{"video", "compress", "ffmpeg"}, Inputs: []operation.Parameter{{Name: "input", Type: operation.TypeString, Description: "Input video file", Required: true}}, Options: []operation.Parameter{{Name: "target-mb", Type: operation.TypeInteger, Description: "Approximate target size in MiB", Default: 20}, {Name: "audio-kbps", Type: operation.TypeInteger, Description: "Audio bitrate in kbit/s", Default: 128}, {Name: "output", Type: operation.TypeString, Description: "Output MP4 file", Required: true}}, Requirements: []operation.Requirement{{Package: "ffmpeg"}}, Source: "ffmpeg"}, Runner: operation.Func(provider.compress)},
		{Definition: operation.Definition{ID: "audio.extract", Summary: "Extract audio from media", Description: "Extract or transcode the audio stream from a media file.", Aliases: []string{"video to audio", "提取音频"}, Tags: []string{"audio", "video", "ffmpeg"}, Inputs: []operation.Parameter{{Name: "input", Type: operation.TypeString, Description: "Input media file", Required: true}}, Options: []operation.Parameter{{Name: "format", Type: operation.TypeString, Description: "mp3, m4a, wav, or flac", Default: "mp3"}, {Name: "output", Type: operation.TypeString, Description: "Output audio file", Required: true}}, Requirements: []operation.Requirement{{Package: "ffmpeg"}}, Source: "ffmpeg"}, Runner: operation.Func(provider.extractAudio)},
	} {
		if err := registry.Register(capability); err != nil {
			return err
		}
	}
	return provider.registerExtended(registry)
}

func requiredOutput(request operation.Request) (string, error) {
	output, err := operation.StringOption(request, "output", "")
	if err != nil {
		return "", err
	}
	if output == "" {
		return "", &operation.Error{Code: operation.CodeInvalidInput, Message: "option \"output\" is required"}
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil && filepath.Dir(output) != "." {
		return "", fmt.Errorf("create output directory: %w", err)
	}
	return output, nil
}

func (p *Provider) command(ctx context.Context, args ...string) error {
	binary, err := p.resolver.Executable("ffmpeg", "ffmpeg")
	if err != nil {
		return err
	}
	command := exec.CommandContext(ctx, binary, args...)
	output := newLimitedOutput(maxDiagnosticBytes)
	command.Stdout = output
	command.Stderr = output
	err = command.Run()
	if err != nil {
		return &operation.Error{Code: operation.CodeExecutionFailed, Message: "FFmpeg operation failed", Details: map[string]any{"output": strings.TrimSpace(output.String())}, Err: err}
	}
	return nil
}

func (p *Provider) trim(ctx context.Context, request operation.Request) (operation.Result, error) {
	if err := operation.RequireInputs(request, 1); err != nil {
		return operation.Result{}, err
	}
	start, err := operation.StringOption(request, "start", "0s")
	if err != nil {
		return operation.Result{}, err
	}
	duration, err := operation.StringOption(request, "duration", "")
	if err != nil {
		return operation.Result{}, err
	}
	if duration == "" {
		return operation.Result{}, &operation.Error{Code: operation.CodeInvalidInput, Message: "option \"duration\" is required"}
	}
	output, err := requiredOutput(request)
	if err != nil {
		return operation.Result{}, err
	}
	if err := p.command(ctx, "-hide_banner", "-loglevel", "error", "-y", "-i", request.Inputs[0], "-ss", start, "-t", duration, "-map", "0:v:0", "-map", "0:a?", "-c:v", "libx264", "-c:a", "aac", "-movflags", "+faststart", output); err != nil {
		return operation.Result{}, err
	}
	return operation.Result{Operation: "video.trim", Outputs: []string{output}}, nil
}

var durationPattern = regexp.MustCompile(`Duration: (\d+):(\d+):(\d+(?:\.\d+)?)`)

func (p *Provider) duration(ctx context.Context, input string) (float64, error) {
	binary, err := p.resolver.Executable("ffmpeg", "ffmpeg")
	if err != nil {
		return 0, err
	}
	command := exec.CommandContext(ctx, binary, "-hide_banner", "-i", input)
	output := newLimitedOutput(maxDiagnosticBytes)
	command.Stdout = output
	command.Stderr = output
	_ = command.Run()
	match := durationPattern.FindStringSubmatch(output.String())
	if len(match) != 4 {
		return 0, &operation.Error{Code: operation.CodeExecutionFailed, Message: "could not determine video duration", Details: map[string]any{"output": strings.TrimSpace(output.String())}}
	}
	hours, _ := strconv.ParseFloat(match[1], 64)
	minutes, _ := strconv.ParseFloat(match[2], 64)
	seconds, _ := strconv.ParseFloat(match[3], 64)
	return hours*3600 + minutes*60 + seconds, nil
}

func (p *Provider) compress(ctx context.Context, request operation.Request) (operation.Result, error) {
	if err := operation.RequireInputs(request, 1); err != nil {
		return operation.Result{}, err
	}
	targetMB, err := operation.IntOption(request, "target-mb", 20)
	if err != nil {
		return operation.Result{}, err
	}
	audioKbps, err := operation.IntOption(request, "audio-kbps", 128)
	if err != nil {
		return operation.Result{}, err
	}
	if targetMB < 1 || audioKbps < 0 {
		return operation.Result{}, &operation.Error{Code: operation.CodeInvalidInput, Message: "target-mb must be positive and audio-kbps cannot be negative"}
	}
	output, err := requiredOutput(request)
	if err != nil {
		return operation.Result{}, err
	}
	duration, err := p.duration(ctx, request.Inputs[0])
	if err != nil {
		return operation.Result{}, err
	}
	videoKbps := int((float64(targetMB)*8192/duration)*0.97) - audioKbps
	if videoKbps < 100 {
		return operation.Result{}, &operation.Error{Code: operation.CodeInvalidInput, Message: "target size is too small for this video's duration"}
	}
	passDir, err := os.MkdirTemp("", "finishbit-ffmpeg-pass-")
	if err != nil {
		return operation.Result{}, err
	}
	defer os.RemoveAll(passDir)
	passlog := filepath.Join(passDir, "pass")
	nullDevice := "/dev/null"
	if os.PathSeparator == '\\' {
		nullDevice = "NUL"
	}
	bitrate := strconv.Itoa(videoKbps) + "k"
	common := []string{"-hide_banner", "-loglevel", "error", "-y", "-i", request.Inputs[0], "-c:v", "libx264", "-b:v", bitrate, "-passlogfile", passlog}
	first := append(append([]string{}, common...), "-pass", "1", "-an", "-f", "null", nullDevice)
	if err := p.command(ctx, first...); err != nil {
		return operation.Result{}, err
	}
	second := append(append([]string{}, common...), "-pass", "2", "-c:a", "aac", "-b:a", strconv.Itoa(audioKbps)+"k", "-movflags", "+faststart", output)
	if err := p.command(ctx, second...); err != nil {
		return operation.Result{}, err
	}
	return operation.Result{Operation: "video.compress", Outputs: []string{output}, Data: map[string]any{"target_mb": targetMB, "video_kbps": videoKbps, "audio_kbps": audioKbps}}, nil
}

func (p *Provider) extractAudio(ctx context.Context, request operation.Request) (operation.Result, error) {
	if err := operation.RequireInputs(request, 1); err != nil {
		return operation.Result{}, err
	}
	format, err := operation.StringOption(request, "format", "mp3")
	if err != nil {
		return operation.Result{}, err
	}
	output, err := requiredOutput(request)
	if err != nil {
		return operation.Result{}, err
	}
	codec := map[string][]string{"mp3": {"-c:a", "libmp3lame", "-q:a", "2"}, "m4a": {"-c:a", "aac", "-b:a", "192k"}, "wav": {"-c:a", "pcm_s16le"}, "flac": {"-c:a", "flac"}}[format]
	if codec == nil {
		return operation.Result{}, &operation.Error{Code: operation.CodeInvalidInput, Message: fmt.Sprintf("unsupported audio format %q", format)}
	}
	args := []string{"-hide_banner", "-loglevel", "error", "-y", "-i", request.Inputs[0], "-vn"}
	args = append(args, codec...)
	args = append(args, output)
	if err := p.command(ctx, args...); err != nil {
		return operation.Result{}, err
	}
	return operation.Result{Operation: "audio.extract", Outputs: []string{output}}, nil
}
