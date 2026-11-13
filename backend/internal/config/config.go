// Package config loads, layers and validates the service configuration.
package config

import (
	"strings"
	"time"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
)

// Config is the typed, validated configuration read once at boot. Code reads
// fields from it directly instead of fetching keys by string at the point of use.
type Config struct {
	App        AppConfig        `mapstructure:"app"`
	Server     ServerConfig     `mapstructure:"server"`
	CORS       CORSConfig       `mapstructure:"cors"`
	Calculator CalculatorConfig `mapstructure:"calculator"`
	RateLimit  RateLimitConfig  `mapstructure:"rate_limit"`
}

// AppConfig describes the running deployment. Env is derived from APP_ENV
// rather than a file so a file can never claim to be a different environment.
type AppConfig struct {
	Env constants.Environment `mapstructure:"env"`
}

// ServerConfig holds HTTP server limits. Timeouts guard against slow clients
// holding connections open; RequestTimeout bounds one handler via its context; MaxBodyBytes caps what a single request can cost.
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

// normalize strips the trailing "/" from each origin. Browsers send Origin
// without one and the CORS middleware compares strings exactly, so
// "https://a.com/" would otherwise pass validation yet never match.
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

// RateLimitConfig sets the per-client limit on POST /api/v1/calculate. Each
// client IP gets a token bucket of Burst tokens that refills at RequestsPerSecond.
// ClientIPHeader names a header (for example X-Real-IP) that holds the client
// address when the service sits behind a proxy. Leave it empty when clients
// connect directly, and set it only when the proxy overwrites the header, or a
// caller could pick its own bucket.
type RateLimitConfig struct {
	Enabled           bool    `mapstructure:"enabled"`
	RequestsPerSecond float64 `mapstructure:"requests_per_second"`
	Burst             int     `mapstructure:"burst"`
	ClientIPHeader    string  `mapstructure:"client_ip_header"`
}
