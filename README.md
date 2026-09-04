# FinishBit

> Small deterministic bits for finishing AI tasks.

[简体中文](docs/README.zh-CN.md) · [Architecture](docs/architecture.md) · [Extension protocol](docs/extension-protocol.md) · [Security](SECURITY.md)

FinishBit is a local, task-oriented capability runtime for AI agents. Its `fnsh` CLI lets an agent discover a small relevant operation, inspect its contract, and run tested functionality instead of generating another one-off script.

```text
Search → Describe → Execute → Finish
```

## Why FinishBit?

AI can write it. That does not mean it should rewrite it. Stable operations reduce token use, avoid repeated parameter mistakes, and keep common transformations reproducible.

- **Task-oriented:** ask to trim a video or format JSON, not to operate FFmpeg or `jq`.
- **Progressive discovery:** search first; only the selected operation enters the agent context.
- **Deterministic core:** lightweight operations are implemented in Go and tested.
- **Managed providers:** large tools are downloaded on demand at a runtime-owned version and verified with pinned SHA-256 digests.
- **Language-neutral extensions:** community capabilities communicate over a versioned JSON process protocol.
- **Transport-independent:** the CLI is an adapter over `pkg/app`; a future web/API adapter reuses the same application service.

## Status

FinishBit is pre-release software. Version `v0.1.0` establishes the first operation and extension contracts, but compatible evolution is not guaranteed until `v1.0.0`.

## Install from source

Go 1.24 or newer is required.

```bash
go install github.com/fu9zhou/finish-bit/cmd/fnsh@latest
fnsh version
```

Release archives and installation scripts will be attached to tagged GitHub releases.

## Quick start

```bash
fnsh search "裁剪并压缩视频"
fnsh describe video.trim --json

# FFmpeg is installed into FinishBit's private data directory.
fnsh pkg add ffmpeg

fnsh video trim input.mp4 --start 10s --duration 20s -o clip.mp4
fnsh video compress clip.mp4 --target-mb 20 -o compressed.mp4
```

Every discovery and execution command supports `--json` for agents and automation:

```bash
fnsh json format data.json --indent 2 --json
fnsh file checksum artifact.zip --json
```

Run `fnsh capabilities` for the full catalog or see [the operation guide](docs/operations.md).

## Agent Skill

The distributable Codex skill lives at [`skills/finishbit`](skills/finishbit). It teaches agents to retrieve only the capability they need and to recover from a missing managed dependency once.

## Extension example

An extension contains an executable and `finishbit-extension.json`:

```bash
fnsh ext add ./my-extension
fnsh search "my new capability"
```

See the [extension protocol](docs/extension-protocol.md) and [example manifest](examples/extensions/echo/finishbit-extension.json).

## Development

```bash
go test ./...
go vet ./...
go build ./cmd/fnsh
```

Contribution and review expectations are documented in [CONTRIBUTING.md](CONTRIBUTING.md). Security issues must follow [SECURITY.md](SECURITY.md), not public issues.

## License

Copyright © 2026 fu9zhou and contributors. Licensed under the [GNU Affero General Public License v3.0](LICENSE).
