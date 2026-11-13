// Package config loads, layers and validates the service configuration.
package config

import (
	"strings"
	"time"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
)

// Config is the typed, validated configuration, read once at boot.
type Config struct {
	App        AppConfig        `mapstructure:"app"`
	Server     ServerConfig     `mapstructure:"server"`
	CORS       CORSConfig       `mapstructure:"cors"`
	Calculator CalculatorConfig `mapstructure:"calculator"`
	RateLimit  RateLimitConfig  `mapstructure:"rate_limit"`
}

// AppConfig describes the deployment. Env comes from APP_ENV, not a file, so a file cannot claim another environment.
type AppConfig struct {
	Env constants.Environment `mapstructure:"env"`
}

// ServerConfig holds HTTP server limits: timeouts guard against slow clients, MaxBodyBytes caps what one request can cost.
type ServerConfig struct {
	Port                int           `mapstructure:"port"`
	ReadTimeout         time.Duration `mapstructure:"read_timeout"`
	WriteTimeout        time.Duration `mapstructure:"write_timeout"`
	IdleTimeout         time.Duration `mapstructure:"idle_timeout"`
	RequestTimeout      time.Duration `mapstructure:"request_timeout"`
	ShutdownGracePeriod time.Duration `mapstructure:"shutdown_grace_period"`
	// ShutdownDrainDelay keeps serving after readiness turns 503, so a load
	// balancer polling /health/ready can stop routing before the listener closes.
	ShutdownDrainDelay time.Duration `mapstructure:"shutdown_drain_delay"`
	MaxBodyBytes       int64         `mapstructure:"max_body_bytes"`
}

// CORSConfig lists the browser origins allowed to call the API.
type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

// normalize strips the trailing "/" from each origin: browsers send Origin without one and CORS compares strings
// exactly, so "https://a.com/" would pass validation yet never match.
func (c *CORSConfig) normalize() {
	for i, origin := range c.AllowedOrigins {
		c.AllowedOrigins[i] = strings.TrimSuffix(origin, constants.OriginPathRoot)
	}
}

// CalculatorConfig bounds the work one calculation may demand, so a single
// request cannot pin a CPU with a huge exponent or operand.
type CalculatorConfig struct {
	MaxOperandLength int `mapstructure:"max_operand_length"`
	MaxExponent      int `mapstructure:"max_exponent"`
	DecimalPrecision int `mapstructure:"decimal_precision"`
}

// RateLimitConfig sets the per-client token bucket on POST /api/v1/calculate. ClientIPHeader (for example X-Real-IP)
// names the proxy header holding the client address. Leave it empty for direct connections, and set it only when the
// proxy overwrites the header, or a caller could pick its own bucket.
type RateLimitConfig struct {
	Enabled           bool    `mapstructure:"enabled"`
	RequestsPerSecond float64 `mapstructure:"requests_per_second"`
	Burst             int     `mapstructure:"burst"`
	ClientIPHeader    string  `mapstructure:"client_ip_header"`
}
