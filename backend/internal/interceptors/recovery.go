package interceptors

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/httpx"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/logger"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

// Recovery turns a handler panic into a logged stack and a JSON INTERNAL
// error, so one bad request cannot crash the process or leak a stack trace.
// If the response had already started, the panic is logged and the connection
// is aborted with http.ErrAbortHandler.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := wrapWriter(w)
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			// ErrAbortHandler is net/http's deliberate "drop this connection"
			// signal; swallowing it would break that contract.
			if rec == http.ErrAbortHandler {
				panic(rec)
			}
			handlePanic(rw, r, rec)
		}()
		next.ServeHTTP(rw, r)
	})
}

func handlePanic(rw *responseWriter, r *http.Request, rec any) {
	ctx := r.Context()
	logger.Ctx(ctx).Errorw(constants.LogMsgRequestPanicked,
		constants.LogKeyPanic, fmt.Sprint(rec),
		constants.LogKeyMethod, r.Method,
		constants.LogKeyPath, r.URL.Path,
		constants.LogKeyStack, string(debug.Stack()),
	)
	// Once the header is out the status is fixed and a second body would
	// corrupt the first. Aborting the connection makes the client see a failed
	// response instead of a truncated body that looks complete.
	if rw.wroteHeader {
		panic(http.ErrAbortHandler)
	}
	httpx.WriteError(ctx, rw, apperror.New(apperror.CodeInternal, fmt.Errorf(panicMessageFmt, rec)))
}
