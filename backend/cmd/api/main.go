// Command api is the calculator HTTP service entry point.
package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/config"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/logger"
)

func main() {
	os.Exit(run(os.Stderr))
}

// run returns an exit code instead of calling os.Exit so the failure path
// stays testable.
func run(stderr io.Writer) int {
	cfg, err := config.Load()
	if err != nil {
		// A failed stderr write leaves nothing else to report to, so the error is dropped.
		_, _ = fmt.Fprintf(stderr, constants.MsgConfigLoadFailed, err)
		return constants.ExitCodeConfigInvalid
	}
	if err := logger.Initialize(cfg.App.Env); err != nil {
		_, _ = fmt.Fprintf(stderr, constants.MsgLoggerInitFailed, err)
		return constants.ExitCodeLoggerInvalid
	}
	logger.Ctx(context.Background()).Infow(constants.LogMsgConfigLoaded,
		constants.LogKeyEnv, string(cfg.App.Env),
		constants.LogKeyPort, cfg.Server.Port,
	)
	// A flush failure at exit must not turn a clean run into a failure.
	if err := logger.Sync(); err != nil {
		_, _ = fmt.Fprintf(stderr, constants.MsgLoggerSyncFailed, err)
	}
	return constants.ExitCodeOK
}
