# Examples

All of these are made up for the demo.

- `proxy/` is a Go module proxy made of plain files: a `.mod`, `.info` and `.zip` for each version of six made-up `example.com` modules. The Makefile and tests point `GOPROXY` at it with a `file://` URL, so the demos need no network.
- `checkout-v1/` is a service that requires `example.com/httpkit v1.2.0` and `example.com/payclient v1.4.0`, with go 1.21.
- `checkout-v2/` is `checkout-v1` after `go get example.com/payclient@v1.5.0 example.com/textutil@v1.1.0`, run with this proxy. payclient v1.5.0 requires a newer `retry` and `telemetry`, `retry` v1.3.0 adds `backoff`, and `telemetry` v0.6.0 declares go 1.22, so the go line moved too. Run `make demo-bump` and `make demo-expect`, which exit 1.
- `checkout-v1-tidy/` has the same versions as `checkout-v1` in one reordered require block with a comment. Run `make demo-quiet`, which must stay quiet and exit 0.

The services are separate modules with only `go.mod` and `go.sum`, so the repo's own `go vet ./...` and `go test ./...` don't touch them.

| Module | Version | go | Requires |
|---|---|---|---|
| example.com/payclient | v1.4.0 | 1.21 | retry v1.0.0 |
| example.com/payclient | v1.5.0 | 1.21 | retry v1.3.0, telemetry v0.6.0 |
| example.com/retry | v1.0.0 | 1.20 | |
| example.com/retry | v1.3.0 | 1.21 | backoff v0.2.0 |
| example.com/backoff | v0.2.0 | 1.20 | |
| example.com/telemetry | v0.4.0, v0.6.0 | 1.20, 1.22 | |
| example.com/httpkit | v1.2.0 | 1.20 | telemetry v0.4.0, textutil v1.0.0 |
| example.com/textutil | v1.0.0, v1.1.0 | 1.20 | |
