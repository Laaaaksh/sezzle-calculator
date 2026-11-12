package interceptors

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/contextkeys"
)

// IDGenerator produces a new request ID.
type IDGenerator func() (string, error)

// RequestID attaches a request ID to the context and echoes it on the
// response. It reuses a well-formed X-Request-ID from the caller so a trace can
// span services, and generates a UUID v4 otherwise.
func RequestID(next http.Handler) http.Handler {
	return RequestIDWith(generateUUID)(next)
}

// RequestIDWith is RequestID with an injectable generator, so the generator
// failure path can be exercised.
func RequestIDWith(generate IDGenerator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get(constants.HeaderRequestID)
			if !validRequestID(id) {
				id = newRequestID(generate)
			}
			w.Header().Set(constants.HeaderRequestID, id)
			next.ServeHTTP(w, r.WithContext(contextkeys.WithRequestID(r.Context(), id)))
		})
	}
}

func generateUUID() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

func newRequestID(generate IDGenerator) string {
	id, err := generate()
	if err != nil {
		return requestIDFallback
	}
	return id
}

// validRequestID accepts 1..128 characters of [A-Za-z0-9._-]. The strict
// alphabet keeps log injection (newlines, control characters) out of the logs.
func validRequestID(id string) bool {
	if id == "" || len(id) > constants.RequestIDMaxLength {
		return false
	}
	for i := 0; i < len(id); i++ {
		if !validRequestIDChar(id[i]) {
			return false
		}
	}
	return true
}

func validRequestIDChar(c byte) bool {
	switch {
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		return true
	}
	for i := 0; i < len(constants.RequestIDExtraChars); i++ {
		if c == constants.RequestIDExtraChars[i] {
			return true
		}
	}
	return false
}
