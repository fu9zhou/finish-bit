# Contributing to FinishBit

Thank you for helping make reusable, deterministic capabilities available to AI agents. Contributions may include code, tests, documentation, issue triage, protocol feedback, and extension examples.

By participating, you agree to follow the [Code of Conduct](CODE_OF_CONDUCT.md). By submitting a contribution, you agree that it may be distributed under the repository's [AGPL-3.0 license](LICENSE). The project does not currently require a CLA or DCO sign-off.

## Before starting

Search existing issues and pull requests before opening a duplicate. Use an issue first for:

- a new Operation or managed package;
- an extension protocol, schema, or persistent-data change;
- a compatibility break or deprecation;
- a new external dependency;
- a broad refactor or security-sensitive design.

Small bug fixes, tests, and documentation corrections may go directly to a pull request. Security vulnerabilities must follow [SECURITY.md](SECURITY.md), not the public issue tracker.

Maintainer agreement on a proposal is not a promise that it will be merged; implementation and review may reveal new constraints.

## Development setup

FinishBit requires Go 1.25.13 or later. CI and release archives currently use Go 1.26.8:

```bash
git clone https://github.com/fu9zhou/finish-bit.git
cd finish-bit
go test -race ./...
go vet ./...
go build ./cmd/fnsh
```

See the [development guide](docs/development.md) for repository boundaries, Operation design, platform testing, and public-contract changes.

## Make a focused change

1. Fork the repository and create a branch from the current `main`.
2. Keep the change scoped to one purpose; separate unrelated cleanup.
3. Add behavioral tests before or with implementation changes.
4. Update machine-readable metadata and documentation whenever public behavior changes.
5. Run the checks below and inspect the final diff for accidental generated files or local data.
6. Open a pull request using the template and link its issue when applicable.

### Required validation

```bash
go test -race ./...
go vet ./...
go build ./cmd/fnsh
git diff --check
```

Run cross-platform or integration checks in proportion to the change. A documentation-only pull request does not need unrelated runtime tests, but its local links and examples must remain valid.

## Design expectations

- New capabilities implement `operation.Runner` and register through exactly one provider.
- CLI-specific parsing and rendering stay out of Operations and `pkg/app`.
- A future Web/API adapter must call `pkg/app`, not execute `fnsh` as a subprocess or duplicate CLI logic.
- Operation IDs and machine-facing errors are stable contracts, not display strings.
- Dependencies require justification, pinned versions where applicable, HTTPS sources, SHA-256 verification, license review, and explicit platform coverage.
- Extension protocol changes require a schema version, compatibility analysis, fixtures, and migration notes.
- Files, archives, processes, downloads, and output streams need explicit resource bounds and safe failure behavior.

## Commit convention

Commits follow Conventional Commits with:

- an English lowercase type and module;
- a Chinese action summary;
- a detailed Chinese bullet-list body explaining material changes and validation.

The repository's complete, executable convention is [`.agents/skills/commit-messages/SKILL.md`](.agents/skills/commit-messages/SKILL.md). Keep commits reviewable and centered on one primary purpose.

## Pull-request review

Reviewers evaluate both implementation quality and conformance to the originating proposal. A pull request should state:

- the user-visible outcome;
- important design choices and alternatives;
- compatibility and migration effects;
- security or supply-chain effects;
- exact validation performed;
- documentation and schema updates.

Address review comments with new commits while review is active unless a maintainer asks for a rebase or squash. Maintainers may edit titles or squash commits during merge to preserve repository history conventions.

## Documentation and user-facing language

The English README and [documentation index](docs/README.md) are the canonical navigation layer. Keep examples executable, distinguish current behavior from roadmap items, and avoid promises about dates or support levels that the project cannot guarantee.

Machine-facing identifiers and code use English. Chinese summaries are required for repository commit messages; user documentation may be translated when its canonical source is clearly linked.

## Licensing and provenance

Do not contribute code, assets, datasets, binaries, or generated material unless you have the right to license them under the project terms. Preserve required notices, identify copied or adapted material in the pull request, and document licenses for managed tools and bundled examples.
