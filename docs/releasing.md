# Release process

This document is for FinishBit maintainers. Releases are built by GitHub Actions from version tags through GoReleaser; release binaries must not be built and uploaded manually as an undocumented side path.

## Versioning

Choose the version according to the [compatibility policy](compatibility.md):

- patch for backward-compatible fixes and additive capabilities during the planned v0.1.x series;
- minor for a documented pre-1.0 breaking change or a larger feature milestone;
- major for stable-contract breaking changes after `v1.0.0`.

Tags use `vMAJOR.MINOR.PATCH`. A semantic prerelease suffix marks preview releases.

## Preflight checklist

1. Confirm the intended commit is on `main` and CI is green.
2. Ensure the working tree is clean.
3. Move relevant entries from `Unreleased` into a versioned, dated section in `CHANGELOG.md` and restore an empty `Unreleased` section.
4. Confirm compatibility breaks, migrations, security notes, and managed-runtime changes are explicit.
5. Run `go test -race ./...`, `go vet ./...`, and `go build ./cmd/fnsh`.
6. Run `go run golang.org/x/vuln/cmd/govulncheck@v1.7.0 ./...` with the supported Go patch version used by CI.
7. Validate `.goreleaser.yml` with the exact GoReleaser version used by CI.
8. Review third-party artifact URLs, digests, and licenses if the package registry changed.

The changelog commit must land before the release tag.

## Publish

Create and push one annotated version tag pointing at the reviewed commit. The `release.yml` workflow then:

- reruns tests;
- builds Windows, Linux, and macOS archives for amd64 and arm64;
- injects the version into `fnsh`;
- includes the license, README, and Agent Skill;
- publishes `checksums.txt` and a GitHub Release.

No release is complete until the workflow succeeds and the assets have been verified.

## Verify

After publication:

1. Download at least one archive from the public release page and verify it against `checksums.txt`.
2. Run `fnsh version`, `fnsh doctor`, one core Operation, and `fnsh capabilities --json` from the archive.
3. Confirm installer-generated archive names match the published assets.
4. Review the rendered release notes and prerelease/latest designation.
5. If verification fails, document the failure publicly and publish a corrected new version; do not silently replace immutable release artifacts.

## Rollback policy

Git tags and published artifacts are immutable. A bad release is superseded by a new patch release or, when exposure is severe, marked as affected with clear upgrade guidance. Security incidents follow [SECURITY.md](../SECURITY.md).
