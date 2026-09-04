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
