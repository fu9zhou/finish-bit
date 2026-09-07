# Batch B acceptance record

Local acceptance completed on **2026-09-07**, Windows x64, Go **1.26.8**. This is a source-workspace result, not a published release or a claim that hosted CI has run.

| Group | New Operations | Managed runtime | Reproduce |
| --- | --- | --- | --- |
| Tables | 27 | qsv 22.0.1 | `./scripts/test-tables.ps1` |
| Documents | 10 | Pandoc 3.11 | `./scripts/test-documents.ps1` |
| Archives | 10 | 7-Zip Extra 26.03 | `./scripts/test-archives.ps1` |

Together with the original 26 Operations and the [97 additions in batch A](batch-a-acceptance.md), the current application and generated Web catalogs contain **170 Operations** before local extensions.

## Delivery order and evidence

The table group was accepted first. The document group's full gate completed before archive implementation started. The archive gate then ran with all earlier managed-runtime test variables enabled, including FFmpeg/ffprobe, PDF, ImageMagick, qsv and Pandoc.

Local machine evidence, generated under `.finishbit-test-batch-b/evidence/` (ignored runtime data):

- Table gate: `37d9871f4bc9462fa6a42e43dd316f58/acceptance.json`.
- Document gate: `46b4e884df5a43a997133bfb88e24216/acceptance.json`, including a structured CLI Markdown→DOCX request and direct CLI text round-trip.
- Archive and combined regression gate: `3f3c8cc543aa4cee98bc49831f28ad7c/acceptance.json`, including structured CLI ZIP creation, direct extraction, SHA-256 comparison and integrity validation.

Every successful script run creates its own evidence directory; these identifiers describe this local run only.

## Checks passed

- Uncached `go test -count=1 ./...` with every new provider's actual managed runtime enabled. Real provider suites completed: archives 15.6 s, documents 148.5 s, FFmpeg 120.6 s, PDF 28.6 s, images 42.4 s, tables 87.0 s. Go runs packages concurrently, so these are not total elapsed times.
- `go vet ./...` and `go build ./cmd/fnsh`.
- `govulncheck` v1.7.0: no reachable Go vulnerabilities found. This scan does not audit external executables.
- Generated website catalog matches all 170 registered Operation contracts; JavaScript syntax checks passed.
- Browser checks covered both new categories (10 cards each), parameter/required-package pages, correct guide links, language switching and both document/archive guides. No browser console errors were reported.
- Repository documentation links checked across README, Chinese README, the documentation index, catalog, guides, package management and roadmap; no missing local links.

The [document guide](documents.md) and [archive guide](archives.md) define exact supported formats, resource rules and limits. The suites verify file content, not merely exit codes. A discovered citeproc network escape was fixed by rejecting implicit citation resource metadata and dependent CSL styles; a local HTTP fixture now checks that these requests do not occur. Archive tests include traversal, expanded-size limits, missing/wrong passwords, marked downloads, cancellation and preservation of original outputs.

## Remaining scope

Excelize remains deferred under the previously recorded upstream parsing-advisory decision. It has not been added as a dependency or advertised as implemented. Pandoc and 7-Zip are registered only for Windows x64 in this delivery. Their documented format/structure limits remain part of the contract. The website remains a static catalog and usage guide; it does not execute local CLI operations. No release, push or website deployment was performed by this acceptance run.

## Pre-commit verification

Before the toolbox commit on 2026-09-07, the combined managed-runtime gate was rerun successfully. Evidence: `.finishbit-test-batch-b/evidence/f19d0e9622ba49e6959c504dad64919a/acceptance.json`; the local console log is `.finishbit-test-batch-b/precommit-validation.log`. Repository tests, vet/build, Go vulnerability scanning and the 170-Operation website catalog check passed.

Browser verification exercised all eight category filters, all twelve guide routes, capability details, Chinese language switching, guide search and a narrow viewport without horizontal overflow. Missing document/archive category icons and stale roadmap delivery statements were corrected. The next planned product release remains `v0.1.4`; Excelize is excluded. No tag, push or deployment is part of this commit.
