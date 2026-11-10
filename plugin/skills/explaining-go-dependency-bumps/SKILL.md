---
name: explaining-go-dependency-bumps
description: Explains why Go module dependency versions changed between two revisions (a git ref or two directories). For every module that was raised, added, lowered or removed in the build list, it names the dependency that required the new version and the chain back to go.mod, and it lists the dependencies that forced the go directive up. Use when reviewing a Dependabot or Renovate pull request for a Go module, when go get or go mod tidy changed more modules than asked, when the go line in go.mod went up after adding a dependency, or when checking that a dependency upgrade only moved what was intended.
---

# Explaining Go dependency bumps

The bundled tool needs Go 1.22 or newer, and git for `--base`. It runs `go mod graph` and `go list -m -mod=readonly -json all` on each revision, with `GOWORK=off`. These read `go.mod` files of the module and its dependencies; they don't build or run any code, and they don't change `go.mod` or `go.sum`.

## 1. Pick the two revisions

- In a pull request, the old side is the merge base: `git merge-base origin/main HEAD`. Use the branch name only if the user asks for it.
- `--base REF` reads `go.mod` and `go.sum` files of that commit with `git archive`, so uncommitted changes in the work tree are the new side. Use `--old DIR --new DIR` for two checkouts.
- Run it in the module root (the directory with `go.mod`), or pass `--new MODULE_DIR`. In a repo with several modules, run it once per module whose `go.mod` changed.

## 2. Run it

    go run "${CLAUDE_PLUGIN_ROOT}/scripts/whybumped.go" --base "$(git merge-base origin/main HEAD)" --new MODULE_DIR

- If the pull request says what it meant to change (a Dependabot title such as "Bump google.golang.org/grpc from 1.63.2 to 1.67.1"), pass it with `--expect google.golang.org/grpc@v1.67.1`. Then every other change is a WARN, including modules that `go.mod` raised on its own.
- Add `--format json` to parse the result.
- Dependencies must be in the module cache or downloadable from the module proxy in the environment. Don't change `GOPROXY`, `GOFLAGS` or `GONOSUMDB` unless the user asks.

## Exit codes

0 means no WARN (or `--fail-on never`). 1 means at least one WARN. 2 means the result can't be trusted: a side's `go.mod` is missing, empty, cut off, not a `go.mod` or has no go directive; `go mod graph` or `go list -m` failed (often a missing `go.sum` entry, an unknown version or no network); their outputs disagree; the two sides are different modules; one side has no dependencies at all; or the `--base` ref doesn't exist. Treat 2 as unknown, never as a pass, and show the user the error.

## Report

1. Start with `go-raised` if present: the old and new go line, the dependency that needs the new version and how it got into the graph. If go.mod went further than any dependency needs, say so.
2. Group the module changes by the "pulled in through" line: for each root module that `go.mod` requires, list what it raised or added. These are the side effects of that upgrade.
3. `requested` changes are versions only `go.mod` asks for. For a `// indirect` module that's raised above what any dependency needs, say it was raised by hand or by `go get -u`, and that it could stay at the needed version if the user wants a smaller change.
4. `lowered` and `removed` changes say which requirer used to need the old version and where that requirer went. A lowered module is a downgrade in the build; mention it plainly.
5. `requirement-dropped` means go.mod lost a line while the build kept the same version through another module. `go mod tidy` does this when no package needs the module. If the user didn't run tidy, check that go.mod wasn't damaged.
6. A chain through a version that isn't selected is real: in minimal version selection, the requirements of every version in the graph count, even ones that lost to a newer version.
7. Don't edit `go.mod` unless the user asks. If they want a smaller change, the usual fix is `go get module@version` for the root module only, then `go mod tidy`, and run the tool again.
