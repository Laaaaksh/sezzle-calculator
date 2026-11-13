# Design decisions

One entry per decision. Each has the context, the decision and the trade-off.

## 1. Decimal arithmetic, not float

- Context: A calculator must show `0.1 + 0.2` as `0.3`. A 64-bit float stores 0.1 as a close neighbour and returns `0.30000000000000004`.
- Decision: The backend uses `shopspring/decimal` for all math.
- Trade-off: Decimals are slower than floats and need explicit rounding rules. For a calculator the speed does not matter and exact digits do.

## 2. Operands are strings

- Context: JavaScript parses a JSON number into a float before the server sees the value. The digits a user typed are gone on arrival, and a 20-digit integer loses its tail.
- Decision: Operands go in as JSON strings and results come back as strings. A JSON number in `operands` returns `400 BAD_REQUEST`. Top-level keys must be exact lower case and unique, or the request returns `400 BAD_REQUEST` with the key and reason in `fields`. Go's decoder would otherwise match `OPERANDS` to `operands` and keep the last of two duplicates.
- Trade-off: The request is a little less natural to write by hand. Every client must send quotes. In exchange the server sees exactly what the user typed.

## 3. Half-even rounding and a precision of 20

- Context: Divide, square root and negative powers have results with no end, such as `1/3`. Something must cut them off, and the cut must not push totals in one direction.
- Decision: Round to 20 decimal places with half-even rounding. An exact tie goes to the even digit, so `0.125` becomes `0.12`. Add, subtract, multiply and percentage stay exact. The precision is the `decimal_precision` config value.
- Trade-off: Half-even surprises users who learned "round half up". Financial work uses half-even because the errors cancel out over many operations.

## 4. Percentage means x percent of y

- Context: "Percentage" has several meanings. Readings include "x percent of y", "what percent is x of y" and "x increased by y percent".
- Decision: `percentage(x, y)` is `x / 100 * y`. So `percentage(15, 200)` is `30`. The percent comes first and the base second.
- Trade-off: The other readings are not available. The choice is written in the API docs and the operation description, so nobody has to guess.
- In the UI the key reads `% of`, is named "Percent of" for screen readers and has a tooltip that gives the order. The display shows `20% of` after the key is pressed. A bare `%` reads differently on every phone calculator, so the label says which meaning this one has.

## 5. Input and result limits

- Context: A public endpoint with no bounds on numbers invites abuse of CPU and memory. `9^1000` has 955 digits, and larger exponents grow without bound.
- Decision: An operand has at most 64 characters. A `power` exponent is within -1000 to 1000. Any result has at most 640 digits. The core counts the finished result of every operation exactly, and `power` also checks with whole-number arithmetic before it multiplies. The body is at most 1 MiB. An `operation` name has at most 32 bytes and `operands` at most 16 items, both checked before parsing (`422 VALIDATION_FAILED`). The first three are config values. The result digit budget is derived as `max_operand_length * 10`, published as `limits.maxResultDigits`, and the body limit is `max_body_bytes`.
- Trade-off: Some valid maths is refused with a clear error. The limits are generous for a calculator and cheap to raise in config.

## 6. One endpoint with an operation field

- Context: The API needs seven operations. The options are one route per operation or one route with the name in the body.
- Decision: `POST /api/v1/calculate` takes `{"operation": "...", "operands": [...]}`. `GET /api/v1/operations` lists what the server supports.
- Trade-off: Logs, metrics and rate limits keyed on the path cannot tell `add` from `power`. The handler also checks the operation name itself, which a router does for free. In exchange, validation, limits and the error shape live in one place, and a new operation needs no new route.

## 7. 400 for malformed requests, 422 for domain errors

- Context: The frontend needs one simple rule for which errors to show to a user.
- Decision: `400` means the server cannot read the request, so the client code has a bug. `422` means the request was read but its content breaks a rule, such as division by zero. A missing field is `422` because the user fixes the problem by filling in a field.
- Trade-off: Some teams use `400` for everything. The split costs a little more thought on the server and gives the client a rule to follow: show `message` on `422`, show a generic failure otherwise.

## 8. No database

- Context: A calculation is a pure function of its input. Nothing in the assignment needs saved data.
- Decision: The server keeps no state. The history in the UI lives in memory and is gone on reload.
- Trade-off: There is no history across sessions or devices. The server scales to any number of copies with no coordination and has nothing to back up or migrate.

## 9. An operation registry, open for extension

- Context: Adding an operation should not mean editing a switch statement in five places.
- Decision: Each operation is one type with `Name`, `Arity`, `Symbol`, `Description` and `Apply`. A registry looks them up by name, lists them in a fixed order and panics at boot on a duplicate name. `GET /api/v1/operations` reports the registered operations and the configured limits. The web UI keeps a static key list, disables keys the server does not report and reads `limits.maxOperandLength` from the response, falling back to 64 if the request fails. A new operation needs a key in the UI as well.
- Trade-off: More types and one level of indirection than a single function with a `switch`. A new operation is one file and one line of wiring, and the core never changes.

## 10. An error model with client-safe fields

- Context: Error bodies are read by users and by logs. Leaking internal text or echoing user input is a security and support problem.
- Decision: Every error has `code`, `message`, `requestId` and an optional `fields` object. The message is fixed text per code. The cause stays in the server log under the same request ID. A test checks the code tables in the docs against the code.
- Layering: `pkg/apperror` holds only the generic codes and `Register`. The calculator codes are defined in `internal/modules/calculator/entities/codes.go` and registered in that package's `init`, so the shared package stays free of one module's vocabulary. `Register` panics on a duplicate code or a status outside 400-599.
- Trade-off: Each new failure needs a code, a status and a message. Clients get a stable contract, and one request ID connects a user report to a log line.

## 11. The frontend never computes

- Context: If the browser also did arithmetic, the two sides might disagree, and the float problem would return.
- Decision: The UI builds an expression and sends the expression. Every result comes from the API.
- Trade-off: The calculator needs the network. A failed request shows an error with a retry button. The answer is always the one the backend computed.

## 12. Keyboard-first UI and accessibility

- Context: A calculator is used by keyboard as much as by touch, and some users depend on screen readers.
- Decision: Every key works from the keyboard. Errors use `role="alert"`. The active operator uses `aria-pressed` and a ring, not colour alone. Touch targets are at least 44px. Contrast is tested for both themes, and axe runs in the component tests.
- Trade-off: More tests and more CSS rules to maintain. A regression now fails the build. Details are in [ACCESSIBILITY.md](ACCESSIBILITY.md).

## 13. Go with chi, zap and viper

- Context: The backend needs routing, structured logs and layered config.
- Decision: chi for routing, because chi is small and uses standard `net/http` types. zap for structured logs. viper for TOML files with environment overrides. golangci-lint runs 12 linters.
- Trade-off: Three dependencies where the standard library covers part of the need. Each one is common, well documented and easy to replace behind a thin wrapper.

## 14. TypeScript pinned to 6.0

- Context: typescript-eslint supports only certain TypeScript versions. A newer compiler leaves the linter unable to parse the code.
- Decision: Pin `typescript` to `~6.0`, the range the installed typescript-eslint supports.
- Trade-off: The project does not get newer compiler features until typescript-eslint catches up. A working lint beats a compiler one release ahead.

## 15. Shutdown timing and the container stop period

- Context: On shutdown the backend turns readiness to 503, keeps serving for `shutdown_drain_delay`, then waits up to `shutdown_grace_period` for open requests. Docker sends SIGKILL once `stop_grace_period` ends. A kill in the middle of a drain drops live requests.
- Decision: `stop_grace_period` in `docker-compose.yml` must be longer than `shutdown_drain_delay + shutdown_grace_period`. Production uses 5s + 10s, so compose sets 20s for the backend. If you raise either config value, raise `stop_grace_period` with it.
- Trade-off: A slow stop takes up to 20s in the worst case. A normal stop ends as soon as the last request finishes after the drain delay.

## 16. Money versus calculator numbers

- Context: This API does arithmetic on whatever a user types. A real buy now, pay later amount API is a different problem.
- A money API would send integer minor units plus an ISO 4217 currency code, for example `{"amount": 1999, "currency": "USD"}`.
- Rounding would follow the currency, since JPY has no minor unit and KWD has three, instead of one global 20 places.
- It would also refuse floats and exponent notation, refuse sums across currencies, and say how the odd cent is shared when a total splits into instalments.
- Decision: The calculator takes arbitrary-precision decimal strings with no currency, because the brief asks for a general calculator with division, roots and powers.
- Trade-off: `1/3` as `0.33333333333333333333` suits a calculator and is wrong for a ledger. Do not reuse this contract to move money.
