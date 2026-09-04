# Operations

An Operation is a stable task contract: identifier, searchable metadata, ordered inputs, typed options, runtime requirements, and one runner.

## Contract anatomy

- `id` is the lowercase `<domain>.<action>` machine identifier.
- `summary`, `description`, `aliases`, and `tags` support human and agent discovery.
- `inputs` are ordered positional values.
- `options` are named, typed values with required/default semantics.
- `requirements` identify managed packages needed at execution time.
- `source` identifies the core, a managed provider, or an installed extension.

Use `fnsh describe <id> --json` instead of hard-coding display text. Consumers should treat IDs, parameter names/types, result fields, and error codes according to the [compatibility policy](compatibility.md).

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

Registration of a duplicate or invalid definition fails instead of choosing one implementation implicitly. An Operation must return an `operation.Result` on success or a structured `operation.Error` on failure.

## Choosing an implementation type

| Type | Use when | Examples |
| --- | --- | --- |
| Core | Behavior is small, deterministic, and practical in Go | JSON, text, hash |
| Managed provider | A mature external runtime is large or specialized | FFmpeg media Operations |
| Extension | Capability lifecycle or dependencies should stay outside core | Community integrations |

All three types share discovery and execution contracts. Adding a provider must not require a new CLI dispatch path.

## Discoverability requirements

Operation names describe user outcomes rather than implementation tools. Summaries should be concise, aliases should include common task language, and tags should improve filtering without repeating every word. Search ranking is intentionally not a stable API; the set of registered definitions is.

See the [development guide](development.md) before adding or changing an Operation.
