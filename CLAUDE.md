# whybumped

Claude Code skill and Go CLI that compares the Go module graph of two revisions of a module and explains each change in the selected dependency versions: the dependency that required the new version, the chain back to go.mod, and the dependencies that forced the go directive up.

## Architecture
`plugin/scripts/whybumped.go` holds all the logic in one standard-library file, so `go run` needs no setup. `loadSide` checks `go.mod` as text (`checkGoMod`), runs `go mod graph` (`parseGraph`) and `go list -m -mod=readonly -json all` (`parseModules`) with `GOWORK=off`, and `crossCheck` recomputes minimal version selection from the graph and compares it with go list. `compare` diffs the selected versions; `explainUp`, `explainDown` and `explainRemoved` find the requirers of each version, and `chainFor` walks requirers back to a module that only go.mod requires. `requirementLines` reports go.mod lines that changed without changing the selection. `goDirective` explains a changed go line. `extractGitRev` reads `--base` with `git archive`. `plugin/skills/explaining-go-dependency-bumps/SKILL.md` tells Claude how to use it. See `docs/architecture.md`.

## Commands
- `make test`: gofmt, vet, unit tests and the integration tests in `tests/` (no network: they use the file module proxy in `examples/proxy`)
- `make test-real`: runs the tool on real releases fetched from GitHub and the module proxy (needs network)
- `make lint`: plugin and marketplace manifest validation
- `make demo`: the README examples on `examples/checkout-v1`, `checkout-v2` and `checkout-v1-tidy`

## Conventions
- Standard library only. The tool only runs `go mod graph`, `go list -m` and git; it never builds or runs the module's code and never writes to `go.mod` or `go.sum`.
- An input it can't fully trust (a `go.mod` that is missing, empty, cut off, not text or has no go directive, a failed or malformed go command output, graph and list that disagree, different modules, one side with no dependencies, a bad `--base` ref) exits 2 and prints nothing on stdout, whatever `--fail-on` says.
- A new finding needs an integration test that must report it and a close case that must stay quiet.
- Tests must not assume the host GOOS or GOARCH, and must not need network unless they're gated like `TestRealReleases`.

## Do and don't
- Do keep README examples as real output from `make demo-bump`, `make demo-expect` and `make demo-quiet`.
- Don't claim a change is safe. The tool explains what moved and why; whether the new versions work is for tests and review.

## Definition of done
`make test`, `make lint` and `make demo` pass, and `ROADMAP.md`, `README.md` and docs are updated when relevant.
