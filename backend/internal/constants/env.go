package constants

// Environment names a deployment target. A named type stops an arbitrary
// string being passed where a validated environment is expected.
type Environment string

const (
	EnvDev  Environment = "dev"
	EnvTest Environment = "test"
	EnvProd Environment = "prod"
)

// DefaultEnv is used when APP_ENV is unset, so a fresh clone runs locally
// with no setup.
const DefaultEnv = EnvDev

// Process environment variable names.
const (
	EnvVarEnvironment = "APP_ENV"
	EnvVarConfigDir   = "APP_CONFIG_DIR"
)

// Overrides use APP_ plus the dotted config key with dots turned into
// underscores, e.g. server.port becomes APP_SERVER_PORT.
const (
	EnvPrefix         = "APP"
	EnvKeyDelimiter   = "."
	EnvKeyReplacement = "_"
)
