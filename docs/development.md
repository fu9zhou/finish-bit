# Development guide

This guide covers changes to the FinishBit runtime, built-in Operations, providers, CLI adapter, and documentation. Extension authors can instead begin with the [extension example](../examples/extensions/echo/README.md).

## Prerequisites

- Go 1.25.13 or later; CI currently builds with Go 1.26.8
- Git
- Node.js when regenerating or checking the static website catalog (`node scripts/website-catalog.mjs --check` after building `fnsh`)
- A supported Windows, Linux, or macOS development environment
- Document and archive acceptance: `./scripts/test-documents.ps1` and `./scripts/test-archives.ps1` install their managed runtimes, run actual file workflows, and verify repository/Web contracts.
- Global FFmpeg is optional for older media tests. The [batch A acceptance script](batch-a-acceptance.md) installs pinned managed runtimes and enables the real media/PDF/image suites on Windows x64.

Clone the repository and run the baseline checks:

```bash
git clone https://github.com/fu9zhou/finish-bit.git
cd finish-bit
go test -race ./...
go vet ./...
go build ./cmd/fnsh
```

ZIP and tar.gz runtime distributions use the Go standard library; tar.xz uses `github.com/ulikunitz/xz`, locked in `go.mod` and `go.sum`. Pinned 7z distributions use an existing managed 7-Zip reader or the official raw `7zr.exe` bootstrap. FinishBit validates the listing and writes each bounded stdout stream itself. Mature dependencies may be introduced when they enable a coherent capability group; record their purpose and keep heavyweight engines optional.

## Repository map

| Path | Responsibility |
| --- | --- |
| `cmd/fnsh` | Composition root and version injection |
| `internal/cli` | Argument parsing and terminal/JSON rendering |
| `pkg/app` | Transport-independent use cases shared by all adapters |
| `pkg/operation` | Operation contracts, registry, requests, results, and typed errors |
| `pkg/search` | Bounded capability discovery |
| `internal/builtin` | Deterministic Go Operations |
| `internal/ffmpeg` | Media provider implemented through managed FFmpeg |
| `internal/pdf`, `internal/raster`, `internal/tabular` | Managed PDF, image and table capability groups |
| `internal/toolrun` | Shared bounded subprocess execution and staged artifact publication |
| `pkg/packagemanager` | Pinned runtime acquisition and lifecycle |
| `pkg/extension` | Extension installation and process protocol |
| `schemas` | Machine-readable public contracts |
| `skills/finishbit` | Agent-facing discovery and execution guidance |

See [architecture](architecture.md) for dependency direction and adapter boundaries.

## Add a built-in Operation

1. Define stable task-oriented metadata: ID, summary, aliases, tags, inputs, options, and failure behavior.
2. Implement `operation.Runner` in the narrowest appropriate provider.
3. Register it once through that provider. Do not add command-specific execution logic to `internal/cli`.
4. Add behavioral tests for valid input, boundaries, failures, and deterministic output.
5. Update [operations](operations.md) and any affected machine-readable schema.
6. Confirm direct dispatch, `run`, search, describe, and JSON output all expose the same contract.

Prefer a core Go implementation for small deterministic behavior. Use a managed provider for a large established runtime. Use the extension protocol when capability ownership or dependencies should remain outside the core distribution.

## Change a public contract

Operation IDs, parameter names/types, structured error codes, CLI JSON, extension protocol fields, data layout, and package registry behavior may affect downstream agents. Follow the [compatibility policy](compatibility.md), add migration notes, and call out the impact in the pull request.

Protocol changes must update the corresponding JSON Schema and fixtures. A breaking protocol shape requires a new schema/protocol version rather than silent reinterpretation.

## Testing

Before every pull request, run:

```bash
go test -race ./...
go vet ./...
go build ./cmd/fnsh
```

For platform-sensitive changes, cross-build with `CGO_ENABLED=0` for the affected combinations in the support matrix. Tests must use temporary or isolated `FINISHBIT_HOME` directories and must not depend on the developer's installed extensions or packages.

Documentation-only changes should at minimum pass `git diff --check` and local-link validation.

## Commits and pull requests

Commits use Conventional Commits with an English type and module, a Chinese action summary, and a detailed Chinese bullet-list body. The repository copy of the full convention is `.agents/skills/commit-messages/SKILL.md`.

Keep pull requests focused on one purpose, explain compatibility and security effects, and record the exact validation performed. The complete workflow is in [CONTRIBUTING.md](../CONTRIBUTING.md).

The document provider in `internal/document` registers managed Pandoc workflows through the same application registry; see [document contracts](documents.md).

The archive provider in `internal/archive` wraps managed 7-Zip. Shared `toolrun.RunToFile` bounds streamed output, and `DirectoryTree` stages nested results before publishing a new destination.
