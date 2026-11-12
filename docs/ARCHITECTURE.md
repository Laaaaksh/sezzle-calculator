# Architecture

The system is two containers. nginx serves the React bundle and forwards API calls to a Go service. The Go service holds all the arithmetic. There is no database.

## Request flow

```mermaid
flowchart TD
    B[Browser] -->|"GET /, /assets/*"| N[nginx]
    B -->|"POST /api/v1/calculate"| N
    N -->|static files, index.html fallback| B
    N -->|"proxy /api and /health"| R[chi router]

    subgraph Backend [Go backend]
        R --> M1[1. Request ID]
        M1 --> M2[2. Recovery]
        M2 --> M3[3. Access log]
        M3 --> M4[4. CORS]
        M4 --> M5[5. Body limit]
        M5 --> M6[6. Timeout]
        M6 --> H[Calculator handler]
        H -->|"decode, validate, parse operands"| C[Core]
        C -->|"lookup, arity, limits"| OP[Operation]
        OP -->|"Apply"| C
        C -->|"normalise result"| H
    end

    H -->|"JSON result or error"| B
```

Read the diagram from top to bottom. The same path carries the response back to the browser.

## Backend

The router lives in `backend/internal/boot/router.go`. The middleware order matters, so the code comments each position.

| Step | Middleware | Reason for the position |
| --- | --- | --- |
| 1 | Request ID | First, so every later log line and error body carries the ID. |
| 2 | Recovery | Before the access log, so a panic still gets a 500 and a log line. |
| 3 | Access log | Wraps everything below, to time the real work. |
| 4 | CORS | Before the limits, so a rejected preflight costs nothing. |
| 5 | Body limit | Before any handler reads the body. |
| 6 | Timeout | Innermost, so the deadline covers only handler work. |

The calculator module has three layers.

- Handler (`server.go`, `keys.go`, `parse.go`). The handler checks the media type and rejects repeated or wrongly cased top-level keys. Then it decodes the JSON, checks the required fields and the operand and name length bounds, and parses each operand string into a decimal. The handler knows HTTP and nothing about arithmetic.
- Core (`core.go`, `validate.go`). The core looks up the operation, checks the operand count (`checkArity`), checks the limits (`checkLimits`), applies the operation and normalises the result. The core knows arithmetic rules and nothing about HTTP.
- Operations (`operations/`). Each operation is one small type with `Name`, `Arity`, `Symbol`, `Description` and `Apply`. The registry maps names to operations and keeps their order.

Other packages:

| Path | Role |
| --- | --- |
| `cmd/api` | The service entry point |
| `cmd/healthcheck` | A tiny probe binary for the Docker healthcheck |
| `internal/boot` | Wiring: config, logger, registry, modules, router, graceful shutdown |
| `internal/config` | Layered TOML config with environment overrides (viper) |
| `internal/constants` | Shared names, so no string appears twice |
| `internal/interceptors` | The six middleware |
| `internal/httpx` | JSON writers, and the 404 and 405 handlers (the echoed path is cut to 256 bytes) |
| `internal/logger` | Structured logging (zap) |
| `internal/modules/calculator` | The calculator module |
| `internal/modules/health` | Liveness and readiness probes |
| `pkg/apperror` | The generic error codes, a registry with `Register`, and the client-safe error body |

Error codes are layered. `pkg/apperror` knows only the generic codes (`BAD_REQUEST`, `NOT_FOUND`, and so on). The calculator codes live in `internal/modules/calculator/entities/codes.go` and register themselves with `apperror.Register` in package `init`, so `pkg/apperror` never learns calculator vocabulary and a second module could add its own codes the same way.

Every failure becomes an `apperror` with a code. One function maps the code to the HTTP status and builds the response body, so no handler writes an error by hand.

## Frontend

The frontend has three layers, and each one only talks to the layer below.

```mermaid
flowchart TD
    UI["components: Calculator, Display, Keypad, HistoryDrawer, ErrorBanner"] --> HK[useCalculator hook]
    HK --> RD["reducer: pure state transitions"]
    HK --> API["api client: calculate, getOperations"]
    API -->|fetch| NG[nginx and backend]
```

- Components in `src/features/calculator/components` draw the state and send actions. They hold no logic.
- `useCalculator` owns the state and the history. The hook sends a request when you press equals, when you chain a second operator onto a pending operation, and when you press the square root key.
- The reducer in `state/` is a pure function. The reducer applies typing rules such as one decimal point and a length cap.
- The API client in `src/api` wraps `fetch`, checks response shapes and turns error bodies into typed errors.

The keypad is a static key list. The hook also fetches `GET /api/v1/operations` once. Keys for operations the server does not report are disabled, and `limits.maxOperandLength` caps the operand length. If the request fails, every key stays enabled and the server still validates each calculation.

The frontend never computes a result. The frontend builds an operation name and operand strings, sends them and shows the reply.

## Deployment

`docker-compose.yml` starts `backend` and `frontend`. The frontend waits until the backend health check passes. Only the frontend publishes a port (8080). The backend image is a static Go binary on distroless, running as a non-root user. The frontend image is Node for the build and unprivileged nginx for serving.

nginx does three jobs. nginx serves hashed assets with a one-year immutable cache, falls back to `index.html` for the single page app and proxies `/api` and `/health` to the backend with `Cache-Control: no-store`.

Docker stops each container with SIGTERM and kills it when `stop_grace_period` ends. The backend needs `shutdown_drain_delay + shutdown_grace_period` to stop cleanly (5s + 10s in prod), so compose gives it 20s. Keep the stop period longer than this sum. See decision 15 in [DECISIONS.md](DECISIONS.md).
