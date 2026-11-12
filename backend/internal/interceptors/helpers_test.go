package interceptors_test

import (
	"net/http"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/logger"
)

// captureLogs swaps the base logger for an in-memory observer and returns the
// observed entries plus a function that restores the previous logger.
func captureLogs() (*observer.ObservedLogs, func()) {
	core, observed := observer.New(zapcore.DebugLevel)
	return observed, logger.SetForTest(zap.New(core).Sugar())
}

// okHandler answers 200 with no body.
var okHandler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
})
