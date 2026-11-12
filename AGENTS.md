# Agent instructions

Calculator take-home: a Go REST API (`backend/`) and a React TypeScript UI (`frontend/`). Operands are strings and all math uses exact decimals on the server. Read `README.md` for the run commands and `docs/` for the contract and decisions.

## Layout

- `backend/`: Go. Entry points in `cmd/`, feature code in `internal/modules/<name>/` (server, core, entities, operations), shared error type in `pkg/apperror`.
- `frontend/`: Vite, React, TypeScript. Feature code in `src/features/calculator/`, API client in `src/api/`.
- `docs/`: `API.md`, `ERROR_CODES.md`, `DECISIONS.md`, `ARCHITECTURE.md`, `ACCESSIBILITY.md`, `REQUIREMENTS.md`, `PROMPTS.md`.

## Checks

Run these before every commit. All must pass.

- Backend: `make -C backend lint` and `make -C backend test`.
- Frontend, from `frontend/`: `npm run lint -- --max-warnings=0`, `npm run typecheck`, `npm run format:check`, `npm test`.
- Everything at once: `make lint` and `make test` from the repo root.
- The gomock mocks in `backend/**/mock/` are committed. After changing an interface that has `//go:generate`, run `make -C backend mock` and commit the result. CI fails on a stale mock.

## Test conventions

- Go: testify suites with gomock. One method per behaviour, named `TestSubjectWhenConditionReturnsOutcome`. No table-driven tests. No `gomock.Any()` or `AnyTimes()`: pass exact arguments and use `.Times(1)`. Use named constants, not magic values.
- Frontend: Vitest, React Testing Library and MSW. Query by accessible role and name. Every UI state is covered, and axe checks the calculator.
- Fix lint errors and flaky tests you meet, even when your change did not cause them.

## Commits and writing

- One step per commit, as a Conventional Commit (`feat:`, `fix:`, `docs:`, `test:`, `build:`, `chore:`).
- Never add an agent as co-author or add any AI attribution line to a commit or pull request.
- Never use the em dash (Unicode U+2014), in code, docs or commit messages. Use a plain dash "-".
- Log every user prompt word for word in `docs/PROMPTS.md`, with one line on what it produced.
- Track the assignment requirements in `docs/REQUIREMENTS.md`.
