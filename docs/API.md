# API contract

This is the contract for the calculator REST API. The handler is written to match it, and the frontend builds against it. If the two disagree, this file is wrong or the code is wrong, and one of them gets fixed.

Error codes are also listed with examples in [ERROR_CODES.md](ERROR_CODES.md). A test checks that both files list every code with the same HTTP status.

## Overview

| Method | Path | Purpose |
| --- | --- | --- |
| POST | `/api/v1/calculate` | Run one operation on its operands |
| GET | `/api/v1/operations` | List the operations the server supports and the limits it enforces |
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

These come from the `[calculator]` section of the config, except `max_body_bytes`, which is in `[server]`. The values below are the defaults.

| Limit | Default | What it bounds | Error |
| --- | --- | --- | --- |
| `max_operand_length` | 64 | The length in bytes of each operand string, including `-` and `.`. Checked before parsing. The significant digits and decimal places of the parsed value are held to the same number. | `422 OPERAND_TOO_LARGE` |
| `max_exponent` | 1000 | The absolute value of the exponent in `power`. | `422 EXPONENT_TOO_LARGE` |
| result digit budget | `max_operand_length * 10` (640) | The number of digits any result may have, for every operation, counting integer and fraction digits (not the sign, not trailing fraction zeros). Zeros inside the number count: `10^1000` (1001 digits) and `0.1^1000` (1000 fraction digits) are refused even though they have one significant digit. The core counts the finished result exactly after each operation. `power` also checks with whole-number arithmetic before it multiplies, so `9^1000` (955 digits) is refused without building the number, while `2^1000` (302 digits) and `7^757` (exactly 640 digits) are computed. Add, subtract, multiply, divide, percentage and square root cannot reach the budget with operands of at most 64 characters (the largest product has 128 digits), but they pass through the same check. | `422 RESULT_TOO_LARGE` |
| `request_timeout` | 5s | How long one request may run. The calculation checks the deadline before it starts, inside the power and square root loops, and after it finishes. See the note below the error table. | `503 REQUEST_TIMEOUT` |
| `max_body_bytes` | 1048576 (1 MiB) | The size of the request body. | `413 PAYLOAD_TOO_LARGE` |

## POST /api/v1/calculate

Runs one operation.

### Request

```json
{"operation": "divide", "operands": ["10", "4"]}
```

| Field | Type | Rule |
| --- | --- | --- |
| `operation` | string | Required, not empty and at most 32 bytes. One of the names returned by `GET /api/v1/operations`. Names are lower case and exact: `"Add"` is `UNKNOWN_OPERATION`. |
| `operands` | array of strings | Required. At most 16 items. The count must match the operation's arity. Order matters. |

No other fields are allowed. An unknown field is a `400 BAD_REQUEST`, so a misspelled field (`"operand"`) fails loudly and is never ignored.

Field names are exact lower case. The handler rejects a top-level key that repeats, or that matches a field name in another case (`"Operation"`, `"OPERANDS"`), with `400 BAD_REQUEST`. `fields` holds `key`, the offending key, and `reason`, either `duplicate` or `wrong_case`. Two readers of one body then never see different values.

```json
{"error": {"code": "BAD_REQUEST", "message": "The request could not be understood.", "requestId": "f2920dad-6fb6-457f-a23c-3f6df63887b4", "fields": {"key": "OPERANDS", "reason": "wrong_case"}}}
```

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

`percentage` has several meanings in everyday use: "what percent is x of y", and "x increased by y percent" are two others. This API means only "x percent of y", with the percent first and the base second. The web keypad says the same: the key reads `% of` (accessible name "Percent of"), has a tooltip "x percent of y: type x, press % of, then type y", and the display shows `20% of` while the second number is typed.

The symbol is for display. The frontend may draw `*` as `×`. Clients never send a symbol, only the name.

### Order of checks

The server checks in this order and returns the first failure:

1. `Content-Type` (`415`)
2. Body size (`413`)
3. Top-level keys are exact, lower case and not repeated (`400 BAD_REQUEST`). This scan reads the object's keys in order and stops quietly at the first JSON error, so a wrong-case or repeated key before a syntax error or an unknown field is the one reported
4. Empty body, JSON syntax, unknown fields, field types and data after the JSON object (`400`)
5. `operation` and `operands` are present, `operation` is at most 32 bytes and `operands` has at most 16 items (`422 VALIDATION_FAILED`)
6. Each operand's length, then its syntax, in operand order (`422 OPERAND_TOO_LARGE`, `422 INVALID_OPERAND`)
7. The operation exists, the operand count matches its arity, then the operation runs (`422`)

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
| `operation` longer than 32 bytes | `422 VALIDATION_FAILED`, with `fields.operation` (`is too long`) |
| `operands` missing or `null` | `422 VALIDATION_FAILED`, with `fields.operands` (`is required`) |
| `operands` with more than 16 items | `422 VALIDATION_FAILED`, with `fields.operands` (`too many`), before any operand is parsed |
| A top-level key that repeats or is not exact lower case | `400 BAD_REQUEST`, with `fields.key` and `fields.reason` (`duplicate` or `wrong_case`) |
| `"operands": []` | `422 INVALID_OPERAND_COUNT` (the field is present, the count is wrong) |

## GET /api/v1/operations

Lets a client discover which operations the server supports and the limits it enforces. The web UI keeps a static key list. It disables any key whose operation the server does not report, and it reads `limits.maxOperandLength` to cap the operand length. A new operation on the server still needs a key in the UI.

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
  ],
  "limits": {"maxOperandLength": 64, "maxExponent": 1000, "decimalPrecision": 20, "maxResultDigits": 640}
}
```

| Field | Type | Meaning |
| --- | --- | --- |
| `name` | string | The value to send as `operation`. |
| `arity` | number | How many operands the operation takes. |
| `symbol` | string | A short symbol for buttons and for showing an expression. |
| `description` | string | One line of plain text for a tooltip or help text. |

The `limits` object reports the configured limits, so a client does not hard-code them.

| Field | Type | Meaning |
| --- | --- | --- |
| `limits.maxOperandLength` | number | `max_operand_length`. The longest operand string the server accepts. |
| `limits.maxExponent` | number | `max_exponent`. The largest absolute exponent `power` accepts. |
| `limits.decimalPrecision` | number | `decimal_precision`. The decimal places kept when a result is rounded. |
| `limits.maxResultDigits` | number | The most digits any result may have: `max_operand_length * 10`. A result over it returns `RESULT_TOO_LARGE`. |

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
| `fields` | no | Client-safe context about the failure, such as which operand failed and why. Left out when an error has none. Which keys each code carries is listed in [ERROR_CODES.md](ERROR_CODES.md). |

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

An operand that is not a number says which operand and why (`operand_index` counts from 0):

```json
{
  "error": {
    "code": "INVALID_OPERAND",
    "message": "An operand is not a valid number.",
    "requestId": "3f6c1d0e-8a54-4f0b-9a8e-0d5b0c1f7a21",
    "fields": {"operand_index": 1, "reason": "exponent"}
  }
}
```

An operand over the length limit says which one and what the limit is:

```json
{
  "error": {
    "code": "OPERAND_TOO_LARGE",
    "message": "An operand is larger than the allowed limit.",
    "requestId": "3f6c1d0e-8a54-4f0b-9a8e-0d5b0c1f7a21",
    "fields": {"operand_index": 1, "limit": 64}
  }
}
```

`fields` values are strings or numbers. The set of keys is fixed per code. The internal cause of an error (driver text, stack details) is never sent; it stays in the server log with the request ID. `NOT_FOUND` and `METHOD_NOT_ALLOWED` echo the request `method` and `path` in `fields`, since the client sent both. The echoed `path` is cut to 256 bytes.

### Rate limit

`POST /api/v1/calculate` is limited per client IP with a token bucket. Each client starts with `burst` requests and earns `requests_per_second` more every second. `GET /api/v1/operations` and the health probes are not limited. These come from the `[rate_limit]` section of the config.

| Key | Default | Meaning |
| --- | --- | --- |
| `enabled` | true | Turns the limit off for load tests. |
| `requests_per_second` | 10.0 | Refill rate, above 0 and at most 10000. |
| `burst` | 20 | Bucket size, from 1 to 10000. |
| `client_ip_header` | empty | Header that holds the client IP behind a proxy, for example `X-Real-IP`. Empty uses the connection address. Set it only when the proxy overwrites the header, or a caller could pick its own bucket. The compose stack sets it to `X-Real-IP`, which nginx overwrites. |

An invalid value stops the service at startup. A request over the limit gets `429 RATE_LIMITED`, the normal error body with `fields.retryAfterSeconds`, and a `Retry-After` header holding the same number of seconds (rounded up, at least 1). Idle clients are dropped from memory once their bucket is full again, so memory follows the number of recently active clients.

The buckets live in the process. With several instances each one counts on its own, so the real limit is the sum across instances. A multi-instance deployment needs a shared store such as Redis, or a limit at the gateway.

### Error codes and HTTP statuses

| Code | HTTP status | When |
| --- | --- | --- |
| `BAD_REQUEST` | 400 | The body cannot be read as the request in this contract: empty, invalid JSON, an unknown field, a wrong field type, data after the JSON object, or a top-level key that repeats or is not exact lower case (`fields.key`, `fields.reason`). |
| `VALIDATION_FAILED` | 422 | The JSON is fine, but `operation` or `operands` is missing (or `operation` is empty), `operation` is over 32 bytes, or `operands` has more than 16 items. |
| `UNSUPPORTED_MEDIA_TYPE` | 415 | `Content-Type` is missing, is not `application/json` or has a charset other than UTF-8. |
| `PAYLOAD_TOO_LARGE` | 413 | The body is larger than `max_body_bytes`. `fields.limit` is that limit in bytes. |
| `NOT_FOUND` | 404 | No route matches the path. |
| `METHOD_NOT_ALLOWED` | 405 | The path exists but not for this method. |
| `INTERNAL` | 500 | An unexpected server fault. The cause is logged and never sent. |
| `SERVICE_UNAVAILABLE` | 503 | The readiness probe while the service is not ready. |
| `REQUEST_TIMEOUT` | 503 | The calculation was still running when `request_timeout` expired, or the client went away. |
| `RATE_LIMITED` | 429 | The client IP used up its bucket on `POST /api/v1/calculate`. Wait `Retry-After` seconds. |
| `UNKNOWN_OPERATION` | 422 | `operation` is not a name from `GET /api/v1/operations`. The name is echoed in `fields.operation` only when it fits the 32 byte limit. |
| `INVALID_OPERAND_COUNT` | 422 | The number of operands does not match the operation's arity, for example 3 operands to `add`. More than 16 operands is `VALIDATION_FAILED` instead. |
| `OPERAND_TOO_LARGE` | 422 | An operand is longer than `max_operand_length`. |
| `INVALID_OPERAND` | 422 | An operand is not a plain decimal number. |
| `DIVISION_BY_ZERO` | 422 | `divide` by zero, or zero to a negative power. |
| `INVALID_EXPONENT` | 422 | The exponent in `power` is not a whole number. |
| `EXPONENT_TOO_LARGE` | 422 | The exponent in `power` is outside `-max_exponent` to `max_exponent`. |
| `RESULT_TOO_LARGE` | 422 | A result would have more digits than the result digit budget allows. Any operation can return it, `power` is the one that reaches it in practice. |
| `NEGATIVE_SQUARE_ROOT` | 422 | `sqrt` of a negative number. |

### Error examples (real curl output)

Bodies go to `POST /api/v1/calculate` with `Content-Type: application/json` unless noted. The full commands are in [API_SMOKE.md](API_SMOKE.md). The `requestId` values change on every call.

Common errors:

Division by zero (HTTP 422):

```bash
curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" \
  -d '{"operation":"divide","operands":["10","0"]}'
```

```json
{"error":{"code":"DIVISION_BY_ZERO","message":"Cannot divide by zero.","requestId":"139e5ba9-8f0e-4707-b316-ed9c0db14ad4"}}
```

A bad operand, with the field and reason (HTTP 422):

```bash
curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" \
  -d '{"operation":"add","operands":["1e5","2"]}'
```

```json
{"error":{"code":"INVALID_OPERAND","message":"An operand is not a valid number.","requestId":"cce3b486-af9e-4fd9-b0c8-7ca385ab8e90","fields":{"operand_index":0,"reason":"exponent"}}}
```

A missing field (HTTP 422):

```bash
curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" \
  -d '{"operands":["1","2"]}'
```

```json
{"error":{"code":"VALIDATION_FAILED","message":"The request failed validation.","requestId":"17584436-4624-495a-815e-5c1cb2df9b5d","fields":{"operation":"is required"}}}
```

A wrong media type (HTTP 415):

```bash
curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: text/plain" \
  -d '{"operation":"add","operands":["1","2"]}'
```

```json
{"error":{"code":"UNSUPPORTED_MEDIA_TYPE","message":"The content type is not supported.","requestId":"d4f1a0b4-d681-4e5b-8037-6fe33e2d3658"}}
```

Remaining cases:

`INTERNAL`, `SERVICE_UNAVAILABLE` and `REQUEST_TIMEOUT` have no curl example here because a normal request does not trigger them. Tests cover all three.

`NEGATIVE_SQUARE_ROOT`, body `{"operation":"sqrt","operands":["-4"]}`, HTTP 422:

```json
{"error":{"code":"NEGATIVE_SQUARE_ROOT","message":"Cannot take the square root of a negative number.","requestId":"77f8bdf6-c58a-4ba0-aad1-0e13b2cd2658"}}
```

`INVALID_EXPONENT`, body `{"operation":"power","operands":["2","0.5"]}`, HTTP 422:

```json
{"error":{"code":"INVALID_EXPONENT","message":"The exponent must be a whole number.","requestId":"9db8b836-e6db-4946-943f-02b118c352f9","fields":{"operand_index":1}}}
```

`EXPONENT_TOO_LARGE`, body `{"operation":"power","operands":["2","1001"]}`, HTTP 422:

```json
{"error":{"code":"EXPONENT_TOO_LARGE","message":"The exponent is outside the allowed range.","requestId":"b40da9a3-4a9e-4017-b017-3ad05f57a651","fields":{"limit":1000,"operand_index":1}}}
```

`RESULT_TOO_LARGE`, body `{"operation":"power","operands":["9","1000"]}`, HTTP 422:

```json
{"error":{"code":"RESULT_TOO_LARGE","message":"The result would be too large to calculate.","requestId":"59e22c67-ebac-45c2-aca3-280f742d2b77","fields":{"limit":640}}}
```

`UNKNOWN_OPERATION`, body `{"operation":"modulo","operands":["1","2"]}`, HTTP 422:

```json
{"error":{"code":"UNKNOWN_OPERATION","message":"The requested operation is not supported.","requestId":"cf975be3-331f-4534-adc7-34709c3be8ce","fields":{"operation":"modulo"}}}
```

`INVALID_OPERAND_COUNT`, body `{"operation":"add","operands":["1"]}`, HTTP 422:

```json
{"error":{"code":"INVALID_OPERAND_COUNT","message":"The operation received the wrong number of operands.","requestId":"b117774b-72de-476b-921e-819aa2b8a6e8","fields":{"actual":1,"expected":2,"operation":"add"}}}
```

`OPERAND_TOO_LARGE`, with a first operand over 64 characters, HTTP 422:

```json
{"error":{"code":"OPERAND_TOO_LARGE","message":"An operand is larger than the allowed limit.","requestId":"f5084070-7c0e-40a4-9185-d3a73b56ae96","fields":{"limit":64,"operand_index":0}}}
```

`BAD_REQUEST`, body `{"operation":` (cut-off JSON), HTTP 400. An empty body, an unknown field (`"extra":true`) and data after the object return the same code and message:

```json
{"error":{"code":"BAD_REQUEST","message":"The request could not be understood.","requestId":"3ce3b06b-9bac-40e1-bd32-8e2b416e716e"}}
```

`BAD_REQUEST` for a key in the wrong case, body `{"operation":"add","operands":["1","2"],"OPERANDS":["9","9"]}`, HTTP 400. A repeated key returns the same code with `"reason":"duplicate"`. Only exact lowercase `operation` and `operands` are accepted, so two readers of one body never see different values:

```json
{"error":{"code":"BAD_REQUEST","message":"The request could not be understood.","requestId":"f2920dad-6fb6-457f-a23c-3f6df63887b4","fields":{"key":"OPERANDS","reason":"wrong_case"}}}
```

`VALIDATION_FAILED` for 17 operands to `add`, HTTP 422. The count is checked before any operand is parsed. Three operands to `add` still return `INVALID_OPERAND_COUNT`:

```json
{"error":{"code":"VALIDATION_FAILED","message":"The request failed validation.","requestId":"7882a092-c476-40f8-8f66-6e03ac726183","fields":{"operands":"too many"}}}
```

`PAYLOAD_TOO_LARGE`, a 1.1 MB body, HTTP 413:

```json
{"error":{"code":"PAYLOAD_TOO_LARGE","message":"The request body is too large.","requestId":"838bf0e4-2942-47a1-a7eb-eaab52de325f","fields":{"limit":1048576}}}
```

`METHOD_NOT_ALLOWED`, `curl -X GET localhost:8080/api/v1/calculate`, HTTP 405:

```json
{"error":{"code":"METHOD_NOT_ALLOWED","message":"The method is not allowed for this resource.","requestId":"98173eaa-117a-4cb4-ab2a-f53c9947081a","fields":{"method":"GET","path":"/api/v1/calculate"}}}
```

`NOT_FOUND`, `curl localhost:8080/api/v1/nope`, HTTP 404:

```json
{"error":{"code":"NOT_FOUND","message":"The requested resource was not found.","requestId":"1c82b3b7-3071-48f1-8ae9-6f18931b9149","fields":{"method":"GET","path":"/api/v1/nope"}}}
```

### Why 400 for malformed and 422 for domain errors

The split is about what the client should do next.

- `400` means the server could not read the request as a calculate request at all. The client has a bug or sent garbage, and sending the same request again will never work. The fix is in the code that builds the request.
- `422` means the request was understood and had the right shape, but its content fails a rule: a missing value, an operand that is not a number, a division by zero. This is usually something the user typed. The fix is a different input, and the `message` is written to be shown to that user.

So the frontend can follow one rule. On `422`, show `message` next to the input. On `400`, `413` and `415`, show a failure and log `requestId`, since sending the same request again will not help. On `5xx`, the fault may be temporary, so the frontend offers a retry of the same request.

`VALIDATION_FAILED` is a judgement call and is `422`, not `400`. The JSON parsed and had the right shape, the value is just absent, and the user can fix it by filling in a field. `400` is kept for requests that fail before the server knows what they say.

`415`, `413`, `404`, `405`, `500` and `503` use their standard statuses, because HTTP already says what they mean.

`REQUEST_TIMEOUT` is `503` and not `408` or `504`. `408` says the client was too slow to send its request, and `504` says a gateway timed out waiting for an upstream; neither describes this case. The request was valid and the server's own deadline ran out, which is temporary and safe to retry, and that is what `503` tells a client.

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
