package interceptors_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/httpx"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/interceptors"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/interceptors/mock"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

const (
	rateRemoteAddr      = "198.51.100.4:51234"
	rateRemoteHost      = "198.51.100.4"
	rateProxyHeader     = "X-Real-IP"
	rateForwardedHeader = "X-Forwarded-For"
	rateProxyClient     = "203.0.113.9"
	rateForwardedChain  = "203.0.113.9, 10.0.0.2"
	rateHeaderRetry     = "Retry-After"
	rateWaitFraction    = 1500 * time.Millisecond
	rateWaitTiny        = 20 * time.Millisecond
	rateWantRetryTwo    = "2"
	rateWantRetryOne    = "1"
)

type RateLimitTestSuite struct {
	suite.Suite
	ctrl     *gomock.Controller
	limiter  *mock.MockRateLimiter
	recorder *httptest.ResponseRecorder
	request  *http.Request
	reached  bool
}

func TestRateLimitTestSuite(t *testing.T) {
	suite.Run(t, new(RateLimitTestSuite))
}

func (s *RateLimitTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.limiter = mock.NewMockRateLimiter(s.ctrl)
	s.recorder = httptest.NewRecorder()
	s.request = httptest.NewRequest(http.MethodPost, testPath, nil)
	s.request.RemoteAddr = rateRemoteAddr
	s.reached = false
}

func (s *RateLimitTestSuite) handler(header string) http.Handler {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		s.reached = true
		w.WriteHeader(http.StatusOK)
	})
	return interceptors.RateLimit(s.limiter, header)(next)
}

func (s *RateLimitTestSuite) TestRateLimitWhenAllowedCallsNextHandler() {
	s.limiter.EXPECT().Allow(rateRemoteHost).Return(true, time.Duration(0)).Times(1)

	s.handler("").ServeHTTP(s.recorder, s.request)

	s.True(s.reached)
	s.Equal(http.StatusOK, s.recorder.Code)
}

func (s *RateLimitTestSuite) TestRateLimitWhenRejectedAnswers429WithRetryAfter() {
	s.limiter.EXPECT().Allow(rateRemoteHost).Return(false, rateWaitFraction).Times(1)

	s.handler("").ServeHTTP(s.recorder, s.request)

	s.False(s.reached)
	s.Equal(http.StatusTooManyRequests, s.recorder.Code)
	s.Equal(rateWantRetryTwo, s.recorder.Header().Get(rateHeaderRetry))
}

func (s *RateLimitTestSuite) TestRateLimitWhenRejectedUsesErrorShapeWithCode() {
	s.limiter.EXPECT().Allow(rateRemoteHost).Return(false, rateWaitFraction).Times(1)

	s.handler("").ServeHTTP(s.recorder, s.request)

	var body httpx.ErrorResponse
	s.Require().NoError(json.Unmarshal(s.recorder.Body.Bytes(), &body))
	s.Equal(apperror.CodeRateLimited, body.Error.Code)
	s.Equal(apperror.MsgRateLimited, body.Error.Message)
}

func (s *RateLimitTestSuite) TestRateLimitWhenWaitIsTinyRoundsRetryAfterUpToOneSecond() {
	s.limiter.EXPECT().Allow(rateRemoteHost).Return(false, rateWaitTiny).Times(1)

	s.handler("").ServeHTTP(s.recorder, s.request)

	s.Equal(rateWantRetryOne, s.recorder.Header().Get(rateHeaderRetry))
}

func (s *RateLimitTestSuite) TestRateLimitWhenProxyHeaderConfiguredKeysOnHeader() {
	s.request.Header.Set(rateProxyHeader, rateProxyClient)
	s.limiter.EXPECT().Allow(rateProxyClient).Return(true, time.Duration(0)).Times(1)

	s.handler(rateProxyHeader).ServeHTTP(s.recorder, s.request)

	s.True(s.reached)
}

func (s *RateLimitTestSuite) TestRateLimitWhenForwardedChainKeysOnFirstAddress() {
	s.request.Header.Set(rateForwardedHeader, rateForwardedChain)
	s.limiter.EXPECT().Allow(rateProxyClient).Return(true, time.Duration(0)).Times(1)

	s.handler(rateForwardedHeader).ServeHTTP(s.recorder, s.request)

	s.True(s.reached)
}

func (s *RateLimitTestSuite) TestRateLimitWhenConfiguredHeaderMissingFallsBackToConnection() {
	s.limiter.EXPECT().Allow(rateRemoteHost).Return(true, time.Duration(0)).Times(1)

	s.handler(rateProxyHeader).ServeHTTP(s.recorder, s.request)

	s.True(s.reached)
}

func (s *RateLimitTestSuite) TestRateLimitWhenHeaderNotConfiguredIgnoresIt() {
	s.request.Header.Set(rateProxyHeader, rateProxyClient)
	s.limiter.EXPECT().Allow(rateRemoteHost).Return(true, time.Duration(0)).Times(1)

	s.handler("").ServeHTTP(s.recorder, s.request)

	s.True(s.reached)
}

func (s *RateLimitTestSuite) TestRateLimitWhenRemoteAddrHasNoPortUsesItWhole() {
	s.request.RemoteAddr = rateRemoteHost
	s.limiter.EXPECT().Allow(rateRemoteHost).Return(true, time.Duration(0)).Times(1)

	s.handler("").ServeHTTP(s.recorder, s.request)

	s.True(s.reached)
}
