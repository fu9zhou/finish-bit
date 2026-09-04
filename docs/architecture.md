# Architecture

FinishBit applies ports-and-adapters boundaries so every delivery surface shares one implementation.

```text
CLI now ───────┐
               ▼
Web/API later ─► pkg/app ─► Operation Registry ─┬─ Go core runners
                                                ├─ managed providers
                                                └─ extension processes
                     │
                     ├─ search index
                     ├─ package manager
                     └─ extension manager
```

## Stable seams

- `pkg/operation` owns capability metadata, requests, results, runners, and structured errors.
- `pkg/app` owns use cases. It is the only layer a CLI or future HTTP adapter needs.
- `pkg/search` indexes compact operation metadata and returns bounded matches.
- `pkg/packagemanager` owns runtime versions, downloads, verification, and private installation paths.
- `pkg/extension` owns the manifest and process protocol.
- `internal/cli` parses terminal syntax and renders human or JSON output. It contains no capability implementation.
- `internal/builtin` and `internal/ffmpeg` implement providers behind `operation.Runner`.

## Adding a web adapter

A future server should instantiate `app.App` once, map HTTP request models to `operation.Request`, and map application errors to HTTP status codes. Search, validation, execution, package state, and extension registration remain in existing packages. The server must not invoke `fnsh` as a subprocess or duplicate command parsing.

Long-running web execution will need an additional job/authorization adapter. That is intentionally separate from Operation semantics and is not part of v0.1.

## Dependency direction

Adapters may depend inward on `pkg/app` and public domain contracts. The application layer coordinates registries, packages, extensions, and execution but does not depend on terminal or future HTTP request models. Providers implement the Operation port and do not know which adapter initiated a request.

This gives the repository one execution path:

```text
transport input → application use case → Operation runner → structured result → transport output
```

New delivery surfaces must preserve this direction. Shared validation belongs in Operation or application contracts, not in each adapter.

## Composition and lifecycle

`cmd/fnsh` is the composition root. It resolves the FinishBit data directory, creates the package and extension managers, registers providers, builds the search index, and passes the assembled application to the CLI. Tests may construct these components with isolated directories and controlled dependencies.

Operation registration is complete before requests execute. Duplicate IDs fail registration rather than silently shadowing an existing capability.

## Trust boundaries

- Core runners execute in the `fnsh` process and must validate input before filesystem or compute work.
- Managed providers start checksum-verified third-party executables with explicit arguments.
- Extensions cross a JSON process boundary but are not sandboxed; the protocol isolates representation, not operating-system permissions.
- A future network adapter adds authentication, authorization, quotas, request-size limits, and job ownership outside the Operation contract.

See the [security policy](../SECURITY.md) for report scope and the [compatibility policy](compatibility.md) for public surfaces.

## Design constraints

- Capability metadata and execution must be registered together.
- Public failures use typed `operation.Error` codes.
- Search results remain bounded so catalog growth does not flood agent context.
- External I/O has explicit size, time, or process-lifecycle bounds.
- Web work must deepen `pkg/app` where a shared use case is missing instead of copying CLI behavior.

## Related documents

- [Development guide](development.md)
- [Operations](operations.md)
- [Extension protocol](extension-protocol.md)
- [Package management](package-management.md)
