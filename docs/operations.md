# Operations

An Operation is a stable task contract: identifier, searchable metadata, ordered inputs, typed options, runtime requirements, and one runner.

## Core catalog in v0.1

| Domain | Operations |
| --- | --- |
| Base64 | `base64.encode`, `base64.decode` |
| Hash | `hash.calculate` |
| JSON | `json.format`, `json.minify`, `json.query`, `json.validate` |
| Text | `text.count`, `text.replace`, `text.sort`, `text.unique` |
| Time | `time.convert` |
| UUID | `uuid.generate` |
| File | `file.info`, `file.checksum` |
| URL | `url.encode`, `url.decode` |
| Video | `video.trim`, `video.compress` |
| Audio | `audio.extract` |

The machine-readable catalog from the built binary is authoritative:

```bash
fnsh capabilities --json
fnsh describe json.query --json
```

## Registration invariant

Each ID uses lowercase `<domain>.<action>` segments and is unique across core and extensions. Registering once automatically enables search, describe, direct CLI dispatch, generic `run`, JSON schemas in responses, and future adapters through `pkg/app`.
