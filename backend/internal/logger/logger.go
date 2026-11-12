// Package logger builds the service's structured zap logger and hands out
// request-scoped copies of it.
package logger

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/contextkeys"
)

// base is the process-wide logger. An atomic pointer keeps Ctx safe to call
// from request goroutines while Initialize or a test swaps the value.
var base atomic.Pointer[zap.SugaredLogger]

// Initialize builds the base logger for env: JSON in prod and test, colourless
// console output in dev. An unknown env is an error and leaves the previous
// logger untouched.
func Initialize(env constants.Environment) error {
	cfg, err := configFor(env)
	if err != nil {
		return err
	}
	built, err := cfg.Build()
	if err != nil {
		return fmt.Errorf("logger: build: %w", err)
	}
	base.Store(built.Sugar())
	return nil
}

func configFor(env constants.Environment) (zap.Config, error) {
	switch env {
	case constants.EnvProd:
		return jsonConfig(levelProd), nil
	case constants.EnvTest:
		return jsonConfig(levelTest), nil
	case constants.EnvDev:
		return consoleConfig(levelDev), nil
	default:
		return zap.Config{}, fmt.Errorf(ErrFmtUnknownEnv, ErrUnknownEnvironment, string(env))
	}
}

func jsonConfig(level zapcore.Level) zap.Config {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(level)
	cfg.EncoderConfig.TimeKey = timeKey
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return cfg
}

func consoleConfig(level zapcore.Level) zap.Config {
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(level)
	// Colours corrupt output when it is piped to a file or CI log.
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return cfg
}

// Ctx returns the base logger, with request_id attached when ctx carries one.
// Before Initialize it returns a no-op logger rather than nil, so early code
// paths and tests can log without a nil-pointer panic.
func Ctx(ctx context.Context) *zap.SugaredLogger {
	l := current()
	if ctx == nil {
		return l
	}
	if id, ok := contextkeys.RequestID(ctx); ok && id != "" {
		return l.With(constants.LogKeyRequestID, id)
	}
	return l
}

func current() *zap.SugaredLogger {
	if l := base.Load(); l != nil {
		return l
	}
	return zap.NewNop().Sugar()
}

// Sync flushes buffered entries; call it once at shutdown. The "cannot sync a
// terminal or pipe" error stderr always returns is dropped because nothing is
// lost; any other error is returned.
func Sync() error {
	err := current().Sync()
	if err == nil || isHarmlessSyncError(err) {
		return nil
	}
	return err
}

func isHarmlessSyncError(err error) bool {
	for _, errno := range harmlessSyncErrnos {
		if errors.Is(err, errno) {
			return true
		}
	}
	return false
}

// SetForTest swaps the base logger and returns a function that restores the
// previous one. It exists so tests can capture output with an observer core.
func SetForTest(l *zap.SugaredLogger) (restore func()) {
	prev := base.Swap(l)
	return func() { base.Store(prev) }
}
