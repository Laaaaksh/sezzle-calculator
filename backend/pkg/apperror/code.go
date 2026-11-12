// Package apperror is the single error model of the service: a typed code, a
// safe public message, an HTTP status and structured fields for logs.
package apperror

import (
	"fmt"
	"maps"
	"net/http"
	"slices"
	"sync"
)

// Code identifies a class of failure. A named type stops arbitrary strings
// being passed where a known code is expected.
type Code string

// Generic error codes, shared by every feature. A feature owns its own
// domain codes and adds them with Register (the calculator's live in
// internal/modules/calculator/entities), so this package never learns the
// vocabulary of any one module.
const (
	CodeBadRequest           Code = "BAD_REQUEST"
	CodeValidationFailed     Code = "VALIDATION_FAILED"
	CodeUnsupportedMediaType Code = "UNSUPPORTED_MEDIA_TYPE"
	CodePayloadTooLarge      Code = "PAYLOAD_TOO_LARGE"
	CodeNotFound             Code = "NOT_FOUND"
	CodeMethodNotAllowed     Code = "METHOD_NOT_ALLOWED"
	CodeInternal             Code = "INTERNAL"
	CodeServiceUnavailable   Code = "SERVICE_UNAVAILABLE"
	CodeRequestTimeout       Code = "REQUEST_TIMEOUT"
)

// definition is what a code means to a client: its HTTP status and default
// public message.
type definition struct {
	status  int
	message string
}

// registry is the only place a code is tied to a status and a message. It is
// written at start-up (generic codes below, feature codes from Register in
// package init) and read per request; the lock keeps a late Register, such as
// one from a test, from racing a reader.
var (
	registryMu sync.RWMutex
	registry   = map[Code]definition{
		CodeBadRequest:           {http.StatusBadRequest, MsgBadRequest},
		CodeValidationFailed:     {http.StatusUnprocessableEntity, MsgValidationFailed},
		CodeUnsupportedMediaType: {http.StatusUnsupportedMediaType, MsgUnsupportedMediaType},
		CodePayloadTooLarge:      {http.StatusRequestEntityTooLarge, MsgPayloadTooLarge},
		CodeNotFound:             {http.StatusNotFound, MsgNotFound},
		CodeMethodNotAllowed:     {http.StatusMethodNotAllowed, MsgMethodNotAllowed},
		CodeInternal:             {http.StatusInternalServerError, MsgInternal},
		CodeServiceUnavailable:   {http.StatusServiceUnavailable, MsgServiceUnavailable},
		// 503, not 408 or 504. 408 means the client was too slow to send its
		// request, which is not the case here. 504 means a gateway timed out
		// waiting for an upstream, and this service is not a gateway. The request
		// was fine and our own deadline ran out, so "temporarily unavailable, try
		// again" (503) is the honest signal and clients already retry it.
		CodeRequestTimeout: {http.StatusServiceUnavailable, MsgRequestTimeout},
	}
)

// Register adds a feature's code with its HTTP status and default public
// message. It panics on an empty or already registered code and on a status
// outside 400-599: two features fighting for one code, or a success status on
// an error, is a wiring bug that must stop start-up, not be silently resolved.
func Register(code Code, status int, message string) {
	if code == "" || status < http.StatusBadRequest || status > maxErrorStatus {
		panic(fmt.Sprintf(panicFmtInvalidRegistration, code, status))
	}
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, exists := registry[code]; exists {
		panic(fmt.Sprintf(panicFmtDuplicateCode, code))
	}
	registry[code] = definition{status: status, message: message}
}

// HTTPStatus returns the status for c. An unknown code maps to 500 so a
// forgotten mapping fails closed as a server error, never as a success.
func (c Code) HTTPStatus() int {
	registryMu.RLock()
	defer registryMu.RUnlock()
	if def, ok := registry[c]; ok {
		return def.status
	}
	return http.StatusInternalServerError
}

// DefaultMessage returns the public text for c, falling back to the INTERNAL
// message for an unknown code.
func (c Code) DefaultMessage() string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	if def, ok := registry[c]; ok {
		return def.message
	}
	return MsgInternal
}

// Codes lists every registered code in alphabetical order. docs/ERROR_CODES.md
// is checked against it in a test, so a new code cannot ship undocumented. It
// sees a feature's codes only once that feature's package has been imported.
func Codes() []Code {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return slices.Sorted(maps.Keys(registry))
}
