# Error codes

Every error the API returns uses one shape, `{"error": {"code": "...", "message": "...", "requestId": "...", "fields": {...}}}`. The message is fixed per code and never contains user input or internal detail. The `fields` column in the first table names the keys of the `fields` map sent to the client with that error; see [API.md](API.md).

Domain errors (the request was well formed but the calculation cannot be done) are 422. Malformed requests and infrastructure failures use the other statuses.

Calculator codes are defined in `backend/internal/modules/calculator/entities/codes.go` and added to the registry in `pkg/apperror` when that package loads. `pkg/apperror` itself holds only the generic codes and `Register`.

Defaults used in the examples: `max_operand_length = 64`, `max_exponent = 1000`, `decimal_precision = 20`. Operands are always strings.

## Calculator errors

| Code | HTTP status | Example input | Message | Fields |
| --- | --- | --- | --- | --- |
| `UNKNOWN_OPERATION` | 422 | `operation: "modulo"` | The requested operation is not supported. | `operation` (left out when the name is over 32 bytes) |
| `INVALID_OPERAND_COUNT` | 422 | `add` with `["1"]` | The operation received the wrong number of operands. | `operation`, `expected`, `actual` |
| `INVALID_OPERAND` | 422 | `add` with `["1", "1e5"]` | An operand is not a valid number. | `operand_index`, `reason` |
| `OPERAND_TOO_LARGE` | 422 | `add` with `["1", "<65 digits>"]` | An operand is larger than the allowed limit. | `operand_index`, `limit` |
| `DIVISION_BY_ZERO` | 422 | `divide` with `["10", "0"]`, or `power` with `["0", "-1"]` | Cannot divide by zero. | none |
| `INVALID_EXPONENT` | 422 | `power` with `["2", "0.5"]` | The exponent must be a whole number. | `operand_index` |
| `EXPONENT_TOO_LARGE` | 422 | `power` with `["2", "1001"]` | The exponent is outside the allowed range. | `operand_index`, `limit` |
| `RESULT_TOO_LARGE` | 422 | `power` with `["9", "1000"]` | The result would be too large to calculate. | `limit` |
| `NEGATIVE_SQUARE_ROOT` | 422 | `sqrt` with `["-4"]` | Cannot take the square root of a negative number. | none |

`INVALID_OPERAND_COUNT` is for a count that is wrong but small: 3 operands to `add`. More than 16 operands (`entities.MaxOperandCount`) is refused before parsing as `VALIDATION_FAILED` with `fields.operands` set to `too many`.

`INVALID_OPERAND` reasons: `empty`, `whitespace`, `plus_sign`, `exponent`, `hex`, `nan`, `infinity`, `separator`, `multiple_points`, `invalid_format`.

Examples per reason: `""` is `empty`, `" 1"` is `whitespace`, `"+1"` is `plus_sign`, `"1e5"` is `exponent`, `"0x1F"` is `hex`, `"NaN"` is `nan`, `"Infinity"` is `infinity`, `"1,000"` is `separator`, `"1.2.3"` is `multiple_points`, `"abc"` is `invalid_format`.

`RESULT_TOO_LARGE` limit: a power result may have at most `max_operand_length * 10` digits (640 by default), counting integer and fraction digits. Trailing zeros of the base and its decimal places count too: `power 10 1000` (1001 digits), `power 0.1 1000` (1000 fraction digits) and `power 0.1 -1000` (1001 digits) are refused. The check is an upper bound with a margin of one digit, made before any multiplication, and the finished result is counted exactly as well. So `power 2 1000` (302 digits) is computed and `power 9 1000` (955 digits) is refused.

## Generic errors

| Code | HTTP status | Example input | Message | Fields |
| --- | --- | --- | --- | --- |
| `BAD_REQUEST` | 400 | A body that is not valid JSON, or a body with `"OPERANDS"` or a repeated `"operation"` key | The request could not be understood. | `key`, `reason` (`duplicate` or `wrong_case`), only for a bad key |
| `VALIDATION_FAILED` | 422 | A request missing the `operation` or `operands` field, an `operation` over 32 bytes, or more than 16 `operands` | The request failed validation. | `operation` (`is required`, `is too long`), `operands` (`is required`, `too many`) |
| `UNSUPPORTED_MEDIA_TYPE` | 415 | `Content-Type: text/plain` | The content type is not supported. | none |
| `PAYLOAD_TOO_LARGE` | 413 | A body over `max_body_bytes` (1 MiB) | The request body is too large. | `limit` (bytes) |
| `NOT_FOUND` | 404 | `GET /api/v1/missing` | The requested resource was not found. | `method`, `path` (path cut to 256 bytes) |
| `METHOD_NOT_ALLOWED` | 405 | `DELETE /api/v1/calculate` | The method is not allowed for this resource. | `method`, `path` (path cut to 256 bytes) |
| `INTERNAL` | 500 | An unexpected server fault, such as a recovered panic | An internal error occurred. | none |
| `SERVICE_UNAVAILABLE` | 503 | The readiness probe while the service is not ready | The service is temporarily unavailable. | none |
| `REQUEST_TIMEOUT` | 503 | A calculation still running when `request_timeout` expires | The request took too long to process. | none |

`REQUEST_TIMEOUT` is 503 and not 408 or 504 on purpose. 408 means the client was too slow to send its request, and 504 means a gateway timed out waiting for an upstream; neither is true here. The request was valid and the server's own deadline (`request_timeout`) ran out, which is a temporary condition the client can retry. A client that disconnects mid-calculation also ends the work with this code, though nobody reads that response.

`PAYLOAD_TOO_LARGE` carries one field, `limit`, the body size limit in bytes. It is the same whether the declared `Content-Length` is refused up front or a body of unknown length is cut off while it is read.

`BAD_REQUEST`, `VALIDATION_FAILED`, `UNSUPPORTED_MEDIA_TYPE` and `PAYLOAD_TOO_LARGE` are raised by the calculate handler as well as by the router. [API.md](API.md) describes the request cases and the order of the checks.

`INTERNAL` can also come from the core in one case: the square root iteration hits its fixed cap. The cap is far above what any accepted input needs, so this is a safety net and not an expected result.

The code list in this file is checked against `pkg/apperror` by a test, so a code added without a row here fails the build.
