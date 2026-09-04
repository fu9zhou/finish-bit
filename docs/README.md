# FinishBit documentation

This directory is the documentation entry point for FinishBit. Each topic has one canonical page so that command behavior, compatibility promises, and maintainer procedures do not drift between documents.

## Use FinishBit

| Goal | Document |
| --- | --- |
| Install, upgrade, or remove `fnsh` | [Installation](installation.md) |
| Find commands, JSON behavior, and exit codes | [CLI reference](cli-reference.md) |
| Browse built-in capabilities and Operation rules | [Operations](operations.md) |
| Install and understand managed runtimes | [Package management](package-management.md) |
| Read the Chinese project introduction | [简体中文](README.zh-CN.md) |

## Extend FinishBit

| Goal | Document |
| --- | --- |
| Understand module boundaries and the future Web seam | [Architecture](architecture.md) |
| Build a local extension in any language | [Extension protocol v1](extension-protocol.md) |
| Set up a development environment and add capabilities | [Development guide](development.md) |
| Understand versioning and stability guarantees | [Compatibility policy](compatibility.md) |

Machine-readable contracts live in [`schemas/`](../schemas/). The catalog returned by `fnsh capabilities --json` is authoritative for the installed binary.

## Maintain the project

| Goal | Document |
| --- | --- |
| Prepare and publish a release | [Release process](releasing.md) |
| See planned direction without date commitments | [Roadmap](roadmap.md) |
| Propose changes or submit code | [Contributing](../CONTRIBUTING.md) |
| Understand decision rights | [Governance](../GOVERNANCE.md) |
| Request help | [Support](../SUPPORT.md) |
| Report a vulnerability | [Security policy](../SECURITY.md) |

## Documentation conventions

- Commands are written for `fnsh`; platform-specific shell details are labeled.
- Relative links target the version of the documentation checked out with the code.
- Examples show public behavior, not internal implementation contracts.
- Planned features appear only in the roadmap and must not be interpreted as release commitments.
