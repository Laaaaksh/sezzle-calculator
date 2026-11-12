# Prompt log

Every prompt I sent to the AI assistant while building this project, in order and word for word. Each entry ends with one line on what it produced.

---

## 1. Ground rules and requirements

```
We're building a take-home for a payments company: a full-stack calculator. The backend is a Go REST microservice, and the frontend is React with TypeScript. The frontend calls the backend for every calculation and never does the maths itself.

Rules for the whole build:
1. We go one step at a time. Build only what each prompt asks for, then stop. Show me the files changed and why, and run whatever proves it works.
2. Log every prompt I send for this project in docs/PROMPTS.md, word for word, numbered, each with one line on what it produced.
3. Keep docs/REQUIREMENTS.md as a checklist of every requirement in the assignment. Tick an item only when it's done and proven.
4. Every approved step is one Conventional Commit. No AI co-author lines and no em dashes, anywhere.
5. Correctness and clarity beat features. If something's out of scope, put it under "next steps" instead of building it.

Start by creating docs/REQUIREMENTS.md from the assignment text I'll paste below, and docs/PROMPTS.md with this prompt as entry 1.

Build a full-stack calculator application with a React frontend and a backend microservice. The frontend should consume the backend API to perform basic and advanced arithmetic operations. Focus on clean design, maintainable code, and testable architecture.Requirements
Functional
Operations:

* Addition, Subtraction, Multiplication, Division
* Optional: Exponentiation, Square Root, Percentage

Frontend (React):

* Intuitive UI for entering input and displaying results
* Input validation and error handling
* Responsive design (basic mobile support)

Backend (REST API):

* Expose endpoints for calculator operations
* Validate input and handle edge cases (division by zero, invalid data)
* Return results in JSON format

Non-Functional

* Clean, readable, and idiomatic code (frontend and backend)
* Unit tests covering key functionality for both layers
* Documentation: setup instructions, API usage, and design rationale
* Optional: Dockerfile for full-stack deployment

Constraints

* Frontend: React (TypeScript preferred)
* Backend: Go is perferred

Deliverables

1. Git repository with frontend and backend code
2. README with setup instructions, API examples, and design decisions
3. Unit tests and coverage report
4. Optional: Dockerfile to run frontend + backend together

Instructions

1. Use any AI tooling you would like
2. Spend ~2–4 hours on this assignment. Prioritize correctness, clarity, and maintainability over extra features.
3. Push your solution to GitHub, GitLab, or another Git repository.
4. Share the repository link with us for evaluation.
5. Share any prompts that you used in your work
6. Make sure your README includes:
   * Setup instructions
   * How to run the frontend and backend
   * Examples of API calls (if using REST)
   * Design decisions or assumptions
```

Produced: `docs/REQUIREMENTS.md`, a 41-item checklist with IDs and a proof column, and this prompt log.

---

## 2. Repo skeleton

```
Create the repo skeleton, with no app logic yet:
- backend/ for the Go service, and frontend/ for the React app
- docs/ (already started)
- A root Makefile that will delegate to backend and frontend. For now it only has a help target that lists the targets.
- A .gitignore covering Go builds, coverage output, generated mocks, node_modules, dist, .env files and OS files
- An .editorconfig: LF line endings, a final newline, tabs for Go, 2 spaces for TS, JSON, YAML and Markdown
- An MIT LICENSE in my name
- A minimal README stub with the project name and one line describing it

Show me the tree when you're done.
```

Produced: `backend/` and `frontend/` placeholders, a self-documenting root `Makefile` with `help`, `.gitignore`, `.editorconfig`, MIT `LICENSE` and a README stub.

---

## 3. CI pipeline

```
Add a GitHub Actions workflow at .github/workflows/ci.yml. It runs on pushes and pull requests to main and has two jobs.
- backend: set up Go from the version in go.mod, cache modules, generate mocks, run golangci-lint, run the tests with the race detector and coverage, then build.
- frontend: set up Node LTS with an npm cache, npm ci, then type-check, lint, run the tests with coverage, and build.
Both jobs should fail loudly on any warning that matters. Upload both coverage reports as artifacts. Use pinned action versions. The jobs can't pass yet; that's expected. Explain each step in one line.
```

Produced: `.github/workflows/ci.yml` with backend and frontend jobs, actions pinned to commit SHAs, golangci-lint pinned, zero-warning lint, and coverage artifacts. Validated with actionlint.

---

## 4. Go module, config and constants

```
Start the Go service in backend/. Use the latest stable Go. Use this layout, which I use for all my services:
- cmd/api/main.go: a thin entry point. All it does is build the app and run it.
- internal/config: a typed Config struct, loaded with viper from config/default.toml, with per-environment files (dev, test, prod) layered on top and env var overrides with an APP_ prefix. Put everything in config: server port, read/write/idle timeouts, the shutdown grace period, max request body size, CORS allowed origins, and the calculator limits (max operand length, max exponent, decimal precision for division and sqrt).
- internal/constants: every string and number the code uses, including header names, log messages, log keys and the API version. No magic strings or numbers anywhere in the code.
Validate the config at startup and fail fast with a clear message when a value is missing or invalid. Add unit tests for config loading and validation, using testify suites with one test method per behaviour and camelCase names. No table-driven tests.
```

Produced: Go module in `backend/`, layered TOML config (default plus dev, test, prod) with `APP_` env overrides and fail-fast validation, `internal/constants`, a thin `cmd/api/main.go`, and config tests at 100% coverage.

---

## 5. Logger and error model

```
Add two packages.

internal/logger: structured logging with zap. Use JSON in prod and console output in dev. Add a Ctx(ctx) helper that returns a logger carrying the request ID from the context. All log calls use constants for messages and keys.

pkg/apperror: the single error model for the service.
- An IError interface with Code(), PublicMessage(), HTTPStatus(), Fields(), Unwrap() and WithField(key, value).
- Error codes as typed constants: BAD_REQUEST, VALIDATION_FAILED, UNSUPPORTED_MEDIA_TYPE, PAYLOAD_TOO_LARGE, NOT_FOUND, METHOD_NOT_ALLOWED, INTERNAL, plus the calculator codes we'll add later.
- Each code maps to an HTTP status in one place.
- Constructors New(code, cause) and NewWithMessage(code, cause, msg).
- Internal causes are logged but never leak to the client.
Test the status mapping, the field handling and that unwrap works with errors.Is and errors.As.
```

Produced: `internal/logger` (zap, JSON in prod and test, console in dev, request-scoped `Ctx`), `internal/contextkeys` for a typed request ID key, and `pkg/apperror` with typed codes, one status map, copy-on-write fields and errors.Is/As support. Coverage 97.6 to 100% per package.

> Note: from entry 6 on, I queued the remaining prompts in advance (written before the build started) and let the assistant run them back to back while I was away, committing each step and shipping each phase through the no-mistakes pipeline. The prompts below are exactly what was queued.

---

## 6. Boot, router, middleware and health

```
Wire the service together.
- internal/boot builds the app in a fixed order: config, then logger, then modules, then router. It runs the HTTP server and shuts it down gracefully on SIGINT or SIGTERM within the configured grace period.
- Use chi for the router. Mount the API under /api/v1.
- Middleware, each in its own file under internal/interceptors:
  - Request ID: read X-Request-ID, or generate one, put it in the context and echo it on the response.
  - Structured access log: method, path, status, duration and request ID.
  - Panic recovery that logs the stack and returns our JSON INTERNAL error.
  - CORS, with origins from config.
  - Max body size.
  - A per-request timeout.
- JSON 404 and 405 responses in our error shape.
- A health module with GET /health/live and GET /health/ready, following the same module layout we'll use for features: init.go wires it, server.go has the handler and RegisterRoutes, core.go has the interface and logic, and entities/ holds request and response types and constants.
Test every middleware in isolation, plus the health endpoints. Then run the server and show me curl output for health, a 404, a 405, and a response with the X-Request-ID header.
```

Produced: `internal/boot` with ordered startup and graceful shutdown, chi router under `/api/v1`, six middleware in `internal/interceptors`, a JSON error writer in `internal/httpx`, JSON 404 and 405, and a health module with live and ready probes. Verified with a real server and curl.

---

## 7. Makefile and lint

```
Add backend/Makefile with these targets: help, deps, fmt, lint, mock, test (with the race detector), test-coverage (writes coverage.out and coverage.html and prints the total percentage), build, run and clean. Make mock run before lint and test, because test files import the generated mocks. Add backend/.golangci.yml with errcheck, govet, staticcheck, revive, gosec, gocyclo, dupl, goconst, gofumpt, goimports, misspell and unparam, and justify any setting that isn't a default. The root Makefile should call these. Run make lint and make test and show me that both are clean.
```

Produced: `backend/Makefile` with pinned tool versions, `backend/.golangci.yml` (12 linters, every non-default setting justified), root Makefile delegation and `.no-mistakes.yaml`. Lint 0 issues, total coverage 92.9%.

---

## 8. Operation contract and registry

```
Create the calculator module at internal/modules/calculator, using the same layout as health. Start with the domain only, no HTTP.
- Use github.com/shopspring/decimal for all maths. Never use float64, because a payments reviewer will check that 0.1 + 0.2 equals exactly 0.3.
- Define an Operation interface with Name() OperationType, Arity() int and Apply(operands []decimal.Decimal) (decimal.Decimal, apperror.IError).
- Keep a Registry that maps OperationType to Operation. Lookup returns an UNKNOWN_OPERATION error for anything unregistered. Adding an operation should mean adding one file and one registry line, and nothing else changes.
- The core service, Calculate(ctx, op, operands), looks up the operation, checks the arity (INVALID_OPERAND_COUNT), checks each operand against the configured limits, applies the operation, and normalises the result: trims trailing zeros and turns "-0" into "0".
Don't add any operations yet except one fake in the tests. Test the registry, the arity check and the normalisation.
```

Produced: The calculator module domain with a decimal-only Operation interface, a registry behind one registration point, and a core that checks arity and operand limits and normalises results. Tests use a fake operation. Calculator coverage 100%.

---

## 9. Basic operations

```
Add add, subtract, multiply and divide, each in its own file, each registered.
- Division by zero returns a DIVISION_BY_ZERO error with the message "Cannot divide by zero."
- Division that doesn't terminate rounds to the configured precision with half-even (banker's) rounding. Document why: it's the standard for financial rounding because it doesn't bias the totals.
Write one test suite per operation, with one method per behaviour. Cover at least: positive and negative values, zero, very large and very small numbers, many decimal places, 0.1 + 0.2 = 0.3, 1 / 3 at the configured precision, 2 / 3 rounding, -0 normalisation, and divide by zero. Every test name should read like a sentence, for example TestDivideWhenDivisorIsZeroReturnsDivisionByZero.
```

Produced: Add, subtract, multiply and divide, one file and one registry line each. Divide uses QuoRem for exact half-even rounding at the configured precision and returns DIVISION_BY_ZERO ("Cannot divide by zero."). One suite per operation. Calculator coverage 100% (core) and 97.8% (operations).

---

## 10. Advanced operations

```
Add power, sqrt and percentage, each in its own file.
- power(base, exponent): integer exponents only, between -max and +max from config. Reject a fractional exponent with INVALID_EXPONENT, because decimal maths can't do it exactly and we don't want a silent float. Zero to a negative power returns DIVISION_BY_ZERO. Any number to the power 0 is 1.
- sqrt(x): one operand. A negative x returns NEGATIVE_SQUARE_ROOT. Compute it with a decimal Newton's method to the configured precision, and stop after a bounded number of iterations so it can never hang. sqrt(0) is 0. Perfect squares are exact.
- percentage(x, y): x percent of y, so x / 100 * y. Write the meaning down in a code comment and in the API docs, because "percentage" is ambiguous.
Test every rule above, plus results that are large and results that are tiny.
```

Produced: Power (integer exponents within the configured range, digit-budget guard, exact squaring), sqrt (bounded decimal Newton's method with guard digits and half-even rounding) and percentage (x / 100 * y, exact), each in its own file and registered. Adds INVALID_EXPONENT, EXPONENT_TOO_LARGE, RESULT_TOO_LARGE and NEGATIVE_SQUARE_ROOT (422).

---

## 11. Input limits and parsing

```
Add a parser that turns operand strings into decimals, with strict rules:
- Accept an optional leading minus, digits, and one optional decimal point.
- Reject an empty string, whitespace, a plus sign, exponent notation, hex, NaN, Infinity, thousands separators and more than one decimal point. Each fails with INVALID_OPERAND, and the error fields say which operand index failed and why.
- Reject operands longer than the configured max length with OPERAND_TOO_LARGE, before parsing, so a huge string can't eat CPU.
Explain why we take strings and not JSON numbers: JSON numbers get parsed as float64 by most clients and lose precision.
Test every accepted and rejected form.
```

Produced: A strict operand parser in the calculator module: optional minus, digits and one optional point, a precompiled regexp, a reason for every rejected form (empty, whitespace, plus_sign, exponent, hex, nan, infinity, separator, multiple_points, invalid_format) and a byte-length check before parsing. Adds INVALID_OPERAND (422). The package comment explains why operands are strings and not JSON numbers.

---

## 12. Core coverage check

```
Run the calculator tests with coverage. Show me the per-file numbers. Anything in the calculator module under 90 percent needs either a test or a reason why not. Then list every error code the core can return, with the input that triggers each one, as a table I can reuse in the README.
```

Produced: Per-file coverage for the calculator module (99.5% overall, every file 96.9% or higher) and docs/ERROR_CODES.md, a table of every error code with its HTTP status, an example input and its message, ready for the README. A test keeps that table in sync with pkg/apperror.

---

## 13. API contract

```
Before writing the handler, write the API contract as docs/API.md for my review:
- POST /api/v1/calculate
  - Request: {"operation": "divide", "operands": ["10", "4"]}
  - Response 200: {"operation": "divide", "operands": ["10", "4"], "result": "2.5"}
- GET /api/v1/operations: lists each operation with its name, arity, symbol and a one-line description, so the frontend can build its UI from it.
- One error shape everywhere: {"error": {"code": "DIVISION_BY_ZERO", "message": "Cannot divide by zero.", "requestId": "...", "fields": {...}}}
- A table mapping every error code to its HTTP status. Domain errors are 422 and malformed requests are 400. Explain that choice.
Also explain why I chose one endpoint with an operation field over one route per operation, and what the trade-off is. Don't write any code in this step.
```

Produced: docs/API.md, the REST contract for review before any handler code: POST /api/v1/calculate, GET /api/v1/operations, one error shape, a table of every error code with its HTTP status, why domain errors are 422 and malformed requests are 400, and why one endpoint with an operation field beats one route per operation, with the trade-off. No code.

---

## 14. Handler and validation

```
Implement the contract from docs/API.md in the calculator module's server.go.
- Decode with DisallowUnknownFields.
- Reject a missing or wrong Content-Type with UNSUPPORTED_MEDIA_TYPE, an empty body with BAD_REQUEST, malformed JSON with BAD_REQUEST, and trailing data after the JSON object.
- Validate that operation and operands are present before calling core.
- Map every apperror to the JSON error shape with the request ID.
- Set Content-Type and Cache-Control: no-store on every response.
The handler depends on the ICore interface, not the concrete type. Generate the mocks with mockgen and test the handler with gomock using exact arguments and .Times(1), never gomock.Any() or AnyTimes(). Cover every success and failure path.
```

Produced: The calculate and operations endpoints in the calculator module's server.go, mounted under /api/v1. Strict decoding (unknown fields, empty body, malformed JSON, trailing data, oversize body), Content-Type check, required-field validation with a fields map, and every apperror mapped to the shared error shape with the request ID. Operations are listed from the registry in a fixed order. Handler tests use gomock with exact arguments and .Times(1), plus one end-to-end suite through the real router and core.

---

## 15. Backend end-to-end check

```
Run make lint, make test and make test-coverage, and show me the totals. Then start the server and use curl to hit every operation once, every error code once, and the health and operations endpoints. Paste the real request and response pairs. If anything looks wrong, reproduce it first, then fix it and add a test that would have caught it. Update docs/REQUIREMENTS.md for everything the backend now covers.
```

Produced: lint 0 issues, race tests green, backend coverage 95.7%, and a real-server smoke test of all 7 operations and every error code ([API_SMOKE.md](API_SMOKE.md)). Backend requirements ticked in REQUIREMENTS.md with proof.

---

## 16. Vite app and tooling

```
Scaffold frontend/ with Vite, React and TypeScript.
- Strict TypeScript, including noUncheckedIndexedAccess.
- ESLint with the typescript, react-hooks and jsx-a11y plugins, plus Prettier.
- Vitest with jsdom, React Testing Library, user-event and MSW for mocking the API.
- Coverage with v8, plus scripts for dev, build, typecheck, lint, test and test:coverage.
- A Vite dev proxy that sends /api to the Go backend, so local dev has no CORS issues.
- Folders: src/api, src/features/calculator, src/components, src/constants, src/styles and src/test.
Remove the Vite demo content. Run typecheck, lint and test and show me they pass.
```

Produced: frontend/ scaffolded with Vite, React and TypeScript in strict mode (noUncheckedIndexedAccess included). ESLint (typescript-eslint strict, react-hooks, jsx-a11y) with Prettier, Vitest with jsdom, Testing Library, user-event and MSW (unhandled requests fail tests), v8 coverage with an 80% gate, a /api dev proxy to the Go backend, the agreed folder layout, frontend-* Makefile targets, and the Vite demo content removed. Typecheck, lint, tests and build pass.

---

## 17. Typed API client

```
Build src/api.
- types.ts mirrors docs/API.md exactly: OperationType as a string union, CalculateRequest, CalculateResponse, OperationInfo and ApiErrorBody.
- client.ts has calculate(request, { signal }) and getOperations().
  - The base URL comes from import.meta.env with a sensible default.
  - Every request has a timeout through AbortController.
  - Failures are thrown as typed errors: an ApiError carrying the code, message, status and request ID; a NetworkError; and a TimeoutError.
- No any types.
Test with MSW: success, each error kind, a non-JSON error body, a timeout, and an aborted request.
```

Produced: frontend/src/api with types.ts mirroring docs/API.md, a fetch client (calculate, getOperations) with an env-driven base URL, a per-request timeout via AbortController, runtime response guards, and typed ApiError, NetworkError, TimeoutError and CanceledError. MSW tests cover success, each error kind, a non-JSON error body, a timeout and an aborted request, plus a test that fails when a code or operation in docs/API.md is missing from the types. Coverage is 100%.

---

## 18. State machine

```
Model the calculator state as a pure reducer in src/features/calculator/state.
- States: entering the first operand, operator chosen, entering the second operand, waiting for the result, showing the result, and error.
- Actions: digit, decimal point, operator, unary operator (sqrt), equals, backspace, clear, toggle sign, result received, request failed, and reuse from history.
- Rules:
  - Only one decimal point per operand.
  - A max length per operand, matching the backend limit.
  - Leading zeros collapse.
  - Pressing an operator after a result chains from that result.
  - Pressing an operator twice replaces it.
  - Equals with nothing to compute does nothing.
- The result is kept as the backend's string and never parsed into a JS number.
Test each rule as its own test, and include the edge cases.
```

Produced: frontend/src/features/calculator/state: a pure reducer over a status-tagged union (entering the first operand, operator chosen, entering the second operand, waiting, showing the result, error) with actions for digits, decimal point, operators, sqrt, equals, backspace, clear, sign, result, failure, retry and history reuse. Operands stay strings (max 64 characters, one decimal point, collapsing leading zeros, chained and replaced operators), each waiting state carries a request id so stale responses are ignored, and every rule has its own test. Also adds toCalculatorError, which maps client errors into the reducer's error shape.

---

## 19. Hook, keyboard and history

```
Add a useCalculator hook. It wires the reducer to the API client, cancels an in-flight request if the user clears, ignores a stale response, and blocks double submits while a request is pending.
Add keyboard support in its own hook: digits, the dot, + - * / ^ %, Enter or = for equals, Backspace, Escape to clear, and r for sqrt. Typing inside the input should not trigger the shortcuts twice.
Add a history list held in memory: each entry shows the expression and the result, the newest goes on top, it's capped at 20, and clicking one reuses its result. Test the hook with MSW, the keyboard mapping, and the history behaviour.
```

Produced: frontend/src/features/calculator: useCalculator wires the reducer to an injected API client (default: the real one) with one AbortController per request, so clear and unmount cancel it, a canceled request never shows as an error, a stale response is dropped and double submits are ignored while pending. useCalculatorKeyboard maps digits, dot, operators, Enter or =, Backspace, Escape and r, and stays out of text fields, focused buttons and modified keys so nothing fires twice. An in-memory history keeps the newest 20 entries on top and reuses a result on click. Tests use MSW for the hook, a table for the key mapping and cover the history cap.

---

## 20. Components and error UX

```
Build the components: Display (expression line and result line), Keypad, Key, HistoryTape and an ErrorBanner.
- Map every backend error code to friendly copy in one constants file. For example, DIVISION_BY_ZERO becomes "You can't divide by zero."
- Unknown codes fall back to a generic message that includes the request ID.
- Network and timeout errors show a Retry button that resends the same request.
- While a request is pending, show a subtle loading state on the result line and disable equals.
Use semantic HTML: real buttons, an aria-label on every symbol key, and a display with role="status" and aria-live="polite" so results and errors are announced.
Write integration tests for the main journeys: 12.5 + 7.25 = 19.75, divide by zero shows the error, a network failure then retry succeeds, a full keyboard-only calculation, and reusing a result from history.
```

Produced: Calculator components (Display, Keypad, Key, HistoryTape, ErrorBanner, plus a Calculator that wires the hook and keyboard) rendered by App, with CSS modules and semantic buttons. Keys are data-driven from keys.ts with an aria-label on every symbol key; the display is a polite role=status region that also holds error text. All error codes get friendly copy in constants/errorMessages.ts (type-checked Record, generic fallback with the request ID), network and timeout errors offer Retry, and equals is disabled with a loading state while pending. Integration tests run the real hook against MSW for the add, divide by zero, network retry, keyboard-only and history reuse journeys.
