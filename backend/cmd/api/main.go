// Command api is the calculator HTTP service entry point.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/boot"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/logger"
)

func main() {
	os.Exit(run(context.Background(), os.Stderr))
}

// run builds the app, serves until shutdown and maps the outcome to an exit
// code. It returns the code instead of calling os.Exit so the logic stays
// testable. Wiring lives in internal/boot; this function only translates.
func run(ctx context.Context, stderr io.Writer) int {
	app, err := boot.Initialize(ctx)
	if err != nil {
		return reportStartupFailure(stderr, err)
	}
	code := constants.ExitCodeOK
	if err := app.Run(ctx); err != nil {
		report(stderr, constants.MsgRunFailed, err)
		code = constants.ExitCodeRuntime
	}
	// A flush failure at exit must not turn a clean run into a failure.
	if err := logger.Sync(); err != nil {
		report(stderr, constants.MsgLoggerSyncFailed, err)
	}
	return code
}

// reportStartupFailure picks the message and exit code by the stage that
// failed. The logger may not exist yet, so everything goes to stderr.
func reportStartupFailure(stderr io.Writer, err error) int {
	switch {
	case errors.Is(err, boot.ErrConfigLoad):
		report(stderr, constants.MsgConfigLoadFailed, err)
		return constants.ExitCodeConfigInvalid
	case errors.Is(err, boot.ErrLoggerInit):
		report(stderr, constants.MsgLoggerInitFailed, err)
		return constants.ExitCodeLoggerInvalid
	default:
		report(stderr, constants.MsgRunFailed, err)
		return constants.ExitCodeRuntime
	}
}

func report(stderr io.Writer, format string, err error) {
	// A failed stderr write leaves nothing else to report to, so the error is dropped.
	_, _ = fmt.Fprintf(stderr, format, err)
}
