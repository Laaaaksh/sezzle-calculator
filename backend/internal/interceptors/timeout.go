package interceptors

import (
	"context"
	"net/http"
	"time"
)

// Timeout gives each request a deadline on its context. It does not stop
// anything by itself: it only makes the deadline visible. The calculator core
// and the looping operations (power, square root) check the context and
// return REQUEST_TIMEOUT (503, JSON), so the work stops and the client gets the
// normal error shape. http.TimeoutHandler is avoided because it replies in
// plain text. A new handler that does long work must check ctx the same way,
// or this deadline is decoration for it.
func Timeout(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
