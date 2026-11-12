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
| F8 | Frontend consumes the backend API for every calculation | [ ] | |

## Frontend (React)

| ID | Requirement | Status | Proof |
|---|---|---|---|
| FE1 | Intuitive UI for entering input and displaying results | [ ] | |
| FE2 | Input validation | [ ] | |
| FE3 | Error handling | [ ] | |
| FE4 | Responsive design (basic mobile support) | [ ] | |

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
| NF1 | Clean, readable, idiomatic frontend code | [ ] | |
| NF2 | Clean, readable, idiomatic backend code | [ ] | |
| NF3 | Unit tests covering key frontend functionality | [ ] | |
| NF4 | Unit tests covering key backend functionality | [x] | testify suites per package, 95.7% backend coverage (`make test-coverage`) |
| NF5 | Documentation: setup instructions | [ ] | |
| NF6 | Documentation: API usage | [ ] | |
| NF7 | Documentation: design rationale | [ ] | |
| NF8 | Dockerfile for full-stack deployment (optional) | [ ] | |

## Constraints

| ID | Requirement | Status | Proof |
|---|---|---|---|
| C1 | Frontend in React, TypeScript preferred | [ ] | |
| C2 | Backend in Go (preferred) | [x] | `backend/` Go 1.26 |

## Deliverables

| ID | Requirement | Status | Proof |
|---|---|---|---|
| D1 | Git repository with frontend and backend code | [ ] | |
| D2 | README with setup instructions, API examples and design decisions | [ ] | |
| D3 | Unit tests | [ ] | |
| D4 | Coverage report | [ ] | |
| D5 | Dockerfile to run frontend and backend together (optional) | [ ] | |

## Instructions

| ID | Requirement | Status | Proof |
|---|---|---|---|
| I1 | AI tooling allowed (used, and disclosed in the README) | [ ] | |
| I2 | Roughly 2 to 4 hours; correctness, clarity and maintainability over extra features | [ ] | |
| I3 | Solution pushed to a Git host | [ ] | |
| I4 | Repository link shared with the reviewers | [ ] | |
| I5 | Prompts used are shared (docs/PROMPTS.md) | [ ] | |
| I6 | README includes setup instructions | [ ] | |
| I7 | README includes how to run the frontend and backend | [ ] | |
| I8 | README includes examples of API calls | [ ] | |
| I9 | README includes design decisions or assumptions | [ ] | |
