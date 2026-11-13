package constants

// Rate limit config keys, as they appear in the TOML files.
const (
	KeyRateLimitEnabled           = "rate_limit.enabled"
	KeyRateLimitRequestsPerSecond = "rate_limit.requests_per_second"
	KeyRateLimitBurst             = "rate_limit.burst"
)

// Rate limit bounds. A typo in an environment file cannot lift the limit
// beyond these.
const (
	RateLimitRateMax  = 10000
	RateLimitBurstMin = 1
	RateLimitBurstMax = 10000
)

// HeaderRetryAfter tells a limited client how many seconds to wait.
const HeaderRetryAfter = "Retry-After"

// RetryAfterMinSeconds keeps Retry-After at one second or more, so a wait of
// a few milliseconds is not sent as "0" (retry at once).
const RetryAfterMinSeconds = 1

// ErrFmtRateLimitRate is the message for a rate that is zero, negative or too big.
const ErrFmtRateLimitRate = "%s must be greater than zero and at most %d"
