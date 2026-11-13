# Sezzle Calculator

A calculator with a Go REST API and a React TypeScript UI. The browser never does the math. Every result comes from the API, in exact decimals, so `0.1 + 0.2` is `0.3`.

```bash
docker compose up --build   # needs Docker running
# open http://localhost:8080
# stop with Ctrl+C, then: docker compose down
```

![Calculator in light mode on desktop](docs/screenshots/desktop-light.jpg)

## What it does

- Add, subtract, multiply, divide, power, square root and percentage. The `% of` key means x percent of y.
- Dark mode, an in-memory history drawer and full keyboard use.
- One port. nginx serves the frontend and forwards `/api` to the Go service.
- `POST /api/v1/calculate` takes `{"operation":"add","operands":["0.1","0.2"]}` and returns `{"operation":"add","operands":["0.1","0.2"],"result":"0.3"}`. Every error has one shape with a `code`, `message` and `requestId`.

## Run

Go 1.26 and Node 24 or newer.

```bash
make -C backend run                    # API on :8080
cd frontend && npm ci && npm run dev   # UI on :5173, proxies /api to :8080
```

`make help` lists every target.

## Docs

- [docs/API.md](docs/API.md): the REST contract, limits, rate limit and every error code.
- [docs/DECISIONS.md](docs/DECISIONS.md): design decisions, architecture and accessibility.
- [docs/PROCESS.md](docs/PROCESS.md): how AI was used, how I directed it and the requirements checklist.
- [docs/PROMPTS.md](docs/PROMPTS.md): the prompt log.

## Key decisions

- Decimal, not float. The backend uses `shopspring/decimal`, and operands travel as strings so no digits are lost in JSON.
- Half-even rounding at 20 places for divide, square root and negative powers. Add, subtract, multiply and percentage are exact.
- Limits on input and output: 64 characters per operand, exponents within 1000, results up to 640 digits, a 5 second deadline.
- Rate limit of 10 requests a second per client IP on `/calculate`, answered with `429` and `Retry-After`.
- `400` means the request cannot be read. `422` means it was read and breaks a rule.

## Tests

```bash
make test             # backend (race detector) and frontend
make test-coverage    # both layers, with coverage reports
```

Backend: <!--cov-backend-->97.5%<!--/cov-backend--> coverage, with division, multiplication, square root and power also checked against an exact `math/big` reference. Frontend: <!--cov-frontend-->506<!--/cov-frontend--> tests at 100% coverage. CI runs lint, tests, `govulncheck`, `npm audit` and a Docker smoke test. The dev-only advisories a full `npm audit` still shows are in [DECISIONS.md](docs/DECISIONS.md#18-known-audit-note).

## AI usage

AI wrote much of the code, mainly Claude Code, one prompt per step under my direction. I reviewed each result before the next. [docs/PROCESS.md](docs/PROCESS.md) covers what I decided up front and what I overruled.
