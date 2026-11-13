package interceptors

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/httpx"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

//go:generate go tool mockgen -source=rate_limit.go -destination=mock/mock_rate_limit.go -package=mock

// RateLimiter decides whether one more request from key may proceed. When it
// may not, it also says how long until one can.
type RateLimiter interface {
	Allow(key string) (bool, time.Duration)
}

// RateLimit answers 429 RATE_LIMITED with a Retry-After header when the client
// has used up its bucket. clientIPHeader, when set, names the proxy header
// that carries the client address; otherwise the connection address is used.
func RateLimit(limiter RateLimiter, clientIPHeader string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowed, wait := limiter.Allow(clientKey(r, clientIPHeader))
			if allowed {
				next.ServeHTTP(w, r)
				return
			}
			seconds := retryAfterSeconds(wait)
			w.Header().Set(constants.HeaderRetryAfter, strconv.Itoa(seconds))
			httpx.WriteError(r.Context(), w, apperror.New(apperror.CodeRateLimited, nil).
				WithField(apperror.FieldRetryAfterSeconds, seconds))
		})
	}
}

// retryAfterSeconds rounds up, because Retry-After takes whole seconds and an
// early retry would be refused again.
func retryAfterSeconds(wait time.Duration) int {
	return max(constants.RetryAfterMinSeconds, int(math.Ceil(wait.Seconds())))
}

// clientKey takes the first address of the configured proxy header, then falls
// back to the connection's host so a missing header never lets a client skip
// the limit. Only point the header at one the proxy overwrites (X-Real-IP): the
// first X-Forwarded-For entry is whatever the caller sent, so keying on it
// would let a client pick its own bucket.
func clientKey(r *http.Request, header string) string {
	if header != "" {
		first, _, _ := strings.Cut(r.Header.Get(header), ",")
		if value := strings.TrimSpace(first); value != "" {
			return value
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
