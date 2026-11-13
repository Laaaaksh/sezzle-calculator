# How this was built

How I used AI on this project and how I checked the result. The prompts are in [PROMPTS.md](PROMPTS.md). The design is in [DECISIONS.md](DECISIONS.md).

## AI usage

AI wrote much of the code, mainly Claude Code, under my step-by-step direction. I set the rules, gave one prompt per step and reviewed each result before the next. The 26 build prompts and 3 operational entries are in [PROMPTS.md](PROMPTS.md), in order, with a line on what each produced. An automated review pipeline and CI checked each pull request.

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
- Tests that only restated constants. I removed them and kept the behaviour tests.

Why this went past 2 to 4 hours: this is a payments role, so I wanted to show how I would ship to production. That means input limits, exact rounding, clear errors, CI and a container. AI made the breadth cheap to write. My time went into reviewing and deciding.

In a strict 2 to 4 hours I would keep the decimal core, the parser, the API contract and the maths tests. I would cut the Docker hardening and nginx security policy, the custom healthcheck, the 100% frontend coverage gate, the CSS-parsing and constants tests, and the motion and design polish.

## How correctness is checked

- Unit tests for every operation, the parser, the handler and the router, plus end-to-end tests against the real router.
- Divide, multiply, square root and power are also checked against an exact reference built on `math/big` rationals with the same half-even rounding. A seeded run of 400 cases per operation is part of `go test`. Native fuzz targets search for more with `make -C backend fuzz`. CI runs only the seed corpus.
- Component tests with Testing Library and MSW, and axe on every calculator state. A token test guards WCAG contrast.
- Docs tests compare the error code tables and the frontend types with the Go registry.

## Requirements checklist

Every requirement from the assignment, one per line. An item is ticked only when it is done and proven.

Optional items are marked (optional).

### Functional: operations

| ID | Requirement | Status | Proof |
|---|---|---|---|
| F1 | Addition | [x] | `operations/add.go`, tests, [API.md](API.md) |
| F2 | Subtraction | [x] | `operations/subtract.go`, tests, [API.md](API.md) |
| F3 | Multiplication | [x] | `operations/multiply.go`, tests, [API.md](API.md) |
| F4 | Division | [x] | `operations/divide.go` (half-even), tests, [API.md](API.md) |
| F5 | Exponentiation (optional) | [x] | `operations/power.go`, tests, [API.md](API.md) |
| F6 | Square root (optional) | [x] | `operations/sqrt.go`, tests, [API.md](API.md) |
| F7 | Percentage (optional) | [x] | `operations/percentage.go` (x percent of y), tests, [API.md](API.md) |
| F8 | Frontend consumes the backend API for every calculation | [x] | `src/api/client.ts` used by `useCalculator`; the UI never computes, `Calculator.test.tsx` journeys ([screenshots](screenshots/)) |

### Frontend (React)

| ID | Requirement | Status | Proof |
|---|---|---|---|
| FE1 | Intuitive UI for entering input and displaying results | [x] | keypad, display, history, keyboard support ([desktop](screenshots/desktop-light.jpg)) |
| FE2 | Input validation | [x] | reducer rules (one decimal point, max length, leading zeros), `state/reducer.test.ts` |
| FE3 | Error handling | [x] | friendly copy per error code, retry on network errors ([error](screenshots/mobile-error.jpg)), `ErrorBanner` tests, `Calculator.test.tsx` divide by zero and network retry journeys |
| FE4 | Responsive design (basic mobile support) | [x] | checked at 320 to 1440 px, history drawer below 900 px ([mobile](screenshots/mobile-dark.jpg)) |

### Backend (REST API)

| ID | Requirement | Status | Proof |
|---|---|---|---|
| BE1 | Expose endpoints for calculator operations | [x] | `POST /api/v1/calculate`, `GET /api/v1/operations` ([API.md](API.md)) |
| BE2 | Validate input | [x] | `calculator/parse.go`, `server.go` validation, [API.md](API.md) |
| BE3 | Handle division by zero | [x] | `DIVISION_BY_ZERO` 422, divide and power tests, [API.md](API.md) |
| BE4 | Handle invalid data | [x] | 400/413/415/422 paths in `server.go` and `parse.go`, [API.md](API.md) |
| BE5 | Return results in JSON format | [x] | `httpx` JSON writer, one error shape, [API.md](API.md) |

### Non-functional

| ID | Requirement | Status | Proof |
|---|---|---|---|
| NF1 | Clean, readable, idiomatic frontend code | [x] | strict TS, ESLint strict type-checked + jsx-a11y, 0 warnings in CI |
| NF2 | Clean, readable, idiomatic backend code | [x] | feature-module layout, golangci-lint (12 linters) 0 issues in CI |
| NF3 | Unit tests covering key frontend functionality | [x] | Vitest + RTL + MSW, current test count and coverage in [README: tests](../README.md#tests) (`npm run test:coverage`) |
| NF4 | Unit tests covering key backend functionality | [x] | testify suites per package, current backend coverage in [README: tests](../README.md#tests) (`make test-coverage`) |
| NF5 | Documentation: setup instructions | [x] | [README: run](../README.md#run) |
| NF6 | Documentation: API usage | [x] | [API.md](API.md), [README](../README.md) |
| NF7 | Documentation: design rationale | [x] | [DECISIONS.md](DECISIONS.md) |
| NF8 | Dockerfile for full-stack deployment (optional) | [x] | `backend/Dockerfile`, `frontend/Dockerfile` (multi-stage, non-root), verified with `docker compose up --build --wait` |

### Constraints

| ID | Requirement | Status | Proof |
|---|---|---|---|
| C1 | Frontend in React, TypeScript preferred | [x] | `frontend/` React 19 + TypeScript strict |
| C2 | Backend in Go (preferred) | [x] | `backend/` Go 1.26 |

### Deliverables

| ID | Requirement | Status | Proof |
|---|---|---|---|
| D1 | Git repository with frontend and backend code | [x] | `backend/` and `frontend/` in this repo |
| D2 | README with setup instructions, API examples and design decisions | [x] | [README.md](../README.md) |
| D3 | Unit tests | [x] | Go testify suites and Vitest suites, run in CI |
| D4 | Coverage report | [x] | `make test-coverage` (backend/coverage.html), `npm run test:coverage` (frontend/coverage/), CI artifacts |
| D5 | Dockerfile to run frontend and backend together (optional) | [x] | `docker-compose.yml`: one port, healthchecks, read-only containers; smoke-tested in CI |

### Instructions

| ID | Requirement | Status | Proof |
|---|---|---|---|
| I1 | AI tooling allowed (used, and disclosed in the README) | [x] | [AI usage](#ai-usage) |
| I2 | Roughly 2 to 4 hours; correctness, clarity and maintainability over extra features | [x] | scope held to the brief, extras listed in DECISIONS.md |
| I3 | Solution pushed to a Git host | [x] | [github.com/Laaaaksh/sezzle-calculator](https://github.com/Laaaaksh/sezzle-calculator), PRs #1 to #8 (and later PRs) each passed an automated review pipeline and CI |
| I4 | Repository link shared with the reviewers | [ ] | |
| I5 | Prompts used are shared (docs/PROMPTS.md) | [x] | [PROMPTS.md](PROMPTS.md) |
| I6 | README includes setup instructions | [x] | [README: run](../README.md#run) |
| I7 | README includes how to run the frontend and backend | [x] | [README: run](../README.md#run) |
| I8 | README includes examples of API calls | [x] | [README](../README.md), [API.md](API.md) |
| I9 | README includes design decisions or assumptions | [x] | [README: key decisions](../README.md#key-decisions) |
