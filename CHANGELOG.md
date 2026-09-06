# Changelog

All notable changes to FinishBit will be documented here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and releases use [Semantic Versioning](https://semver.org/).

This file records user-visible changes, compatibility breaks, migrations, and security-relevant release notes. Internal refactors without user impact may be omitted. See the [compatibility policy](docs/compatibility.md) and [release process](docs/releasing.md).

## [Unreleased]

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

[Unreleased]: https://github.com/fu9zhou/finish-bit/compare/v0.1.1...HEAD
[0.1.0]: https://github.com/fu9zhou/finish-bit/releases/tag/v0.1.0

[0.1.1]: https://github.com/fu9zhou/finish-bit/compare/v0.1.0...v0.1.1
