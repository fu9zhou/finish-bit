# `fnsh` agent contract

## Discovery

```bash
fnsh search "<natural-language task>" --limit 5 --json
fnsh describe <operation-id> --json
fnsh capabilities --json
```

Use `capabilities` only for audits or catalog generation. Ordinary task execution starts with `search`.

## Execution

Both forms call the same Operation runtime:

```bash
fnsh video trim input.mp4 --start 10s --duration 20s --output clip.mp4 --json
fnsh run video.trim input.mp4 --start 10s --duration 20s --output clip.mp4 --json
```

Positional values map to the described `inputs` in order. Long flags map to `options`; `-o` aliases `--output`. Boolean flags may be passed without a value.

Success writes one JSON value to stdout. Failure writes `{"error": ...}` to stderr and uses these exit codes:

- `2`: invalid input or unknown operation.
- `3`: required managed package is missing.
- `1`: execution, integrity, platform, or other runtime failure.

Relevant structured error codes include `invalid_input`, `operation_not_found`, `dependency_missing`, `execution_failed`, `integrity_failure`, and `unsupported_platform`.

## Managed dependencies

```bash
fnsh pkg add ffmpeg --json
fnsh pkg info ffmpeg --json
fnsh pkg repair ffmpeg --json
```

FinishBit selects the compatible version and verifies its pinned SHA-256. Do not append a user-selected version.
