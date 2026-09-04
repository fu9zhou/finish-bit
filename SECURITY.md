# Security policy

The FinishBit maintainers take reports about the runtime, CLI, managed-package pipeline, official Agent Skill, and extension protocol seriously.

## Supported versions

| Version | Security fixes |
| --- | --- |
| Latest published `0.x` release | Yes |
| Earlier `0.x` releases | No |
| Unreleased `main` branch | Investigated, but not a supported release |

Before `v1.0.0`, users should upgrade to the latest published release to receive security fixes. A security release may include a narrowly scoped compatibility break when required to remove exposure.

## Report privately

Do not open a public issue, pull request, or discussion for a suspected vulnerability. Use GitHub's [private vulnerability reporting](https://github.com/fu9zhou/finish-bit/security/advisories/new) for this repository.

Include, when available:

- affected FinishBit versions, operating systems, and architectures;
- the affected Operation, package, extension, or protocol surface;
- reproduction steps or a minimal proof of concept;
- realistic impact and required attacker capabilities;
- whether exploitation has been observed publicly;
- suggested remediation or disclosure constraints.

Remove secrets and unrelated personal data. If GitHub private reporting is unavailable, contact the repository owner through the private contact method published on their GitHub profile and reference FinishBit security in the subject.

## What to expect

The targets below are best-effort, not service-level guarantees:

- acknowledgement within 72 hours;
- initial assessment or request for more information within 7 days;
- periodic updates while a confirmed report remains unresolved;
- coordinated disclosure after a fix or mitigation is available.

The maintainer will validate the report, determine affected versions, prepare a fix and advisory, and coordinate credit and disclosure timing with the reporter. Please allow a reasonable remediation window before public disclosure. Reports made in good faith will not be penalized by the project.

## In scope

- arbitrary code execution, path traversal, or sandbox-boundary assumptions caused by FinishBit itself;
- checksum, archive extraction, download, or activation flaws in managed packages;
- extension installation or protocol flaws that violate documented trust boundaries;
- command/argument injection in official providers;
- unintended disclosure, overwrite, or corruption of local data;
- denial of service that bypasses documented resource bounds.

## Usually out of scope

- malicious behavior intentionally implemented by a third-party extension;
- vulnerabilities in upstream FFmpeg or another third-party tool without a FinishBit-specific exploitation path;
- social engineering, physical access, or unsupported modified builds;
- reports based only on automated scanner output without a reproducible security impact;
- missing hardening that the documentation never claims to provide.

Upstream vulnerabilities may still warrant a FinishBit package-registry update. Report them when the pinned artifact exposes FinishBit users.

## Trust boundaries

- Managed package downloads require HTTPS and a registry-pinned SHA-256 digest, but the downloaded program still runs with the current user's permissions.
- Extensions are third-party executable code. Installation validates packaging; it does not establish publisher trust or sandbox execution.
- Operation input and output paths are caller-controlled. Callers must choose intended locations and protect sensitive files.
- The Agent Skill guides tool use but does not create a security boundary around the CLI.

For non-security defects, use the process in [SUPPORT.md](SUPPORT.md).
