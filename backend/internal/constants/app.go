package constants

// APIVersion is the path segment every public route is mounted under, so a
// breaking change can ship as v2 without touching v1 clients.
const APIVersion = "v1"

// Process exit codes. Distinct non-zero codes let a supervisor tell a bad
// configuration (do not restart-loop) apart from a runtime crash.
const (
	ExitCodeOK            = 0
	ExitCodeRuntime       = 1
	ExitCodeConfigInvalid = 2
)

// MsgConfigLoadFailed is printed to stderr when startup configuration is
// unusable. The logger is not built yet at that point, so plain stderr is the
// only reliable channel.
const MsgConfigLoadFailed = "failed to load configuration: %v\n"

// Startup failures after configuration loads. The logger may be unusable at
// these points, so they are reported on stderr.
const (
	ExitCodeLoggerInvalid = 3

	MsgLoggerInitFailed = "failed to initialise logger: %v\n"
	MsgLoggerSyncFailed = "failed to flush logger: %v\n"
	MsgRunFailed        = "server failed: %v\n"
)
