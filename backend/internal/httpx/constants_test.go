package httpx_test

import (
	"errors"
	"strings"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/httpx"
)

const (
	testRequestID    = "req-123"
	testPath         = "/missing"
	testMethod       = "PUT"
	testSecretCause  = "pq: password authentication failed for user admin"
	testFieldKey     = "account"
	testFieldValue   = "acc-42"
	testIndexKey     = "operand_index"
	testIndexValue   = 1
	testStatusKey    = "status"
	testStatusValue  = "ok"
	jsonKeyRequestID = `"requestId"`
	jsonKeyFields    = `"fields"`
	wantFallbackBody = `{"error":{"code":"INTERNAL","message":"An internal error occurred."}}`
	wantContentType  = "application/json"
	wantNoStore      = "no-store"
)

const (
	// testMultiByteRune is two bytes in UTF-8, so a cut at an odd offset
	// would land inside it.
	testMultiByteRune   = "é"
	testMultiByteRunes  = 200
	testPathPrefix      = "/"
	testPathOverrunByte = 44
)

var (
	testLongPath         = testPathPrefix + strings.Repeat("a", httpx.MaxEchoedPathBytes+testPathOverrunByte)
	testMultiBytePath    = testPathPrefix + strings.Repeat(testMultiByteRune, testMultiByteRunes)
	testMultiBytePathCut = testPathPrefix + strings.Repeat(testMultiByteRune, (httpx.MaxEchoedPathBytes-1)/2)
)

var errTestSecret = errors.New(testSecretCause)
