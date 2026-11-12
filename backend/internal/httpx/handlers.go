package httpx

import (
	"net/http"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

// NotFound answers an unrouted path with the standard JSON error. The router
// would otherwise reply in plain text, which clients cannot parse.
func NotFound(w http.ResponseWriter, r *http.Request) {
	WriteError(r.Context(), w, apperror.New(apperror.CodeNotFound, nil).
		WithField(constants.LogKeyMethod, r.Method).
		WithField(constants.LogKeyPath, r.URL.Path))
}

// MethodNotAllowed answers a known path with an unsupported method using the
// standard JSON error.
func MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	WriteError(r.Context(), w, apperror.New(apperror.CodeMethodNotAllowed, nil).
		WithField(constants.LogKeyMethod, r.Method).
		WithField(constants.LogKeyPath, r.URL.Path))
}
