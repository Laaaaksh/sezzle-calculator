# Sezzle Calculator

```bash
docker compose up --build   # needs Docker running
# open http://localhost:8080
# stop with Ctrl+C, then: docker compose down
```

## What the app does

A calculator with a Go REST API and a React TypeScript frontend. The browser never does the math. Every result comes from the API, which uses exact decimal arithmetic, so `0.1 + 0.2` is `0.3`.

![Calculator in light mode on desktop](docs/screenshots/desktop-light.jpg)

Operations: add, subtract, multiply, divide, power, square root and percentage. The `% of` key means x percent of y: type x, press `% of`, type y. Dark mode and an in-memory history drawer are included. More screenshots: [mobile dark](docs/screenshots/mobile-dark.jpg) and [mobile error](docs/screenshots/mobile-error.jpg).

Only one port is exposed. nginx serves the static frontend and forwards `/api` to the backend.

## Layout

```text
backend/              Go API: cmd/, internal/modules/calculator, pkg/apperror
frontend/             React, TypeScript and Vite UI
docs/                 API, error codes, decisions, architecture, accessibility, prompt log
docker-compose.yml    backend plus nginx serving the built frontend
```

## Run the parts separately

| Tool | Version |
| --- | --- |
| Go | 1.26 (`backend/go.mod`) |
| Node | 24 or newer (`frontend/.nvmrc`) |

```bash
make -C backend run                    # API on :8080
cd frontend && npm ci && npm run dev   # UI on :5173, proxies /api to :8080
```

Run `make help` for every target.

## API

Operands go in as strings and results come back as strings. The full contract is in [docs/API.md](docs/API.md). Every error code, with real curl output for each, is in [docs/API.md](docs/API.md#error-examples-real-curl-output) and [docs/ERROR_CODES.md](docs/ERROR_CODES.md). Real server responses for every operation are in [docs/API_SMOKE.md](docs/API_SMOKE.md).

| Method | Path | Purpose |
| --- | --- | --- |
| POST | `/api/v1/calculate` | Run one operation |
| GET | `/api/v1/operations` | List the operations and the limits |
| GET | `/health/live` | Liveness probe |
| GET | `/health/ready` | Readiness probe |

```bash
curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" \
  -d '{"operation":"add","operands":["0.1","0.2"]}'
```

```json
{"operation":"add","operands":["0.1","0.2"],"result":"0.3"}
```

Every error has one shape: `code`, `message`, `requestId` and an optional `fields` object. A result over 640 digits, for example `10^1000`, returns `RESULT_TOO_LARGE`.

```json
{"error":{"code":"DIVISION_BY_ZERO","message":"Cannot divide by zero.","requestId":"139e5ba9-8f0e-4707-b316-ed9c0db14ad4"}}
```

Limits, all published by `GET /api/v1/operations`:

| Limit | Value | Error |
| --- | --- | --- |
| Operand length | 64 characters | `OPERAND_TOO_LARGE` |
| Power exponent | -1000 to 1000 | `EXPONENT_TOO_LARGE` |
| Result size | 640 digits | `RESULT_TOO_LARGE` |

## Tests

```bash
cd backend && go test ./...    # works on a fresh clone, the gomock mocks are committed
make test                      # backend (race detector) and frontend
make test-coverage             # both layers, with coverage reports
```

| Layer | Result | HTML report |
| --- | --- | --- |
| Backend (Go) | <!--cov-backend-->97.5%<!--/cov-backend--> total coverage, race detector on, mocks excluded | `backend/coverage.html` |
| Frontend (React) | <!--cov-frontend-->506<!--/cov-frontend--> tests, 100% statements, branches, functions and lines | `frontend/coverage/index.html` |

CI runs both suites on every push and pull request, regenerates the mocks and fails on any diff, then builds the Docker stack and smoke tests the full path through nginx. Lint gates: golangci-lint with 12 linters, ESLint in strict mode with jsx-a11y, Prettier, `govulncheck` and `npm audit` for production dependencies. axe-core checks the calculator in every state, and a token test guards WCAG contrast ([docs/ACCESSIBILITY.md](docs/ACCESSIBILITY.md)).

Known audit note: `npm audit` reports 7 moderate advisories in `minimatch`, inside `typescript-eslint`. This is dev tooling and does not ship. The production bundle has 0.

## Key design decisions

Each decision has a longer entry in [docs/DECISIONS.md](docs/DECISIONS.md). The structure is in [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

- Decimal, not float. The backend uses `shopspring/decimal`. A float stores `0.1` as a near neighbour and gives `0.30000000000000004`.
- Operands are strings. JavaScript turns a JSON number into a float before the server sees the value, so digits are lost. A string arrives untouched.
- Half-even rounding at 20 places. Divide, square root and negative powers never end, so they round. A tie goes to the even digit: `0.125` becomes `0.12`. Add, subtract, multiply and percentage are exact.
- Percentage means x percent of y. `percentage(15, 200)` is `30`. The UI key says `% of` so the meaning is plain.
- Input and result limits. An operand has at most 64 characters, an exponent stays within 1000, and any result has at most 640 digits. The limits are published at `GET /api/v1/operations`. Calculations stop at the 5 second request deadline. The body is at most 1 MiB.
- One endpoint with an `operation` field. Validation, limits and the error shape live in one handler. The cost: logs and rate limits keyed on the path do not separate operations.
- 400 versus 422. `400` means the server cannot read the request. `422` means the request parsed but the content breaks a rule.
- No database. A calculation is a pure function of its input. History lives in the browser and is gone on reload.
- Operation registry. Each operation is one type with a name, arity, symbol and `Apply`. A new operation is one file and one line of wiring.

## Beyond the brief

The brief asked for a working calculator, clear code and tests. I added a few production habits on purpose, to show how I would ship this: a theme toggle and history drawer, accessibility tests, a hardened Docker image with a strict Content-Security-Policy, a CI smoke test, and vulnerability and format checks. Each one is small and listed in [docs/DECISIONS.md](docs/DECISIONS.md).

## Next steps

- Generate an OpenAPI spec from the code so `docs/API.md` stays in step with the handler.
- Add Playwright end-to-end tests in Chromium, Firefox and WebKit.
- Add metrics and tracing, using the request ID already on every log line.
- Format numbers by locale in the UI.
- Parse full expressions with operator precedence, instead of one operation at a time.

## AI usage

AI wrote much of the code, mainly Claude Code, under my step-by-step direction. I set the rules, gave one prompt per step and reviewed each result before the next. The 26 build prompts and 3 operational entries are in [docs/PROMPTS.md](docs/PROMPTS.md), in order, with a line on what each produced. An automated review pipeline and CI checked each pull request.

`AGENTS.md` and `CLAUDE.md` hold the working rules I gave the AI tools. They are not part of the app.

## How I directed this

Decided before any code:

- The frontend never does maths. The API is the only source of truth.
- Numbers travel as strings and are computed with decimals, never floats. In payments, `0.1 + 0.2` must be `0.3`.
- Division rounds half-even at 20 places.
- I wrote the API contract and error codes before the handlers: `400` for malformed, `422` for valid but refused. I listed the edge cases up front: divide by zero, negative square root, fractional exponents and huge inputs.
- Power takes integer exponents only, so nothing falls back to floats silently.

Rejected or overruled:

- Generating the app from one big prompt. I split the work into 26 build steps, plus 3 operational entries. Each build step stopped for proof and a commit.
- Letting the builder review its own work. A fresh, separate review audited the code and I fixed every finding.
- A later hiring-style review found real gaps. Fresh clones failed `go test`, only power had a result-size limit, the `%` key was ambiguous and the README was bloated. I fixed all four.
- Deleting tests to look leaner. I kept them.

Why this went past 2 to 4 hours: this is a payments role, so I wanted to show how I would ship to production. That means input limits, exact rounding, clear errors, CI and a container. AI made the breadth cheap to write. My time went into reviewing and deciding.

In a strict 2 to 4 hours I would keep the decimal core, the parser, the API contract and the maths tests. I would cut the Docker hardening and nginx security policy, the custom healthcheck, the 100% frontend coverage gate, the CSS-parsing and constants tests, and the motion and design polish.
