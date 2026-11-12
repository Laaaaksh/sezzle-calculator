package interceptors

import (
	"net/http"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/httpx"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

// BodyLimit caps the request body at limit bytes. A declared Content-Length
// over the limit is refused immediately with a JSON 413; a body of unknown or
// understated length is cut off by http.MaxBytesReader when a handler reads
// past the limit, and the handler maps that read error to PAYLOAD_TOO_LARGE.
func BodyLimit(limit int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > limit {
				httpx.WriteError(r.Context(), w, apperror.New(apperror.CodePayloadTooLarge, nil).
					WithField(constants.LogKeyContentLength, r.ContentLength))
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, limit)
			next.ServeHTTP(w, r)
		})
	}
}
