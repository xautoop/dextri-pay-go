STATICCHECK ?= go run honnef.co/go/tools/cmd/staticcheck@v0.7.0

.PHONY: list fmt-check tidy-check test test-race test-sandbox vet staticcheck diff-check check

list:
	go list ./...

fmt-check:
	@files="$$(gofmt -l $$(find . -type f -name '*.go' -not -path './vendor/*'))"; \
	if [ -n "$$files" ]; then echo "Go files require formatting:"; echo "$$files"; exit 1; fi

tidy-check:
	go mod tidy -diff

test:
	go test ./...

test-race:
	go test -race ./...

test-sandbox:
	DEXTRI_PAY_RUN_SANDBOX=1 go test ./internal/integration -run TestSandboxChannelsSmoke -count=1

vet:
	go vet ./...

staticcheck:
	$(STATICCHECK) ./...

diff-check:
	git diff --check

check: list fmt-check tidy-check test test-race vet staticcheck diff-check
