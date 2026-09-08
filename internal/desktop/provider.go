// Package desktop supplies explicit, bounded local OS operations.
package desktop

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/fu9zhou/finish-bit/internal/toolrun"
	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func Register(reg *operation.Registry, resolver toolrun.Resolver) error {
	defs := []operation.Definition{
		{ID: "speech.voices", Summary: "List installed offline Windows speech voices", Aliases: []string{"本地语音列表"}, Tags: []string{"speech", "offline"}, Source: "windows"},
		{ID: "speech.synthesize", Summary: "Synthesize text into a WAV file using an installed offline Windows voice", Aliases: []string{"文本转语音"}, Tags: []string{"speech", "offline"}, Source: "windows", Inputs: []operation.Parameter{toolrun.Param("text", "Literal UTF-8 text up to 32768 bytes", true)}, Options: append([]operation.Parameter{toolrun.Option("voice", operation.TypeString, "Installed voice name; blank selects the OS default", ""), toolrun.Option("rate", operation.TypeInteger, "Speaking rate -10 to 10", 0), toolrun.Option("volume", operation.TypeInteger, "Volume 0 to 100", 100)}, toolrun.OutputOptions()...)},
		{ID: "screen.record", Summary: "Record the Windows desktop or an explicit rectangle into a silent MP4", Aliases: []string{"本地屏幕录制"}, Tags: []string{"screen", "video", "offline"}, Source: "ffmpeg", Requirements: []operation.Requirement{{Package: "ffmpeg"}}, Options: append([]operation.Parameter{toolrun.Option("seconds", operation.TypeInteger, "Recording duration, 1 to 300 seconds", 10), toolrun.Option("fps", operation.TypeInteger, "Frame rate", 15), toolrun.Option("width", operation.TypeInteger, "Rectangle width; 0 captures the desktop", 0), toolrun.Option("height", operation.TypeInteger, "Rectangle height; must accompany width", 0), toolrun.Option("x", operation.TypeInteger, "Rectangle left offset", 0), toolrun.Option("y", operation.TypeInteger, "Rectangle top offset", 0), toolrun.Option("cursor", operation.TypeBoolean, "Include mouse cursor", true)}, toolrun.OutputOptions()...)},
	}
	for _, def := range defs {
		def.Description = def.Summary + ". Windows only. Explicit user invocation; no resident service, no uploads and no voice or video playback."
		if e := reg.Register(operation.Capability{Definition: def, Runner: operation.Func(func(ctx context.Context, r operation.Request) (operation.Result, error) {
			if e := operation.ValidateRequest(def, r); e != nil {
				return operation.Result{}, e
			}
			if runtime.GOOS != "windows" {
				return operation.Result{}, &operation.Error{Code: operation.CodeUnsupportedPlatform, Message: def.ID + " currently supports Windows only"}
			}
			if def.ID == "screen.record" {
				return record(ctx, resolver, r)
			}
			return speak(ctx, def.ID, r)
		})}); e != nil {
			return e
		}
	}
	return nil
}
func recordArguments(r operation.Request) ([]string, int, error) {
	v := &toolrun.Values{Request: r}
	seconds := v.Int("seconds", 10, 1, 300)
	fps := v.Int("fps", 15, 1, 60)
	w, h := v.Int("width", 0, 0, 8192), v.Int("height", 0, 0, 8192)
	x, y := v.Int("x", 0, -32768, 32768), v.Int("y", 0, -32768, 32768)
	cursor := v.Bool("cursor", true)
	v.Check((w == 0 && h == 0) || (w >= 2 && h >= 2), "provide both width and height, at least 2 pixels")
	if h > 0 {
		v.Check(w <= 25000000/h, "capture exceeds 25 million pixels")
	}
	if w == 0 {
		v.Check(x == 0 && y == 0, "offsets require explicit rectangle dimensions")
	}
	if v.Err != nil {
		return nil, 0, v.Err
	}
	mouse := "0"
	if cursor {
		mouse = "1"
	}
	args := []string{"-hide_banner", "-loglevel", "error", "-nostdin", "-y", "-f", "gdigrab", "-framerate", strconv.Itoa(fps), "-draw_mouse", mouse}
	if w > 0 {
		args = append(args, "-offset_x", strconv.Itoa(x), "-offset_y", strconv.Itoa(y), "-video_size", fmt.Sprintf("%dx%d", w, h))
	}
	args = append(args, "-i", "desktop", "-t", strconv.Itoa(seconds), "-an", "-vf", "scale=trunc(iw/2)*2:trunc(ih/2)*2", "-c:v", "libx264", "-pix_fmt", "yuv420p", "-movflags", "+faststart")
	return args, seconds, nil
}
func record(ctx context.Context, resolver toolrun.Resolver, r operation.Request) (operation.Result, error) {
	v := &toolrun.Values{Request: r}
	out := v.String("output", "")
	overwrite := v.Bool("overwrite", false)
	args, seconds, e := recordArguments(r)
	if e != nil {
		return operation.Result{}, e
	}
	if v.Err != nil {
		return operation.Result{}, v.Err
	}
	if !strings.EqualFold(filepath.Ext(out), ".mp4") {
		return operation.Result{}, toolrun.Invalid("output must be .mp4")
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(seconds+30)*time.Second)
	defer cancel()
	e = toolrun.Artifact(ctx, out, overwrite, func(p string) error {
		_, e := toolrun.Run(ctx, resolver, "ffmpeg", "ffmpeg", "", append(args, p)...)
		return e
	})
	return operation.Result{Operation: "screen.record", Outputs: []string{out}, Data: map[string]any{"seconds": seconds, "audio": false, "capture": "explicit Windows desktop or rectangle"}}, e
}

// All request data goes through JSON on stdin, never through PowerShell source.
const speechScript = `$ErrorActionPreference='Stop'; [Console]::InputEncoding=[Text.UTF8Encoding]::new($false); [Console]::OutputEncoding=[Text.UTF8Encoding]::new($false); $r=([Console]::In.ReadToEnd()|ConvertFrom-Json); Add-Type -AssemblyName System.Speech; $s=New-Object System.Speech.Synthesis.SpeechSynthesizer; try { if($r.action -eq 'voices'){ $voices=@($s.GetInstalledVoices()|ForEach-Object { @{name=$_.VoiceInfo.Name;culture=$_.VoiceInfo.Culture.Name;enabled=$_.Enabled} }); [Console]::Out.Write((ConvertTo-Json -InputObject $voices -Compress)) } else { if($r.voice){$s.SelectVoice($r.voice)}; $s.Rate=$r.rate; $s.Volume=$r.volume; $s.SetOutputToWaveFile($r.output); $s.Speak($r.text); $s.SetOutputToNull() } } finally { $s.Dispose() }`

func runSpeech(ctx context.Context, payload map[string]any) ([]byte, error) {
	data, e := json.Marshal(payload)
	if e != nil {
		return nil, e
	}
	binary := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	if !filepath.IsAbs(binary) {
		return nil, toolrun.Invalid("Windows PowerShell is unavailable")
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, "-NoLogo", "-NoProfile", "-NonInteractive", "-Command", speechScript)
	cmd.Stdin = bytes.NewReader(data)
	var out, diagnostic bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &diagnostic
	if e = cmd.Run(); e != nil {
		return nil, &operation.Error{Code: operation.CodeExecutionFailed, Message: "local Windows speech failed", Details: map[string]any{"diagnostic": diagnostic.String()}, Err: e}
	}
	return out.Bytes(), nil
}
func speak(ctx context.Context, id string, r operation.Request) (operation.Result, error) {
	if id == "speech.voices" {
		b, e := runSpeech(ctx, map[string]any{"action": "voices"})
		if e != nil {
			return operation.Result{}, e
		}
		var voices []map[string]any
		if e = json.Unmarshal(b, &voices); e != nil {
			return operation.Result{}, e
		}
		return operation.Result{Operation: id, Data: map[string]any{"voices": voices}}, nil
	}
	v := &toolrun.Values{Request: r}
	out := v.String("output", "")
	overwrite := v.Bool("overwrite", false)
	voice := v.String("voice", "")
	rate := v.Int("rate", 0, -10, 10)
	volume := v.Int("volume", 100, 0, 100)
	v.Check(len(r.Inputs[0]) > 0 && len(r.Inputs[0]) <= 32768, "text must be 1 to 32768 UTF-8 bytes")
	v.Check(len(voice) <= 256, "voice name too long")
	if v.Err != nil {
		return operation.Result{}, v.Err
	}
	if !strings.EqualFold(filepath.Ext(out), ".wav") {
		return operation.Result{}, toolrun.Invalid("output must be .wav; use audio.convert for other formats")
	}
	e := toolrun.Artifact(ctx, out, overwrite, func(p string) error {
		_, e := runSpeech(ctx, map[string]any{"action": "speak", "text": r.Inputs[0], "output": p, "voice": voice, "rate": rate, "volume": volume})
		return e
	})
	return operation.Result{Operation: id, Outputs: []string{out}, Data: map[string]any{"voice": voice, "engine": "installed Windows System.Speech"}}, e
}
