# Security

Report vulnerabilities through [GitHub's private vulnerability reporting](https://github.com/GillMacedo/whybumped/security/advisories/new), not in a public issue.

`whybumped` runs `go mod graph` and `go list -m -mod=readonly -json all` in the module directories you point it at. These download `go.mod` files of dependencies through your configured module proxy and check them against `go.sum`, as any go command does; they don't build or run code. With `--base` it runs `git rev-parse` and `git archive`, and extracts only `go.mod` and `go.sum` files into a temporary directory, refusing paths that leave it and skipping links. Please report privately if a crafted module, go.mod or repository could make it do anything else. A wrong explanation is a normal bug.
