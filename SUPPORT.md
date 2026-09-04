# Support

FinishBit is maintained as an open-source project. Support is best-effort and no response-time or service-level guarantee is provided.

## Where to ask

- **Reproducible bug:** open a [bug report](https://github.com/fu9zhou/finish-bit/issues/new?template=bug.yml).
- **New Operation or design proposal:** open a [capability proposal](https://github.com/fu9zhou/finish-bit/issues/new?template=feature.yml).
- **Security vulnerability:** do not open an issue; follow [SECURITY.md](SECURITY.md).
- **Usage question:** first check the [documentation index](docs/README.md), CLI `help`, and `describe` output. If the answer is missing and the problem may indicate a documentation defect, open a bug report with the question and context.

Please do not use private maintainer contact details for routine support. Public issues make answers searchable and allow others to help.

## Information to include

Provide the smallest safe reproduction and include:

- `fnsh version` output;
- operating system and architecture;
- the exact command with secrets and sensitive paths removed;
- expected and actual behavior;
- structured error output from `--json` when available;
- whether `FINISHBIT_HOME`, a managed package, or an extension is involved.

For large media or private data, create a minimal non-sensitive sample or describe its relevant properties. Never post credentials, personal data, private source, or confidential files.

## Scope

Maintainers can help with FinishBit behavior, documented installation paths, built-in providers, and the extension protocol. Support for third-party extensions, custom builds, downstream packaging, modified source, or the internals of managed third-party tools belongs with their respective maintainers unless the issue is reproducible in an official FinishBit release.

Issues may be closed when they cannot be reproduced, lack requested information, fall outside project scope, duplicate an existing report, or have been inactive after a clear request for follow-up. They can be reopened when the missing evidence becomes available.
