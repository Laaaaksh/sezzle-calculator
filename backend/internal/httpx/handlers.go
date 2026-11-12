package httpx

import (
	"net/http"
	"unicode/utf8"

	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

// MaxEchoedPathBytes caps how much of the request path a 404 or 405 repeats in
// its error fields. Those fields reach the client and the warn log, and a path
// can be as long as the server's header limit, so an unbounded echo would let
// every mistyped URL cost a log line that size.
const MaxEchoedPathBytes = 256

// NotFound answers an unrouted path with the standard JSON error. The router
// would otherwise reply in plain text, which clients cannot parse.
func NotFound(w http.ResponseWriter, r *http.Request) {
	WriteError(r.Context(), w, apperror.New(apperror.CodeNotFound, nil).
		WithField(apperror.FieldMethod, r.Method).
		WithField(apperror.FieldPath, echoablePath(r.URL.Path)))
}

// MethodNotAllowed answers a known path with an unsupported method using the
// standard JSON error.
func MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	WriteError(r.Context(), w, apperror.New(apperror.CodeMethodNotAllowed, nil).
		WithField(apperror.FieldMethod, r.Method).
		WithField(apperror.FieldPath, echoablePath(r.URL.Path)))
}

// echoablePath cuts path to MaxEchoedPathBytes on a UTF-8 boundary, so the
// result is still valid text and the JSON encoder never has to replace a
// split character.
func echoablePath(path string) string {
	if len(path) <= MaxEchoedPathBytes {
		return path
	}
	cut := MaxEchoedPathBytes
	for cut > 0 && !utf8.RuneStart(path[cut]) {
		cut--
	}
	return path[:cut]
}
