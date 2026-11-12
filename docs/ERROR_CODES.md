# Error codes

Every error the API returns uses one shape, `{"error": {"code": "...", "message": "...", "requestId": "...", "fields": {...}}}`. The message is fixed per code and never contains user input or internal detail. The `fields` column in the first table names the extra context an error carries.

Domain errors (the request was well formed but the calculation cannot be done) are 422. Malformed requests and infrastructure failures use the other statuses.

Defaults used in the examples: `max_operand_length = 64`, `max_exponent = 1000`, `decimal_precision = 20`. Operands are always strings.

## Calculator errors

| Code | HTTP status | Example input | Message | Fields |
| --- | --- | --- | --- | --- |
| `UNKNOWN_OPERATION` | 422 | `operation: "modulo"` | The requested operation is not supported. | `operation` |
| `INVALID_OPERAND_COUNT` | 422 | `add` with `["1"]` | The operation received the wrong number of operands. | `operation`, `expected`, `actual` |
| `INVALID_OPERAND` | 422 | `add` with `["1", "1e5"]` | An operand is not a valid number. | `operand_index`, `reason` |
| `OPERAND_TOO_LARGE` | 422 | `add` with `["1", "<65 digits>"]` | An operand is larger than the allowed limit. | `operand_index`, `limit` |
| `DIVISION_BY_ZERO` | 422 | `divide` with `["10", "0"]`, or `power` with `["0", "-1"]` | Cannot divide by zero. | none |
| `INVALID_EXPONENT` | 422 | `power` with `["2", "0.5"]` | The exponent must be a whole number. | `operand_index` |
| `EXPONENT_TOO_LARGE` | 422 | `power` with `["2", "1001"]` | The exponent is outside the allowed range. | `operand_index`, `limit` |
| `RESULT_TOO_LARGE` | 422 | `power` with `["2", "1000"]` | The result would be too large to calculate. | `limit` |
| `NEGATIVE_SQUARE_ROOT` | 422 | `sqrt` with `["-4"]` | Cannot take the square root of a negative number. | none |

`INVALID_OPERAND` reasons: `empty`, `whitespace`, `plus_sign`, `exponent`, `hex`, `nan`, `infinity`, `separator`, `multiple_points`, `invalid_format`.

Examples per reason: `""` is `empty`, `" 1"` is `whitespace`, `"+1"` is `plus_sign`, `"1e5"` is `exponent`, `"0x1F"` is `hex`, `"NaN"` is `nan`, `"Infinity"` is `infinity`, `"1,000"` is `separator`, `"1.2.3"` is `multiple_points`, `"abc"` is `invalid_format`.

`RESULT_TOO_LARGE` limit: a power result may have at most `max_operand_length * 10` digits (640 by default). The check is an upper-bound estimate made before any multiplication.

## Generic errors

| Code | HTTP status | Example input | Message |
| --- | --- | --- | --- |
| `BAD_REQUEST` | 400 | A body that is not valid JSON | The request could not be understood. |
| `VALIDATION_FAILED` | 422 | A request missing the `operation` or `operands` field | The request failed validation. |
| `UNSUPPORTED_MEDIA_TYPE` | 415 | `Content-Type: text/plain` | The content type is not supported. |
| `PAYLOAD_TOO_LARGE` | 413 | A body over `max_body_bytes` (1 MiB) | The request body is too large. |
| `NOT_FOUND` | 404 | `GET /api/v1/missing` | The requested resource was not found. |
| `METHOD_NOT_ALLOWED` | 405 | `DELETE /api/v1/calculate` | The method is not allowed for this resource. |
| `INTERNAL` | 500 | An unexpected server fault, such as a recovered panic | An internal error occurred. |
| `SERVICE_UNAVAILABLE` | 503 | The readiness probe while the service is not ready | The service is temporarily unavailable. |

`BAD_REQUEST`, `VALIDATION_FAILED` and `UNSUPPORTED_MEDIA_TYPE` are raised by the calculate handler, which arrives in a later step. Until then only `PAYLOAD_TOO_LARGE`, `NOT_FOUND`, `METHOD_NOT_ALLOWED`, `INTERNAL` and `SERVICE_UNAVAILABLE` can occur.

`INTERNAL` can also come from the core in one case: the square root iteration hits its fixed cap. The cap is far above what any accepted input needs, so this is a safety net and not an expected result.

The code list in this file is checked against `pkg/apperror` by a test, so a code added without a row here fails the build.
