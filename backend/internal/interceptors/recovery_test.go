package interceptors_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap/zaptest/observer"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/httpx"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/interceptors"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

type RecoveryTestSuite struct {
	suite.Suite
	recorder *httptest.ResponseRecorder
	request  *http.Request
	observed *observer.ObservedLogs
	restore  func()
}

func TestRecoveryTestSuite(t *testing.T) {
	suite.Run(t, new(RecoveryTestSuite))
}

func (s *RecoveryTestSuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
	s.request = httptest.NewRequest(http.MethodGet, testPath, nil)
	s.observed, s.restore = captureLogs()
}

func (s *RecoveryTestSuite) TearDownTest() {
	s.restore()
}

// panickingHandler is a named function so its name must appear in the logged stack.
func panickingHandler(_ http.ResponseWriter, _ *http.Request) {
	panic(testPanicValue)
}

func (s *RecoveryTestSuite) serve(h http.HandlerFunc) {
	chain := interceptors.RequestID(interceptors.Recovery(h))
	chain.ServeHTTP(s.recorder, s.request)
}

func (s *RecoveryTestSuite) TestPanicBecomesJSONInternalError() {
	s.serve(panickingHandler)

	var got httpx.ErrorResponse
	s.Require().NoError(json.Unmarshal(s.recorder.Body.Bytes(), &got))
	s.Equal(http.StatusInternalServerError, s.recorder.Code)
	s.Equal(apperror.CodeInternal, got.Error.Code)
	s.Equal(testWantSecretMsg, got.Error.Message)
	s.Equal(s.recorder.Header().Get(constants.HeaderRequestID), got.Error.RequestID)
}

func (s *RecoveryTestSuite) TestPanicValueNeverReachesTheClient() {
	s.serve(panickingHandler)

	s.NotContains(s.recorder.Body.String(), testPanicValue)
}

func (s *RecoveryTestSuite) TestPanicLogsValueAndStackOfThePanickingHandler() {
	s.serve(panickingHandler)

	entries := s.observed.FilterMessage(constants.LogMsgRequestPanicked).All()
	s.Require().Len(entries, 1)
	fields := entries[0].ContextMap()
	s.Equal(testPanicValue, fields[constants.LogKeyPanic])
	s.Contains(fields[constants.LogKeyStack], "panickingHandler")
	s.Equal(testPath, fields[constants.LogKeyPath])
	s.Equal(http.MethodGet, fields[constants.LogKeyMethod])
}

func (s *RecoveryTestSuite) TestPanicLogCarriesRequestID() {
	s.serve(panickingHandler)

	entries := s.observed.FilterMessage(constants.LogMsgRequestPanicked).All()
	s.Require().Len(entries, 1)
	s.Equal(s.recorder.Header().Get(constants.HeaderRequestID), entries[0].ContextMap()[constants.LogKeyRequestID])
}

func (s *RecoveryTestSuite) TestPanicAfterHeaderWrittenAbortsTheResponse() {
	s.PanicsWithValue(http.ErrAbortHandler, func() {
		s.serve(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(testPartialBody))
			panic(testPanicValue)
		})
	})

	s.Equal(http.StatusOK, s.recorder.Code)
	s.Equal(testPartialBody, s.recorder.Body.String())
	s.Equal(1, s.observed.FilterMessage(constants.LogMsgRequestPanicked).Len())
}

func (s *RecoveryTestSuite) TestNoPanicPassesResponseThrough() {
	s.serve(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	s.Equal(http.StatusAccepted, s.recorder.Code)
	s.Equal(0, s.observed.FilterMessage(constants.LogMsgRequestPanicked).Len())
}

func (s *RecoveryTestSuite) TestAbortHandlerPanicIsRaisedAgain() {
	handler := interceptors.Recovery(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic(http.ErrAbortHandler)
	}))

	s.PanicsWithValue(http.ErrAbortHandler, func() {
		handler.ServeHTTP(s.recorder, s.request)
	})
	s.Equal(0, s.observed.FilterMessage(constants.LogMsgRequestPanicked).Len())
}
