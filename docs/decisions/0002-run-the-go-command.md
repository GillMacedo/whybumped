# 0002: Ask the go command, don't read saved graphs

**Decision:** whybumped takes module directories (or a git ref) and runs `go mod graph` and `go list -m -json all` itself. It doesn't accept a saved `go mod graph` file.

**Why:** `go mod graph` prints the part of the graph it loaded before failing, then exits 1. A file saved with `go mod graph > graph.txt` from a failed run looks valid, and a file cut at a line break can't be told from a complete one. Running the commands lets the tool check both exit codes and compare the graph's selection with `go list -m all`, which catches both cases.

**Alternatives:** reading graph files would work without the module cache, but there's no way to trust them. It's on the roadmap only if a safe check is found.
