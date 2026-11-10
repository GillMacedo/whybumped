# 0001: One standard-library Go file

**Decision:** the tool is a single Go file, `plugin/scripts/whybumped.go`, using only the standard library.

**Why:** everyone who needs it already has Go, so `go run` is the whole install, with no npm package or release binaries. A single file also runs with `go run path/to/whybumped.go` from inside any other module, which is how the skill calls it from the plugin folder.

**Alternatives:** `golang.org/x/mod` has a semver package and a go.mod parser, but would add a dependency for about 150 lines of version comparison. An MCP server would give an agent the same command with more setup and nothing a script can't do.
