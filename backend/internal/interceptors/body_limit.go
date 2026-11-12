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
					WithField(constants.LogKeyLimit, limit))
				return
			}
			r.Body = http.MaxBytesReader(innermostWriter(w), r.Body, limit)
			next.ServeHTTP(w, r)
		})
	}
}

// innermostWriter unwraps middleware writers down to the one net/http created.
// MaxBytesReader can only tell the server to close an over-limit connection
// through an unexported hook on that writer, which wrappers cannot forward.
func innermostWriter(w http.ResponseWriter) http.ResponseWriter {
	for {
		u, ok := w.(interface{ Unwrap() http.ResponseWriter })
		if !ok {
			return w
		}
		w = u.Unwrap()
	}
}
