package boot_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/boot"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/config"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/httpx"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/health"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

const (
	limitBurst         = 2
	limitSlowRate      = 0.001
	limitProxyHeader   = "X-Real-IP"
	limitClientOne     = "203.0.113.1"
	limitClientTwo     = "203.0.113.2"
	limitWantRetryWait = "1000"
)

// RateLimitE2ETestSuite drives the real router with the limiter switched on.
// The refill rate is so slow that no token comes back during a test.
type RateLimitE2ETestSuite struct {
	suite.Suite
	router chi.Router
}

func TestRateLimitE2ETestSuite(t *testing.T) {
	suite.Run(t, new(RateLimitE2ETestSuite))
}

func (s *RateLimitE2ETestSuite) SetupTest() {
	calc := config.CalculatorConfig{
		MaxOperandLength: e2eMaxOperandLength,
		MaxExponent:      e2eMaxExponent,
		DecimalPrecision: e2ePrecision,
	}
	cfg := &config.Config{
		Server:     config.ServerConfig{MaxBodyBytes: e2eBodyLimit, RequestTimeout: testTimeout},
		CORS:       config.CORSConfig{AllowedOrigins: []string{testOrigin}},
		Calculator: calc,
		RateLimit: config.RateLimitConfig{
			Enabled:           true,
			RequestsPerSecond: limitSlowRate,
			Burst:             limitBurst,
			ClientIPHeader:    limitProxyHeader,
		},
	}
	healthCore := health.NewCore()
	s.router = boot.NewRouter(cfg, boot.Modules{
		Health:     &health.Module{Core: healthCore, Handler: health.NewHTTPHandler(healthCore)},
		Calculator: calculator.NewModule(context.Background(), calc),
	})
}

func (s *RateLimitE2ETestSuite) do(method, path, client string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(testAddBody))
	req.Header.Set(constants.HeaderContentType, wantContentType)
	req.Header.Set(limitProxyHeader, client)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, req)
	return recorder
}

func (s *RateLimitE2ETestSuite) spendBurst(client string) {
	for range limitBurst {
		s.Require().Equal(http.StatusOK, s.do(http.MethodPost, testCalculatePath, client).Code)
	}
}

func (s *RateLimitE2ETestSuite) TestCalculateWhenBurstIsSpentReturns429RateLimited() {
	s.spendBurst(limitClientOne)

	recorder := s.do(http.MethodPost, testCalculatePath, limitClientOne)

	var body httpx.ErrorResponse
	s.Require().NoError(json.Unmarshal(recorder.Body.Bytes(), &body))
	s.Equal(http.StatusTooManyRequests, recorder.Code)
	s.Equal(apperror.CodeRateLimited, body.Error.Code)
	s.NotEmpty(body.Error.RequestID)
	s.Equal(limitWantRetryWait, recorder.Header().Get(constants.HeaderRetryAfter))
}

func (s *RateLimitE2ETestSuite) TestCalculateWhenOneClientIsLimitedAnotherStillWorks() {
	s.spendBurst(limitClientOne)

	recorder := s.do(http.MethodPost, testCalculatePath, limitClientTwo)

	s.Equal(http.StatusOK, recorder.Code)
}

func (s *RateLimitE2ETestSuite) TestOperationsWhenClientIsLimitedStillAnswers() {
	s.spendBurst(limitClientOne)

	recorder := s.do(http.MethodGet, testOperationsPath, limitClientOne)

	s.Equal(http.StatusOK, recorder.Code)
}
