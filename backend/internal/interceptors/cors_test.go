package interceptors_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/interceptors"
)

type CORSTestSuite struct {
	suite.Suite
	recorder *httptest.ResponseRecorder
	handler  http.Handler
	reached  bool
}

func TestCORSTestSuite(t *testing.T) {
	suite.Run(t, new(CORSTestSuite))
}

func (s *CORSTestSuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
	s.reached = false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		s.reached = true
		w.WriteHeader(http.StatusOK)
	})
	s.handler = interceptors.CORS([]string{testAllowedOrigin})(next)
}

func (s *CORSTestSuite) preflight(origin, method string) *http.Request {
	req := httptest.NewRequest(http.MethodOptions, testPath, nil)
	req.Header.Set(headerOrigin, origin)
	req.Header.Set(headerACRMethod, method)
	req.Header.Set(headerACRHeaders, constants.HeaderContentType)
	return req
}

func (s *CORSTestSuite) TestAllowedOriginIsEchoed() {
	req := httptest.NewRequest(http.MethodGet, testPath, nil)
	req.Header.Set(headerOrigin, testAllowedOrigin)

	s.handler.ServeHTTP(s.recorder, req)

	s.Equal(testAllowedOrigin, s.recorder.Header().Get(headerACAOrigin))
	s.True(s.reached)
}

func (s *CORSTestSuite) TestRequestIDHeaderIsExposed() {
	req := httptest.NewRequest(http.MethodGet, testPath, nil)
	req.Header.Set(headerOrigin, testAllowedOrigin)

	s.handler.ServeHTTP(s.recorder, req)

	// Header names are case-insensitive and the library canonicalises them.
	s.Contains(strings.ToLower(s.recorder.Header().Get(headerACExpose)), strings.ToLower(constants.HeaderRequestID))
}

func (s *CORSTestSuite) TestOtherOriginGetsNoAllowOriginHeader() {
	req := httptest.NewRequest(http.MethodGet, testPath, nil)
	req.Header.Set(headerOrigin, testOtherOrigin)

	s.handler.ServeHTTP(s.recorder, req)

	s.Empty(s.recorder.Header().Get(headerACAOrigin))
}

func (s *CORSTestSuite) TestRequestWithoutOriginPassesThroughWithoutCORSHeaders() {
	s.handler.ServeHTTP(s.recorder, httptest.NewRequest(http.MethodGet, testPath, nil))

	s.Empty(s.recorder.Header().Get(headerACAOrigin))
	s.True(s.reached)
}

func (s *CORSTestSuite) TestPreflightFromAllowedOriginIsAnsweredWithoutReachingHandler() {
	s.handler.ServeHTTP(s.recorder, s.preflight(testAllowedOrigin, http.MethodPost))

	// go-chi/cors answers an approved preflight with 200 and an empty body.
	s.Equal(http.StatusOK, s.recorder.Code)
	s.Equal(testAllowedOrigin, s.recorder.Header().Get(headerACAOrigin))
	s.Contains(s.recorder.Header().Get(headerACAMethods), wantPreflightMethods)
	s.Contains(s.recorder.Header().Get(headerACAHeaders), constants.HeaderContentType)
	s.False(s.reached)
}

func (s *CORSTestSuite) TestPreflightAdvertisesMaxAge() {
	s.handler.ServeHTTP(s.recorder, s.preflight(testAllowedOrigin, http.MethodPost))

	s.Equal(wantMaxAge, s.recorder.Header().Get(headerACMaxAge))
}

func (s *CORSTestSuite) TestPreflightForUnlistedMethodIsNotApproved() {
	s.handler.ServeHTTP(s.recorder, s.preflight(testAllowedOrigin, http.MethodPut))

	s.Empty(s.recorder.Header().Get(headerACAOrigin))
}

func (s *CORSTestSuite) TestPreflightFromOtherOriginIsNotApproved() {
	s.handler.ServeHTTP(s.recorder, s.preflight(testOtherOrigin, http.MethodPost))

	s.Empty(s.recorder.Header().Get(headerACAOrigin))
}
