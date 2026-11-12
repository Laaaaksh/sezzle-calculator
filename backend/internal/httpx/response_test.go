package httpx_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/contextkeys"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/httpx"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/logger"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

type ResponseTestSuite struct {
	suite.Suite
	recorder *httptest.ResponseRecorder
	observed *observer.ObservedLogs
	restore  func()
}

func TestResponseTestSuite(t *testing.T) {
	suite.Run(t, new(ResponseTestSuite))
}

func (s *ResponseTestSuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
	core, observed := observer.New(zapcore.DebugLevel)
	s.observed = observed
	s.restore = logger.SetForTest(zap.New(core).Sugar())
}

func (s *ResponseTestSuite) TearDownTest() {
	s.restore()
}

func (s *ResponseTestSuite) decodeError() httpx.ErrorResponse {
	var got httpx.ErrorResponse
	s.Require().NoError(json.Unmarshal(s.recorder.Body.Bytes(), &got))
	return got
}

func (s *ResponseTestSuite) TestWriteJSONSetsStatusHeadersAndBody() {
	httpx.WriteJSON(context.Background(), s.recorder, http.StatusCreated, map[string]string{testStatusKey: testStatusValue})

	s.Equal(http.StatusCreated, s.recorder.Code)
	s.Equal(wantContentType, s.recorder.Header().Get(constants.HeaderContentType))
	s.Equal(wantNoStore, s.recorder.Header().Get(constants.HeaderCacheControl))
	s.JSONEq(`{"status":"ok"}`, s.recorder.Body.String())
}

func (s *ResponseTestSuite) TestWriteJSONUnencodableValueSendsInternalFallback() {
	httpx.WriteJSON(context.Background(), s.recorder, http.StatusOK, make(chan int))

	s.Equal(http.StatusInternalServerError, s.recorder.Code)
	s.Equal(wantFallbackBody, s.recorder.Body.String())
	s.Equal(wantContentType, s.recorder.Header().Get(constants.HeaderContentType))
	s.Equal(1, s.observed.FilterMessage(constants.LogMsgResponseEncode).Len())
}

func (s *ResponseTestSuite) TestWriteJSONEncodeFailureLogsRequestID() {
	ctx := contextkeys.WithRequestID(context.Background(), testRequestID)

	httpx.WriteJSON(ctx, s.recorder, http.StatusOK, make(chan int))

	entries := s.observed.FilterMessage(constants.LogMsgResponseEncode).All()
	s.Require().Len(entries, 1)
	s.Equal(testRequestID, entries[0].ContextMap()[constants.LogKeyRequestID])
}

func (s *ResponseTestSuite) TestWriteErrorUsesStatusAndPublicFields() {
	httpx.WriteError(context.Background(), s.recorder, apperror.New(apperror.CodeNotFound, nil))

	got := s.decodeError()
	s.Equal(http.StatusNotFound, s.recorder.Code)
	s.Equal(apperror.CodeNotFound, got.Error.Code)
	s.Equal(apperror.MsgNotFound, got.Error.Message)
	s.Equal(wantContentType, s.recorder.Header().Get(constants.HeaderContentType))
	s.Equal(wantNoStore, s.recorder.Header().Get(constants.HeaderCacheControl))
}

func (s *ResponseTestSuite) TestWriteErrorNeverSendsTheCause() {
	err := apperror.New(apperror.CodeInternal, errTestSecret).WithField(testFieldKey, testFieldValue)

	httpx.WriteError(context.Background(), s.recorder, err)

	s.NotContains(s.recorder.Body.String(), testSecretCause)
}

func (s *ResponseTestSuite) TestWriteErrorSendsFieldsOfEveryCode() {
	err := apperror.New(apperror.CodeBadRequest, errTestSecret).
		WithField(testFieldKey, testFieldValue).
		WithField(testIndexKey, testIndexValue)

	httpx.WriteError(context.Background(), s.recorder, err)

	s.Equal(map[string]any{testFieldKey: testFieldValue, testIndexKey: float64(testIndexValue)}, s.decodeError().Error.Fields)
}

func (s *ResponseTestSuite) TestWriteErrorOmitsFieldsWhenErrorHasNone() {
	httpx.WriteError(context.Background(), s.recorder, apperror.New(apperror.CodeBadRequest, nil))

	s.NotContains(s.recorder.Body.String(), jsonKeyFields)
}

func (s *ResponseTestSuite) TestWriteErrorIncludesRequestIDFromContext() {
	ctx := contextkeys.WithRequestID(context.Background(), testRequestID)

	httpx.WriteError(ctx, s.recorder, apperror.New(apperror.CodeBadRequest, nil))

	s.Equal(testRequestID, s.decodeError().Error.RequestID)
}

func (s *ResponseTestSuite) TestWriteErrorOmitsRequestIDWhenContextHasNone() {
	httpx.WriteError(context.Background(), s.recorder, apperror.New(apperror.CodeBadRequest, nil))

	s.NotContains(s.recorder.Body.String(), jsonKeyRequestID)
}

func (s *ResponseTestSuite) TestWriteErrorLogsServerFaultAtErrorLevelWithCauseAndFields() {
	err := apperror.New(apperror.CodeInternal, errTestSecret).WithField(testFieldKey, testFieldValue)

	httpx.WriteError(context.Background(), s.recorder, err)

	entries := s.observed.FilterMessage(constants.LogMsgErrorResponse).All()
	s.Require().Len(entries, 1)
	s.Equal(zapcore.ErrorLevel, entries[0].Level)
	fields := entries[0].ContextMap()
	s.Equal(string(apperror.CodeInternal), fields[constants.LogKeyCode])
	s.Contains(fields[constants.LogKeyCause], testSecretCause)
	s.Equal(map[string]any{testFieldKey: testFieldValue}, fields[constants.LogKeyFields])
}

func (s *ResponseTestSuite) TestWriteErrorLogsClientFaultAtWarnLevel() {
	httpx.WriteError(context.Background(), s.recorder, apperror.New(apperror.CodeNotFound, nil))

	entries := s.observed.FilterMessage(constants.LogMsgErrorResponse).All()
	s.Require().Len(entries, 1)
	s.Equal(zapcore.WarnLevel, entries[0].Level)
}

func (s *ResponseTestSuite) TestWriteErrorLogsServiceUnavailableAtWarnLevel() {
	httpx.WriteError(context.Background(), s.recorder, apperror.New(apperror.CodeServiceUnavailable, nil))

	s.Equal(http.StatusServiceUnavailable, s.recorder.Code)
	entries := s.observed.FilterMessage(constants.LogMsgErrorResponse).All()
	s.Require().Len(entries, 1)
	s.Equal(zapcore.WarnLevel, entries[0].Level)
}

func (s *ResponseTestSuite) TestErrorBodyIncludesFieldsWhenSet() {
	body := httpx.ErrorBody{
		Code:    apperror.CodeValidationFailed,
		Message: apperror.MsgValidationFailed,
		Fields:  map[string]any{testFieldKey: testFieldValue},
	}

	raw, err := json.Marshal(httpx.ErrorResponse{Error: body})

	s.Require().NoError(err)
	s.JSONEq(`{"error":{"code":"VALIDATION_FAILED","message":"The request failed validation.","fields":{"account":"acc-42"}}}`, string(raw))
}
