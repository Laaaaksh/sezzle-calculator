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
	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/boot"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/config"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/httpx"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/health"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

// CalculateE2ETestSuite drives the real router, the real handler and the real
// core with the default operations. Nothing is mocked, so it proves the pieces
// agree with each other and with docs/API.md.
type CalculateE2ETestSuite struct {
	suite.Suite
	router   chi.Router
	recorder *httptest.ResponseRecorder
}

func TestCalculateE2ETestSuite(t *testing.T) {
	suite.Run(t, new(CalculateE2ETestSuite))
}

func (s *CalculateE2ETestSuite) SetupTest() {
	calc := config.CalculatorConfig{
		MaxOperandLength: e2eMaxOperandLength,
		MaxExponent:      e2eMaxExponent,
		DecimalPrecision: e2ePrecision,
	}
	cfg := &config.Config{
		Server:     config.ServerConfig{MaxBodyBytes: e2eBodyLimit, RequestTimeout: testTimeout},
		CORS:       config.CORSConfig{AllowedOrigins: []string{testOrigin}},
		Calculator: calc,
	}
	healthCore := health.NewCore()
	modules := boot.Modules{
		Health:     &health.Module{Core: healthCore, Handler: health.NewHTTPHandler(healthCore)},
		Calculator: calculator.NewModule(context.Background(), calc),
	}
	s.router = boot.NewRouter(cfg, modules)
	s.recorder = httptest.NewRecorder()
}

func (s *CalculateE2ETestSuite) postJSON(contentType, body string) {
	req := httptest.NewRequest(http.MethodPost, testCalculatePath, strings.NewReader(body))
	if contentType != "" {
		req.Header.Set(constants.HeaderContentType, contentType)
	}
	s.router.ServeHTTP(s.recorder, req)
}

func (s *CalculateE2ETestSuite) errorBody() httpx.ErrorBody {
	var got httpx.ErrorResponse
	s.Require().NoError(json.Unmarshal(s.recorder.Body.Bytes(), &got))
	return got.Error
}

func (s *CalculateE2ETestSuite) TestPointOnePlusPointTwoIsExactlyPointThree() {
	s.postJSON(wantContentType, e2eBodyPointOnePlusPointTwo)

	s.Equal(http.StatusOK, s.recorder.Code)
	s.JSONEq(e2eWantPointThree, s.recorder.Body.String())
	s.Equal(wantContentType, s.recorder.Header().Get(constants.HeaderContentType))
	s.Equal(e2eNoStore, s.recorder.Header().Get(constants.HeaderCacheControl))
}

func (s *CalculateE2ETestSuite) TestDivideReturnsTheContractExample() {
	s.postJSON(wantContentType, e2eBodyDivide)

	s.Equal(http.StatusOK, s.recorder.Code)
	s.JSONEq(e2eWantDivide, s.recorder.Body.String())
}

func (s *CalculateE2ETestSuite) TestDivisionByZeroReturns422WithRequestID() {
	s.postJSON(wantContentType, e2eBodyDivideByZero)

	body := s.errorBody()
	s.Equal(http.StatusUnprocessableEntity, s.recorder.Code)
	s.Equal(entities.CodeDivisionByZero, body.Code)
	s.Equal(entities.MsgDivisionByZero, body.Message)
	s.NotEmpty(body.RequestID)
	s.Equal(body.RequestID, s.recorder.Header().Get(wantHeaderTemplate))
}

func (s *CalculateE2ETestSuite) TestUnknownFieldReturns400() {
	s.postJSON(wantContentType, e2eBodyUnknownField)

	s.Equal(http.StatusBadRequest, s.recorder.Code)
	s.Equal(apperror.CodeBadRequest, s.errorBody().Code)
}

func (s *CalculateE2ETestSuite) TestWrongContentTypeReturns415() {
	s.postJSON(e2eTextPlain, e2eBodyDivide)

	s.Equal(http.StatusUnsupportedMediaType, s.recorder.Code)
	s.Equal(apperror.CodeUnsupportedMediaType, s.errorBody().Code)
}

func (s *CalculateE2ETestSuite) TestMissingFieldsReturn422WithFields() {
	s.postJSON(wantContentType, e2eBodyEmptyObject)

	body := s.errorBody()
	s.Equal(http.StatusUnprocessableEntity, s.recorder.Code)
	s.Equal(apperror.CodeValidationFailed, body.Code)
	s.Len(body.Fields, e2eExpectedFieldCount)
}

func (s *CalculateE2ETestSuite) TestThreeOperandsForAddReturnInvalidOperandCount() {
	s.postJSON(wantContentType, e2eBodyThreeOperands)

	s.Equal(http.StatusUnprocessableEntity, s.recorder.Code)
	s.Equal(entities.CodeInvalidOperandCount, s.errorBody().Code)
}

func (s *CalculateE2ETestSuite) TestSeventeenOperandsReturn422BeforeAnyParsing() {
	s.postJSON(wantContentType, e2eBodyTooManyOperands)

	body := s.errorBody()
	s.Equal(http.StatusUnprocessableEntity, s.recorder.Code)
	s.Equal(apperror.CodeValidationFailed, body.Code)
}

func (s *CalculateE2ETestSuite) TestCaseVariantDuplicateKeyReturns400AndComputesNothing() {
	s.postJSON(wantContentType, e2eBodyCaseVariantKey)

	body := s.errorBody()
	s.Equal(http.StatusBadRequest, s.recorder.Code)
	s.Equal(apperror.CodeBadRequest, body.Code)
	s.Contains(body.Fields, e2eDuplicateKeyField)
}

func (s *CalculateE2ETestSuite) TestOperationsPublishesTheConfiguredLimits() {
	s.router.ServeHTTP(s.recorder, httptest.NewRequest(http.MethodGet, testOperationsPath, nil))

	var got map[string]json.RawMessage
	s.Require().NoError(json.Unmarshal(s.recorder.Body.Bytes(), &got))
	s.JSONEq(e2eWantLimitsJSON, string(got[e2eLimitsKey]))
}

func (s *CalculateE2ETestSuite) TestOperationsListsEverySupportedOperationInOrder() {
	s.router.ServeHTTP(s.recorder, httptest.NewRequest(http.MethodGet, testOperationsPath, nil))

	var got entities.OperationsResponse
	s.Require().NoError(json.Unmarshal(s.recorder.Body.Bytes(), &got))
	names := make([]entities.OperationType, len(got.Operations))
	for i, info := range got.Operations {
		names[i] = info.Name
	}
	s.Equal(http.StatusOK, s.recorder.Code)
	s.Equal([]entities.OperationType{
		entities.OpAdd, entities.OpSubtract, entities.OpMultiply, entities.OpDivide,
		entities.OpPower, entities.OpSqrt, entities.OpPercentage,
	}, names)
}

func (s *CalculateE2ETestSuite) TestGetOnCalculateReturns405() {
	s.router.ServeHTTP(s.recorder, httptest.NewRequest(http.MethodGet, testCalculatePath, nil))

	s.Equal(http.StatusMethodNotAllowed, s.recorder.Code)
	s.Equal(apperror.CodeMethodNotAllowed, s.errorBody().Code)
}

func (s *CalculateE2ETestSuite) TestShortUnknownOperationIsEchoedInFields() {
	s.postJSON(wantContentType, e2eBodyUnknownOperation)

	body := s.errorBody()
	s.Equal(http.StatusUnprocessableEntity, s.recorder.Code)
	s.Equal(entities.CodeUnknownOperation, body.Code)
	s.Equal(map[string]any{entities.FieldOperation: e2eUnknownOperationName}, body.Fields)
}

func (s *CalculateE2ETestSuite) TestHugeOperationNameIsRejectedWithoutEchoingIt() {
	huge := strings.Repeat("x", e2eHugeOperationLength)
	s.postJSON(wantContentType, `{"operation":"`+huge+`","operands":["1","2"]}`)

	body := s.errorBody()
	s.Equal(http.StatusUnprocessableEntity, s.recorder.Code)
	s.Equal(apperror.CodeValidationFailed, body.Code)
	s.Equal(map[string]any{entities.FieldNameOperation: entities.ProblemTooLong}, body.Fields)
	s.NotContains(s.recorder.Body.String(), huge[:entities.MaxOperationNameLength+1])
	s.Less(s.recorder.Body.Len(), e2eSmallErrorBodyBytes)
}

func (s *CalculateE2ETestSuite) TestExpiredDeadlineReturns503RequestTimeoutAsJSON() {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	req := httptest.NewRequest(http.MethodPost, testCalculatePath, strings.NewReader(e2eBodyDivide)).WithContext(ctx)
	req.Header.Set(constants.HeaderContentType, wantContentType)

	s.router.ServeHTTP(s.recorder, req)

	body := s.errorBody()
	s.Equal(http.StatusServiceUnavailable, s.recorder.Code)
	s.Equal(apperror.CodeRequestTimeout, body.Code)
	s.Equal(apperror.MsgRequestTimeout, body.Message)
	s.Equal(wantContentType, s.recorder.Header().Get(constants.HeaderContentType))
}
