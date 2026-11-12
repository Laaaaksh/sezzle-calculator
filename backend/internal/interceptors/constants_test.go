package interceptors_test

import (
	"errors"
	"strings"
	"time"
)

const (
	testPath          = "/things"
	testPipeURL       = "http://pipe/things"
	testQuery         = "secret=hunter2"
	testIncomingID    = "abc-123"
	testDotID         = "a.b_c-D9"
	testInvalidID     = "has space"
	testSlashID       = "a/b"
	testFallbackID    = "unavailable"
	testPanicValue    = "boom-value"
	testPartialBody   = "partial"
	testBody          = "hello"
	testAllowedOrigin = "http://localhost:5173"
	testOtherOrigin   = "http://evil.example.com"
	testBodyLimit     = int64(8)
	testOverLimit     = "123456789"
	testWantSecretMsg = "An internal error occurred."

	headerOrigin         = "Origin"
	headerACRMethod      = "Access-Control-Request-Method"
	headerACRHeaders     = "Access-Control-Request-Headers"
	headerACAOrigin      = "Access-Control-Allow-Origin"
	headerACAMethods     = "Access-Control-Allow-Methods"
	headerACAHeaders     = "Access-Control-Allow-Headers"
	headerACExpose       = "Access-Control-Expose-Headers"
	headerACMaxAge       = "Access-Control-Max-Age"
	wantMaxAge           = "600"
	wantPreflightMethods = "POST"
	wantRequestIDLength  = 36
	wantUUIDVersion      = 4

	testTimeout   = 50 * time.Millisecond
	testWaitLimit = 2 * time.Second
)

var (
	testLongID    = strings.Repeat("a", 128)
	testTooLongID = strings.Repeat("a", 129)
	errTestIDGen  = errors.New("entropy exhausted")
)
