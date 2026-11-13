package interceptors

import (
	"context"
	"net/http"
	"time"
)

// Timeout gives each request a deadline on its context. It stops nothing itself: the core and the looping operations
// check ctx and return REQUEST_TIMEOUT (503, JSON). http.TimeoutHandler is avoided because it replies in plain text.
// A new handler that does long work must check ctx the same way, or this deadline is decoration.
func Timeout(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
