package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
)

// Load builds the configuration from default.toml, layers the file for the
// active environment on top, applies APP_ environment variable overrides and
// validates the result. Precedence, lowest to highest: default file,
// environment file, environment variables.
//
// Viper's default decode hooks turn "5s" into a time.Duration and a
// comma-separated string into a slice, which is what lets APP_CORS_ALLOWED_ORIGINS
// carry a list.
func Load() (*Config, error) {
	env, err := resolveEnvironment(os.Getenv(constants.EnvVarEnvironment))
	if err != nil {
		return nil, err
	}

	v := viper.New()
	v.SetConfigType(constants.ConfigFileType)
	dir := resolveConfigDir(os.Getenv(constants.EnvVarConfigDir))
	if err := readLayers(v, dir, env); err != nil {
		return nil, err
	}

	// Env binding must be configured after the files are read so viper knows
	// every key; it only maps variables onto keys it has already seen.
	v.SetEnvPrefix(constants.EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(constants.EnvKeyDelimiter, constants.EnvKeyReplacement))
	v.AutomaticEnv()

	var cfg Config
	// UnmarshalExact turns a misspelled key in a TOML file into a startup
	// error instead of a silently ignored setting.
	if err := v.UnmarshalExact(&cfg); err != nil {
		return nil, fmt.Errorf(constants.ErrFmtUnmarshal, err)
	}
	cfg.App.Env = env

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf(constants.ErrFmtInvalid, err)
	}
	return &cfg, nil
}

// resolveEnvironment treats an empty value as unset so that exporting an
// empty APP_ENV falls back to the default instead of failing.
func resolveEnvironment(raw string) (constants.Environment, error) {
	if raw == "" {
		return constants.DefaultEnv, nil
	}
	switch env := constants.Environment(raw); env {
	case constants.EnvDev, constants.EnvTest, constants.EnvProd:
		return env, nil
	default:
		return "", fmt.Errorf(constants.ErrFmtUnknownEnv, ErrUnknownEnvironment, raw)
	}
}

func resolveConfigDir(raw string) string {
	if raw == "" {
		return constants.DefaultConfigDir
	}
	return raw
}

func readLayers(v *viper.Viper, dir string, env constants.Environment) error {
	basePath := configFilePath(dir, constants.ConfigBaseName)
	v.SetConfigFile(basePath)
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf(constants.ErrFmtReadFile, basePath, err)
	}

	envPath := configFilePath(dir, string(env))
	v.SetConfigFile(envPath)
	if err := v.MergeInConfig(); err != nil {
		return fmt.Errorf(constants.ErrFmtReadFile, envPath, err)
	}
	return nil
}

func configFilePath(dir, name string) string {
	return filepath.Join(dir, fmt.Sprintf(constants.ConfigFileNameFmt, name))
}
