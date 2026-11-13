package constants

// Container healthcheck probe (cmd/healthcheck). The runtime image is distroless
// and has no shell or curl, so the probe ships as its own small binary.
const (
	// HealthcheckHost is loopback: the probe always asks the process in its own container.
	HealthcheckHost = "127.0.0.1"
	// HealthcheckScheme and HealthcheckURLFmt build the probe URL.
	HealthcheckScheme = "http"
	HealthcheckURLFmt = "%s://%s:%d%s"
	// HealthcheckTimeoutSeconds is below the Docker healthcheck timeout so the probe
	// reports its own failure instead of being killed.
	HealthcheckTimeoutSeconds = 2

	MsgHealthcheckConfigFailed = "healthcheck: load configuration: %v\n"
	MsgHealthcheckFailed       = "healthcheck: %v\n"
	ErrFmtHealthcheckRequest   = "build request: %w"
	ErrFmtHealthcheckDo        = "request %s: %w"
	ErrFmtHealthcheckStatus    = "request %s: unexpected status %d"
)
