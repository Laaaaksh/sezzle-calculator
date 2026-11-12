.DEFAULT_GOAL := help

.PHONY: help
help: ## List available targets
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-24s %s\n", $$1, $$2}'

# Aggregate targets run every part of the project. The frontend is added here when it has a Makefile.
.PHONY: lint
lint: backend-lint ## Lint everything

.PHONY: test
test: backend-test ## Test everything

.PHONY: test-coverage
test-coverage: backend-test-coverage ## Test everything and write coverage reports

.PHONY: build
build: backend-build ## Build everything

.PHONY: hooks
hooks: ## Enable the git hooks in .githooks (no em dash, no agent co-author)
	git config core.hooksPath .githooks

# Each backend-* target delegates to the matching target in backend/Makefile.
.PHONY: backend-deps backend-fmt backend-lint backend-mock backend-test backend-test-coverage backend-build backend-run backend-clean
backend-deps: ## Backend: download modules and install tools
	$(MAKE) -C backend deps

backend-fmt: ## Backend: format code
	$(MAKE) -C backend fmt

backend-lint: ## Backend: run golangci-lint
	$(MAKE) -C backend lint

backend-mock: ## Backend: generate mocks
	$(MAKE) -C backend mock

backend-test: ## Backend: run tests with the race detector
	$(MAKE) -C backend test

backend-test-coverage: ## Backend: run tests and write coverage reports
	$(MAKE) -C backend test-coverage

backend-build: ## Backend: build bin/api
	$(MAKE) -C backend build

backend-run: ## Backend: run the API locally
	$(MAKE) -C backend run

backend-clean: ## Backend: remove build and coverage output
	$(MAKE) -C backend clean

.PHONY: clean
clean: backend-clean ## Clean everything
