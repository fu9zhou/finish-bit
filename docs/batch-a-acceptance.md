# Batch A acceptance

Status: implemented and tested on Windows x64 on 2026-09-07; unreleased.

The catalog grows from 26 to 123 Operations. New capabilities comprise 33 FFmpeg Operations, 35 pdfcpu Operations, 7 Poppler Operations and 22 ImageMagick Operations. Three existing image APIs additionally accept `--engine imagemagick`; they are not counted again.

## Reproduce the gate

From the repository on Windows x64 with Go installed:

```powershell
./scripts/test-batch-a.ps1
```

The script selects Go 1.26.8 (matching CI), builds the CLI, installs or verifies the five pinned managed packages, resolves their executable paths from installed metadata, enables every real-provider suite, runs uncached `go test ./...`, `go vet ./...`, `go build ./cmd/fnsh` and govulncheck, then checks both structured and direct CLI execution. Any failure stops the script. It restores the caller's environment and writes a successful acceptance report and CLI artifacts under `.finishbit-test-batch-a/evidence/<unique-id>/`.

The initial local vulnerability scan used Go 1.26.4 and reported six reachable standard-library advisories. The Go 1.26.8 scan passed with no vulnerabilities found; use the pinned acceptance toolchain when reproducing this result.

CI includes a dedicated Windows job running this gate. Ordinary unit tests skip real-provider suites unless the runtime environment variables are set; a plain green unit-test run alone does not establish batch acceptance. The CI job has been configured; local results do not claim that a remote workflow has run.

## What is checked

| Capability group | Real acceptance evidence |
| --- | --- |
| Media | Every new Operation runs against generated video/audio/images/subtitles; ffprobe checks streams, size and duration, FFmpeg decodes outputs; silence removal, audio-only rejection, invalid parameters, cancellation and output preservation are covered. |
| PDF | Every new Operation runs on actual PDFs, including a fillable form. Independent Poppler checks page counts, extracted text/coordinates, form values, attachment bytes and rendered dimensions. Image/PDF stamps and Unicode paths are included. |
| Images | All 22 new Operations and three enhanced existing APIs run through the shared application service. Checks include pixels, alpha, EXIF orientation, animation frames, icon variants and WebP/TIFF/BMP/AVIF round trips. |
| CLI | A JSON request crops an image to independently verified dimensions; direct commands create and stamp a PDF, then recover the stamp text. Catalog completeness and dependency doctor are checked. |
| Packaging | Real archives install with preserved DLL/configuration layouts. Unit tests cover unsafe paths, links, cache verification, repair failures and preservation of existing installations. |

The real binaries are FFmpeg/ffprobe 6.1.1, pdfcpu 0.15.0, Poppler 26.07.0-0 and ImageMagick 7.1.2-31. Package registry platform coverage is not equivalent to actual platform acceptance: Poppler currently has a Windows x64 package; ImageMagick has Windows x64/arm64 packages, with x64 tested here. Other registered FFmpeg/pdfcpu platforms still need equivalent real-runtime acceptance.

Go 1.26.8 builds also passed for Linux amd64/arm64, macOS amd64/arm64 and Windows arm64 with CGO disabled, in addition to the Windows amd64 acceptance build. These are compilation checks, not real-provider tests on those platforms.

## Next batch

Batch B is spreadsheet cleaning, Excel workbooks, document conversion and archives. Batch C contains OCR, Office export, metadata and configuration/data-query candidates. Each group must have complete public contracts, managed dependency handling where applicable, real input/output assertions, failure-path tests and the repository checks before the following batch proceeds. The implemented catalogs document supported modes and limits; wrapping every upstream flag is not an acceptance criterion.
