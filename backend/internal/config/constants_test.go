package config_test

import "time"

// Fixture file contents and expected values are kept apart on purpose: a test
// that derives its expectation from the fixture string would pass even if
// loading silently ignored the file.
const (
	fixtureDefaultToml = `
[server]
port = 8080
read_timeout = "5s"
write_timeout = "10s"
idle_timeout = "60s"
request_timeout = "5s"
shutdown_grace_period = "10s"
max_body_bytes = 1048576

[cors]
allowed_origins = ["http://localhost:5173"]

[calculator]
max_operand_length = 64
max_exponent = 1000
decimal_precision = 20
`
	fixtureDevToml = `
[server]
port = 8081
`
	fixtureTestToml = `
[server]
port = 8082
read_timeout = "2s"

[cors]
allowed_origins = ["http://test.example.com", "http://test2.example.com"]
`
	fixtureProdToml = `
[server]
port = 8083

[calculator]
max_exponent = 500
`
	fixtureMalformedToml    = "[server\nport = "
	fixtureUnknownKeyToml   = "[server]\nprot = 1\n"
	fixtureBadDurationToml  = "[server]\nread_timeout = \"soon\"\n"
	fixtureInvalidValueToml = "[calculator]\nmax_operand_length = 0\n"
)

const (
	fileDefault = "default.toml"
	fileDev     = "dev.toml"
	fileTest    = "test.toml"
	fileProd    = "prod.toml"

	envNameDev     = "dev"
	envNameTest    = "test"
	envNameProd    = "prod"
	envNameUnknown = "staging"

	envVarPort           = "APP_SERVER_PORT"
	envVarReadTimeout    = "APP_SERVER_READ_TIMEOUT"
	envVarMaxBody        = "APP_SERVER_MAX_BODY_BYTES"
	envVarCORSOrigins    = "APP_CORS_ALLOWED_ORIGINS"
	envVarDecimalPrec    = "APP_CALCULATOR_DECIMAL_PRECISION"
	envVarEnvironment    = "APP_ENV"
	envVarConfigDir      = "APP_CONFIG_DIR"
	envValuePort         = "9090"
	envValueZero         = "0"
	envValueReadTimeout  = "7s"
	envValueMaxBody      = "2048"
	envValueDecimalPrec  = "30"
	envValueCORSOrigins  = "https://a.example.com,https://b.example.com"
	envValueCORSSlashed  = "https://a.example.com/,https://b.example.com"
	envValueMissingDir   = "does-not-exist"
	envValueEmpty        = ""
	envValueDefaultNoDir = "config"
)

const (
	wantDefaultPort           = 8080
	wantDevPort               = 8081
	wantTestPort              = 8082
	wantProdPort              = 8083
	wantDefaultWriteTimeout   = 10 * time.Second
	wantDefaultIdleTimeout    = 60 * time.Second
	wantDefaultRequestTimeout = 5 * time.Second
	wantDefaultGracePeriod    = 10 * time.Second
	wantDefaultReadTimeout    = 5 * time.Second
	wantTestReadTimeout       = 2 * time.Second
	wantOverrideReadTimeout   = 7 * time.Second
	wantDefaultMaxBody        = int64(1048576)
	wantOverrideMaxBody       = int64(2048)
	wantOverridePort          = 9090
	wantDefaultMaxOperand     = 64
	wantDefaultMaxExponent    = 1000
	wantProdMaxExponent       = 500
	wantDefaultDecimalPrec    = 20
	wantOverrideDecimalPrec   = 30
	wantDefaultOrigin         = "http://localhost:5173"
	wantTestOriginOne         = "http://test.example.com"
	wantTestOriginTwo         = "http://test2.example.com"
	wantOverrideOriginOne     = "https://a.example.com"
	wantOverrideOriginTwo     = "https://b.example.com"
	wantPortMessage           = "server.port must be between 1 and 65535, got 0"
	wantOperandMessage        = "calculator.max_operand_length must be between 1 and 1000, got 0"
	wantUnknownEnvMessage     = "staging"
	wantExponentMessageFormat = "calculator.max_exponent"
)

// Values for hand-built Config structs passed to Validate.
const (
	validPort             = 8080
	validTimeout          = 5 * time.Second
	validRequestTimeout   = 2 * time.Second
	validMaxBody          = int64(1024)
	validOrigin           = "https://calculator.example.com"
	validOriginWithPort   = "http://localhost:5173"
	validOriginRootSlash  = "https://calculator.example.com/"
	validMaxOperand       = 64
	validMaxExponent      = 1000
	validDecimalPrecision = 20

	invalidPortLow        = 0
	invalidPortHigh       = 65536
	invalidDuration       = time.Duration(0)
	invalidNegDuration    = -time.Second
	invalidMaxBody        = int64(0)
	invalidOperandLow     = 0
	invalidOperandHigh    = 1001
	invalidExponentLow    = 0
	invalidExponentHigh   = 100001
	invalidPrecisionLow   = 0
	invalidPrecisionHigh  = 101
	originWildcard        = "*"
	originWithPath        = "https://calculator.example.com/app"
	originWithoutScheme   = "calculator.example.com"
	originWrongScheme     = "ftp://calculator.example.com"
	originWithQuery       = "https://calculator.example.com?x=1"
	originWithoutHost     = "https://"
	originUnparsable      = "http://[::1"
	originWithCredentials = "https://user:pass@calculator.example.com"
	originWithFragment    = "https://calculator.example.com#frag"
)
