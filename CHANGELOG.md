# Changelog

All notable changes are documented here, in [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) format. This project uses [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.1.0] - 2026-11-10

### Added
- Compares the selected dependency versions of a Go module at two revisions, from `--old DIR --new DIR` or `--base GITREF`, using `go mod graph` and `go list -m -json all`.
- For every raised, added, lowered or removed module, names the dependency that required the version and the chain back to `go.mod`, or says that only `go.mod` asks for it.
- Explains a raised `go` directive by the selected dependencies that need a newer Go, and reports `toolchain` line changes.
- `requirement-dropped` when go.mod loses a line but the build stays the same, so a go.mod that lost lines is never reported as unchanged.
- A "pulled in through" summary of the root modules behind the changes.
- `--expect PATH[@VERSION]` to mark intended changes, `--format json`, `--fail-on`, and exit 2 for any input it can't trust.
- `explaining-go-dependency-bumps` skill and plugin manifest.
