package boot_test

import (
	"time"

	"github.com/shopspring/decimal"
)

const (
	wantOperationsBodyStart = `{"operations":[{"name":"add","arity":2`

	testOrigin         = "http://localhost:5173"
	testOtherOrigin    = "http://evil.example.com"
	testRequestID      = "client-id-7"
	testMissingPath    = "/nope"
	testAPIPath        = "/api/v1/missing"
	testLivePath       = "/health/live"
	testReadyPath      = "/health/ready"
	testCalculatePath  = "/api/v1/calculate"
	testCalculateBare  = "/calculate"
	testOperationsPath = "/api/v1/operations"
	testAddBody        = `{"operation":"add","operands":["1","2"]}`
	testAddResult      = "3"
	testBodyLimit      = int64(64)
	testOversizedBody  = "this body is longer than sixty-four bytes: " + "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	testTimeout        = 3 * time.Second

	headerOrigin = "Origin"
	headerACAO   = "Access-Control-Allow-Origin"
	headerAllow  = "Allow"

	wantNotFoundCode     = "NOT_FOUND"
	wantMethodCode       = "METHOD_NOT_ALLOWED"
	wantOperationsAllow  = "GET"
	wantLiveAllow        = "GET"
	wantTooLargeCode     = "PAYLOAD_TOO_LARGE"
	wantContentType      = "application/json"
	wantLiveBody         = `{"status":"ok"}`
	wantReadyBody        = `{"status":"ready"}`
	wantAddResponse      = `{"operation":"add","operands":["1","2"],"result":"3"}`
	wantNoOperationsBody = `{"operations":[],"limits":{"maxOperandLength":64,"maxExponent":0,"decimalPrecision":0,"maxResultDigits":0}}`
	wantHeaderTemplate   = "X-Request-ID"
)

// Config fixtures for Initialize. Ports are placeholders: Initialize builds the
// server but does not bind it.
const (
	fixtureDefaultToml = `
[server]
port = 8080
read_timeout = "5s"
write_timeout = "10s"
idle_timeout = "60s"
request_timeout = "5s"
shutdown_grace_period = "2s"
max_body_bytes = 1048576

[cors]
allowed_origins = ["http://localhost:5173"]

[calculator]
max_operand_length = 64
max_exponent = 1000
decimal_precision = 20
`
	fixtureTestToml = "# no overrides\n"
	fixtureBadToml  = "[server\nport = "

	fileDefault = "default.toml"
	fileTest    = "test.toml"

	envVarEnv       = "APP_ENV"
	envVarConfigDir = "APP_CONFIG_DIR"
	envVarPort      = "APP_SERVER_PORT"
	envNameTest     = "test"
	envValueEmpty   = ""
	missingDirName  = "does-not-exist"
)

func testAddOperands() []decimal.Decimal {
	return []decimal.Decimal{decimal.NewFromInt(1), decimal.NewFromInt(2)}
}

const (
	e2eMaxOperandLength   = 64
	e2eMaxExponent        = 1000
	e2ePrecision          = 20
	e2eBodyLimit          = int64(1048576)
	e2eExpectedFieldCount = 2

	e2eNoStore   = "no-store"
	e2eTextPlain = "text/plain"

	e2eBodyPointOnePlusPointTwo = `{"operation":"add","operands":["0.1","0.2"]}`
	e2eWantPointThree           = `{"operation":"add","operands":["0.1","0.2"],"result":"0.3"}`
	e2eBodyDivide               = `{"operation":"divide","operands":["10","4"]}`
	e2eWantDivide               = `{"operation":"divide","operands":["10","4"],"result":"2.5"}`
	e2eBodyDivideByZero         = `{"operation":"divide","operands":["10","0"]}`
	e2eBodyUnknownField         = `{"operation":"add","operands":["1","2"],"extra":1}`
	e2eBodyUnknownOperation     = `{"operation":"modulo","operands":["1","2"]}`
	e2eUnknownOperationName     = "modulo"
	e2eHugeOperationLength      = 1 << 19
	e2eSmallErrorBodyBytes      = 1 << 10
	e2eBodyEmptyObject          = `{}`
	e2eBodyTooManyOperands      = `{"operation":"add","operands":["1","2","3","4","5","6","7","8","9","10","11","12","13","14","15","16","17"]}`
	e2eBodyThreeOperands        = `{"operation":"add","operands":["1","2","3"]}`
	e2eBodyCaseVariantKey       = `{"operation":"add","operands":["1","2"],"OPERANDS":["9","9"]}`
	e2eWantLimitsJSON           = `{"maxOperandLength":64,"maxExponent":1000,"decimalPrecision":20,"maxResultDigits":640}`
	e2eDuplicateKeyField        = "key"
	e2eLimitsKey                = "limits"
)
