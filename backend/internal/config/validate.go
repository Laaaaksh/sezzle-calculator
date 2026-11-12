package config

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
)

// Validate checks every rule and reports all violations together, so an
// operator fixes a broken file in one pass instead of one restart per mistake.
func (c Config) Validate() error {
	var errs []error
	errs = append(errs, c.Server.validate()...)
	errs = append(errs, c.CORS.validate()...)
	errs = append(errs, c.Calculator.validate()...)
	return errors.Join(errs...)
}

func (s ServerConfig) validate() []error {
	var errs []error
	if s.Port < constants.PortMin || s.Port > constants.PortMax {
		errs = append(errs, withValue(ErrPortOutOfRange, s.Port))
	}
	errs = appendIfNotPositive(errs, s.ReadTimeout, ErrReadTimeoutNotPositive)
	errs = appendIfNotPositive(errs, s.WriteTimeout, ErrWriteTimeoutNotPositive)
	errs = appendIfNotPositive(errs, s.IdleTimeout, ErrIdleTimeoutNotPositive)
	errs = appendIfNotPositive(errs, s.RequestTimeout, ErrRequestTimeoutNotPositive)
	// Only compare when both are valid, so one typo is not reported twice.
	if s.RequestTimeout > 0 && s.WriteTimeout > 0 && s.RequestTimeout >= s.WriteTimeout {
		errs = append(errs, withValue(ErrRequestTimeoutNotBelowWrite, s.RequestTimeout))
	}
	errs = appendIfNotPositive(errs, s.ShutdownGracePeriod, ErrShutdownGraceNotPositive)
	if s.MaxBodyBytes <= 0 {
		errs = append(errs, withValue(ErrMaxBodyNotPositive, s.MaxBodyBytes))
	}
	return errs
}

func (c CORSConfig) validate() []error {
	if len(c.AllowedOrigins) == 0 {
		return []error{ErrCORSOriginsEmpty}
	}
	var errs []error
	for _, origin := range c.AllowedOrigins {
		if err := validateOrigin(origin); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (c CalculatorConfig) validate() []error {
	var errs []error
	if c.MaxOperandLength < constants.OperandLengthMin || c.MaxOperandLength > constants.OperandLengthMax {
		errs = append(errs, withValue(ErrOperandLengthOutOfRange, c.MaxOperandLength))
	}
	if c.MaxExponent < constants.ExponentMin || c.MaxExponent > constants.ExponentMax {
		errs = append(errs, withValue(ErrExponentOutOfRange, c.MaxExponent))
	}
	if c.DecimalPrecision < constants.DecimalPrecisionMin || c.DecimalPrecision > constants.DecimalPrecisionMax {
		errs = append(errs, withValue(ErrDecimalPrecisionOutOfRange, c.DecimalPrecision))
	}
	return errs
}

// validateOrigin rejects "*" explicitly: a wildcard origin would let any site
// call the API from a visitor's browser, so it gets its own clear message.
func validateOrigin(origin string) error {
	if origin == constants.CORSWildcard {
		return withValue(ErrCORSWildcard, origin)
	}
	parsed, err := url.Parse(origin)
	if err != nil {
		return withValue(ErrCORSOriginInvalid, origin)
	}
	if parsed.Scheme != constants.SchemeHTTP && parsed.Scheme != constants.SchemeHTTPS {
		return withValue(ErrCORSOriginInvalid, origin)
	}
	if parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return withValue(ErrCORSOriginInvalid, origin)
	}
	if parsed.Path != "" && parsed.Path != constants.OriginPathRoot {
		return withValue(ErrCORSOriginInvalid, origin)
	}
	return nil
}

func appendIfNotPositive(errs []error, value time.Duration, rule error) []error {
	if value > 0 {
		return errs
	}
	return append(errs, withValue(rule, value))
}

func withValue(rule error, got any) error {
	return fmt.Errorf(constants.ErrFmtGot, rule, got)
}
