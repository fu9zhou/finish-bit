# Roadmap

The roadmap is capability-driven rather than a promise of dates.

## Available in v0.1

- Stable Operation registry, progressive search, describe, and execution.
- Core Go operations for encoding, JSON, text, time, UUID, files, and URLs.
- Managed FFmpeg installation plus video trim/compress and audio extraction.
- Local directory/ZIP extension installation and process protocol v1.
- Agent Skill and structured CLI output.

## Candidate next work

- Image resize/convert providers with runtime selection.
- PDF merge/split and document conversion Operations.
- Signed remote extension registry and publisher identity.
- Hybrid lexical/vector search when catalog size justifies it.
- HTTP/job adapter built solely on `pkg/app`.

Protocol stability, package supply-chain security, and cross-platform behavior take precedence over operation count.

## How priorities are chosen

Roadmap order reflects user task coverage, reuse across agents, deterministic behavior, maintenance cost, supply-chain risk, and fit with the existing Operation model. A listed item is a direction, not a commitment to a version or date.

Concrete proposals should use the repository's capability issue template and include the proposed Operation contract. Accepted proposals may change as implementation reveals compatibility or security constraints. See [governance](../GOVERNANCE.md) for the decision process.
