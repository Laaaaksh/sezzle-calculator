package boot_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/boot"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/config"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/httpx"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/operations"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/health"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

const (
	slowBodyLimit      = 1 << 20
	slowRequestTimeout = 20 * time.Millisecond
	slowOperationName  = entities.OperationType("spin")
	slowSymbol         = "~"
	slowDescription    = "Runs until its context ends."
	slowBody           = `{"operation":"spin","operands":["1"]}`
)

var slowLimits = entities.Limits{MaxOperandLength: 64, MaxExponent: 1000, DecimalPrecision: 20, MaxResultDigits: 640}

// spinOperation works until its context ends, the way power and square root
// do. It can only return once the request deadline cancels the context, so
// the test needs no stopwatch: if the timeout never reached the work, the run
// would hang and the test binary's own timeout would fail it.
type spinOperation struct{}

func (spinOperation) Name() entities.OperationType { return slowOperationName }
func (spinOperation) Arity() int                   { return entities.ArityUnary }
func (spinOperation) Symbol() string               { return slowSymbol }
func (spinOperation) Description() string          { return slowDescription }

func (spinOperation) Apply(ctx context.Context, _ []decimal.Decimal) (decimal.Decimal, apperror.IError) {
	for {
		if aerr := operations.CheckContext(ctx); aerr != nil {
			return decimal.Zero, aerr
		}
	}
}

// SlowCalculationTimeoutE2ETestSuite drives the real router, with its real
// Timeout interceptor, real handler and real core, into an operation that
// never finishes by itself, and checks the work is stopped with the JSON
// REQUEST_TIMEOUT error.
type SlowCalculationTimeoutE2ETestSuite struct {
	suite.Suite
	router chi.Router
}

func TestSlowCalculationTimeoutE2ETestSuite(t *testing.T) {
	suite.Run(t, new(SlowCalculationTimeoutE2ETestSuite))
}

func (s *SlowCalculationTimeoutE2ETestSuite) SetupTest() {
	cfg := &config.Config{
		Server: config.ServerConfig{MaxBodyBytes: slowBodyLimit, RequestTimeout: slowRequestTimeout},
		CORS:   config.CORSConfig{AllowedOrigins: []string{testOrigin}},
	}
	core := calculator.NewCore(operations.NewRegistry(spinOperation{}), slowLimits)
	healthCore := health.NewCore()
	s.router = boot.NewRouter(cfg, boot.Modules{
		Health:     &health.Module{Core: healthCore, Handler: health.NewHTTPHandler(healthCore)},
		Calculator: &calculator.Module{Core: core, Handler: calculator.NewHTTPHandler(core, slowLimits)},
	})
}

func (s *SlowCalculationTimeoutE2ETestSuite) TestEndlessOperationIsStoppedByTheRequestTimeout() {
	req := httptest.NewRequest(http.MethodPost, testCalculatePath, strings.NewReader(slowBody))
	req.Header.Set(constants.HeaderContentType, wantContentType)
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, req)

	var got httpx.ErrorResponse
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &got), rec.Body.String())
	s.Equal(http.StatusServiceUnavailable, rec.Code)
	s.Equal(apperror.CodeRequestTimeout, got.Error.Code)
	s.Equal(wantContentType, rec.Header().Get(constants.HeaderContentType))
}
