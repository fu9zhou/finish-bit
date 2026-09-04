# Repository guidance

Before creating, amending, rewording, squashing, merging, or reverting a Git commit, read `.agents/skills/commit-messages/SKILL.md` and apply its commit format.

FinishBit keeps delivery adapters thin. Put reusable use cases in `pkg/app`, capability contracts in `pkg/operation`, and implementations behind `operation.Runner`; the CLI and future web adapter must share those paths.

Before completing a code change, run `go test ./...`, `go vet ./...`, and `go build ./cmd/fnsh`.
