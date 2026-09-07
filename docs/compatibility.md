# Compatibility policy

FinishBit uses [Semantic Versioning](https://semver.org/) for releases and treats compatibility as part of the product contract.

## Before v1.0.0

The current release sequence is `v0.1.3` → `v0.1.4` → `v0.1.5`, continuing patch increments for backward-compatible fixes and additive capabilities. Adding dependency groups or many Operations does not change this sequence. Moving to another series requires an explicit project-owner decision; a requested `v0.2.x` series can continue as `v0.2.1`, `v0.2.2`. No automatic jump to a minor or major series is authorized by feature size.

Every known breaking change must be called out in the changelog with a migration path when one exists. Resolve its scope with the owner instead of silently changing the release series.

Users who require reproducible automation should pin a complete version rather than following `latest`.

## Public contracts

The following surfaces are compatibility-sensitive:

- command names, documented flags, exit-code categories, and JSON output envelopes;
- Operation IDs, inputs, options, result fields, requirements, and structured error codes;
- extension manifest and process protocol versions;
- published JSON Schemas;
- managed-package names and persisted data required across upgrades;
- exported Go packages under `pkg/`.

Human-readable messages, search ranking among similarly relevant results, internal package structure, and undocumented implementation details are not stable APIs.

## Operation evolution

Compatible changes include adding a new Operation, alias, tag, optional field, or optional parameter with a safe default. Removing or renaming an Operation or field, changing a type or meaning, making optional input required, or changing deterministic output semantics is breaking.

Deprecated behavior should remain available through at least the next minor release when practical. Deprecation notices belong in `describe` metadata or documentation, the changelog, and the replacement Operation's guidance.

## Extension protocol

The integer `schema` and `protocol` fields identify the contract version. Implementations reject unsupported versions instead of guessing. Additive changes that older readers safely ignore may remain within a version; incompatible message or manifest changes require a new version and parallel migration support where practical.

See [extension protocol v1](extension-protocol.md) for the current wire contract.

## Platform support

CI tests Windows, Linux, and macOS. Release builds target amd64 and arm64 with `CGO_ENABLED=0`. A platform is supported only when it appears in the [installation matrix](installation.md) and in CI/release configuration.

Managed providers may support a narrower matrix based on verified upstream artifacts. `fnsh pkg info <name>` is authoritative for the current binary and platform.

## Security exception

A security fix may intentionally break compatibility when preserving old behavior would leave users exposed. Maintainers will minimize the scope and document the reason and migration.
