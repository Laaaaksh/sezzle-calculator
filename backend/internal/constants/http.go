package constants

// HTTP header names and values shared by middleware and handlers.
const (
	HeaderRequestID   = "X-Request-ID"
	HeaderContentType = "Content-Type"
	HeaderAllow       = "Allow"

	ContentTypeJSON = "application/json"
)

// Response caching. Every API body is dynamic, so intermediaries must not
// store it.
const (
	HeaderCacheControl  = "Cache-Control"
	CacheControlNoStore = "no-store"
)

// Request ID rules. An incoming X-Request-ID is trusted only when it is short
// and made of log-safe characters; anything else is replaced, so a client
// cannot inject newlines or megabytes into log lines.
const (
	RequestIDMaxLength = 128
	// RequestIDExtraChars are accepted besides ASCII letters and digits.
	RequestIDExtraChars = "._-"
)

// CORS policy. Methods and headers are fixed by the API contract; only the
// origin list comes from configuration.
const (
	CORSMaxAgeSeconds = 600
)

// Router layout. Operational endpoints sit at the root so load balancers do
// not depend on the API version; features mount under APIBasePath.
const (
	APIBasePath = "/api/" + APIVersion
)

// Server wiring.
const (
	ListenAddrFmt = ":%d"
)

// MillisecondsPerDuration converts a duration for the access log: milliseconds as a float so
// sub-millisecond requests do not all read as zero.
const MillisecondsPerDuration = float64(1e6)
