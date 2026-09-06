# Changelog

All notable changes to FinishBit will be documented here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and releases use [Semantic Versioning](https://semver.org/).

This file records user-visible changes, compatibility breaks, migrations, and security-relevant release notes. Internal refactors without user impact may be omitted. See the [compatibility policy](docs/compatibility.md) and [release process](docs/releasing.md).

## [Unreleased]

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

[Unreleased]: https://github.com/fu9zhou/finish-bit/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/fu9zhou/finish-bit/releases/tag/v0.1.0
