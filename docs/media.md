# Media operations

FinishBit retains `video.trim`, `video.compress`, and `audio.extract`. The expanded catalog adds task-oriented local media editing, previews, subtitles and audio processing through the same `pkg/app` execution path.

## Install and discover

```sh
fnsh pkg add ffmpeg
fnsh pkg add ffprobe
fnsh search "视频 拼接"
fnsh describe video.concat --json
```

FFmpeg and ffprobe are independently managed, pinned to 6.1.1 with platform-specific SHA-256 checksums. Only `media.info` and `audio.silence-detect` require ffprobe; other new media operations require FFmpeg. No PATH fallback is used in production.

## Catalog

| Area | Operations |
| --- | --- |
| Inspection | `media.info` |
| Video conversion | `video.convert`, `video.remux`, `video.resize`, `video.framerate` |
| Video editing | `video.crop`, `video.rotate`, `video.flip`, `video.speed`, `video.concat`, `video.watermark`, `video.mute`, `video.replace-audio`, `video.add-music` |
| Previews | `video.thumbnail`, `video.frames`, `video.contact-sheet`, `video.gif`, `video.from-images` |
| Subtitles | `subtitle.extract`, `subtitle.add`, `subtitle.burn` |
| Audio | `audio.convert`, `audio.trim`, `audio.concat`, `audio.mix`, `audio.volume`, `audio.normalize`, `audio.denoise`, `audio.silence-detect`, `audio.silence-remove`, `audio.fade`, `audio.waveform` |

## Examples

```sh
fnsh video concat first.mp4 --files second.mp4 --width 1280 --height 720 -o joined.mp4
fnsh video frames input.mp4 --interval 2 --count 12 --width 320 -o frames
fnsh video contact-sheet input.mp4 --interval 5 --columns 4 --rows 3 -o overview.png
fnsh audio normalize input.wav --lufs -16 -o normalized.wav
fnsh subtitle add input.mp4 captions.srt -o subtitled.mp4
fnsh media info input.mp4 --json
```

Use `fnsh describe <id> --json` for the exact typed contract. Decimal quantities are string options containing finite decimal seconds/multipliers (for example `"0.5"`, not `"0.5s"`). This does not change the older trim operation's time syntax. Multi-input operations accept a first positional input and ordered `--files` options, or a second positional input for two-input tasks.

## Output and resource behavior

- New single-output operations default to refusing replacement. Use `--overwrite` explicitly. Output is staged beside its destination and only published after a successful process; failed/cancelled runs preserve an existing destination.
- `video.frames` requires a new directory and returns an ordered list of generated PNG files. It accepts at most 500 frames. Multi-input operations accept at most 32 files. Every input must be a regular local file; URLs and generated filter inputs are not accepted as input paths.
- New video editing operations produce H.264/AAC MP4 unless their contract says otherwise. `video.convert` also offers MKV and WebM; remux accepts MP4/MKV/MOV and fails on incompatible streams instead of silently transcoding. Output extensions must match the selected format.
- Audio transformations support MP3/M4A/WAV/FLAC/Opus, selected by output extension; `audio.convert` additionally requires its format option to agree. Waveforms, contact sheets and frame sequences use PNG. Thumbnails also support JPEG, and `video.gif` requires GIF.
- Video concatenation normalizes dimensions, sample aspect ratio and frame rate. By default every input needs audio; use `--audio false` for video-only concatenation. The same explicit audio switch exists on video speed changes. Background-music mixing requires an existing soundtrack. Audio replacement stops at the shorter stream.
- Contact sheets sample at a fixed interval and expose the interval/grid in result data. An optional local `--font` file enables timestamps; font rendering is a runtime capability. Short inputs can produce incomplete grids. GIF duration and frame rate are bounded. These are previews, not scene understanding.
- Subtitles are existing text streams or supplied SRT files. Bitmap subtitle OCR and speech transcription are not included. Burning uses the managed build's libass/font support. Stream conversion may not preserve all subtitle styling or media metadata.
- Loudness normalization uses dynamic EBU R128 processing rather than a guaranteed two-pass mastering workflow. FFT denoising targets stationary noise. Silence detection returns ffprobe frame tags; trailing silence can have a start without an end tag. Silence removal trims leading silence immediately and applies the configured minimum duration to interior/trailing silence; it does not discard a confirmation interval of audible content.
- Providers bound stdout to 16 MiB and stderr to 1 MiB. Oversized structured output fails explicitly. Commands inherit cancellation from their application context.

## Real acceptance tests

Set `FINISHBIT_TEST_FFMPEG` and `FINISHBIT_TEST_FFPROBE` to absolute paths of managed binaries, then run:

```sh
go test ./internal/ffmpeg -run TestExtendedMediaWithManagedTools -count=1 -v
```

The test generates a video with a tone, partially silent audio, a watermark and subtitles; invokes every new Operation; probes and decodes outputs; checks dimensions, duration, stream removal, subtitle text, frame counts and silence; and checks validation, cancellation and output preservation. Missing test binaries skip this opt-in suite only when the environment variables are absent. Explicitly configured unusable binaries fail.
