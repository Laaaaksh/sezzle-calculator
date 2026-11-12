package boot_test

import "time"

const (
	testOrigin        = "http://localhost:5173"
	testOtherOrigin   = "http://evil.example.com"
	testRequestID     = "client-id-7"
	testMissingPath   = "/nope"
	testAPIPath       = "/api/v1/missing"
	testLivePath      = "/health/live"
	testPingRoute     = "/ping"
	testAPIPingPath   = "/api/v1/ping"
	testPingBody      = "pong"
	testReadyPath     = "/health/ready"
	testBodyLimit     = int64(16)
	testOversizedBody = "this body is longer than sixteen bytes"
	testTimeout       = 3 * time.Second

	headerOrigin = "Origin"
	headerACAO   = "Access-Control-Allow-Origin"
	headerAllow  = "Allow"

	wantNotFoundCode   = "NOT_FOUND"
	wantMethodCode     = "METHOD_NOT_ALLOWED"
	wantPingAllow      = "GET"
	wantLiveAllow      = "GET"
	wantTooLargeCode   = "PAYLOAD_TOO_LARGE"
	wantContentType    = "application/json"
	wantLiveBody       = `{"status":"ok"}`
	wantReadyBody      = `{"status":"ready"}`
	wantHeaderTemplate = "X-Request-ID"
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
