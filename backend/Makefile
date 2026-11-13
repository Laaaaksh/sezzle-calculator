# Backend developer commands. CI calls `make mock`, `make test-coverage` and `make build`.
# Fail fast: -e stops on the first error, -u rejects unset variables, pipefail fails a pipeline on any stage.
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := help

# Must match GOLANGCI_LINT_VERSION in .github/workflows/ci.yml so local lint equals CI lint.
GOLANGCI_LINT_VERSION := v2.15.0
BIN_DIR := bin
GOLANGCI_LINT := $(BIN_DIR)/golangci-lint
API_BIN := $(BIN_DIR)/api
COVERAGE_OUT := coverage.out
COVERAGE_HTML := coverage.html
# Generated mocks (committed, checked by CI) are excluded from the coverage total: they are not hand-written code.
COVERAGE_EXCLUDE := /mock/

.PHONY: help
help: ## List available targets
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-24s %s\n", $$1, $$2}'

.PHONY: deps
deps: lint-tool ## Download Go modules and install the pinned golangci-lint
	go mod download

# Installs golangci-lint into bin/ only when it is missing or not the pinned version.
# mockgen needs no install step: it is pinned by the `tool` directive in go.mod (`go tool mockgen`).
.PHONY: lint-tool
lint-tool:
	@if ! $(GOLANGCI_LINT) version 2>/dev/null | grep -q "version $(GOLANGCI_LINT_VERSION:v%=%) "; then \
		echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION) into $(BIN_DIR)"; \
		GOBIN=$(abspath $(BIN_DIR)) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION); \
	fi

.PHONY: fmt
fmt: lint-tool ## Format code with gofumpt and goimports
	$(GOLANGCI_LINT) fmt ./...

.PHONY: mock
mock: ## Generate gomock mocks (go generate)
	go generate ./...

.PHONY: lint
lint: mock lint-tool ## Run golangci-lint (generates mocks first)
	$(GOLANGCI_LINT) run ./...

.PHONY: test
test: mock ## Run all tests with the race detector (regenerates the committed mocks first; plain `go test ./...` also works)
	go test -race ./...

.PHONY: test-coverage
test-coverage: mock ## Like test, plus coverage.out and coverage.html and the total (also runs `make mock` first)
	go test -race -covermode=atomic -coverprofile=$(COVERAGE_OUT) ./...
	grep -v '$(COVERAGE_EXCLUDE)' $(COVERAGE_OUT) > $(COVERAGE_OUT).filtered
	mv $(COVERAGE_OUT).filtered $(COVERAGE_OUT)
	go tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)
	go tool cover -func=$(COVERAGE_OUT) | tail -1

.PHONY: build
build: ## Build the API binary into bin/api
	go build -trimpath -ldflags="-s -w" -o $(API_BIN) ./cmd/api

.PHONY: run
run: ## Run the API locally with APP_ENV=dev
	APP_ENV=dev go run ./cmd/api

.PHONY: clean
clean: ## Remove binaries and coverage files (mocks are committed, so they stay)
	rm -rf $(BIN_DIR) $(COVERAGE_OUT) $(COVERAGE_OUT).filtered $(COVERAGE_HTML)
