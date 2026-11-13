package interceptors

import (
	"net/http"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
)

// requestIDFallback is used only if the system random source fails, which is
// not expected in practice. A shared ID is worse than a unique one but far
// better than failing the request.
const requestIDFallback = "unavailable"

// panicMessageFmt builds the cause of a recovered panic; it is for logs only and never sent.
const panicMessageFmt = "panic: %v"

// corsAllowedMethods and the header lists below are the API contract, not
// configuration: widening them is a code change that gets reviewed.
var (
	corsAllowedMethods = []string{http.MethodGet, http.MethodPost, http.MethodOptions}
	corsAllowedHeaders = []string{constants.HeaderContentType, constants.HeaderRequestID}
	corsExposedHeaders = []string{constants.HeaderRequestID}
)
