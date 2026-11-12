# Requirements checklist

Every requirement from the assignment, one per line. An item is ticked only when it is done and proven. The "Proof" column links to the test, file or output that shows it.

Legend: `[ ]` not done, `[x]` done and proven. Optional items are marked (optional).

## Functional: operations

| ID | Requirement | Status | Proof |
|---|---|---|---|
| F1 | Addition | [x] | `operations/add.go`, tests, [smoke](API_SMOKE.md) |
| F2 | Subtraction | [x] | `operations/subtract.go`, tests, [smoke](API_SMOKE.md) |
| F3 | Multiplication | [x] | `operations/multiply.go`, tests, [smoke](API_SMOKE.md) |
| F4 | Division | [x] | `operations/divide.go` (half-even), tests, [smoke](API_SMOKE.md) |
| F5 | Exponentiation (optional) | [x] | `operations/power.go`, tests, [smoke](API_SMOKE.md) |
| F6 | Square root (optional) | [x] | `operations/sqrt.go`, tests, [smoke](API_SMOKE.md) |
| F7 | Percentage (optional) | [x] | `operations/percentage.go` (x percent of y), tests, [smoke](API_SMOKE.md) |
| F8 | Frontend consumes the backend API for every calculation | [x] | `src/api/client.ts` used by `useCalculator`; the UI never computes, `Calculator.test.tsx` journeys ([screenshots](screenshots/)) |

## Frontend (React)

| ID | Requirement | Status | Proof |
|---|---|---|---|
| FE1 | Intuitive UI for entering input and displaying results | [x] | keypad, display, history, keyboard support ([desktop](screenshots/desktop-light.jpg)) |
| FE2 | Input validation | [x] | reducer rules (one decimal point, max length, leading zeros), `state/reducer.test.ts` |
| FE3 | Error handling | [x] | friendly copy per error code, retry on network errors ([error](screenshots/mobile-error.jpg)), `ErrorBanner` tests, `Calculator.test.tsx` divide by zero and network retry journeys |
| FE4 | Responsive design (basic mobile support) | [x] | checked at 320 to 1440 px, history drawer below 900 px ([mobile](screenshots/mobile-dark.jpg)) |

## Backend (REST API)

| ID | Requirement | Status | Proof |
|---|---|---|---|
| BE1 | Expose endpoints for calculator operations | [x] | `POST /api/v1/calculate`, `GET /api/v1/operations` ([API.md](API.md)) |
| BE2 | Validate input | [x] | `calculator/parse.go`, `server.go` validation, [smoke](API_SMOKE.md) |
| BE3 | Handle division by zero | [x] | `DIVISION_BY_ZERO` 422, divide and power tests, [smoke](API_SMOKE.md) |
| BE4 | Handle invalid data | [x] | 400/413/415/422 paths in `server.go` and `parse.go`, [smoke](API_SMOKE.md) |
| BE5 | Return results in JSON format | [x] | `httpx` JSON writer, one error shape, [smoke](API_SMOKE.md) |

## Non-functional

| ID | Requirement | Status | Proof |
|---|---|---|---|
| NF1 | Clean, readable, idiomatic frontend code | [x] | strict TS, ESLint strict type-checked + jsx-a11y, 0 warnings in CI |
| NF2 | Clean, readable, idiomatic backend code | [x] | feature-module layout, golangci-lint (12 linters) 0 issues in CI |
| NF3 | Unit tests covering key frontend functionality | [x] | Vitest + RTL + MSW, 520 tests, 100% coverage (`npm run test:coverage`) |
| NF4 | Unit tests covering key backend functionality | [x] | testify suites per package, 94.9% backend coverage (`make test-coverage`) |
| NF5 | Documentation: setup instructions | [x] | [README: run separately](../README.md#run-the-backend-and-frontend-separately) |
| NF6 | Documentation: API usage | [x] | [API.md](API.md), [README: API](../README.md#api) |
| NF7 | Documentation: design rationale | [x] | [DECISIONS.md](DECISIONS.md), [ARCHITECTURE.md](ARCHITECTURE.md) |
| NF8 | Dockerfile for full-stack deployment (optional) | [x] | `backend/Dockerfile`, `frontend/Dockerfile` (multi-stage, non-root), verified with `docker compose up --build --wait` |

## Constraints

| ID | Requirement | Status | Proof |
|---|---|---|---|
| C1 | Frontend in React, TypeScript preferred | [x] | `frontend/` React 19 + TypeScript strict |
| C2 | Backend in Go (preferred) | [x] | `backend/` Go 1.26 |

## Deliverables

| ID | Requirement | Status | Proof |
|---|---|---|---|
| D1 | Git repository with frontend and backend code | [x] | `backend/` and `frontend/` in this repo |
| D2 | README with setup instructions, API examples and design decisions | [x] | [README.md](../README.md) |
| D3 | Unit tests | [x] | Go testify suites and Vitest suites, run in CI |
| D4 | Coverage report | [x] | `make test-coverage` (backend/coverage.html), `npm run test:coverage` (frontend/coverage/), CI artifacts |
| D5 | Dockerfile to run frontend and backend together (optional) | [x] | `docker-compose.yml`: one port, healthchecks, read-only containers; smoke-tested in CI |

## Instructions

| ID | Requirement | Status | Proof |
|---|---|---|---|
| I1 | AI tooling allowed (used, and disclosed in the README) | [x] | [README: AI usage](../README.md#ai-usage) |
| I2 | Roughly 2 to 4 hours; correctness, clarity and maintainability over extra features | [x] | scope held to the brief, extras listed under next steps in the README |
| I3 | Solution pushed to a Git host | [x] | [github.com/Laaaaksh/sezzle-calculator](https://github.com/Laaaaksh/sezzle-calculator), PRs #1 to #7 each passed the no-mistakes pipeline and CI |
| I4 | Repository link shared with the reviewers | [ ] | |
| I5 | Prompts used are shared (docs/PROMPTS.md) | [x] | [PROMPTS.md](PROMPTS.md) |
| I6 | README includes setup instructions | [x] | [README: quick start](../README.md#quick-start) |
| I7 | README includes how to run the frontend and backend | [x] | [README: run separately](../README.md#run-the-backend-and-frontend-separately) |
| I8 | README includes examples of API calls | [x] | [README: API](../README.md#api), [API_SMOKE.md](API_SMOKE.md) |
| I9 | README includes design decisions or assumptions | [x] | [README: design decisions](../README.md#design-decisions-and-assumptions) |
