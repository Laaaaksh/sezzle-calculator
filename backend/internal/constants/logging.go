package constants

// Log messages are static strings so every occurrence can be found with one
// search; variable data travels in the LogKey* fields below.
const (
	LogMsgConfigLoaded      = "configuration loaded"
	LogMsgServerStarting    = "http server starting"
	LogMsgServerStopped     = "http server stopped"
	LogMsgShutdownStarted   = "graceful shutdown started"
	LogMsgShutdownCompleted = "graceful shutdown completed"
	LogMsgShutdownFailed    = "graceful shutdown failed"
	LogMsgRequestCompleted  = "request completed"
	LogMsgRequestPanicked   = "request handler panicked"
	LogMsgErrorResponse     = "error response sent"
	LogMsgRequestRejected   = "request rejected"
	LogMsgResponseEncode    = "response encoding failed"
	LogMsgServerListening   = "http server listening"
)

// Log field keys.
const (
	LogKeyEnv           = "env"
	LogKeyPort          = "port"
	LogKeyError         = "error"
	LogKeyRequestID     = "request_id"
	LogKeyMethod        = "method"
	LogKeyPath          = "path"
	LogKeyStatus        = "status"
	LogKeyDurationMS    = "duration_ms"
	LogKeyCode          = "code"
	LogKeyCause         = "cause"
	LogKeyFields        = "fields"
	LogKeyPanic         = "panic"
	LogKeyStack         = "stack"
	LogKeyAddr          = "addr"
	LogKeyContentLength = "content_length"
	LogKeyLimit         = "limit"
)
