.DEFAULT_GOAL := help

.PHONY: help
help: ## List available targets
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-24s %s\n", $$1, $$2}'

# Aggregate targets run every part of the project: the Go backend and the React frontend.
.PHONY: lint
lint: backend-lint frontend-lint frontend-typecheck ## Lint everything

.PHONY: test
test: backend-test frontend-test ## Test everything

.PHONY: test-coverage
test-coverage: backend-test-coverage frontend-test-coverage ## Test everything and write coverage reports

.PHONY: build
build: backend-build frontend-build ## Build everything

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

# Each frontend-* target runs the matching npm script in frontend/.
.PHONY: frontend-install frontend-lint frontend-typecheck frontend-test frontend-test-coverage frontend-build frontend-clean
frontend-install: ## Frontend: install dependencies from package-lock.json
	npm --prefix frontend ci

frontend-lint: ## Frontend: run ESLint with zero warnings allowed
	npm --prefix frontend run lint -- --max-warnings=0

frontend-typecheck: ## Frontend: run the TypeScript compiler without emitting
	npm --prefix frontend run typecheck

frontend-test: ## Frontend: run tests once
	npm --prefix frontend test

frontend-test-coverage: ## Frontend: run tests and write coverage reports
	npm --prefix frontend run test:coverage

frontend-build: ## Frontend: build the production bundle
	npm --prefix frontend run build

frontend-clean: ## Frontend: remove build and coverage output
	rm -rf frontend/dist frontend/coverage

.PHONY: clean
clean: backend-clean frontend-clean ## Clean everything
