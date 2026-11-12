package logger

import (
	"errors"
	"syscall"

	"go.uber.org/zap/zapcore"
)

// Log levels per environment. Debug stays off outside dev so production logs
// are not flooded and do not carry developer detail.
const (
	levelProd = zapcore.InfoLevel
	levelTest = zapcore.InfoLevel
	levelDev  = zapcore.DebugLevel
)

// timeKey names the timestamp field in JSON output.
const timeKey = "time"

// ErrFmtBuild wraps a zap build failure so the caller can tell which stage failed.
const ErrFmtBuild = "logger: build: %w"

// ErrFmtUnknownEnv wraps ErrUnknownEnvironment with the offending value.
const ErrFmtUnknownEnv = "%w: %q"

// ErrUnknownEnvironment is returned by Initialize for an environment it has no
// encoder for, so a typo fails loudly instead of silently picking a format.
var ErrUnknownEnvironment = errors.New("logger: unknown environment")

// harmlessSyncErrnos are what Sync reports when stderr is a terminal or pipe
// that cannot be fsynced. They carry no lost data, so they must not fail shutdown.
// EBADF is what macOS returns when fsync is called on a pipe (seen when the
// binary runs with stderr piped, as in CI and shell pipelines).
var harmlessSyncErrnos = []syscall.Errno{syscall.ENOTTY, syscall.EINVAL, syscall.EBADF}
