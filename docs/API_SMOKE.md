# API smoke test

Real requests against a running server (`make -C backend run`), captured with curl. Request IDs vary per run.

Two codes are not in this capture because a client request cannot cause them. `INTERNAL` needs a server fault, such as a recovered panic. `SERVICE_UNAVAILABLE` only comes from `GET /health/ready` while the server is shutting down. Both are covered by the handler and router tests instead.

The `PAYLOAD_TOO_LARGE` body below shows the current `fields` shape (`limit`, the body cap in bytes). The router test `TestOversizedBodyIsRejectedBeforeRouting` checks the same body.

```text
### operations
$ curl localhost:8080/api/v1/operations
{"operations":[{"name":"add","arity":2,"symbol":"+","description":"Adds the two operands."},{"name":"subtract","arity":2,"symbol":"-","description":"Subtracts the second operand from the first."},{"name":"multiply","arity":2,"symbol":"*","description":"Multiplies the two operands."},{"name":"divide","arity":2,"symbol":"/","description":"Divides the first operand by the second."},{"name":"power","arity":2,"symbol":"^","description":"Raises the first operand to the whole-number power given by the second."},{"name":"sqrt","arity":1,"symbol":"√","description":"Returns the square root of the operand."},{"name":"percentage","arity":2,"symbol":"%","description":"Returns the first operand percent of the second."}],"limits":{"maxOperandLength":64,"maxExponent":1000,"decimalPrecision":20,"maxResultDigits":640}}
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
{"error":{"code":"DIVISION_BY_ZERO","message":"Cannot divide by zero.","requestId":"19ef11b0-ac55-4672-8f49-22a3f490d133"}}
[HTTP 422]

### NEGATIVE_SQUARE_ROOT
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"sqrt\"\,\"operands\":\[\"-4\"\]\}
{"error":{"code":"NEGATIVE_SQUARE_ROOT","message":"Cannot take the square root of a negative number.","requestId":"f67a87e8-50f4-4a32-87ea-b3160a337d97"}}
[HTTP 422]

### INVALID_EXPONENT
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"power\"\,\"operands\":\[\"2\"\,\"0.5\"\]\}
{"error":{"code":"INVALID_EXPONENT","message":"The exponent must be a whole number.","requestId":"82d9be0c-8456-4a0d-9af4-4a79fc87a8be","fields":{"operand_index":1}}}
[HTTP 422]

### EXPONENT_TOO_LARGE
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"power\"\,\"operands\":\[\"2\"\,\"1001\"\]\}
{"error":{"code":"EXPONENT_TOO_LARGE","message":"The exponent is outside the allowed range.","requestId":"c522e39c-4fc5-41b5-8386-51924c96abfa","fields":{"limit":1000,"operand_index":1}}}
[HTTP 422]

### RESULT_TOO_LARGE
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"power\"\,\"operands\":\[\"9\"\,\"1000\"\]\}
{"error":{"code":"RESULT_TOO_LARGE","message":"The result would be too large to calculate.","requestId":"9244fd95-4cbc-4515-9c0a-90eb59715173","fields":{"limit":640}}}
[HTTP 422]

### UNKNOWN_OPERATION
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"modulo\"\,\"operands\":\[\"1\"\,\"2\"\]\}
{"error":{"code":"UNKNOWN_OPERATION","message":"The requested operation is not supported.","requestId":"73dce24f-d6a5-4148-8fb3-d3f49cdb6a99","fields":{"operation":"modulo"}}}
[HTTP 422]

### INVALID_OPERAND_COUNT
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"add\"\,\"operands\":\[\"1\"\]\}
{"error":{"code":"INVALID_OPERAND_COUNT","message":"The operation received the wrong number of operands.","requestId":"74da693e-8b85-473b-a126-4a4cc516fd7e","fields":{"actual":1,"expected":2,"operation":"add"}}}
[HTTP 422]

### INVALID_OPERAND
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"add\"\,\"operands\":\[\"1e5\"\,\"2\"\]\}
{"error":{"code":"INVALID_OPERAND","message":"An operand is not a valid number.","requestId":"c5c75ee2-52e9-4a56-917f-b3a4016b4299","fields":{"operand_index":0,"reason":"exponent"}}}
[HTTP 422]

### OPERAND_TOO_LARGE
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"add\"\,\"operands\":\[\"99999999999999999999999999999999999999999999999999999999999999999999999999999999\"\,\"1\"\]\}
{"error":{"code":"OPERAND_TOO_LARGE","message":"An operand is larger than the allowed limit.","requestId":"c849990c-7b7e-4409-92e2-8df670078dd0","fields":{"limit":64,"operand_index":0}}}
[HTTP 422]

### VALIDATION_FAILED
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operands\":\[\"1\"\,\"2\"\]\}
{"error":{"code":"VALIDATION_FAILED","message":"The request failed validation.","requestId":"ed551bc6-8f3d-477b-846e-c8661422697c","fields":{"operation":"is required"}}}
[HTTP 422]

### BAD_REQUEST malformed
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":
{"error":{"code":"BAD_REQUEST","message":"The request could not be understood.","requestId":"9549cc8e-2851-4aa5-8622-3e1840642f9e"}}
[HTTP 400]

### BAD_REQUEST empty
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d ''
{"error":{"code":"BAD_REQUEST","message":"The request could not be understood.","requestId":"d41819a5-6822-496e-a525-eae3ba33102a"}}
[HTTP 400]

### BAD_REQUEST unknown field
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"add\"\,\"operands\":\[\"1\"\,\"2\"\]\,\"extra\":true\}
{"error":{"code":"BAD_REQUEST","message":"The request could not be understood.","requestId":"16d4def5-fee9-42fc-b8e1-435eed2d8ca1"}}
[HTTP 400]

### BAD_REQUEST trailing data
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"add\"\,\"operands\":\[\"1\"\,\"2\"\]\}\{\}
{"error":{"code":"BAD_REQUEST","message":"The request could not be understood.","requestId":"91d73c5e-6cb4-472f-828e-635ee2f7f2c3"}}
[HTTP 400]

### BAD_REQUEST case-variant key
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"add\"\,\"operands\":\[\"1\"\,\"2\"\]\,\"OPERANDS\":\[\"9\"\,\"9\"\]\}
{"error":{"code":"BAD_REQUEST","message":"The request could not be understood.","requestId":"f2920dad-6fb6-457f-a23c-3f6df63887b4","fields":{"key":"OPERANDS","reason":"wrong_case"}}}
[HTTP 400]

### VALIDATION_FAILED too many operands
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d \{\"operation\":\"add\"\,\"operands\":\[\"1\"\,\"2\"\,\"3\"\,\"4\"\,\"5\"\,\"6\"\,\"7\"\,\"8\"\,\"9\"\,\"10\"\,\"11\"\,\"12\"\,\"13\"\,\"14\"\,\"15\"\,\"16\"\,\"17\"\]\}
{"error":{"code":"VALIDATION_FAILED","message":"The request failed validation.","requestId":"7882a092-c476-40f8-8f66-6e03ac726183","fields":{"operands":"too many"}}}
[HTTP 422]

### UNSUPPORTED_MEDIA_TYPE
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: text/plain" -d \{\"operation\":\"add\"\,\"operands\":\[\"1\"\,\"2\"\]\}
{"error":{"code":"UNSUPPORTED_MEDIA_TYPE","message":"The content type is not supported.","requestId":"551ac0f2-9198-4332-af06-cd61cc8be4dc"}}
[HTTP 415]

### PAYLOAD_TOO_LARGE
$ curl -X POST localhost:8080/api/v1/calculate -H "Content-Type: application/json" --data-binary @1.1MB-body.json
{"error":{"code":"PAYLOAD_TOO_LARGE","message":"The request body is too large.","requestId":"a556c104-e11f-41de-a52e-04a5f672155e","fields":{"limit":1048576}}}
[HTTP 413]

### METHOD_NOT_ALLOWED
$ curl -X GET localhost:8080/api/v1/calculate
{"error":{"code":"METHOD_NOT_ALLOWED","message":"The method is not allowed for this resource.","requestId":"a5ae0868-af89-4202-b43d-56bbcfd054cf","fields":{"method":"GET","path":"/api/v1/calculate"}}}
[HTTP 405]

### NOT_FOUND
$ curl localhost:8080/api/v1/nope
{"error":{"code":"NOT_FOUND","message":"The requested resource was not found.","requestId":"7d30f8d7-b146-4439-9202-8cb9cd282346","fields":{"method":"GET","path":"/api/v1/nope"}}}
[HTTP 404]

```
