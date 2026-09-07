package ffmpeg

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fu9zhou/finish-bit/internal/toolrun"
	"github.com/fu9zhou/finish-bit/pkg/operation"
)

type mediaSpec struct {
	id, summary, alias string
	options            []operation.Parameter
	inputs             int
	directory, inspect bool
}

func s(name, description, value string) operation.Parameter {
	return toolrun.Option(name, operation.TypeString, description, value)
}
func n(name, description string, value int) operation.Parameter {
	return toolrun.Option(name, operation.TypeInteger, description, value)
}
func b(name, description string, value bool) operation.Parameter {
	return toolrun.Option(name, operation.TypeBoolean, description, value)
}
func dimensions() []operation.Parameter {
	return []operation.Parameter{n("width", "Output width, even pixels", 640), n("height", "Output height, even pixels", 360)}
}
func mediaCatalog() []mediaSpec {
	files := toolrun.Option("files", operation.TypeStrings, "Additional local files in order (1 to 31)", nil)
	return []mediaSpec{
		{id: "media.info", summary: "Inspect media streams, chapters and container metadata", alias: "媒体信息", inputs: 1, inspect: true},
		{id: "video.convert", summary: "Transcode video to MP4, MKV or WebM", alias: "视频转码", inputs: 1, options: []operation.Parameter{s("format", "mp4, mkv or webm", "mp4"), n("crf", "Quality: 0 to 51 (lower is better)", 23)}},
		{id: "video.remux", summary: "Change a media container without re-encoding streams", alias: "无损换封装", inputs: 1, options: []operation.Parameter{s("format", "mp4, mkv or mov; streams must be compatible", "mkv")}},
		{id: "video.resize", summary: "Fit and pad video to an even-sized canvas", alias: "视频缩放", inputs: 1, options: dimensions()},
		{id: "video.crop", summary: "Crop a rectangular area from video", alias: "画面裁剪", inputs: 1, options: []operation.Parameter{n("width", "Crop width in even pixels", 320), n("height", "Crop height in even pixels", 240), n("x", "Left coordinate", 0), n("y", "Top coordinate", 0)}},
		{id: "video.rotate", summary: "Rotate video by a right angle", alias: "视频旋转", inputs: 1, options: []operation.Parameter{s("angle", "90, 180 or 270 clockwise degrees", "90")}},
		{id: "video.flip", summary: "Flip video horizontally or vertically", alias: "视频翻转", inputs: 1, options: []operation.Parameter{s("axis", "horizontal or vertical", "horizontal")}},
		{id: "video.speed", summary: "Change video speed and audio tempo together", alias: "视频变速", inputs: 1, options: []operation.Parameter{s("factor", "Speed multiplier from 0.5 to 2", "1.5"), b("audio", "Input has an audio stream to preserve", true)}},
		{id: "video.framerate", summary: "Convert to a constant frame rate", alias: "调整帧率", inputs: 1, options: []operation.Parameter{n("fps", "Frames per second", 30)}},
		{id: "video.concat", summary: "Normalize and concatenate videos in order", alias: "视频拼接", inputs: 1, options: append([]operation.Parameter{files, b("audio", "Every input has an audio stream", true)}, dimensions()...)},
		{id: "video.thumbnail", summary: "Extract one video frame as PNG or JPEG", alias: "视频截图 封面", inputs: 1, options: []operation.Parameter{s("time", "Seconds from start", "0"), n("width", "Maximum thumbnail width", 640)}},
		{id: "video.frames", summary: "Extract a bounded sequence of PNG frames", alias: "视频抽帧", inputs: 1, directory: true, options: []operation.Parameter{s("interval", "Seconds between frames", "1"), n("count", "Maximum number of frames", 20), n("width", "Frame width", 640)}},
		{id: "video.contact-sheet", summary: "Create a grid of regularly sampled video frames", alias: "视频联系表", inputs: 1, options: []operation.Parameter{s("interval", "Seconds between frames", "1"), n("columns", "Grid columns", 4), n("rows", "Grid rows", 3), n("width", "Width per tile", 240), s("font", "Optional local font for timestamps", "")}},
		{id: "video.gif", summary: "Create a palette-optimized animated GIF from video", alias: "视频转动图", inputs: 1, options: []operation.Parameter{n("fps", "GIF frames per second", 10), n("width", "GIF width", 480), s("duration", "Maximum seconds to include", "10")}},
		{id: "video.from-images", summary: "Create a slideshow from an ordered image list", alias: "图片转视频", inputs: 1, options: append([]operation.Parameter{files, s("seconds", "Seconds per image", "2")}, dimensions()...)},
		{id: "video.watermark", summary: "Overlay a local image on video", alias: "视频水印", inputs: 2, options: []operation.Parameter{n("x", "Left position", 10), n("y", "Top position", 10)}},
		{id: "video.mute", summary: "Remove all audio streams from video", alias: "视频静音", inputs: 1},
		{id: "video.replace-audio", summary: "Replace a video's audio with a local track", alias: "替换音轨", inputs: 2},
		{id: "video.add-music", summary: "Mix background music into an existing video soundtrack", alias: "添加背景音乐", inputs: 2, options: []operation.Parameter{s("volume", "Background music volume from 0 to 2", "0.25")}},
		{id: "subtitle.extract", summary: "Extract an existing text subtitle stream as SRT", alias: "提取字幕", inputs: 1, options: []operation.Parameter{n("index", "Zero-based subtitle stream index", 0)}},
		{id: "subtitle.add", summary: "Add a local SRT subtitle track to video", alias: "添加软字幕", inputs: 2},
		{id: "subtitle.burn", summary: "Render a local SRT subtitle file into video pixels", alias: "烧录字幕", inputs: 2},
		{id: "audio.convert", summary: "Convert audio to MP3, M4A, WAV, FLAC or Opus", alias: "音频转换", inputs: 1, options: []operation.Parameter{s("format", "mp3, m4a, wav, flac or opus", "mp3")}},
		{id: "audio.trim", summary: "Trim an audio track by start and duration", alias: "音频裁剪", inputs: 1, options: []operation.Parameter{s("start", "Start in seconds", "0"), s("duration", "Duration in seconds", "1")}},
		{id: "audio.concat", summary: "Normalize and concatenate audio files", alias: "音频拼接", inputs: 1, options: []operation.Parameter{files}},
		{id: "audio.mix", summary: "Mix two audio files with normalized gain", alias: "音频混音", inputs: 2},
		{id: "audio.volume", summary: "Adjust audio gain", alias: "音量调整", inputs: 1, options: []operation.Parameter{s("gain", "Volume multiplier from 0 to 10", "1")}},
		{id: "audio.normalize", summary: "Normalize loudness using EBU R128 dynamic processing", alias: "响度标准化", inputs: 1, options: []operation.Parameter{n("lufs", "Target integrated loudness, -70 to -5", -16)}},
		{id: "audio.denoise", summary: "Reduce stationary noise with FFT denoising", alias: "音频降噪", inputs: 1, options: []operation.Parameter{n("reduction", "Noise reduction in dB, 0 to 97", 12)}},
		{id: "audio.silence-detect", summary: "Report detected silent intervals", alias: "静音检测", inputs: 1, inspect: true, options: []operation.Parameter{n("threshold", "Noise threshold in dB, -90 to -1", -40), s("duration", "Minimum silence in seconds", "0.5")}},
		{id: "audio.silence-remove", summary: "Remove silence using an explicit threshold", alias: "移除静音", inputs: 1, options: []operation.Parameter{n("threshold", "Noise threshold in dB, -90 to -1", -40), s("duration", "Minimum silence in seconds", "0.5")}},
		{id: "audio.fade", summary: "Apply an audio fade at a specified time", alias: "音频淡入淡出", inputs: 1, options: []operation.Parameter{s("type", "in or out", "in"), s("start", "Fade start in seconds", "0"), s("duration", "Fade duration in seconds", "1")}},
		{id: "audio.waveform", summary: "Render an audio waveform as PNG", alias: "音频波形图", inputs: 1, options: []operation.Parameter{n("width", "Image width", 1000), n("height", "Image height", 240)}},
	}
}

func (p *Provider) registerExtended(registry *operation.Registry) error {
	for _, spec := range mediaCatalog() {
		def := operation.Definition{ID: spec.id, Summary: spec.summary, Description: spec.summary + ". Local files only. Outputs are staged; existing files require overwrite. See docs/media.md for supported formats and limits.", Aliases: []string{spec.alias}, Tags: []string{strings.Split(spec.id, ".")[0], "media"}, Options: append([]operation.Parameter{}, spec.options...), Source: "ffmpeg", Requirements: []operation.Requirement{{Package: "ffmpeg"}}}
		for i := 0; i < spec.inputs; i++ {
			name := "input"
			if i > 0 {
				name = "second"
			}
			def.Inputs = append(def.Inputs, toolrun.Param(name, "Local input file", true))
		}
		if spec.id == "media.info" {
			def.Tags = []string{"media", "metadata"}
		}
		if spec.id == "media.info" || spec.id == "audio.silence-detect" {
			def.Requirements = []operation.Requirement{{Package: "ffprobe"}}
		}
		if !spec.inspect {
			if spec.directory {
				def.Options = append(def.Options, toolrun.Param("output", "New output directory", true))
			} else {
				def.Options = append(def.Options, toolrun.OutputOptions()...)
			}
		}
		runner := operation.Func(func(ctx context.Context, r operation.Request) (operation.Result, error) {
			if err := operation.ValidateRequest(def, r); err != nil {
				return operation.Result{}, err
			}
			return p.runExtended(ctx, spec, r)
		})
		if err := registry.Register(operation.Capability{Definition: def, Runner: runner}); err != nil {
			return err
		}
	}
	return nil
}

func decimal(v *toolrun.Values, name, fallback string, low, high float64) string {
	value := v.String(name, fallback)
	parsed, err := strconv.ParseFloat(value, 64)
	v.Check(err == nil && !math.IsNaN(parsed) && !math.IsInf(parsed, 0) && parsed >= low && parsed <= high, fmt.Sprintf("%s must be a number between %g and %g", name, low, high))
	return strconv.FormatFloat(parsed, 'f', -1, 64)
}
func even(v *toolrun.Values, name string, fallback int) int {
	x := v.Int(name, fallback, 2, 8192)
	v.Check(x%2 == 0, name+" must be even")
	return x
}
func videoCodec() []string { return []string{"-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "aac"} }
func fit(w, h int) string {
	return fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2,setsar=1", w, h, w, h)
}
func audioCodec(format string) []string {
	switch format {
	case "mp3":
		return []string{"-c:a", "libmp3lame", "-q:a", "2"}
	case "m4a":
		return []string{"-c:a", "aac", "-b:a", "192k"}
	case "flac":
		return []string{"-c:a", "flac"}
	case "opus":
		return []string{"-c:a", "libopus", "-b:a", "128k"}
	default:
		return []string{"-c:a", "pcm_s16le"}
	}
}

func (p *Provider) runExtended(ctx context.Context, spec mediaSpec, r operation.Request) (operation.Result, error) {
	v := &toolrun.Values{Request: r}
	paths := []string{}
	for _, path := range r.Inputs {
		paths = append(paths, v.File(path))
	}
	for _, path := range v.Strings("files") {
		paths = append(paths, v.File(path))
	}
	v.Check(len(paths) <= 32, "at most 32 input files are supported")
	if v.Err != nil {
		return operation.Result{}, v.Err
	}
	if spec.id == "media.info" {
		return p.inspect(ctx, paths[0])
	}
	if spec.id == "audio.silence-detect" {
		return p.detectSilence(ctx, paths[0], v)
	}
	output := v.String("output", "")
	overwrite := v.Bool("overwrite", false)
	args := []string{"-hide_banner", "-loglevel", "error", "-nostdin", "-y", "-filter_complex_threads", "1"}
	for _, path := range paths {
		args = append(args, "-i", path)
	}
	tail := []string{}
	format := "mp4"
	data := map[string]any{}
	workFiles := map[string]string{}
	if strings.HasPrefix(spec.id, "audio.") && spec.id != "audio.waveform" {
		format = strings.TrimPrefix(strings.ToLower(filepath.Ext(output)), ".")
		v.Check(format == "mp3" || format == "m4a" || format == "wav" || format == "flac" || format == "opus", "audio output extension must be mp3, m4a, wav, flac or opus")
		tail = append(tail, "-map", "0:a:0", "-vn")
		tail = append(tail, audioCodec(format)...)
	} else {
		tail = append(tail, videoCodec()...)
	}
	switch spec.id {
	case "video.convert":
		format = v.Enum("format", "mp4", "mp4", "mkv", "webm")
		crf := v.Int("crf", 23, 0, 51)
		tail = append(tail, "-crf", strconv.Itoa(crf))
		if format == "webm" {
			tail = []string{"-c:v", "libvpx-vp9", "-crf", strconv.Itoa(crf), "-b:v", "0", "-c:a", "libopus"}
		}
	case "video.remux":
		format = v.Enum("format", "mkv", "mp4", "mkv", "mov")
		tail = []string{"-map", "0", "-c", "copy"}
	case "video.resize":
		tail = append(tail, "-vf", fit(even(v, "width", 640), even(v, "height", 360)))
	case "video.crop":
		tail = append(tail, "-vf", fmt.Sprintf("crop=%d:%d:%d:%d", even(v, "width", 320), even(v, "height", 240), v.Int("x", 0, 0, 32768), v.Int("y", 0, 0, 32768)))
	case "video.rotate":
		angle := v.Enum("angle", "90", "90", "180", "270")
		f := map[string]string{"90": "transpose=1", "180": "hflip,vflip", "270": "transpose=2"}[angle]
		tail = append(tail, "-vf", f)
	case "video.flip":
		axis := v.Enum("axis", "horizontal", "horizontal", "vertical")
		filter := "hflip"
		if axis == "vertical" {
			filter = "vflip"
		}
		tail = append(tail, "-vf", filter)
	case "video.speed":
		factor := decimal(v, "factor", "1.5", 0.5, 2)
		tail = append(tail, "-vf", "setpts=(PTS-STARTPTS)/"+factor)
		if v.Bool("audio", true) {
			tail = append(tail, "-map", "0:v:0", "-map", "0:a:0", "-af", "atempo="+factor)
		} else {
			tail = append(tail, "-an")
		}
	case "video.framerate":
		tail = append(tail, "-vf", "fps="+strconv.Itoa(v.Int("fps", 30, 1, 120)))
	case "video.concat", "audio.concat", "video.from-images":
		v.Check(len(paths) >= 2, "provide at least one additional --files input")
		isVideo := spec.id != "audio.concat"
		audio := !isVideo || v.Bool("audio", true)
		if spec.id == "video.from-images" {
			audio = false
		}
		filters := []string{}
		labels := ""
		w, h := 640, 360
		if isVideo {
			w = even(v, "width", 640)
			h = even(v, "height", 360)
		}
		if spec.id == "video.from-images" {
			seconds := decimal(v, "seconds", "2", 0.04, 3600)
			args = args[:7]
			for _, path := range paths {
				args = append(args, "-loop", "1", "-t", seconds, "-i", path)
			}
		}
		for i := range paths {
			if isVideo {
				filters = append(filters, fmt.Sprintf("[%d:v:0]%s,fps=25,format=yuv420p,setpts=PTS-STARTPTS[v%d]", i, fit(w, h), i))
				labels += fmt.Sprintf("[v%d]", i)
			}
			if audio {
				filters = append(filters, fmt.Sprintf("[%d:a:0]aformat=sample_rates=48000:channel_layouts=stereo,asetpts=PTS-STARTPTS[a%d]", i, i))
				labels += fmt.Sprintf("[a%d]", i)
			}
		}
		vc, ac := 0, 0
		outputs := ""
		if isVideo {
			vc = 1
			outputs += "[v]"
		}
		if audio {
			ac = 1
			outputs += "[a]"
		}
		filters = append(filters, fmt.Sprintf("%sconcat=n=%d:v=%d:a=%d%s", labels, len(paths), vc, ac, outputs))
		tail = []string{"-filter_complex", strings.Join(filters, ";")}
		if isVideo {
			tail = append(tail, "-map", "[v]")
			tail = append(tail, videoCodec()...)
		} else {
			tail = append(tail, audioCodec(format)...)
		}
		if audio {
			tail = append(tail, "-map", "[a]")
		}
	case "video.thumbnail":
		format = "png"
		if strings.EqualFold(filepath.Ext(output), ".jpg") || strings.EqualFold(filepath.Ext(output), ".jpeg") {
			format = "jpeg"
		}
		tail = []string{"-ss", decimal(v, "time", "0", 0, 864000), "-vf", fmt.Sprintf("scale='min(%d,iw)':-1", v.Int("width", 640, 1, 8192)), "-frames:v", "1", "-an"}
	case "video.frames":
		format = "directory"
		interval := decimal(v, "interval", "1", 0.04, 86400)
		tail = []string{"-vf", fmt.Sprintf("fps=1/%s,scale=%d:-1", interval, v.Int("width", 640, 1, 8192)), "-frames:v", strconv.Itoa(v.Int("count", 20, 1, 500)), "-an"}
		data["interval_seconds"] = interval
	case "video.contact-sheet":
		format = "png"
		cols := v.Int("columns", 4, 1, 10)
		rows := v.Int("rows", 3, 1, 10)
		width := v.Int("width", 240, 16, 1000)
		interval := decimal(v, "interval", "1", 0.04, 86400)
		v.Check(cols*rows*width*width <= 25_000_000, "contact sheet exceeds pixel budget")
		filter := fmt.Sprintf("fps=1/%s,scale=%d:-1", interval, width)
		font := v.String("font", "")
		if font != "" {
			workFiles["font.ttf"] = v.File(font)
			filter += ",drawtext=fontfile=font.ttf:text='%{pts\\:hms}':fontcolor=white:box=1:boxcolor=black@0.6:x=4:y=4"
		}
		filter += fmt.Sprintf(",tile=%dx%d:nb_frames=%d", cols, rows, cols*rows)
		tail = []string{"-vf", filter, "-frames:v", "1", "-an"}
		data["interval_seconds"] = interval
		data["columns"] = cols
		data["rows"] = rows
	case "video.gif":
		format = "gif"
		fps := v.Int("fps", 10, 1, 30)
		width := v.Int("width", 480, 16, 1920)
		duration := decimal(v, "duration", "10", 0.04, 300)
		tail = []string{"-t", duration, "-filter_complex", fmt.Sprintf("[0:v]fps=%d,scale=%d:-1:flags=lanczos,split[a][b];[a]palettegen[p];[b][p]paletteuse", fps, width), "-an", "-loop", "0"}
	case "video.watermark":
		tail = append(tail, "-filter_complex", fmt.Sprintf("[0:v][1:v]overlay=%d:%d:format=auto[v]", v.Int("x", 10, 0, 8192), v.Int("y", 10, 0, 8192)), "-map", "[v]", "-map", "0:a?")
	case "video.mute":
		tail = []string{"-map", "0:v:0", "-c:v", "copy", "-an"}
	case "video.replace-audio":
		tail = []string{"-map", "0:v:0", "-map", "1:a:0", "-c:v", "copy", "-c:a", "aac", "-shortest"}
	case "video.add-music":
		gain := decimal(v, "volume", "0.25", 0, 2)
		tail = append(tail, "-filter_complex", "[1:a]volume="+gain+"[music];[0:a][music]amix=inputs=2:duration=first:normalize=0[a]", "-map", "0:v:0", "-map", "[a]")
	case "subtitle.extract":
		format = "srt"
		tail = []string{"-map", fmt.Sprintf("0:s:%d", v.Int("index", 0, 0, 99)), "-c:s", "srt"}
	case "subtitle.add":
		v.Check(strings.EqualFold(filepath.Ext(paths[1]), ".srt"), "subtitle input must be SRT")
		tail = []string{"-map", "0:v:0", "-map", "0:a?", "-map", "1:s:0", "-c:v", "copy", "-c:a", "copy", "-c:s", "mov_text"}
	case "subtitle.burn":
		v.Check(strings.EqualFold(filepath.Ext(paths[1]), ".srt"), "subtitle input must be SRT")
		workFiles["subtitle.srt"] = paths[1]
		tail = append(tail, "-vf", "subtitles=subtitle.srt", "-map", "0:v:0", "-map", "0:a?")
	case "audio.convert":
		format = v.Enum("format", "mp3", "mp3", "m4a", "wav", "flac", "opus")
		tail = append([]string{"-map", "0:a:0", "-vn"}, audioCodec(format)...)
	case "audio.trim":
		tail = append(tail, "-ss", decimal(v, "start", "0", 0, 864000), "-t", decimal(v, "duration", "1", 0.001, 864000))
	case "audio.mix":
		tail = append(audioCodec(format), "-filter_complex", "[0:a][1:a]amix=inputs=2:duration=longest:normalize=1[a]", "-map", "[a]", "-vn")
	case "audio.volume":
		tail = append(tail, "-af", "volume="+decimal(v, "gain", "1", 0, 10))
	case "audio.normalize":
		tail = append(tail, "-af", fmt.Sprintf("loudnorm=I=%d:TP=-1.5:LRA=11", v.Int("lufs", -16, -70, -5)))
	case "audio.denoise":
		tail = append(tail, "-af", fmt.Sprintf("afftdn=nr=%d", v.Int("reduction", 12, 0, 97)))
	case "audio.silence-remove":
		threshold := v.Int("threshold", -40, -90, -1)
		duration := decimal(v, "duration", "0.5", 0.001, 3600)
		tail = append(tail, "-af", fmt.Sprintf("silenceremove=start_periods=1:start_duration=0:start_threshold=%ddB:stop_periods=-1:stop_duration=%s:stop_threshold=%ddB", threshold, duration, threshold))
	case "audio.fade":
		kind := v.Enum("type", "in", "in", "out")
		tail = append(tail, "-af", "afade=t="+kind+":st="+decimal(v, "start", "0", 0, 864000)+":d="+decimal(v, "duration", "1", 0.001, 864000))
	case "audio.waveform":
		format = "png"
		tail = []string{"-filter_complex", fmt.Sprintf("[0:a]showwavespic=s=%dx%d:colors=steelblue[v]", v.Int("width", 1000, 16, 8192), v.Int("height", 240, 16, 2048)), "-map", "[v]", "-frames:v", "1"}
	default:
		return operation.Result{}, toolrun.Invalid("unsupported media operation")
	}
	if strings.HasPrefix(spec.id, "video.") {
		hasMapping := false
		for _, arg := range tail {
			if arg == "-map" || arg == "-filter_complex" {
				hasMapping = true
			}
		}
		if !hasMapping {
			tail = append(tail, "-map", "0:v:0", "-map", "0:a?")
		}
	}
	if !spec.directory {
		ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(output)), ".")
		if ext == "jpg" {
			ext = "jpeg"
		}
		v.Check(ext == format, "output extension must match "+format)
	}
	if v.Err != nil {
		return operation.Result{}, v.Err
	}
	execute := func(destination string) error {
		dir := filepath.Dir(destination)
		for name, path := range workFiles {
			if err := toolrun.CopyFile(path, filepath.Join(dir, name)); err != nil {
				return err
			}
		}
		commandArgs := append(append(append([]string{}, args...), tail...), "-threads", "2", destination)
		_, err := toolrun.Run(ctx, p.resolver, "ffmpeg", "ffmpeg", dir, commandArgs...)
		return err
	}
	var outputs []string
	if spec.directory {
		var err error
		outputs, err = toolrun.Directory(ctx, output, func(dir string) error { return execute(filepath.Join(dir, "frame-%05d.png")) })
		if err != nil {
			return operation.Result{}, err
		}
	} else {
		if err := toolrun.Artifact(ctx, output, overwrite, execute); err != nil {
			return operation.Result{}, err
		}
		outputs = []string{output}
	}
	return operation.Result{Operation: spec.id, Outputs: outputs, Data: data}, nil
}

func (p *Provider) inspect(ctx context.Context, path string) (operation.Result, error) {
	raw, err := toolrun.Run(ctx, p.resolver, "ffprobe", "ffprobe", "", "-v", "error", "-show_format", "-show_streams", "-show_chapters", "-of", "json", path)
	if err != nil {
		return operation.Result{}, err
	}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.UseNumber()
	var data map[string]any
	if err := decoder.Decode(&data); err != nil {
		return operation.Result{}, fmt.Errorf("decode ffprobe output: %w", err)
	}
	return operation.Result{Operation: "media.info", Data: data}, nil
}

func (p *Provider) detectSilence(ctx context.Context, path string, v *toolrun.Values) (operation.Result, error) {
	threshold := v.Int("threshold", -40, -90, -1)
	duration := decimal(v, "duration", "0.5", 0.001, 3600)
	if v.Err != nil {
		return operation.Result{}, v.Err
	}
	// A fixed relative filename avoids injecting paths into the lavfi grammar.
	dir, err := os.MkdirTemp("", "finishbit-silence-")
	if err != nil {
		return operation.Result{}, err
	}
	defer os.RemoveAll(dir)
	local := filepath.Join(dir, "input"+filepath.Ext(path))
	if err := toolrun.CopyFile(path, local); err != nil {
		return operation.Result{}, err
	}
	filter := fmt.Sprintf("amovie=%s,silencedetect=noise=%ddB:d=%s", filepath.Base(local), threshold, duration)
	raw, err := toolrun.Run(ctx, p.resolver, "ffprobe", "ffprobe", dir, "-v", "error", "-f", "lavfi", "-i", filter, "-show_entries", "frame_tags=lavfi.silence_start,lavfi.silence_end,lavfi.silence_duration", "-of", "json")
	if err != nil {
		return operation.Result{}, err
	}
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return operation.Result{}, err
	}
	data["threshold_db"] = threshold
	data["minimum_seconds"] = duration
	data["note"] = "Frame tags mark starts and ends; trailing silence may have no end tag."
	return operation.Result{Operation: "audio.silence-detect", Data: data}, nil
}
