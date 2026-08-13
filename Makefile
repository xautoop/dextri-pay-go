GOCACHE ?= /private/tmp/dextri-pay-go-cache
STATICCHECK ?= go run honnef.co/go/tools/cmd/staticcheck@v0.7.0

.PHONY: list fmt-check tidy-check test test-race test-sandbox vet staticcheck diff-check check

list:
	GOCACHE=$(GOCACHE) go list ./...

fmt-check:
	@files="$$(gofmt -l $$(find . -type f -name '*.go' -not -path './vendor/*'))"; \
	if [ -n "$$files" ]; then echo "Go files require formatting:"; echo "$$files"; exit 1; fi

tidy-check:
	GOCACHE=$(GOCACHE) go mod tidy -diff

test:
	GOCACHE=$(GOCACHE) go test ./...

test-race:
	GOCACHE=$(GOCACHE) go test -race ./...

test-sandbox:
	GOCACHE=$(GOCACHE) DEXTRI_PAY_RUN_SANDBOX=1 go test ./internal/integration -run TestSandboxChannelsSmoke -count=1

vet:
	GOCACHE=$(GOCACHE) go vet ./...

staticcheck:
	GOCACHE=$(GOCACHE) $(STATICCHECK) ./...

diff-check:
	git diff --check

check: list fmt-check tidy-check test test-race vet staticcheck diff-check
