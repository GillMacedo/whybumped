# whybumped plugin

This plugin adds the `explaining-go-dependency-bumps` skill and the `scripts/whybumped.go` tool. The tool runs `go mod graph` and `go list -m -json all` on two revisions of a Go module and explains each change in the selected dependency versions: which dependency required the new version, the chain back to `go.mod`, and which dependencies forced the `go` directive up. It needs Go 1.22 or newer, and git for `--base`. See the repository README for usage, findings, limits and what was tested.

Install it in Claude Code with:

    /plugin install whybumped --marketplace GillMacedo/whybumped
