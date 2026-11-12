package interceptors_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap/zaptest/observer"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/interceptors"
)

type AccessLogTestSuite struct {
	suite.Suite
	recorder *httptest.ResponseRecorder
	request  *http.Request
	observed *observer.ObservedLogs
	restore  func()
}

func TestAccessLogTestSuite(t *testing.T) {
	suite.Run(t, new(AccessLogTestSuite))
}

func (s *AccessLogTestSuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
	s.request = httptest.NewRequest(http.MethodPost, testPath+"?"+testQuery, nil)
	s.observed, s.restore = captureLogs()
}

func (s *AccessLogTestSuite) TearDownTest() {
	s.restore()
}

func (s *AccessLogTestSuite) serve(h http.HandlerFunc) {
	interceptors.RequestID(interceptors.AccessLog(h)).ServeHTTP(s.recorder, s.request)
}

func (s *AccessLogTestSuite) entry() map[string]any {
	entries := s.observed.FilterMessage(constants.LogMsgRequestCompleted).All()
	s.Require().Len(entries, 1)
	return entries[0].ContextMap()
}

func (s *AccessLogTestSuite) TestLogsMethodPathStatusAndRequestID() {
	s.request.Header.Set(constants.HeaderRequestID, testIncomingID)

	s.serve(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) })

	fields := s.entry()
	s.Equal(http.MethodPost, fields[constants.LogKeyMethod])
	s.Equal(testPath, fields[constants.LogKeyPath])
	s.Equal(int64(http.StatusTeapot), fields[constants.LogKeyStatus])
	s.Equal(testIncomingID, fields[constants.LogKeyRequestID])
}

func (s *AccessLogTestSuite) TestLogsDurationInMilliseconds() {
	s.serve(okHandler)

	duration, ok := s.entry()[constants.LogKeyDurationMS].(float64)
	s.True(ok)
	s.GreaterOrEqual(duration, 0.0)
}

func (s *AccessLogTestSuite) TestNeverLogsTheQueryString() {
	s.serve(okHandler)

	for _, value := range s.entry() {
		s.NotContains(fmt.Sprint(value), testQuery)
	}
}

func (s *AccessLogTestSuite) TestStatusDefaultsTo200WhenHandlerOnlyWritesBody() {
	s.serve(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(testBody)) })

	s.Equal(int64(http.StatusOK), s.entry()[constants.LogKeyStatus])
}

func (s *AccessLogTestSuite) TestRecordsFirstStatusWhenHeaderWrittenTwice() {
	s.serve(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.WriteHeader(http.StatusTeapot)
	})

	s.Equal(int64(http.StatusCreated), s.entry()[constants.LogKeyStatus])
}

func (s *AccessLogTestSuite) TestPanicIsLoggedAs500AndRaisedAgain() {
	handler := interceptors.RequestID(interceptors.AccessLog(http.HandlerFunc(
		func(_ http.ResponseWriter, _ *http.Request) { panic(testPanicValue) })))

	s.PanicsWithValue(testPanicValue, func() { handler.ServeHTTP(s.recorder, s.request) })

	s.Equal(int64(http.StatusInternalServerError), s.entry()[constants.LogKeyStatus])
}

func (s *AccessLogTestSuite) TestPanicAfterHeaderLogsTheStatusAlreadySent() {
	handler := interceptors.RequestID(interceptors.AccessLog(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
			panic(testPanicValue)
		})))

	s.PanicsWithValue(testPanicValue, func() { handler.ServeHTTP(s.recorder, s.request) })

	s.Equal(int64(http.StatusCreated), s.entry()[constants.LogKeyStatus])
}

func (s *AccessLogTestSuite) TestWrapperStillAllowsResponseController() {
	var flushErr error
	s.serve(func(w http.ResponseWriter, _ *http.Request) {
		flushErr = http.NewResponseController(w).Flush()
	})

	s.NoError(flushErr)
}

func (s *AccessLogTestSuite) TestRecoveryAndAccessLogShareOneWriter() {
	chain := interceptors.RequestID(interceptors.Recovery(interceptors.AccessLog(http.HandlerFunc(
		func(_ http.ResponseWriter, _ *http.Request) { panic(testPanicValue) }))))

	chain.ServeHTTP(s.recorder, s.request)

	s.Equal(http.StatusInternalServerError, s.recorder.Code)
	s.Equal(int64(http.StatusInternalServerError), s.entry()[constants.LogKeyStatus])
}
