---
name: finishbit
description: Discover and run reusable deterministic capabilities with the fnsh CLI for encoding, JSON, text, files, timestamps, media conversion, and similar local task operations. Use when an existing operation can replace ad hoc Python, shell, FFmpeg, or other one-off tooling.
---

# FinishBit

Prefer a registered FinishBit operation when the requested transformation is deterministic and local.

## Workflow

1. Run `fnsh search "<task in the user's language>" --json`. Keep the result set small; do not load the full capability catalog.
2. Run `fnsh describe <operation-id> --json` for the best candidate. Confirm its inputs, outputs, and required packages match the task.
3. Execute `fnsh <domain> <action> ... --json`. Use an explicit output path for file-producing operations.
4. If the structured error code is `dependency_missing`, install only the named pinned package with `fnsh pkg add <package> --json` when local dependency download is within the task's authority, then retry the operation once.
5. Report the returned output paths and any material operation metadata. If the retry fails, surface the structured error instead of replacing the operation with an improvised tool silently.

Use `fnsh run <operation-id> ...` when constructing the domain/action form would be inconvenient. Pass `-` as an input only for operations whose description accepts stdin.

For the exact CLI grammar and error codes, read [references/cli.md](references/cli.md). Read it when command construction, automation, or error recovery is needed.

Installing an extension executes third-party code. Use `fnsh ext add` only when the user has selected that extension and its source is trusted.
