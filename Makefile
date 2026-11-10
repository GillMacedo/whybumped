.PHONY: test test-real lint demo demo-bump demo-expect demo-quiet tool

# The examples use a module proxy made of files in examples/proxy, so the
# demos and tests need no network.
FIXTURE_ENV := GOPROXY=file://$(CURDIR)/examples/proxy GOSUMDB=off GOTOOLCHAIN=local GOFLAGS=-modcacherw GONOSUMDB= GONOPROXY= GOPRIVATE=

test:
	test -z "$$(gofmt -l plugin tests)"
	go vet ./...
	go test -race -count=1 ./...

# Fetches go.mod and go.sum of real releases from GitHub and their
# dependencies from the module proxy, so it needs network access.
test-real:
	WHYBUMPED_REAL=1 go test -count=1 -run TestRealReleases -v ./tests

lint:
	claude plugin validate ./plugin --strict
	claude plugin validate . --strict

TOOL := bin/whybumped

tool:
	@go build -o $(TOOL) ./plugin/scripts

demo: demo-bump demo-expect demo-quiet

# checkout-v2 raised payclient, which pulled other modules and the go line up,
# so exit 1 is the expected result.
demo-bump: tool
	@cd examples && $(FIXTURE_ENV) ../$(TOOL) --old checkout-v1 --new checkout-v2 || test $$? -eq 1

demo-expect: tool
	@cd examples && $(FIXTURE_ENV) ../$(TOOL) --old checkout-v1 --new checkout-v2 --expect example.com/payclient@v1.5.0 || test $$? -eq 1

# checkout-v1-tidy reorders go.mod without changing a version. It must stay
# quiet and exit 0.
demo-quiet: tool
	@cd examples && $(FIXTURE_ENV) ../$(TOOL) --old checkout-v1 --new checkout-v1-tidy
