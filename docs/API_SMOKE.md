# API smoke test

Real requests against a running server (`make -C backend run`), captured with curl. Request IDs vary per run.

Two codes are not in this capture because no client request can cause them. `INTERNAL` needs a server fault, such as a recovered panic. `SERVICE_UNAVAILABLE` only comes from `GET /readyz` while the server is shutting down. Both are covered by the handler and router tests instead.

The `PAYLOAD_TOO_LARGE` body below shows the current `fields` shape (`limit`, the body cap in bytes). The router test `TestOversizedBodyIsRejectedBeforeRouting` checks the same body.

```text
### operations
$ curl localhost:8080/api/v1/operations
{"operations":[{"name":"add","arity":2,"symbol":"+","description":"Adds the two operands."},{"name":"subtract","arity":2,"symbol":"-","description":"Subtracts the second operand from the first."},{"name":"multiply","arity":2,"symbol":"*","description":"Multiplies the two operands."},{"name":"divide","arity":2,"symbol":"/","description":"Divides the first operand by the second."},{"name":"power","arity":2,"symbol":"^","description":"Raises the first operand to the whole-number power given by the second."},{"name":"sqrt","arity":1,"symbol":"√","description":"Returns the square root of the operand."},{"name":"percentage","arity":2,"symbol":"%","description":"Returns the first operand percent of the second."}]}
[HTTP 200]

### add 0.1 + 0.2
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"add\"\,\"operands\":\[\"0.1\"\,\"0.2\"\]\}
{"operation":"add","operands":["0.1","0.2"],"result":"0.3"}
[HTTP 200]

### subtract
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"subtract\"\,\"operands\":\[\"10\"\,\"4.75\"\]\}
{"operation":"subtract","operands":["10","4.75"],"result":"5.25"}
[HTTP 200]

### multiply
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"multiply\"\,\"operands\":\[\"-2.5\"\,\"4\"\]\}
{"operation":"multiply","operands":["-2.5","4"],"result":"-10"}
[HTTP 200]

### divide 10 / 4
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"divide\"\,\"operands\":\[\"10\"\,\"4\"\]\}
{"operation":"divide","operands":["10","4"],"result":"2.5"}
[HTTP 200]

### divide 1 / 3
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"divide\"\,\"operands\":\[\"1\"\,\"3\"\]\}
{"operation":"divide","operands":["1","3"],"result":"0.33333333333333333333"}
[HTTP 200]

### power 2 ^ 10
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"power\"\,\"operands\":\[\"2\"\,\"10\"\]\}
{"operation":"power","operands":["2","10"],"result":"1024"}
[HTTP 200]

### power 2 ^ -2
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"power\"\,\"operands\":\[\"2\"\,\"-2\"\]\}
{"operation":"power","operands":["2","-2"],"result":"0.25"}
[HTTP 200]

### sqrt 2
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"sqrt\"\,\"operands\":\[\"2\"\]\}
{"operation":"sqrt","operands":["2"],"result":"1.4142135623730950488"}
[HTTP 200]

### sqrt 144
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"sqrt\"\,\"operands\":\[\"144\"\]\}
{"operation":"sqrt","operands":["144"],"result":"12"}
[HTTP 200]

### percentage 15% of 200
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"percentage\"\,\"operands\":\[\"15\"\,\"200\"\]\}
{"operation":"percentage","operands":["15","200"],"result":"30"}
[HTTP 200]

### DIVISION_BY_ZERO
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"divide\"\,\"operands\":\[\"10\"\,\"0\"\]\}
{"error":{"code":"DIVISION_BY_ZERO","message":"Cannot divide by zero.","requestId":"139e5ba9-8f0e-4707-b316-ed9c0db14ad4"}}
[HTTP 422]

### NEGATIVE_SQUARE_ROOT
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"sqrt\"\,\"operands\":\[\"-4\"\]\}
{"error":{"code":"NEGATIVE_SQUARE_ROOT","message":"Cannot take the square root of a negative number.","requestId":"77f8bdf6-c58a-4ba0-aad1-0e13b2cd2658"}}
[HTTP 422]

### INVALID_EXPONENT
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"power\"\,\"operands\":\[\"2\"\,\"0.5\"\]\}
{"error":{"code":"INVALID_EXPONENT","message":"The exponent must be a whole number.","requestId":"9db8b836-e6db-4946-943f-02b118c352f9","fields":{"operand_index":1}}}
[HTTP 422]

### EXPONENT_TOO_LARGE
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"power\"\,\"operands\":\[\"2\"\,\"1001\"\]\}
{"error":{"code":"EXPONENT_TOO_LARGE","message":"The exponent is outside the allowed range.","requestId":"b40da9a3-4a9e-4017-b017-3ad05f57a651","fields":{"limit":1000,"operand_index":1}}}
[HTTP 422]

### RESULT_TOO_LARGE
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"power\"\,\"operands\":\[\"9\"\,\"1000\"\]\}
{"error":{"code":"RESULT_TOO_LARGE","message":"The result would be too large to calculate.","requestId":"59e22c67-ebac-45c2-aca3-280f742d2b77","fields":{"limit":640}}}
[HTTP 422]

### UNKNOWN_OPERATION
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"modulo\"\,\"operands\":\[\"1\"\,\"2\"\]\}
{"error":{"code":"UNKNOWN_OPERATION","message":"The requested operation is not supported.","requestId":"cf975be3-331f-4534-adc7-34709c3be8ce","fields":{"operation":"modulo"}}}
[HTTP 422]

### INVALID_OPERAND_COUNT
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"add\"\,\"operands\":\[\"1\"\]\}
{"error":{"code":"INVALID_OPERAND_COUNT","message":"The operation received the wrong number of operands.","requestId":"b117774b-72de-476b-921e-819aa2b8a6e8","fields":{"actual":1,"expected":2,"operation":"add"}}}
[HTTP 422]

### INVALID_OPERAND
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"add\"\,\"operands\":\[\"1e5\"\,\"2\"\]\}
{"error":{"code":"INVALID_OPERAND","message":"An operand is not a valid number.","requestId":"cce3b486-af9e-4fd9-b0c8-7ca385ab8e90","fields":{"operand_index":0,"reason":"exponent"}}}
[HTTP 422]

### OPERAND_TOO_LARGE
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"add\"\,\"operands\":\[\"99999999999999999999999999999999999999999999999999999999999999999999999999999999\"\,\"1\"\]\}
{"error":{"code":"OPERAND_TOO_LARGE","message":"An operand is larger than the allowed limit.","requestId":"f5084070-7c0e-40a4-9185-d3a73b56ae96","fields":{"limit":64,"operand_index":0}}}
[HTTP 422]

### VALIDATION_FAILED
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operands\":\[\"1\"\,\"2\"\]\}
{"error":{"code":"VALIDATION_FAILED","message":"The request failed validation.","requestId":"17584436-4624-495a-815e-5c1cb2df9b5d","fields":{"operation":"is required"}}}
[HTTP 422]

### BAD_REQUEST malformed
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":
{"error":{"code":"BAD_REQUEST","message":"The request could not be understood.","requestId":"3ce3b06b-9bac-40e1-bd32-8e2b416e716e"}}
[HTTP 400]

### BAD_REQUEST empty
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d ''
{"error":{"code":"BAD_REQUEST","message":"The request could not be understood.","requestId":"780c9dbd-427d-46e8-b8d8-d01372292daf"}}
[HTTP 400]

### BAD_REQUEST unknown field
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"add\"\,\"operands\":\[\"1\"\,\"2\"\]\,\"extra\":true\}
{"error":{"code":"BAD_REQUEST","message":"The request could not be understood.","requestId":"21b94bf4-8234-42cb-a4ca-02392f1af02f"}}
[HTTP 400]

### BAD_REQUEST trailing data
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"add\"\,\"operands\":\[\"1\"\,\"2\"\]\}\{\}
{"error":{"code":"BAD_REQUEST","message":"The request could not be understood.","requestId":"f4315a84-7b55-48b6-ac68-70b353c24813"}}
[HTTP 400]

### UNSUPPORTED_MEDIA_TYPE
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: text/plain" -d \{\"operation\":\"add\"\,\"operands\":\[\"1\"\,\"2\"\]\}
{"error":{"code":"UNSUPPORTED_MEDIA_TYPE","message":"The content type is not supported.","requestId":"d4f1a0b4-d681-4e5b-8037-6fe33e2d3658"}}
[HTTP 415]

### PAYLOAD_TOO_LARGE
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" --data-binary @1.1MB-body.json
{"error":{"code":"PAYLOAD_TOO_LARGE","message":"The request body is too large.","requestId":"838bf0e4-2942-47a1-a7eb-eaab52de325f","fields":{"limit":1048576}}}
[HTTP 413]

### METHOD_NOT_ALLOWED
$ curl -X GET localhost:8080/api/v1/calculate
{"error":{"code":"METHOD_NOT_ALLOWED","message":"The method is not allowed for this resource.","requestId":"98173eaa-117a-4cb4-ab2a-f53c9947081a","fields":{"method":"GET","path":"/api/v1/calculate"}}}
[HTTP 405]

### NOT_FOUND
$ curl localhost:8080/api/v1/nope
{"error":{"code":"NOT_FOUND","message":"The requested resource was not found.","requestId":"1c82b3b7-3071-48f1-8ae9-6f18931b9149","fields":{"method":"GET","path":"/api/v1/nope"}}}
[HTTP 404]

```
