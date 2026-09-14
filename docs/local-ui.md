# Local web UI

Run `fnsh ui` (alias: `fnsh web`) to start the local toolbox and open your default browser. UI assets are embedded in the executable: no Node.js, separate UI installation, or CDN is required.

```sh
fnsh ui
fnsh ui --port 8080
fnsh web --no-open
```

`fnsh ui` defaults to `fnsh ui start`: it ensures one background service is running, prints its live status and URL, and opens the browser. Closing the terminal does not stop the service. `fnsh ui status` checks the authenticated listener and prints its URL; `fnsh ui stop` stops it. Repeated starts reuse the same service and session. Use `fnsh ui run` for foreground debugging (Ctrl+C stops it). This is a per-user background process, not an auto-start service after reboot.

The default port is selected automatically. `--no-open` prints a link to open manually. To change an existing service's port, stop it first. Open the complete link including its fragment to establish a browser session. Each new server process creates a new session secret. State and logs live under `FINISHBIT_HOME/ui` (or the default FinishBit data directory). OS locks serialize lifecycle commands and prevent duplicate daemons; stop uses authenticated loopback HTTP rather than killing a recorded PID. The sidebar has one shared refresh control for all pages; package progress continues updating automatically.

The UI provides capability search and category filtering, forms generated from operation contracts, structured execution results, package installation/repair/removal with download progress, local extension installation/removal, and environment checks. The catalog includes installed and uninstalled dependencies and identifies unsupported platforms.

Each tool has a navigable subpage, for example `#tools/archive.extract`. Browser back/forward and direct page refresh work. Draft inputs and results are retained in memory while navigating or changing language; refreshing the browser starts fresh drafts. Chinese and English can be selected in the header, and the preference is saved locally. Interface labels and choices are translated; canonical provider descriptions, processing limits, third-party diagnostics, and result keys remain available in their original language.

File parameters provide an inline local file browser. Select existing files/directories, append multiple input files, or choose a parent folder and type an output file/new directory name. Selection returns absolute paths and does not create directories, copy files, or upload contents. Text parameters remain text; archive member names, JSON paths, and column names are not filesystem pickers. Explicit `input-mode=file` switches all of that operation's inputs to file selection. Terminal stdin (`-` or `input-mode=stdin`) is not supported in the UI. Relative manually entered paths resolve from the directory where `fnsh ui` was launched. Output files are written by the existing runners, and results display those paths.

Enumerated settings use dropdowns, boolean settings use switches, integer settings use numeric controls, and free-form values remain text. Being optional alone does not make a field an enumeration.

Package installation and repair show live progress inside the package row. Known download sizes show a percentage and transferred/total bytes; unknown sizes show transferred bytes with an indeterminate indicator. Connecting, mirror fallback, verification, extraction, and readiness are separate stages: 100% downloaded does not mean installation has finished. The download source hostname is shown without URL credentials. Both CLI and UI installs publish the same durable package task snapshots through the package manager; `progress.mjs` formats them without duplicating installer logic.

Package tasks are stored as atomic JSON snapshots in `<FINISHBIT_HOME>/tasks` (the normal default FinishBit root is used when the variable is unset). Every install/repair, including CLI calls, records its ID, package, origin, owner PID, timestamps, progress, and terminal state. The authenticated `/api/package-tasks` endpoint reads snapshots once per second in the UI. This works across processes and without a UI server running when the CLI download starts. Both processes must use the same root and a version that supports task records. The view returns the latest 100 records plus active tasks; task files remain local history.

UI package installs/repairs use a background task owned by the UI server context; refreshing or closing the browser does not cancel them. Reopening the page restores running or completed package state. Stopping the owning server or CLI process stops its work; this feature does not resume a killed download. A heartbeat is saved every 5 seconds. After 30 seconds without a heartbeat, readers probe the OS package lock: an occupied lock shows progress recovery pending, a free lock with a still-stale snapshot shows the task stopped without completion confirmation, and a lock inspection error remains unknown. A snapshot is re-read under a free lock to catch concurrent completion; readers never rewrite history or infer success from a previously installed package. The UI polls every second, blocks duplicate actions while work is active or unverified, labels stale progress as historical, shows report/check times, and offers a status check or installation retry. Download-source errors retain their diagnostics. Reopen the new complete startup link after restarting the UI server.

Per-package OS locks prevent CLI/UI install, repair, and removal from modifying the same package simultaneously. Locks are released by the OS on process exit. Separate packages may run concurrently. UI task reads do not take the action mutex, so progress remains observable during execution. General operation drafts/results and non-package actions still use the existing in-memory/NDJSON path; only package install/repair has background ownership and durable progress.

The server binds only to `127.0.0.1` and checks Host, Origin, and a per-launch bearer secret for API access, including directory listings. Treat the complete startup URL as access to your local toolbox. Catalog/action requests are serialized; directory browsing is read-only and remains available independently. CLI processes remain independent and coordinate package mutations through per-package OS locks. Closing the server cancels request contexts; outputs already produced by operations remain on disk. Dependency downloads require network access and use the same pinned registry and integrity checks as the CLI.

Implementation: `internal/web` is a delivery adapter over `pkg/app`; operations continue through `operation.Runner`. Static UI files in `internal/web/static` need no frontend build step. Run `go test ./...`, `go vet ./...`, and `go build ./cmd/fnsh` to validate and package changes.

The package view automatically inspects interrupted history once per page session. **Show history** displays the evidence and checks it if it has not been loaded yet; use the shared refresh control to clear cached evidence before checking again. The authenticated `GET /api/package-tasks?inspect=<task-id>` use case takes the package lock, verifies current installed files and the primary archive cache, and reports remaining staging downloads. Hash checks run on demand rather than in heartbeat polling. Current files and unlinked residuals are evidence about the package, not proof that an old task completed. New installations record their task ID in staging names and installed metadata; matching metadata plus valid file integrity can confirm completion even if the final task snapshot was lost. Historical snapshots remain unchanged. Deleted partial downloads cannot be reconstructed from a progress percentage.

## Architecture

`CLI / Web → pkg/app → operation.Runner → provider`

The two delivery adapters share execution, package/extension management, discovery, and validation. The web adapter handles HTTP authentication and progress streaming; it does not assemble shell commands or implement tool behavior.

| Module | Responsibility |
| --- | --- |
| `pkg/operation.Parameter` | Additive `kind` path semantics and `choices` discovery contract. Runners remain authoritative for validation. |
| `pkg/app/metadata.go` + `catalog-metadata.json` | Explicit compatibility metadata for legacy built-ins, keyed by operation ID and parameter name. Both `Describe` and `Capabilities` apply it without mutating the registry. No UI guesses based on English prose. |
| `pkg/app/files.go` | Portable bounded directory listings with absolute paths; no file reads, uploads, or output creation. |
| `internal/web/server.go` | Loopback HTTP delivery, session/origin checks, static assets, and event streaming. |
| `router.mjs` | Catalog/subpage URL parsing, independent of view rendering. |
| `api.mjs` | Authenticated requests, package task creation/polling, and streamed operation event parsing. |
| `pkg/packagemanager/tasks.go` | Cross-process task snapshots, heartbeat, background ownership, and package mutation locks. |
| `tasks.mjs` | Restores latest package results and preserves active work over rejected duplicate attempts. |
| `progress.mjs` | Download byte formatting and stage/percentage presentation, including unknown totals. |
| `i18n.mjs` | Locale preference, interface copy, parameter labels, and choice labels. |
| `forms.mjs` | Typed draft state, request serialization, and metadata-driven controls. |
| `file-picker.mjs` | Local browsing and path selection, reused for tools and extensions. |
| `app.mjs` | Navigation, catalog views, and in-memory draft/result ownership. |

New providers and extensions should declare `kind` and `choices` directly on parameters. Omitted metadata remains backwards compatible and renders as the existing text/number/boolean type. Legacy records must match actual provider parameters; tests catch stale IDs and missing defaults. Since capabilities JSON gains metadata, regenerate the website catalog with `node scripts/website-catalog.mjs` after building.

The small UI uses native ES modules and embedded assets to preserve the single-binary installation. Node is needed only for developer tests (`node --test internal/web/ui_test.mjs`), not for users. The same tests run in CI alongside the Go and website checks.

## Toolbox discovery

The home page uses a centered 1320px workspace beside the navigation and nine category cards. Navigation has three levels: `#tools` (home), `#categories/media` (category tools), and `#tools/audio.convert` (operation). Category links support refresh and browser history; operation pages link back to their category and expose clickable breadcrumbs. Search is global on the home page and scoped on category pages, with separate in-memory queries. Recently opened tools remain compact shortcuts. Tool lists use four/two/one columns. Search intersects whitespace-separated terms against names, aliases, summaries, tags, and localized category labels. Recently opened tool IDs are stored in browser local storage; no file paths or input values are stored there.

Loading, unavailable catalog, stale catalog, and empty search results are distinct UI states. Expired sessions pause task polling; navigating to a new complete startup link updates the session in the existing page. Offline tool forms disable execution while retaining the loaded catalog and drafts.

Design references reviewed: [iLovePDF](https://www.ilovepdf.com/) for task categories and tool cards, and [UtiliTools](https://utilitools.dev/) for search, category browsing, and quick discovery. FinishBit keeps its local execution and dependency installation model rather than introducing a file-upload flow.

Category illustrations in `internal/web/static/icons/` are nine original SVG assets on a 64-unit grid with rounded 2.2-unit strokes and muted category colors. They appear only on the home category cards; category lists and operation pages use compact text without logos. PDF and documents have separate categories. They are decorative beside accessible text labels and ship inside the executable; no external icon service or font is required.
