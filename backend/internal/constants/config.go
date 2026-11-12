package constants

// Config file layout.
const (
	DefaultConfigDir  = "config"
	ConfigFileType    = "toml"
	ConfigBaseName    = "default"
	ConfigFileNameFmt = "%s.toml"
)

// Config keys, spelled exactly as they appear in the TOML files. Validation
// messages name these so an operator can find the offending line.
const (
	KeyServerPort                = "server.port"
	KeyServerReadTimeout         = "server.read_timeout"
	KeyServerWriteTimeout        = "server.write_timeout"
	KeyServerIdleTimeout         = "server.idle_timeout"
	KeyServerRequestTimeout      = "server.request_timeout"
	KeyServerShutdownGracePeriod = "server.shutdown_grace_period"
	KeyServerMaxBodyBytes        = "server.max_body_bytes"
	KeyCORSAllowedOrigins        = "cors.allowed_origins"
	KeyCalculatorMaxOperandLen   = "calculator.max_operand_length"
	KeyCalculatorMaxExponent     = "calculator.max_exponent"
	KeyCalculatorDecimalPrec     = "calculator.decimal_precision"
)

// Validation bounds. They are hard ceilings that no environment file may
// exceed, which keeps a typo in prod.toml from disabling a safety limit.
const (
	PortMin = 1
	PortMax = 65535

	OperandLengthMin = 1
	OperandLengthMax = 1000

	ExponentMin = 1
	ExponentMax = 100000

	DecimalPrecisionMin = 1
	DecimalPrecisionMax = 100
)

// CORS origin rules.
const (
	CORSWildcard = "*"
	SchemeHTTP   = "http"
	SchemeHTTPS  = "https"
	// OriginPathRoot is the only path tolerated on an origin, because
	// url.Parse reports "http://host/" with a single slash.
	OriginPathRoot = "/"
)

// Error message formats for config loading and validation.
const (
	ErrFmtRange      = "%s must be between %d and %d"
	ErrFmtPositive   = "%s must be greater than zero"
	ErrFmtBelow      = "%s must be less than %s"
	ErrFmtRequired   = "%s must contain at least one origin"
	ErrFmtWildcard   = "%s must not contain the wildcard origin"
	ErrFmtOrigin     = "%s entry must be an absolute http(s) origin without path"
	ErrFmtGot        = "%w, got %v"
	ErrFmtUnknownEnv = "%w: %q (allowed: dev, test, prod)"
	ErrFmtReadFile   = "read config file %s: %w"
	ErrFmtUnmarshal  = "decode config: %w"
	ErrFmtInvalid    = "invalid config: %w"
	ErrMsgUnknownEnv = "unknown environment"
)

// Error message formats for running the server.
const (
	ErrFmtListen   = "listen on %s: %w"
	ErrFmtServe    = "serve http: %w"
	ErrFmtShutdown = "graceful shutdown: %w"
)
