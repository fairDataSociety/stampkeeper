GO ?= go
GOLANGCI_LINT ?= $$($(GO) env GOPATH)/bin/golangci-lint
GOLANGCI_LINT_VERSION ?= v1.43.0

.PHONY: lint
lint: linter
	$(GOLANGCI_LINT) run

.PHONY: linter
linter:
	test -f $(GOLANGCI_LINT) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$($(GO) env GOPATH)/bin $(GOLANGCI_LINT_VERSION)

.PHONY: test-race
test-race:
	$(GO) test -race -timeout 300000ms -v ./...

.PHONY: test
test:
	$(GO) test -v ./...

.PHONY: build
build:
	$(GO) build