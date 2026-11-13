package httpx_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/httpx"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/logger"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

type HandlersTestSuite struct {
	suite.Suite
	recorder *httptest.ResponseRecorder
	request  *http.Request
	observed *observer.ObservedLogs
	restore  func()
}

func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}

func (s *HandlersTestSuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
	s.request = httptest.NewRequest(testMethod, testPath, nil)
	core, observed := observer.New(zapcore.DebugLevel)
	s.observed = observed
	s.restore = logger.SetForTest(zap.New(core).Sugar())
}

func (s *HandlersTestSuite) TearDownTest() {
	s.restore()
}

func (s *HandlersTestSuite) TestNotFoundWritesJSONNotFound() {
	httpx.NotFound(s.recorder, s.request)

	var got httpx.ErrorResponse
	s.Require().NoError(json.Unmarshal(s.recorder.Body.Bytes(), &got))
	s.Equal(http.StatusNotFound, s.recorder.Code)
	s.Equal(apperror.CodeNotFound, got.Error.Code)
	s.Equal(apperror.MsgNotFound, got.Error.Message)
}

func (s *HandlersTestSuite) TestNotFoundLogsMethodAndPath() {
	httpx.NotFound(s.recorder, s.request)

	entries := s.observed.FilterMessage(constants.LogMsgErrorResponse).All()
	s.Require().Len(entries, 1)
	fields := entries[0].ContextMap()[constants.LogKeyFields].(map[string]any)
	s.Equal(testMethod, fields[apperror.FieldMethod])
	s.Equal(testPath, fields[apperror.FieldPath])
}

func (s *HandlersTestSuite) TestMethodNotAllowedWritesJSONMethodNotAllowed() {
	httpx.MethodNotAllowed(s.recorder, s.request)

	var got httpx.ErrorResponse
	s.Require().NoError(json.Unmarshal(s.recorder.Body.Bytes(), &got))
	s.Equal(http.StatusMethodNotAllowed, s.recorder.Code)
	s.Equal(apperror.CodeMethodNotAllowed, got.Error.Code)
	s.Equal(apperror.MsgMethodNotAllowed, got.Error.Message)
}

func (s *HandlersTestSuite) echoedPath() string {
	var got httpx.ErrorResponse
	s.Require().NoError(json.Unmarshal(s.recorder.Body.Bytes(), &got))
	path, ok := got.Error.Fields[apperror.FieldPath].(string)
	s.Require().True(ok)
	return path
}

func (s *HandlersTestSuite) TestNotFoundEchoesAShortPathWhole() {
	httpx.NotFound(s.recorder, s.request)

	s.Equal(testPath, s.echoedPath())
}

func (s *HandlersTestSuite) TestNotFoundCapsAnOverlongPath() {
	s.request.URL.Path = testLongPath

	httpx.NotFound(s.recorder, s.request)

	s.Len(s.echoedPath(), httpx.MaxEchoedPathBytes)
}

func (s *HandlersTestSuite) TestNotFoundCutsAnOverlongPathOnARuneBoundary() {
	s.request.URL.Path = testMultiBytePath

	httpx.NotFound(s.recorder, s.request)

	s.Equal(testMultiBytePathCut, s.echoedPath())
}

func (s *HandlersTestSuite) TestNotFoundLogsTheCappedPath() {
	s.request.URL.Path = testLongPath

	httpx.NotFound(s.recorder, s.request)

	entries := s.observed.FilterMessage(constants.LogMsgErrorResponse).All()
	s.Require().Len(entries, 1)
	fields := entries[0].ContextMap()[constants.LogKeyFields].(map[string]any)
	s.Len(fields[apperror.FieldPath], httpx.MaxEchoedPathBytes)
}

func (s *HandlersTestSuite) TestMethodNotAllowedCapsAnOverlongPath() {
	s.request.URL.Path = testLongPath

	httpx.MethodNotAllowed(s.recorder, s.request)

	s.Len(s.echoedPath(), httpx.MaxEchoedPathBytes)
}
