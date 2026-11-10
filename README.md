# whybumped: explain why Go dependency versions changed

[![CI](https://github.com/GillMacedo/whybumped/actions/workflows/ci.yml/badge.svg)](https://github.com/GillMacedo/whybumped/actions/workflows/ci.yml) [![MIT license](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

**A Claude Code skill and Go CLI for Go developers and reviewers. It compares the Go module graph of two revisions and explains each version change in the build: which dependency required the new version of every module that a Dependabot or Renovate pull request, `go get` or `go mod tidy` raised, added, lowered or removed, the chain back to `go.mod`, and which dependencies forced the `go` directive up.**

A pull request titled "Bump payclient from 1.4.0 to 1.5.0" can change several other require lines in `go.mod` and the go line. `go get` prints `upgraded go 1.21 => 1.22` and `upgraded example.com/retry v1.0.0 => v1.3.0`, but not why. Finding out means searching `go mod graph` for the module that requires each new version; that output is 1,997 lines for caddy v2.9.0 and 4,133 for alertmanager v0.28.0. `whybumped` runs `go mod graph` and `go list -m all` on both revisions and does that for every change: it names the requirer, follows the chain back to the module `go.mod` asked for, and marks the changes nothing but `go.mod` asks for. Pass the intended change with `--expect` and everything else becomes a warning.

## What you can ask Claude

- "This Dependabot PR bumps grpc but go.mod changed eight lines. Why did the other modules move?"
- "After I added a dependency, go.mod now says go 1.23. Which dependency did that?"
- "Did my `go get -u` raise anything that nothing needs?"
- "Why did `go mod tidy` downgrade a module?"

## Install

As a Claude Code plugin:

    claude plugin marketplace add GillMacedo/whybumped
    claude plugin install whybumped@whybumped-marketplace

Or run it directly with Go 1.22 or newer. It has no dependencies:

    go run github.com/GillMacedo/whybumped/plugin/scripts@latest --base main

## Example

`examples/proxy` is a module proxy made of plain files with six made-up `example.com` modules, so this runs without network. `examples/checkout-v1` is a service at go 1.21 that requires `httpkit` and `payclient v1.4.0`. `examples/checkout-v2` is the same service after `go get example.com/payclient@v1.5.0 example.com/textutil@v1.1.0`:

    $ make demo-bump
    whybumped checkout-v1 -> checkout-v2: example.com/checkout, 5 modules before, 6 after, 1 unchanged
      WARN  go-raised go 1.21 -> 1.22
            go directive raised from 1.21 to 1.22; example.com/telemetry@v0.6.0 needs go 1.22
            via example.com/payclient@v1.5.0 <- go.mod
      WARN  added example.com/backoff v0.2.0
            added at v0.2.0 because example.com/retry@v1.3.0 requires it
            via example.com/retry@v1.3.0 <- example.com/payclient@v1.5.0 <- go.mod
      WARN  raised example.com/retry v1.0.0 -> v1.3.0
            raised to v1.3.0 because example.com/payclient@v1.5.0 requires it
            via example.com/payclient@v1.5.0 <- go.mod
      WARN  raised example.com/telemetry v0.4.0 -> v0.6.0
            raised to v0.6.0 because example.com/payclient@v1.5.0 requires it
            via example.com/payclient@v1.5.0 <- go.mod
      INFO  raised example.com/payclient v1.4.0 -> v1.5.0
            go.mod requires this version and no dependency requires it
      INFO  raised example.com/textutil v1.0.0 -> v1.1.0
            go.mod requires this version; dependencies need at most v1.0.0; it is marked // indirect, so an explicit `go get` or `go get -u` raised it
      pulled in through: example.com/payclient@v1.5.0 (3)
      4 warnings, 2 info

Exit code 1. payclient v1.5.0 requires newer `retry` and `telemetry`, `retry` v1.3.0 brings in `backoff`, and `telemetry` v0.6.0 declares go 1.22, which forced the go line up. `textutil` was raised by hand: no dependency needs more than v1.0.0.

When the intended change is known, `--expect` turns every other change into a warning, including `textutil`:

    $ make demo-expect
    whybumped checkout-v1 -> checkout-v2: example.com/checkout, 5 modules before, 6 after, 1 unchanged
      WARN  go-raised go 1.21 -> 1.22
            go directive raised from 1.21 to 1.22; example.com/telemetry@v0.6.0 needs go 1.22
            via example.com/payclient@v1.5.0 <- go.mod
      WARN  added example.com/backoff v0.2.0
            added at v0.2.0 because example.com/retry@v1.3.0 requires it
            via example.com/retry@v1.3.0 <- example.com/payclient@v1.5.0 <- go.mod
      WARN  raised example.com/retry v1.0.0 -> v1.3.0
            raised to v1.3.0 because example.com/payclient@v1.5.0 requires it
            via example.com/payclient@v1.5.0 <- go.mod
      WARN  raised example.com/telemetry v0.4.0 -> v0.6.0
            raised to v0.6.0 because example.com/payclient@v1.5.0 requires it
            via example.com/payclient@v1.5.0 <- go.mod
      WARN  raised example.com/textutil v1.0.0 -> v1.1.0
            not in --expect; go.mod requires this version; dependencies need at most v1.0.0; it is marked // indirect, so an explicit `go get` or `go get -u` raised it
      INFO  raised example.com/payclient v1.4.0 -> v1.5.0
            expected; go.mod requires this version and no dependency requires it
      pulled in through: example.com/payclient@v1.5.0 (3)
      5 warnings, 1 info

`examples/checkout-v1-tidy` has the same versions as `checkout-v1` in one reordered require block. It stays quiet and exits 0:

    $ make demo-quiet
    whybumped checkout-v1 -> checkout-v1-tidy: example.com/checkout, 5 modules before, 5 after, 5 unchanged
      no dependency changes

On a real release, caddy v2.8.4 to v2.9.0 (397 modules before, 459 after), the first finding is:

    WARN  go-raised go 1.21.0 -> 1.22.3
          go directive raised from 1.21.0 to 1.22.3; github.com/quic-go/quic-go@v0.48.2 needs go 1.22 and go.mod requires it directly, and 9 more selected modules need more than go 1.21.0; go.mod went further than any dependency needs (go 1.22)

## Usage

    whybumped (--old DIR | --base GITREF) [--new DIR] [--expect PATH[@VERSION]]...
              [--format text|json] [--fail-on warn|never]

1. **Load both sides.** In each module directory it runs `go mod graph` and `go list -m -mod=readonly -json all`, with `GOWORK=off`. These read `go.mod` files of the module and its dependencies, from the module cache or the module proxy in your environment. They don't build anything. `--base REF` takes the old side from git: it extracts the `go.mod` and `go.sum` files of that commit, keeping the directory layout so `replace => ../lib` still works. The new side is the work tree, uncommitted changes included. `--new` defaults to the current directory.
2. **Check them.** It recomputes minimal version selection from the graph and stops if the result differs from what `go list` selected.
3. **Compare.** Every module whose selected version changed is explained by the modules other than yours that require the new version. It follows those requirers back until it reaches a module that only your `go.mod` requires: the root of the change.

- **`--expect PATH[@VERSION]`** (repeatable) marks a change you intend. It becomes INFO, and every other change only `go.mod` asks for becomes WARN. An expected module that didn't change, or changed to another version, is a WARN.
- **`--format json`** prints the same report for scripts and agents: `severity`, `change`, `module`, `old`, `new`, `cause`, `chain`, `others` and `detail` per finding, and `roots` with the number of changes behind each root.
- **`--fail-on`** sets what exits 1: `warn` (default) or `never`.

Exit 0 means no WARN, or `--fail-on never`. Exit 1 means at least one WARN. Exit 2 means bad usage or a result it can't trust, and it prints nothing on stdout:

- a side that's missing or isn't a directory, a `go.mod` that's missing, empty, not UTF-8 text, cut off (no final newline), has no module line or no go directive, or is something else such as build or test output
- `go mod graph` or `go list -m` failing: a `go.mod` the go command can't parse, a missing, cut-off or wrong `go.sum` entry, a module or version the proxy doesn't have, or no network when the module cache doesn't have what's needed
- output of either command that's empty, garbage, cut off or has errors, or a graph and a build list that disagree
- two sides that are different modules, or one side with no dependencies while the other has some
- a `--base` ref that isn't a commit, looks like a flag, or has no `go.mod` in that directory

Exit 2 ignores `--fail-on`.

## What it reports

| Finding | Severity | Meaning |
|---|---|---|
| `go-raised` | WARN | The go directive went up. Names the selected dependency with the newest go line and how it got into the graph, lists the others that need more than the old line, and says if go.mod went further than any dependency needs |
| `raised`, `added`, cause `pulled` | WARN | Another module requires this version. Shows that requirer and the chain back to the module go.mod requires, plus other requirers of the same version |
| `raised`, `added`, cause `requested` | INFO (WARN with `--expect`) | Only go.mod requires this version. Shows the highest version any dependency needs, and notes `// indirect` modules raised past that |
| `lowered` | WARN | The selected version went down. Shows who required the old version and where those requirers went |
| `removed`, cause `requested` | WARN | go.mod no longer requires a module that only go.mod required. A deliberate removal looks like this, and so does a go.mod that lost lines |
| `removed`, cause `released` | INFO | The module left the build list because the module that required it moved or left |
| `requirement-dropped` | WARN | go.mod no longer lists a module, but the build still selects the same version through another requirer. `go mod tidy` does this when no package needs the module; a go.mod that lost lines looks the same |
| `requirement-added` | INFO | go.mod now lists a module whose selected version didn't change |
| `expect-missed` | WARN | An `--expect` module didn't change |
| `go-lowered`, `go-respelled`, `toolchain-changed` | INFO | The go line went down, was rewritten from `1.21` to `1.21.0` (the same release), or the toolchain line changed |

The "pulled in through" line counts the pulled changes behind each root module.

## How it compares

- **`go mod graph`** prints every requirement edge. `go mod why -m` explains which package imports a module, not why it has a version. Neither compares revisions.
- **`go get`** prints `upgraded X => Y` for each change, without the reason.
- **`go-modiff`** (saschagrunert, MIT) lists modules added, changed and removed between two versions of a repo, without saying what required them.
- **`gosumwhy`** (LeGEC, MIT) and **`gomodwhy`** (ycydsxy, Apache-2.0) print dependency paths to a module in one graph. They don't compare two revisions or explain the go line.
- **Graph visualizers** such as `modgv` and `modview` draw one graph.
- Go issue 33124 (2019) asked for a built-in explanation of which versions minimal version selection picked and why. Its author closed it as stale in 2025, with no command added.

I searched GitHub and the web for a tool that explains each version change between two revisions of a Go module and found none. Keyword search can miss projects.

## Limits

- **Both revisions must load.** Dependencies' `go.mod` files must be in the module cache or downloadable, and `go.sum` must have their entries. It uses your `GOPROXY`, `GOFLAGS` and `GONOSUMDB` as they are.
- **One module at a time.** It runs with `GOWORK=off`, so workspaces are compared module by module. `replace` directives with local paths work with `--old DIR`, and with `--base` when the target is in the same repository.
- **It explains requirements, not imports.** A module in the build list may provide no package your code uses; `go mod why -m` answers that.
- **The chain is one of possibly several.** It prefers selected versions and the shortest path, and lists up to five other requirers of the same version.
- **It doesn't judge the new versions.** Whether v1.3.0 is safe is for tests, changelogs and `govulncheck`.
- **Replaced modules** are shown with the version go.mod requires, not the replacement.

## FAQ

**Why did go get upgrade other modules?** Each module you upgrade requires minimum versions of its own dependencies, and Go picks the highest minimum anyone asks for. Run `whybumped --base HEAD` after the `go get` and each `pulled` finding names the module that asked.

**Why did go mod tidy raise the go version in go.mod?** Since Go 1.21, the go line must be at least the go line of every module in the build list. `whybumped` lists the dependencies whose go line is newer than your old one and how each got into the graph.

**How do I review a Dependabot pull request for a Go module?** Run `whybumped --base "$(git merge-base origin/main HEAD)" --expect MODULE@VERSION` with the module from the title. Any other change is a WARN with its reason.

**Why did a dependency get downgraded?** A module that used to require the higher version was upgraded or removed and no longer asks for it. The `lowered` finding names it and its new version.

**Does it change go.mod or go.sum?** No. It runs `go list -m` with `-mod=readonly` and only reads with git.

**Does it need network access?** Only if the module cache doesn't have the `go.mod` files of both revisions' dependencies. The examples and tests use a proxy made of files in the repo.

**Is it an MCP server?** No. It's a skill with one standard-library Go file, so `go run` is the whole install.

## Tested on

- Unit tests for module and Go version ordering (pre-releases, pseudo-versions, `+incompatible`, `1.21rc1`), graph and `go list` parsing, the graph-versus-list check, chains through unselected versions and cycles, `--expect` parsing, `go.mod` checks, and the archive extractor, which refuses paths that leave its directory and archives that are cut off.
- Integration tests in `tests/` that build the tool and run it on the examples with the file proxy: the bump, the reverse direction (lowered and removed), `--expect` hits, misses and wrong versions, JSON output, each example against itself, the reordered `go.mod` that must stay quiet, `go 1.21` rewritten as `go 1.21.0` (INFO, exit 0), and a `go.mod` that lost its last `require` line while the build stayed the same (WARN, exit 1).
- Damaged inputs, on the old side and on the new side in turn with the other side real, each under the default flags, `--fail-on never`, `--fail-on warn` and JSON output: a missing directory, a file, an empty directory, no `go.mod`; an empty, binary, garbage, cut mid-line, cut inside the require block, cut after the go line and cut before the go line `go.mod`; `go build`, `go test` and `go` command error output in place of `go.mod`; another module; an unknown version and an unknown module; no `go.sum`, an empty, garbage, cut, one-line, build-log and wrong-hash `go.sum`. All exit 2 with nothing on stdout. A damaged `go.sum` with `GOFLAGS=-mod=mod` is left unchanged.
- A fake `go` command on each side: empty, garbage, cut-off and build-log graph output, a graph cut at a line break, a graph without the main module's edges, a partial graph followed by a failure, and `go list` output that's empty, garbage, cut off, failing, with an error entry, missing a module, missing or renaming the main module, or selecting another version. All exit 2.
- `--base` in real git repositories with the module in a subdirectory: a tag against a changed tree, HEAD against an unchanged tree, `HEAD~1`, unknown refs, refs that look like flags, a path instead of a commit, a commit without the module, a directory outside git, and a sibling module through `replace => ../lib`.
- Go 1.27.3 on linux/arm64, with `-race`. The CI workflow runs the same tests on linux/amd64 with Go 1.22 and the latest stable Go.

Run on real releases on 2026-11-10, with `go.mod` and `go.sum` from GitHub and dependencies from proxy.golang.org (`make test-real`):

- **spf13/cobra v1.8.0 to v1.8.1**: one INFO, `go-md2man/v2` raised by go.mod, exit 0.
- **prometheus/alertmanager v0.27.0 to v0.28.0**: go raised from 1.21 to 1.22.0 because `automemlimit` v0.7.0 needs go 1.22.0, and `cloud.google.com/go/compute` lowered from v1.20.1 to v1.6.1 because `golang.org/x/oauth2` moved from v0.16.0, which required v1.20.1, to v0.24.0.
- **caddy v2.8.4 to v2.9.0**: go raised from 1.21.0 to 1.22.3 because `quic-go` v0.48.2 needs go 1.22, and `cloud.google.com/go/compute` lowered from v1.24.0 to v1.23.3 because `grpc` moved off v1.63.2. I checked these requirements against the published `go.mod` files on proxy.golang.org.
- Each release against itself: quiet, exit 0. `--base v1.7.0` on a fresh clone of cobra: four INFO changes, exit 0.

Not tested: Windows or macOS as the machine running the tool, workspaces, private module proxies, very large monorepos, and use in a fresh Claude session.

## Contributing

Contributions are welcome, from people and from people working with AI agents. See [CONTRIBUTING.md](CONTRIBUTING.md). Good first issues are [listed here](https://github.com/GillMacedo/whybumped/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22).

### How pull requests work

1. **Open a PR.** Fork the repo, work on a branch, and keep it to one change. Write a title that says what changes for users, because it becomes the changelog line: `feat:` goes under Added and `fix:` under Fixed.
2. **Automatic check.** Code changes need tests, and `make test` must pass. Changes over 30 lines need a linked open issue (`Closes #<number>`), and a PR can have at most 800 changed lines. If something is missing, a comment tells you what to fix.
3. **Review.** Claude reviews each PR automatically. Code PRs also get a security review and are tested in a sandbox. The review then merges the PR or asks for changes, and a small problem may become a follow-up issue instead. A PR with no new commit 14 days after a change request is closed, and you can reopen it. PRs that change CI, agent settings, npm or Python dependencies, vendored code or build files can't be merged this way, and a comment says so. The same goes for anything the security review doesn't clear.
4. **Release.** A merge that changes what users install becomes a new version within about two hours, with a GitHub Release and, if the project has one, its npm package. Docs and CI changes wait for the next release.

## AI disclosure

This project was largely generated with Claude Code and reviewed by its maintainer.

## About me

I'm Giulliano Macedo, and I build developer tools at GoTech Solutions. Follow along or get in touch:

- GitHub: [GillMacedo](https://github.com/GillMacedo)
- LinkedIn: [giulliano](https://www.linkedin.com/in/giulliano/)
- X: [@giumacedo](https://x.com/giumacedo)

If `whybumped` saved you some prompts, a star helps others find it.

---

Built by [Giulliano Macedo](https://github.com/GillMacedo), GoTech Solutions · [LinkedIn](https://www.linkedin.com/in/giulliano/) · [X](https://x.com/giumacedo)
