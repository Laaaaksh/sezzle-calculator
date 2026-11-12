package interceptors_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/contextkeys"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/interceptors"
)

type RequestIDTestSuite struct {
	suite.Suite
	recorder *httptest.ResponseRecorder
	request  *http.Request
	seenID   string
	seenOK   bool
	handler  http.Handler
}

func TestRequestIDTestSuite(t *testing.T) {
	suite.Run(t, new(RequestIDTestSuite))
}

func (s *RequestIDTestSuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
	s.request = httptest.NewRequest(http.MethodGet, testPath, nil)
	s.seenID, s.seenOK = "", false
	s.handler = interceptors.RequestID(s.capture())
}

func (s *RequestIDTestSuite) capture() http.Handler {
	return http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		s.seenID, s.seenOK = contextkeys.RequestID(r.Context())
	})
}

func (s *RequestIDTestSuite) TestGeneratesUUIDv4WhenHeaderAbsent() {
	s.handler.ServeHTTP(s.recorder, s.request)

	parsed, err := uuid.Parse(s.seenID)
	s.Require().NoError(err)
	s.Equal(uuid.Version(wantUUIDVersion), parsed.Version())
	s.Len(s.seenID, wantRequestIDLength)
}

func (s *RequestIDTestSuite) TestEchoesGeneratedIDOnResponseAndContext() {
	s.handler.ServeHTTP(s.recorder, s.request)

	s.True(s.seenOK)
	s.Equal(s.seenID, s.recorder.Header().Get(constants.HeaderRequestID))
}

func (s *RequestIDTestSuite) TestAcceptsWellFormedIncomingID() {
	s.request.Header.Set(constants.HeaderRequestID, testIncomingID)

	s.handler.ServeHTTP(s.recorder, s.request)

	s.Equal(testIncomingID, s.seenID)
	s.Equal(testIncomingID, s.recorder.Header().Get(constants.HeaderRequestID))
}

func (s *RequestIDTestSuite) TestAcceptsDotUnderscoreAndMixedCase() {
	s.request.Header.Set(constants.HeaderRequestID, testDotID)

	s.handler.ServeHTTP(s.recorder, s.request)

	s.Equal(testDotID, s.seenID)
}

func (s *RequestIDTestSuite) TestAcceptsIDOfMaximumLength() {
	s.request.Header.Set(constants.HeaderRequestID, testLongID)

	s.handler.ServeHTTP(s.recorder, s.request)

	s.Equal(testLongID, s.seenID)
}

func (s *RequestIDTestSuite) TestReplacesIDOneCharacterTooLong() {
	s.request.Header.Set(constants.HeaderRequestID, testTooLongID)

	s.handler.ServeHTTP(s.recorder, s.request)

	s.NotEqual(testTooLongID, s.seenID)
	s.Len(s.seenID, wantRequestIDLength)
}

func (s *RequestIDTestSuite) TestReplacesIDContainingSpace() {
	s.request.Header.Set(constants.HeaderRequestID, testInvalidID)

	s.handler.ServeHTTP(s.recorder, s.request)

	s.NotEqual(testInvalidID, s.seenID)
	s.Equal(s.seenID, s.recorder.Header().Get(constants.HeaderRequestID))
}

func (s *RequestIDTestSuite) TestReplacesIDContainingSlash() {
	s.request.Header.Set(constants.HeaderRequestID, testSlashID)

	s.handler.ServeHTTP(s.recorder, s.request)

	s.NotEqual(testSlashID, s.seenID)
}

func (s *RequestIDTestSuite) TestGeneratesDistinctIDsPerRequest() {
	s.handler.ServeHTTP(s.recorder, s.request)
	first := s.seenID

	s.handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, testPath, nil))

	s.NotEqual(first, s.seenID)
}

func (s *RequestIDTestSuite) TestUsesFallbackWhenGeneratorFails() {
	failing := func() (string, error) { return "", errTestIDGen }
	handler := interceptors.RequestIDWith(failing)(s.capture())

	handler.ServeHTTP(s.recorder, s.request)

	s.Equal(testFallbackID, s.seenID)
	s.Equal(testFallbackID, s.recorder.Header().Get(constants.HeaderRequestID))
}
