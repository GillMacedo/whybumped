# Roadmap

## Next
- Show the replacement directory or module for replaced modules in the chain
- Group findings by root module in the text output, not only in the "pulled in through" line
- An `--only-roots` summary for very large upgrades
- A GitHub Action example that runs `whybumped --base` on Dependabot and Renovate pull requests and posts nothing without a person approving it

## Later
- Workspaces (`go.work`) with several main modules
- Explain `toolchain` line changes made by the go command
- Read saved `go mod graph` output, once there's a safe way to tell a complete file from one cut at a line break
