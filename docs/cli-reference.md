# CLI reference

`fnsh` is both the human interface and the stable local adapter for agents. Run `fnsh help` for the concise command summary and `fnsh describe <operation>` for an Operation's live contract.

## Command model

```text
fnsh search <query> [--limit N] [--json]
fnsh describe <operation> [--json]
fnsh <domain> <action> [inputs...] [options...] [--json]
fnsh run <operation> [inputs...] [options...] [--json]
fnsh capabilities [--json]
fnsh pkg <add|remove|ls|info|repair> [name] [--json]
fnsh ext <add|remove|ls|info> [name-or-path] [--json]
fnsh doctor [--json]
fnsh version [--json]
```

`--json` may appear anywhere in the argument list. Operation inputs are positional and options use `--name value` or `--name=value`. Boolean options may be passed as flags. Options with the `strings` type may be repeated to collect multiple values. `-o` is an alias for `--output`.

Paths containing spaces must be quoted according to the current shell.

## Discovery

### `search`

Searches IDs, summaries, aliases, and tags. Results are relevance-ranked and bounded; the default limit is 5.

```bash
fnsh search "裁剪视频" --limit 3 --json
```

### `describe`

Returns the selected Operation's inputs, typed options, defaults, runtime requirements, and source.

```bash
fnsh describe video.trim --json
```

### `capabilities`

Lists every registered core, managed-provider, and installed-extension Operation. This output is the authoritative catalog for the current binary and local environment.

## Execution

An Operation can use its direct two-segment command or the generic `run` form:

```bash
fnsh json format data.json --indent 2
fnsh run json.format data.json --indent 2
```

The forms resolve to the same Operation and application service. Use `-` as an input where the Operation description explicitly supports stdin.

## Managed packages

| Command | Purpose |
| --- | --- |
| `fnsh pkg ls` | List installed managed packages |
| `fnsh pkg info <name>` | Show registry and installation state |
| `fnsh pkg add <name>` | Download, verify, and activate a package |
| `fnsh pkg repair <name>` | Reinstall the pinned package artifact |
| `fnsh pkg remove <name>` | Remove the package from FinishBit's data directory |

See [package management](package-management.md) for provenance and integrity guarantees.

## Extensions

| Command | Purpose |
| --- | --- |
| `fnsh ext ls` | List installed extensions |
| `fnsh ext info <name>` | Show an installed manifest |
| `fnsh ext add <directory-or-zip>` | Validate and install an extension |
| `fnsh ext remove <name>` | Remove an installed extension |

Installing an extension does not execute it. Invoking one of its Operations does. See the [extension protocol](extension-protocol.md) for the trust model.

## Diagnostics and version

`fnsh doctor` reports whether configured runtime dependencies are ready. A failed check produces a non-zero exit status. `fnsh version` reports the build version; source builds may report a development value.

## Output contract

Human mode favors readable text and paths. JSON mode is intended for agents and automation:

- successful output is one JSON value on stdout;
- errors are one object of the form `{"error": {"code": "...", "message": "..."}}` on stderr;
- diagnostic and error text never shares stdout with a successful JSON result.

Consumers should branch on `error.code`, not parse English messages.

## Exit codes

| Code | Meaning |
| ---: | --- |
| `0` | Success |
| `1` | Execution, integrity, platform, or unexpected failure |
| `2` | Invalid input or unknown Operation |
| `3` | Required managed package is missing |

The structured error code provides more detail than the process exit code. Current codes are `invalid_input`, `operation_not_found`, `dependency_missing`, `execution_failed`, `integrity_failure`, and `unsupported_platform`.

## Literal inputs and binary output (v0.1.1)

Use `--` before positional inputs containing flag-like text. All remaining arguments are literal inputs; place options before this delimiter:

```bash
fnsh text replace -- abc abc --json
fnsh --json base64 encode -- --hello
fnsh json format --help
```

Named option values are consumed as data, including a value equal to `--json`.
Base64 decoding writes exact bytes in human mode without adding a newline. For non-UTF-8 binary data, JSON mode returns `data.base64` (standard Base64) and `data.encoding: "base64"`; it does not return lossy text. Use `-o decoded.bin` for binary files. UTF-8 data continues to use `data.text` in JSON mode.

Damaged or conflicting extensions are excluded from discovery and reported as failed doctor checks. Core Operations and `ext remove <name>` remain available. A newly installed or removed extension is reflected immediately in the application registry.
