package boot_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/boot"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/config"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/httpx"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/operations"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/health"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

const (
	// Limits wide enough that 7^4000000 (about 3.4 million digits) passes every
	// size guard, so only the request deadline can stop it.
	slowMaxOperandLength = 1_000_000
	slowMaxExponent      = 5_000_000
	slowPrecision        = 20
	slowBodyLimit        = 1 << 20
	slowRequestTimeout   = 20 * time.Millisecond
	// Far below the time the full power takes, far above the deadline.
	slowMaxElapsed = 1500 * time.Millisecond
	slowPowerBody  = `{"operation":"power","operands":["7","4000000"]}`
)

// SlowCalculationTimeoutE2ETestSuite drives the real router, with its real
// Timeout interceptor, into a power computation that runs far longer than the
// configured request timeout, and checks the work is stopped with the JSON
// REQUEST_TIMEOUT error instead of running to completion.
type SlowCalculationTimeoutE2ETestSuite struct {
	suite.Suite
	router chi.Router
}

func TestSlowCalculationTimeoutE2ETestSuite(t *testing.T) {
	suite.Run(t, new(SlowCalculationTimeoutE2ETestSuite))
}

func (s *SlowCalculationTimeoutE2ETestSuite) SetupTest() {
	calc := config.CalculatorConfig{
		MaxOperandLength: slowMaxOperandLength,
		MaxExponent:      slowMaxExponent,
		DecimalPrecision: slowPrecision,
	}
	cfg := &config.Config{
		Server:     config.ServerConfig{MaxBodyBytes: slowBodyLimit, RequestTimeout: slowRequestTimeout},
		CORS:       config.CORSConfig{AllowedOrigins: []string{testOrigin}},
		Calculator: calc,
	}
	core := calculator.NewCore(operations.NewDefaultRegistry(calc), calc)
	healthCore := health.NewCore()
	s.router = boot.NewRouter(cfg, boot.Modules{
		Health:     &health.Module{Core: healthCore, Handler: health.NewHTTPHandler(healthCore)},
		Calculator: &calculator.Module{Core: core, Handler: calculator.NewHTTPHandler(core, calc)},
	})
}

func (s *SlowCalculationTimeoutE2ETestSuite) TestLongPowerIsStoppedByTheRequestTimeout() {
	req := httptest.NewRequest(http.MethodPost, testCalculatePath, strings.NewReader(slowPowerBody))
	req.Header.Set(constants.HeaderContentType, wantContentType)
	rec := httptest.NewRecorder()

	start := time.Now()
	s.router.ServeHTTP(rec, req)
	elapsed := time.Since(start)

	var got httpx.ErrorResponse
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &got), rec.Body.String())
	s.Equal(http.StatusServiceUnavailable, rec.Code)
	s.Equal(apperror.CodeRequestTimeout, got.Error.Code)
	s.Equal(wantContentType, rec.Header().Get(constants.HeaderContentType))
	s.Less(elapsed, slowMaxElapsed)
}
