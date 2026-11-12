package interceptors

import (
	"context"
	"net/http"
	"time"
)

// Timeout gives each request a deadline on its context. http.TimeoutHandler is
// avoided because it replies in plain text; here handlers observe ctx.Done()
// and answer with the normal JSON error themselves.
func Timeout(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
