package interceptors

import (
	"net/http"
	"time"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/logger"
)

// AccessLog writes one structured line per request: method, path, status and
// duration. The request ID is attached by the logger from the context. Only
// the path is logged, never the query string, which may carry user data.
func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := wrapWriter(w)
		// A panic skips the normal return, so it is caught here only to record
		// status 500 and is then re-raised for Recovery, which sits outside.
		// If the header already went out, the client got that status instead.
		defer func() {
			rec := recover()
			if rec != nil && !rw.wroteHeader {
				rw.status = http.StatusInternalServerError
			}
			logRequest(r, rw.status, time.Since(start))
			if rec != nil {
				panic(rec)
			}
		}()
		next.ServeHTTP(rw, r)
	})
}

func logRequest(r *http.Request, status int, elapsed time.Duration) {
	logger.Ctx(r.Context()).Infow(constants.LogMsgRequestCompleted,
		constants.LogKeyMethod, r.Method,
		constants.LogKeyPath, r.URL.Path,
		constants.LogKeyStatus, status,
		constants.LogKeyDurationMS, float64(elapsed)/constants.MillisecondsPerDuration,
	)
}
