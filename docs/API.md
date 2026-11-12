# API contract

This is the contract for the calculator REST API. The handler is written to match it, and the frontend builds against it. If the two disagree, this file is wrong or the code is wrong, and one of them gets fixed.

Error codes are also listed with examples in [ERROR_CODES.md](ERROR_CODES.md). A test checks that both files list every code with the same HTTP status.

## Overview

| Method | Path | Purpose |
| --- | --- | --- |
| POST | `/api/v1/calculate` | Run one operation on its operands |
| GET | `/api/v1/operations` | List the operations the server supports |
| GET | `/health/live` | Liveness probe |
| GET | `/health/ready` | Readiness probe |

Every route under `/api/v1` is versioned. A breaking change ships as `/api/v2` and `v1` keeps working. The health routes sit at the root so load balancers do not depend on the API version.

All bodies are JSON in UTF-8. Any other method on these paths returns `405 METHOD_NOT_ALLOWED`, and any other path returns `404 NOT_FOUND`, both in the standard error shape below.

## Headers

| Header | Direction | Rule |
| --- | --- | --- |
| `Content-Type: application/json` | request | Required on `POST`. A `charset=utf-8` parameter is accepted. A missing header, another media type or another charset returns `415 UNSUPPORTED_MEDIA_TYPE`. |
| `X-Request-ID` | request | Optional. Used as the request ID when it is at most 128 characters of letters, digits, `.`, `_` and `-`. Anything else is ignored and the server generates a UUID. |
| `X-Request-ID` | response | Always present. It holds the request ID and is repeated as `requestId` in every error body, so a user can quote one string to support. |
| `Content-Type: application/json` | response | Always present, on success and on error. |
| `Cache-Control: no-store` | response | Always present. Every response is computed for one request and must not be stored by a proxy or browser. |

Browsers may call the API from the origins listed in `cors.allowed_origins`. Allowed methods are `GET`, `POST` and `OPTIONS`. Cookies are not used.

## Numbers are strings

Operands go in as JSON strings and the result comes back as a string.

```json
{"operation": "add", "operands": ["0.1", "0.2"]}
```

The reason is precision. Most JSON parsers, JavaScript's included, turn a JSON number into a 64-bit float before the application sees it. A float cannot hold `0.1` exactly, and it cannot hold an integer with 20 digits. By the time the server received `0.1` as a number, the digits the user typed would already be gone, and `0.1 + 0.2` could never be exactly `0.3`. A string reaches the server untouched. The server reads it with exact decimal arithmetic.

A JSON number in `operands` (`[10, 4]`) is rejected with `400 BAD_REQUEST`. It is never accepted and rounded.

### Operand format

An operand is an optional leading `-`, then digits with at most one `.`. So `12`, `-3.5`, `.5`, `1.` and `007` are valid. These are rejected with `422 INVALID_OPERAND`: an empty string, whitespace anywhere, a leading `+`, exponent notation (`1e5`), hex (`0x1F`), `NaN`, `Infinity`, thousands separators (`1,000`, `1_000`), more than one `.` and anything else that is not a plain decimal. Digits must be ASCII.

### Result format

The result has no trailing zeros, no exponent notation and no `-0`. `2.50 + 2.50` returns `"5"`, and `-5 * 0` returns `"0"`.

### Precision and rounding

Add, subtract, multiply and percentage are exact. Divide, square root and a power with a negative exponent can have results that never end (`1/3`), so they are rounded to `decimal_precision` decimal places (20 by default). Rounding is half-even, also called banker's rounding: an exact tie goes to the even digit, so `0.125` becomes `0.12` and `0.135` becomes `0.14`. Ties round up half the time and down half the time, so totals are not pushed in one direction.

### Limits

These come from the `[calculator]` section of the config. The values below are the defaults.

| Limit | Default | What it bounds | Error |
| --- | --- | --- | --- |
| `max_operand_length` | 64 | The length in bytes of each operand string, including `-` and `.`. Checked before parsing. The significant digits and decimal places of the parsed value are held to the same number. | `422 OPERAND_TOO_LARGE` |
| `max_exponent` | 1000 | The absolute value of the exponent in `power`. | `422 EXPONENT_TOO_LARGE` |
| result digit budget | `max_operand_length * 10` (640) | The number of digits a `power` result may have. The check uses an upper bound and runs before any multiplication, so `2^1000` (302 digits) is computed and `9^1000` (955 digits) is refused. | `422 RESULT_TOO_LARGE` |
| `max_body_bytes` | 1048576 (1 MiB) | The size of the request body. | `413 PAYLOAD_TOO_LARGE` |

## POST /api/v1/calculate

Runs one operation.

### Request

```json
{"operation": "divide", "operands": ["10", "4"]}
```

| Field | Type | Rule |
| --- | --- | --- |
| `operation` | string | Required and not empty. One of the names returned by `GET /api/v1/operations`. Names are lower case and exact: `"Add"` is `UNKNOWN_OPERATION`. |
| `operands` | array of strings | Required. The count must match the operation's arity. Order matters. |

No other fields are allowed. An unknown field is a `400 BAD_REQUEST`, so a misspelled field (`"operand"`) fails loudly and is never ignored.

### Response 200

```json
{"operation": "divide", "operands": ["10", "4"], "result": "2.5"}
```

`operation` and `operands` repeat the request exactly as sent, so a client can show the expression next to the result without keeping its own state. `result` is the normalised string described above.

### Operations

Operands are used in the order given.

| Name | Arity | Symbol | Meaning |
| --- | --- | --- | --- |
| `add` | 2 | `+` | `a + b` |
| `subtract` | 2 | `-` | `a - b` |
| `multiply` | 2 | `*` | `a * b` |
| `divide` | 2 | `/` | `a / b`, rounded to `decimal_precision` places. `b = 0` is `DIVISION_BY_ZERO`. |
| `power` | 2 | `^` | `a` raised to `b`. `b` must be a whole number, so `2.0` is accepted and `0.5` is `INVALID_EXPONENT`. Anything to the power 0 is 1, including `0^0`. `0` to a negative power is `DIVISION_BY_ZERO`. A negative exponent is `1 / a^|b|`, rounded like divide. |
| `sqrt` | 1 | `√` | The square root of `a`. A negative `a` is `NEGATIVE_SQUARE_ROOT`. A perfect square is exact: `sqrt(144)` is `12`. |
| `percentage` | 2 | `%` | `a` percent of `b`, which is `a / 100 * b`. So `percentage(20, 50)` is `10`. |

`percentage` has several meanings in everyday use: "what percent is x of y", and "x increased by y percent" are two others. This API means only "x percent of y", with the percent first and the base second.

The symbol is for display. The frontend may draw `*` as `×`. Clients never send a symbol, only the name.

### Order of checks

The server checks in this order and returns the first failure:

1. `Content-Type` (`415`)
2. Body size, then empty body, JSON syntax, unknown fields, field types and data after the JSON object (`413`, `400`)
3. `operation` and `operands` are present (`422 VALIDATION_FAILED`)
4. Each operand's length, then its syntax, in operand order (`422 OPERAND_TOO_LARGE`, `422 INVALID_OPERAND`)
5. The operation exists, the operand count matches its arity, then the operation runs (`422`)

One consequence: an unknown operation sent with a malformed operand reports the operand, not the operation.

### Malformed request cases

| Request | Response |
| --- | --- |
| No `Content-Type`, or `text/plain` | `415 UNSUPPORTED_MEDIA_TYPE` |
| Empty body, or only whitespace | `400 BAD_REQUEST` |
| `{"operation": ` (invalid or cut-off JSON) | `400 BAD_REQUEST` |
| `{"operation": "add", "operands": ["1", "2"], "extra": 1}` | `400 BAD_REQUEST` (unknown field) |
| `{"operation": "add", "operands": [1, 2]}` | `400 BAD_REQUEST` (operands must be strings) |
| `{"operation": "add", "operands": ["1", "2"]} {"x": 1}` or any text after the object | `400 BAD_REQUEST` (trailing data) |
| A body over `max_body_bytes` | `413 PAYLOAD_TOO_LARGE` |
| `{}`, or `operation` missing, empty or `null` | `422 VALIDATION_FAILED`, with `fields.operation` |
| `operands` missing or `null` | `422 VALIDATION_FAILED`, with `fields.operands` |
| `"operands": []` | `422 INVALID_OPERAND_COUNT` (the field is present, the count is wrong) |

## GET /api/v1/operations

Lists every operation, so the frontend can build its keypad and its operand inputs from the server and does not hard-code them. Adding an operation on the server shows up in the UI without a frontend change.

### Response 200

```json
{
  "operations": [
    {"name": "add", "arity": 2, "symbol": "+", "description": "Adds the two operands."},
    {"name": "subtract", "arity": 2, "symbol": "-", "description": "Subtracts the second operand from the first."},
    {"name": "multiply", "arity": 2, "symbol": "*", "description": "Multiplies the two operands."},
    {"name": "divide", "arity": 2, "symbol": "/", "description": "Divides the first operand by the second."},
    {"name": "power", "arity": 2, "symbol": "^", "description": "Raises the first operand to the whole-number power given by the second."},
    {"name": "sqrt", "arity": 1, "symbol": "√", "description": "Returns the square root of the operand."},
    {"name": "percentage", "arity": 2, "symbol": "%", "description": "Returns the first operand percent of the second."}
  ]
}
```

| Field | Type | Meaning |
| --- | --- | --- |
| `name` | string | The value to send as `operation`. |
| `arity` | number | How many operands the operation takes. |
| `symbol` | string | A short symbol for buttons and for showing an expression. |
| `description` | string | One line of plain text for a tooltip or help text. |

The list is in a fixed order (the order above), and the response does not change while the server runs. The array sits inside an object so a field can be added later without breaking clients.

## Health

`GET /health/live` returns `200 {"status":"ok"}` while the process can serve a request. It checks no dependency, so a failing dependency does not get the process restarted.

`GET /health/ready` returns `200 {"status":"ready"}` when the service should receive traffic. Before startup finishes and once shutdown has begun it returns `503 SERVICE_UNAVAILABLE` in the standard error shape.

## Errors

Every non-2xx response, from every route, has this one shape:

```json
{
  "error": {
    "code": "DIVISION_BY_ZERO",
    "message": "Cannot divide by zero.",
    "requestId": "3f6c1d0e-8a54-4f0b-9a8e-0d5b0c1f7a21"
  }
}
```

| Field | Always present | Meaning |
| --- | --- | --- |
| `code` | yes | A stable, machine-readable name. Clients branch on this. |
| `message` | yes | Fixed text per code, safe to show a user. It never contains the user's input or any internal detail. |
| `requestId` | yes | The same value as the `X-Request-ID` response header. |
| `fields` | no | A map from an input field to a short problem description. It is sent only with `VALIDATION_FAILED` and left out otherwise. |

A validation error:

```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "The request failed validation.",
    "requestId": "3f6c1d0e-8a54-4f0b-9a8e-0d5b0c1f7a21",
    "fields": {"operation": "is required", "operands": "is required"}
  }
}
```

Extra detail about a domain error (which operand failed, which limit) is written to the server log with the request ID. It is not sent to the client, because the message is enough to tell a calculator user what to change, and the log has the rest.

### Error codes and HTTP statuses

| Code | HTTP status | When |
| --- | --- | --- |
| `BAD_REQUEST` | 400 | The body cannot be read as the request in this contract: empty, invalid JSON, an unknown field, a wrong field type or data after the JSON object. |
| `VALIDATION_FAILED` | 422 | The JSON is fine, but `operation` or `operands` is missing or empty. |
| `UNSUPPORTED_MEDIA_TYPE` | 415 | `Content-Type` is missing, is not `application/json` or has a charset other than UTF-8. |
| `PAYLOAD_TOO_LARGE` | 413 | The body is larger than `max_body_bytes`. |
| `NOT_FOUND` | 404 | No route matches the path. |
| `METHOD_NOT_ALLOWED` | 405 | The path exists but not for this method. |
| `INTERNAL` | 500 | An unexpected server fault. The cause is logged and never sent. |
| `SERVICE_UNAVAILABLE` | 503 | The readiness probe while the service is not ready. |
| `UNKNOWN_OPERATION` | 422 | `operation` is not a name from `GET /api/v1/operations`. |
| `INVALID_OPERAND_COUNT` | 422 | The number of operands does not match the operation's arity. |
| `OPERAND_TOO_LARGE` | 422 | An operand is longer than `max_operand_length`. |
| `INVALID_OPERAND` | 422 | An operand is not a plain decimal number. |
| `DIVISION_BY_ZERO` | 422 | `divide` by zero, or zero to a negative power. |
| `INVALID_EXPONENT` | 422 | The exponent in `power` is not a whole number. |
| `EXPONENT_TOO_LARGE` | 422 | The exponent in `power` is outside `-max_exponent` to `max_exponent`. |
| `RESULT_TOO_LARGE` | 422 | A `power` result would exceed the digit budget. |
| `NEGATIVE_SQUARE_ROOT` | 422 | `sqrt` of a negative number. |

### Why 400 for malformed and 422 for domain errors

The split is about what the client should do next.

- `400` means the server could not read the request as a calculate request at all. The client has a bug or sent garbage, and sending the same request again will never work. The fix is in the code that builds the request.
- `422` means the request was understood and had the right shape, but its content fails a rule: a missing value, an operand that is not a number, a division by zero. This is usually something the user typed. The fix is a different input, and the `message` is written to be shown to that user.

So the frontend can follow one rule. On `422`, show `message` next to the input. On `400`, `413`, `415` and `5xx`, show a generic failure and log `requestId`.

`VALIDATION_FAILED` is a judgement call and is `422`, not `400`. The JSON parsed and had the right shape, the value is just absent, and the user can fix it by filling in a field. `400` is kept for requests that fail before the server knows what they say.

`415`, `413`, `404`, `405`, `500` and `503` use their standard statuses, because HTTP already says what they mean.

## Why one endpoint and not one route per operation

The design is `POST /calculate` with an `operation` field, and not `/add`, `/subtract` and so on.

Reasons for one endpoint:

- Adding an operation is a server change only. The route table, the request type, the handler and the docs stay the same, and the operation list endpoint tells the frontend about it.
- Validation, the error shape, limits, logging and tests live in one handler. Seven handlers would repeat them seven times, and they would drift apart.
- The request says what to do in its body, so one client function can call the API with any operation, and an expression history can be stored as plain request bodies.
- The operation is data, so it can be chosen at runtime (a button the user pressed) without building a URL.

The trade-off:

- The operation is not in the URL. Access logs, metrics and rate limits keyed on the path cannot tell `add` from `power` without reading the body, and the cost of `power` and `add` is very different. Per-operation limits would need the handler to look inside the body.
- The server must validate the `operation` string itself, which a router does for free when each operation is a route. That is the `UNKNOWN_OPERATION` error.
- A per-operation route could have a request type that fits it exactly (`sqrt` with a single `operand`). Here `operands` is an array for all operations, so the arity check is done at run time and not by the type.
- Per-route caching and permissions are harder. This API caches nothing (`Cache-Control: no-store`) and has no permissions, so that cost is zero today.

For seven operations that share rules and one error shape, the single handler is the simpler option. If one operation later needed its own limits or permissions, a route for it could be added next to `/calculate` without breaking `v1`.
