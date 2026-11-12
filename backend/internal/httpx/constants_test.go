package httpx_test

import "errors"

const (
	testRequestID    = "req-123"
	testPath         = "/missing"
	testMethod       = "PUT"
	testSecretCause  = "pq: password authentication failed for user admin"
	testFieldKey     = "account"
	testFieldValue   = "acc-42"
	testStatusKey    = "status"
	testStatusValue  = "ok"
	jsonKeyRequestID = `"requestId"`
	jsonKeyFields    = `"fields"`
	wantFallbackBody = `{"error":{"code":"INTERNAL","message":"An internal error occurred."}}`
	wantContentType  = "application/json"
	wantNoStore      = "no-store"
)

var errTestSecret = errors.New(testSecretCause)
