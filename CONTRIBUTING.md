# Contributing to FinishBit

Thank you for helping build reusable capabilities for AI agents.

## Before opening a change

Use an issue for significant new Operations, protocol changes, new managed packages, or behavior that affects compatibility. Small bug fixes and documentation corrections may go directly to a pull request.

## Development workflow

1. Fork the repository and create a focused branch.
2. Add or update tests with behavioral assertions.
3. Run `go test ./...`, `go vet ./...`, and `go build ./cmd/fnsh`.
4. Update machine-readable metadata and relevant documentation when an Operation contract changes.
5. Open a pull request using the repository template.

Commits follow Conventional Commits with an English type/module, Chinese summary, and concrete Chinese bullet-list body. See `.agents/skills/commit-messages/SKILL.md`.

## Design expectations

- New capabilities implement `operation.Runner` and register through one provider.
- CLI-specific logic stays out of Operations and `pkg/app`.
- Dependencies need pinned versions, HTTPS sources, SHA-256 values, licensing review, and platform coverage.
- Extension protocol changes require a schema version and migration notes.
- Errors exposed to agents use stable structured codes.

By contributing, you agree that your contribution is licensed under AGPL-3.0. Please follow the [Code of Conduct](CODE_OF_CONDUCT.md).
