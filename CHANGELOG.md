# Changelog

All notable changes to FinishBit will be documented here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and releases use [Semantic Versioning](https://semver.org/).

This file records user-visible changes, compatibility breaks, migrations, and security-relevant release notes. Internal refactors without user impact may be omitted. See the [compatibility policy](docs/compatibility.md) and [release process](docs/releasing.md).

## [Unreleased]

### Changed

- Reuse managed 7-Zip for runtime installation, automatically downloading the official standalone bootstrap on cold installs. Remove the embedded Go sevenzip decoder and nine transitive modules while retaining bounded, exclusive writes for 7z entries.
- Upgrade the managed Windows x64 OCR engine to Tesseract 5.5.3, keeping the pinned English and simplified Chinese fast models. The extracted runtime and models occupy approximately 103.3 MiB. Existing users can install the registered version with `fnsh pkg add tesseract` after updating the CLI.

## [0.1.6] - 2026-09-08

### Added

- 107 local Operations (277 total), including QR generation/decoding, bounded text/JSON diffs and RE2 tools, YAML/TOML/XML, Chinese text conversion, date/calendar/unit/math utilities, cryptographic random and AES-GCM helpers, and image/document utilities.
- Managed Windows x64 Tesseract with pinned Chinese/English fast models, searchable PDF and OCR-to-document workflows; NSIS archive extraction and separately checksummed package resources without running installers.
- Shared application workflows for PDF-to-DOCX text reflow, page-image PPTX, long images, raster PDFs and JPEG page compression; Windows installed-voice WAV generation and explicit silent screen recording.
- Offline handwriting worksheets, LED HTML, collapsible Markdown outlines, OOXML image/ZIP compression and a 166-entry implementation audit with explicit limits.

### Changed

- PDF crop supports independent margins; added metadata fields, dynamic page numbers, paper resizing, image signatures and image filter presets.
- Release archives include third-party notices for the new embedded libraries and dictionaries.

## [0.1.5] - 2026-09-07

### Added

- Optional `pdf.extract-text --unwrap` post-processing for paragraph-aware plain text, including end-of-line dehyphenation and conservative cross-column continuation.
- Dedicated searchable Operation catalog at `#/operations`, hierarchical detail routes at `#/operations/:id`, breadcrumbs, category summaries, and an explicit not-found page on the project website.

### Changed

- `document.split` now asks Pandoc for unwrapped output by default (`--wrap none`); callers can select `auto` or `preserve` through the new `wrap` option and tune wrapping through `columns`.
- The website homepage now presents the product workflow and capability categories instead of rendering all 170 Operations inline. Catalog filters and search terms remain shareable in the URL, while legacy `#/operation/:id` links continue to work.

### Fixed

- Plain-text extraction can now remove PDF visual line wrapping without collapsing headings, lists, or completed paragraph boundaries.
- Heading-based TXT splits no longer acquire Pandoc's default hard wrapping, avoiding false paragraph breaks in downstream processing.

## [0.1.4] - 2026-09-07

### Added

- Archive group: 10 managed 7-Zip Operations for creation, inspection, validation, selected extraction, update/removal/rename, repacking and volumes; 8 writable formats and ZIP/7z encryption bring the source catalog to 170 Operations.

- Document group: 10 managed Pandoc Operations spanning 25 output formats, merge, text/structure/media extraction, heading splits, templates, Office references and bibliography conversion; the catalog reached 160 Operations after this group.

- Table cleaning group: 27 managed qsv Operations for selection, filtering, sorting, deduplication, joins, differences, statistics, reshaping, schema validation and JSONL conversion; the catalog reached 150 Operations after this group.
- Generated bilingual website catalog with per-operation contracts, PDF/image/table guides and a CI check against the actual application registry; README capability descriptions match the expanded groups.
- First capability batch: 33 media Operations, 42 PDF Operations, and 22 image Operations, expanding the original 26 Operations to 123 before the table group.
- Managed ffprobe, pdfcpu, Poppler and ImageMagick runtimes with pinned release artifacts and integrity checks; ZIP/tar.gz/tar.xz/7z installation preserves multi-file runtime layouts and license files.
- Optional `--engine imagemagick` for the original image info/convert/resize APIs, preserving their default core behavior.
- Shared bounded process output, staged provider artifacts, and verified download caching. New file-writing operations require explicit overwrite; multi-file outputs require a new directory.
- Opt-in real-runtime acceptance suites covering every new Operation, with content/dimension/duration checks, Unicode paths, cancellation and failed-output preservation.

### Notes

- Full new-provider acceptance currently runs on Windows x64. Poppler installation is registered for Windows x64 only; ImageMagick for Windows x64/arm64. See each capability guide for exact runtime and format boundaries.
- OCR, Excel and later dependency batches remain planned, not implemented in this batch.
- qsv currently supports Windows x64. Excelize was evaluated but not introduced: the latest stable version has an upstream parsing-panic advisory without a patched release as of 2026-09-07.

## [0.1.3] - 2026-09-06

### Added

- Core `image.info`, `image.resize`, and `image.convert` Operations for PNG/JPEG, with aspect-ratio-preserving resize, opt-in upscaling, JPEG quality and white transparency compositing.
- Core `csv.info`, `csv.to-json`, and `json.to-csv` Operations with UTF-8 CSV validation, lossless string cells, stable or explicit column order, custom delimiters and scalar JSON values.
- Input, decoded-image, cell-count and serialized-output bounds; existing-destination protection and opt-in overwrite for the new Operations.
- Pixel, transparency, CSV round-trip, malformed-data, expansion-limit and structured-request integration tests.

### Notes

- Image transforms operate on encoded pixels without applying EXIF orientation and do not copy image metadata. PNG and JPEG are the supported formats for this version.
- Table conversion supports CSV/TSV and JSON object arrays; it does not read Excel workbooks or infer cell types.

## [0.1.2] - 2026-09-06

### Added

- Structured invocation through `fnsh run --request <file|->`, with the same Operation validation and output contract as positional execution.
- Public `operation.Call`, bounded UTF-8 JSON decoding, `app.ExecuteCall`, and a request JSON Schema for future adapters.
- Strict envelope validation, lossless integer option decoding, and regression tests for file/stdin input and malformed requests.

## [0.1.1] - 2026-09-06

### Fixed

- Honor the `--` end-of-options delimiter and preserve named option values that equal `--json`.
- Preserve binary Base64 decode output: raw stdout contains exact bytes; non-UTF-8 JSON results use lossless `data.base64` plus `data.encoding` instead of corrupted `data.text`. UTF-8 JSON text and file output remain supported.
- Keep the CLI usable when installed extensions conflict or become damaged; doctor reports the affected extension and removal remains available.
- Refresh the application registry after extension installation and removal.
- Reap extension processes when request transmission fails.

### Added

- Operation-level `--help` and visible parameter defaults in human-readable descriptions.

## [0.1.0] - 2026-09-04

### Added

- Initial task-oriented Operation runtime and `fnsh` CLI.
- Progressive capability search and structured JSON interface.
- Managed FFmpeg provider and local extension protocol.
- Codex Skill for agent-driven discovery and execution.

### Fixed

- Enforced Operation request contracts consistently at the application boundary.
- Bounded core input reads, FFmpeg diagnostics, package downloads, and decompression.
- Rejected invalid extension response envelopes and inconsistent manifest sources.
- Added macOS-compatible installer checksum verification and hardened release validation.

[Unreleased]: https://github.com/fu9zhou/finish-bit/compare/v0.1.6...HEAD
[0.1.6]: https://github.com/fu9zhou/finish-bit/compare/v0.1.5...v0.1.6
[0.1.0]: https://github.com/fu9zhou/finish-bit/releases/tag/v0.1.0

[0.1.1]: https://github.com/fu9zhou/finish-bit/compare/v0.1.0...v0.1.1
[0.1.2]: https://github.com/fu9zhou/finish-bit/compare/v0.1.1...v0.1.2
[0.1.3]: https://github.com/fu9zhou/finish-bit/compare/v0.1.2...v0.1.3

[0.1.4]: https://github.com/fu9zhou/finish-bit/compare/v0.1.3...v0.1.4
[0.1.5]: https://github.com/fu9zhou/finish-bit/compare/v0.1.4...v0.1.5
