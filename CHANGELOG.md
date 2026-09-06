# Changelog

All notable changes to FinishBit will be documented here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and releases use [Semantic Versioning](https://semver.org/).

This file records user-visible changes, compatibility breaks, migrations, and security-relevant release notes. Internal refactors without user impact may be omitted. See the [compatibility policy](docs/compatibility.md) and [release process](docs/releasing.md).

## [Unreleased]

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

[Unreleased]: https://github.com/fu9zhou/finish-bit/compare/v0.1.3...HEAD
[0.1.0]: https://github.com/fu9zhou/finish-bit/releases/tag/v0.1.0

[0.1.1]: https://github.com/fu9zhou/finish-bit/compare/v0.1.0...v0.1.1
[0.1.2]: https://github.com/fu9zhou/finish-bit/compare/v0.1.1...v0.1.2
[0.1.3]: https://github.com/fu9zhou/finish-bit/compare/v0.1.2...v0.1.3
