# Security policy

## Supported versions

Until the first stable release, security fixes are provided for the latest `0.x` release only.

## Reporting a vulnerability

Do not open a public issue. Use GitHub's **Report a vulnerability** private advisory flow for `fu9zhou/finish-bit`. Include affected versions, reproduction steps, impact, and any suggested remediation.

You should receive an acknowledgement within 72 hours and a status update within seven days. Disclosure timing will be coordinated after a fix is available.

## Trust boundaries

- Managed package downloads are pinned to HTTPS URLs and SHA-256 digests.
- Extensions are third-party executable code. Install only reviewed, trusted extensions.
- Operation input paths are user-controlled. Callers remain responsible for choosing intended inputs and outputs.
