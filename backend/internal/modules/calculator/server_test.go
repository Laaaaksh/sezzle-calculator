package calculator_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/contextkeys"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/httpx"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/mock"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

var testHandlerLimits = entities.Limits{
	MaxOperandLength: testHandlerMaxLen,
	MaxExponent:      testHandlerMaxExponent,
	DecimalPrecision: testHandlerPrecision,
	MaxResultDigits:  testHandlerResultDigits,
}

type ServerTestSuite struct {
	suite.Suite
	ctx      context.Context
	ctrl     *gomock.Controller
	mockCore *mock.MockICore
	handler  *calculator.HTTPHandler
	recorder *httptest.ResponseRecorder
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}

func (s *ServerTestSuite) SetupTest() {
	s.ctx = contextkeys.WithRequestID(context.Background(), testRequestID)
	s.ctrl = gomock.NewController(s.T())
	s.mockCore = mock.NewMockICore(s.ctrl)
	s.recorder = httptest.NewRecorder()
	s.handler = calculator.NewHTTPHandler(s.mockCore, testHandlerLimits)
}

func (s *ServerTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

// post calls Calculate directly so the mock sees the exact request context;
// going through chi would wrap it with routing state.
func (s *ServerTestSuite) post(contentType, body string) {
	s.postReader(contentType, strings.NewReader(body))
}

func (s *ServerTestSuite) postReader(contentType string, body io.Reader) {
	req := httptest.NewRequest(http.MethodPost, testCalcPath, body).WithContext(s.ctx)
	if contentType != contentTypeAbsent {
		req.Header.Set(constants.HeaderContentType, contentType)
	}
	s.handler.Calculate(s.recorder, req)
}

func (s *ServerTestSuite) errorBody() httpx.ErrorBody {
	var got httpx.ErrorResponse
	s.Require().NoError(json.Unmarshal(s.recorder.Body.Bytes(), &got))
	return got.Error
}

func (s *ServerTestSuite) requireError(status int, code apperror.Code) httpx.ErrorBody {
	body := s.errorBody()
	s.Equal(status, s.recorder.Code)
	s.Equal(code, body.Code)
	s.Equal(code.DefaultMessage(), body.Message)
	s.Equal(testRequestID, body.RequestID)
	s.requireJSONHeaders()
	return body
}

func (s *ServerTestSuite) requireJSONHeaders() {
	s.Equal(contentTypeJSON, s.recorder.Header().Get(constants.HeaderContentType))
	s.Equal(cacheNoStore, s.recorder.Header().Get(constants.HeaderCacheControl))
}

func (s *ServerTestSuite) expectCalculate(op entities.OperationType, operands []decimal.Decimal, text string) {
	s.mockCore.EXPECT().
		Calculate(gomock.Eq(s.ctx), gomock.Eq(op), gomock.Eq(operands)).
		Return(entities.Result{Text: text}, nil).
		Times(1)
}

func (s *ServerTestSuite) expectCalculateError(op entities.OperationType, operands []decimal.Decimal, aerr apperror.IError) {
	s.mockCore.EXPECT().
		Calculate(gomock.Eq(s.ctx), gomock.Eq(op), gomock.Eq(operands)).
		Return(entities.Result{}, aerr).
		Times(1)
}

func (s *ServerTestSuite) TestCalculateReturns200WithResultAndEchoedRequest() {
	s.expectCalculate(entities.OpDivide, []decimal.Decimal{mustDecimal(handlerNum10), mustDecimal(handlerNum4)}, wantResultText)

	s.post(contentTypeJSON, bodyDivide)

	s.Equal(http.StatusOK, s.recorder.Code)
	s.JSONEq(wantDivideResp, s.recorder.Body.String())
}

func (s *ServerTestSuite) TestCalculateSuccessSetsContentTypeAndNoStore() {
	s.expectCalculate(entities.OpDivide, []decimal.Decimal{mustDecimal(handlerNum10), mustDecimal(handlerNum4)}, wantResultText)

	s.post(contentTypeJSON, bodyDivide)

	s.requireJSONHeaders()
}

func (s *ServerTestSuite) TestCalculateEchoesTheOriginalOperandStrings() {
	s.expectCalculate(entities.OpDivide, []decimal.Decimal{mustDecimal(handlerNum10Pt0), mustDecimal(handlerNum4Pt00)}, wantResultText)

	s.post(contentTypeJSON, bodyDivideSpaced)

	s.JSONEq(wantSpacedResp, s.recorder.Body.String())
}

func (s *ServerTestSuite) TestCalculateAcceptsOneOperandForUnaryOperation() {
	s.expectCalculate(entities.OpSqrt, []decimal.Decimal{mustDecimal(handlerNum144)}, wantSqrtText)

	s.post(contentTypeJSON, bodySqrt)

	s.Equal(http.StatusOK, s.recorder.Code)
	s.JSONEq(wantSqrtResp, s.recorder.Body.String())
}

func (s *ServerTestSuite) TestCalculateAcceptsUTF8CharsetParameter() {
	s.expectCalculate(entities.OpDivide, []decimal.Decimal{mustDecimal(handlerNum10), mustDecimal(handlerNum4)}, wantResultText)

	s.post(contentTypeJSONUTF8, bodyDivide)

	s.Equal(http.StatusOK, s.recorder.Code)
}

func (s *ServerTestSuite) TestCalculateAcceptsMediaTypeInAnyCase() {
	s.expectCalculate(entities.OpDivide, []decimal.Decimal{mustDecimal(handlerNum10), mustDecimal(handlerNum4)}, wantResultText)

	s.post(contentTypeMixedCase, bodyDivide)

	s.Equal(http.StatusOK, s.recorder.Code)
}

func (s *ServerTestSuite) TestCalculateAcceptsWhitespaceAfterTheObject() {
	s.expectCalculate(entities.OpDivide, []decimal.Decimal{mustDecimal(handlerNum10), mustDecimal(handlerNum4)}, wantResultText)

	s.post(contentTypeJSON, bodyTrailingSpace)

	s.Equal(http.StatusOK, s.recorder.Code)
}

func (s *ServerTestSuite) TestCalculateMissingContentTypeReturns415() {
	s.post(contentTypeAbsent, bodyDivide)

	s.requireError(http.StatusUnsupportedMediaType, apperror.CodeUnsupportedMediaType)
}

func (s *ServerTestSuite) TestCalculateWrongContentTypeReturns415() {
	s.post(contentTypePlain, bodyDivide)

	s.requireError(http.StatusUnsupportedMediaType, apperror.CodeUnsupportedMediaType)
}

func (s *ServerTestSuite) TestCalculateUnreadableContentTypeReturns415() {
	s.post(contentTypeUnreadable, bodyDivide)

	s.requireError(http.StatusUnsupportedMediaType, apperror.CodeUnsupportedMediaType)
}

func (s *ServerTestSuite) TestCalculateNonUTF8CharsetReturns415() {
	s.post(contentTypeJSONLatin, bodyDivide)

	s.requireError(http.StatusUnsupportedMediaType, apperror.CodeUnsupportedMediaType)
}

func (s *ServerTestSuite) TestCalculateEmptyBodyReturns400() {
	s.post(contentTypeJSON, bodyEmpty)

	s.requireError(http.StatusBadRequest, apperror.CodeBadRequest)
}

func (s *ServerTestSuite) TestCalculateWhitespaceOnlyBodyReturns400() {
	s.post(contentTypeJSON, bodyBlank)

	s.requireError(http.StatusBadRequest, apperror.CodeBadRequest)
}

func (s *ServerTestSuite) TestCalculateMalformedJSONReturns400() {
	s.post(contentTypeJSON, bodyNotJSON)

	s.requireError(http.StatusBadRequest, apperror.CodeBadRequest)
}

func (s *ServerTestSuite) TestCalculateCutOffJSONReturns400() {
	s.post(contentTypeJSON, bodyCutOff)

	s.requireError(http.StatusBadRequest, apperror.CodeBadRequest)
}

func (s *ServerTestSuite) TestCalculateUnknownFieldReturns400() {
	s.post(contentTypeJSON, bodyUnknownField)

	s.requireError(http.StatusBadRequest, apperror.CodeBadRequest)
}

func (s *ServerTestSuite) TestCalculateThreeOperandsForAddStillReachesTheCore() {
	operands := []decimal.Decimal{mustDecimal("1"), mustDecimal("2"), mustDecimal("3")}
	s.expectCalculateError(entities.OpAdd, operands, apperror.New(entities.CodeInvalidOperandCount, nil))

	s.post(contentTypeJSON, bodyThreeOperands)

	s.requireError(http.StatusUnprocessableEntity, entities.CodeInvalidOperandCount)
}

func (s *ServerTestSuite) TestCalculateMoreOperandsThanTheCapReturns422BeforeParsing() {
	s.post(contentTypeJSON, bodyTooManyOperands)

	body := s.requireError(http.StatusUnprocessableEntity, apperror.CodeValidationFailed)
	s.Equal(map[string]any{fieldOperands: problemMany}, body.Fields)
}

func (s *ServerTestSuite) TestCalculateRepeatedKeyReturns400NamingTheKey() {
	s.post(contentTypeJSON, bodyDuplicateKey)

	body := s.requireError(http.StatusBadRequest, apperror.CodeBadRequest)
	s.Equal(map[string]any{fieldKey: fieldOperands, entities.FieldReason: reasonDup}, body.Fields)
}

func (s *ServerTestSuite) TestCalculateKeyRepeatedInAnotherCaseReturns400NamingTheVariant() {
	s.post(contentTypeJSON, bodyCaseVariantDup)

	body := s.requireError(http.StatusBadRequest, apperror.CodeBadRequest)
	s.Equal(map[string]any{fieldKey: "OPERANDS", entities.FieldReason: reasonCase}, body.Fields)
}

func (s *ServerTestSuite) TestCalculateKeyInTheWrongCaseReturns400() {
	s.post(contentTypeJSON, bodyWrongCaseKey)

	body := s.requireError(http.StatusBadRequest, apperror.CodeBadRequest)
	s.Equal(map[string]any{fieldKey: "Operation", entities.FieldReason: reasonCase}, body.Fields)
}

func (s *ServerTestSuite) TestCalculateRepeatedUnknownKeyIsLeftToTheUnknownFieldCheck() {
	s.post(contentTypeJSON, bodyDuplicateUnk)

	body := s.requireError(http.StatusBadRequest, apperror.CodeBadRequest)
	s.Empty(body.Fields)
}

func (s *ServerTestSuite) TestCalculateBodyWithEachKeyOnceStillSucceeds() {
	s.expectCalculate(entities.OpAdd, []decimal.Decimal{mustDecimal(handlerNum1), mustDecimal(handlerNum2)}, wantAddText)

	s.post(contentTypeJSON, bodyAdd)

	s.Equal(http.StatusOK, s.recorder.Code)
	s.JSONEq(wantAddEcho, s.recorder.Body.String())
}

func (s *ServerTestSuite) TestCalculateNumericOperandsReturn400() {
	s.post(contentTypeJSON, bodyNumericOperand)

	s.requireError(http.StatusBadRequest, apperror.CodeBadRequest)
}

func (s *ServerTestSuite) TestCalculateNumericOperationReturns400() {
	s.post(contentTypeJSON, bodyNumericOp)

	s.requireError(http.StatusBadRequest, apperror.CodeBadRequest)
}

func (s *ServerTestSuite) TestCalculateTopLevelArrayReturns400() {
	s.post(contentTypeJSON, bodyArray)

	s.requireError(http.StatusBadRequest, apperror.CodeBadRequest)
}

func (s *ServerTestSuite) TestCalculateSecondJSONObjectReturns400() {
	s.post(contentTypeJSON, bodyTwoObjects)

	s.requireError(http.StatusBadRequest, apperror.CodeBadRequest)
}

func (s *ServerTestSuite) TestCalculateTrailingTextReturns400() {
	s.post(contentTypeJSON, bodyTrailingText)

	s.requireError(http.StatusBadRequest, apperror.CodeBadRequest)
}

func (s *ServerTestSuite) TestCalculateBodyOverTheLimitInsideTheObjectReturns413() {
	limited := http.MaxBytesReader(s.recorder, io.NopCloser(strings.NewReader(bodyDivide)), testOversizeLimit)

	s.postReader(contentTypeJSON, limited)

	body := s.requireError(http.StatusRequestEntityTooLarge, apperror.CodePayloadTooLarge)
	s.Equal(map[string]any{entities.FieldLimit: float64(testOversizeLimit)}, body.Fields)
}

func (s *ServerTestSuite) TestCalculateBodyOverTheLimitAfterTheObjectReturns413() {
	padded := bodyDivide + strings.Repeat(" ", testPadding)
	limited := http.MaxBytesReader(s.recorder, io.NopCloser(strings.NewReader(padded)), testTrailingLimit)

	s.postReader(contentTypeJSON, limited)

	s.requireError(http.StatusRequestEntityTooLarge, apperror.CodePayloadTooLarge)
}

func (s *ServerTestSuite) TestCalculateEmptyObjectReturns422NamingBothFields() {
	s.post(contentTypeJSON, bodyNoFields)

	body := s.requireError(http.StatusUnprocessableEntity, apperror.CodeValidationFailed)
	s.Equal(map[string]any{fieldOperation: problemText, fieldOperands: problemText}, body.Fields)
}

func (s *ServerTestSuite) TestCalculateNullBodyReturns422NamingBothFields() {
	s.post(contentTypeJSON, bodyNull)

	body := s.requireError(http.StatusUnprocessableEntity, apperror.CodeValidationFailed)
	s.Equal(map[string]any{fieldOperation: problemText, fieldOperands: problemText}, body.Fields)
}

func (s *ServerTestSuite) TestCalculateMissingOperationReturns422NamingOperation() {
	s.post(contentTypeJSON, bodyNoOperation)

	body := s.requireError(http.StatusUnprocessableEntity, apperror.CodeValidationFailed)
	s.Equal(map[string]any{fieldOperation: problemText}, body.Fields)
}

func (s *ServerTestSuite) TestCalculateBlankOperationReturns422NamingOperation() {
	s.post(contentTypeJSON, bodyBlankOperation)

	body := s.requireError(http.StatusUnprocessableEntity, apperror.CodeValidationFailed)
	s.Equal(map[string]any{fieldOperation: problemText}, body.Fields)
}

func (s *ServerTestSuite) postOperationOfLength(n int) {
	s.post(contentTypeJSON, bodyOperationHead+strings.Repeat("a", n)+bodyOperationTail)
}

func (s *ServerTestSuite) TestCalculateOperationOverTheLengthBoundReturns422WithoutCallingCore() {
	s.postOperationOfLength(testLongOperationLen)

	body := s.requireError(http.StatusUnprocessableEntity, apperror.CodeValidationFailed)
	s.Equal(map[string]any{fieldOperation: problemTooLong}, body.Fields)
}

func (s *ServerTestSuite) TestCalculateOneMiBOperationReturns422WithASmallBody() {
	s.postOperationOfLength(testHugeOperationLen)

	body := s.requireError(http.StatusUnprocessableEntity, apperror.CodeValidationFailed)
	s.Equal(map[string]any{fieldOperation: problemTooLong}, body.Fields)
	s.Less(s.recorder.Body.Len(), testSmallBodyLimit)
}

func (s *ServerTestSuite) TestCalculateOperationAtTheLengthBoundReachesCore() {
	name := entities.OperationType(strings.Repeat("a", entities.MaxOperationNameLength))
	s.expectCalculate(name, []decimal.Decimal{decimal.NewFromInt(1), decimal.NewFromInt(2)}, wantResultText)

	s.postOperationOfLength(entities.MaxOperationNameLength)

	s.Equal(http.StatusOK, s.recorder.Code)
}

func (s *ServerTestSuite) TestCalculateMissingOperandsReturns422NamingOperands() {
	s.post(contentTypeJSON, bodyNoOperands)

	body := s.requireError(http.StatusUnprocessableEntity, apperror.CodeValidationFailed)
	s.Equal(map[string]any{fieldOperands: problemText}, body.Fields)
}

func (s *ServerTestSuite) TestCalculateNullOperandsReturns422NamingOperands() {
	s.post(contentTypeJSON, bodyNullOperands)

	body := s.requireError(http.StatusUnprocessableEntity, apperror.CodeValidationFailed)
	s.Equal(map[string]any{fieldOperands: problemText}, body.Fields)
}

func (s *ServerTestSuite) TestCalculateInvalidOperandReturns422WithoutCallingCore() {
	s.post(contentTypeJSON, bodyBadOperand)

	body := s.requireError(http.StatusUnprocessableEntity, entities.CodeInvalidOperand)
	s.Equal(map[string]any{entities.FieldOperandIndex: float64(testIndexSecond), entities.FieldReason: entities.ReasonExponent}, body.Fields)
	s.NotContains(s.recorder.Body.String(), calculator.ErrOperandSyntax.Error())
}

func (s *ServerTestSuite) TestCalculateOperandLongerThanConfiguredLimitReturns422() {
	s.post(contentTypeJSON, bodyLongOperand)

	body := s.requireError(http.StatusUnprocessableEntity, entities.CodeOperandTooLarge)
	s.Equal(map[string]any{entities.FieldOperandIndex: float64(testIndexSecond), entities.FieldLimit: float64(testHandlerMaxLen)}, body.Fields)
	s.NotContains(s.recorder.Body.String(), calculator.ErrOperandTooLong.Error())
}

func (s *ServerTestSuite) TestCalculateEmptyOperandsArrayReachesCoreAndReturnsOperandCountError() {
	s.expectCalculateError(entities.OpAdd, []decimal.Decimal{}, apperror.New(entities.CodeInvalidOperandCount, nil))

	s.post(contentTypeJSON, bodyEmptyOperands)

	s.requireError(http.StatusUnprocessableEntity, entities.CodeInvalidOperandCount)
}

func (s *ServerTestSuite) TestCalculateMapsDomainErrorFromCoreTo422() {
	divisor := []decimal.Decimal{mustDecimal(handlerNum10), mustDecimal(handlerNum4)}
	s.expectCalculateError(entities.OpDivide, divisor, apperror.New(entities.CodeDivisionByZero, errFakeCause))

	s.post(contentTypeJSON, bodyDivide)

	body := s.requireError(http.StatusUnprocessableEntity, entities.CodeDivisionByZero)
	s.Equal(entities.MsgDivisionByZero, body.Message)
}

func (s *ServerTestSuite) TestCalculateMapsUnknownOperationFromCoreTo422() {
	operands := []decimal.Decimal{mustDecimal(handlerNum1), mustDecimal(handlerNum2)}
	s.expectCalculateError(entities.OperationType(opModulo), operands, apperror.New(entities.CodeUnknownOperation, nil))

	s.post(contentTypeJSON, bodyUnknownOp)

	s.requireError(http.StatusUnprocessableEntity, entities.CodeUnknownOperation)
}

func (s *ServerTestSuite) TestCalculateMapsInternalErrorFromCoreTo500() {
	operands := []decimal.Decimal{mustDecimal(handlerNum10), mustDecimal(handlerNum4)}
	s.expectCalculateError(entities.OpDivide, operands, apperror.New(apperror.CodeInternal, errors.New(secretCause)))

	s.post(contentTypeJSON, bodyDivide)

	s.requireError(http.StatusInternalServerError, apperror.CodeInternal)
}

func (s *ServerTestSuite) TestCalculateNeverSendsTheInternalCause() {
	operands := []decimal.Decimal{mustDecimal(handlerNum10), mustDecimal(handlerNum4)}
	s.expectCalculateError(entities.OpDivide, operands, apperror.New(apperror.CodeInternal, errors.New(secretCause)))

	s.post(contentTypeJSON, bodyDivide)

	s.NotContains(s.recorder.Body.String(), secretCause)
}

func (s *ServerTestSuite) TestCalculateErrorOmitsRequestIDWhenContextHasNone() {
	s.ctx = context.Background()

	s.post(contentTypePlain, bodyDivide)

	s.Empty(s.errorBody().RequestID)
	s.Equal(http.StatusUnsupportedMediaType, s.recorder.Code)
}

func (s *ServerTestSuite) TestOperationsReturnsTheCoreListUnderOperationsKey() {
	s.mockCore.EXPECT().ListOperations(gomock.Eq(s.ctx)).Return([]entities.OperationInfo{
		{Name: entities.OpAdd, Arity: entities.ArityBinary, Symbol: "+", Description: "Adds the two operands."},
		{Name: entities.OpSqrt, Arity: entities.ArityUnary, Symbol: "√", Description: "Returns the square root of the operand."},
	}).Times(1)

	s.handler.Operations(s.recorder, httptest.NewRequest(http.MethodGet, testOperationsPath, nil).WithContext(s.ctx))

	s.Equal(http.StatusOK, s.recorder.Code)
	s.JSONEq(wantOperationsJS, s.recorder.Body.String())
	s.requireJSONHeaders()
}

func (s *ServerTestSuite) TestOperationsEncodesNilListAsEmptyArray() {
	s.mockCore.EXPECT().ListOperations(gomock.Eq(s.ctx)).Return(nil).Times(1)

	s.handler.Operations(s.recorder, httptest.NewRequest(http.MethodGet, testOperationsPath, nil).WithContext(s.ctx))

	s.JSONEq(wantEmptyOpsJS, s.recorder.Body.String())
}

func (s *ServerTestSuite) TestOperationsPublishesTheConfiguredLimits() {
	s.mockCore.EXPECT().ListOperations(gomock.Eq(s.ctx)).Return(nil).Times(1)

	s.handler.Operations(s.recorder, httptest.NewRequest(http.MethodGet, testOperationsPath, nil).WithContext(s.ctx))

	var got entities.OperationsResponse
	s.Require().NoError(json.Unmarshal(s.recorder.Body.Bytes(), &got))
	s.Equal(testHandlerLimits, got.Limits)
}

// routed sends req through a chi router so route registration and method
// matching are covered; the mock matches on the request ID in the context.
func (s *ServerTestSuite) routed(req *http.Request) {
	router := chi.NewRouter()
	router.MethodNotAllowed(httpx.MethodNotAllowed)
	router.NotFound(httpx.NotFound)
	s.handler.RegisterRoutes(router)
	router.ServeHTTP(s.recorder, req.WithContext(s.ctx))
}

func (s *ServerTestSuite) hasRequestID() gomock.Matcher {
	return gomock.Cond(func(x any) bool {
		ctx, ok := x.(context.Context)
		if !ok {
			return false
		}
		id, found := contextkeys.RequestID(ctx)
		return found && id == testRequestID
	})
}

func (s *ServerTestSuite) TestRegisterRoutesServesPostCalculate() {
	operands := []decimal.Decimal{mustDecimal(handlerNum10), mustDecimal(handlerNum4)}
	s.mockCore.EXPECT().
		Calculate(s.hasRequestID(), gomock.Eq(entities.OpDivide), gomock.Eq(operands)).
		Return(entities.Result{Text: wantResultText}, nil).
		Times(1)
	req := httptest.NewRequest(http.MethodPost, testCalcPath, strings.NewReader(bodyDivide))
	req.Header.Set(constants.HeaderContentType, contentTypeJSON)

	s.routed(req)

	s.Equal(http.StatusOK, s.recorder.Code)
	s.JSONEq(wantDivideResp, s.recorder.Body.String())
}

func (s *ServerTestSuite) TestRegisterRoutesServesGetOperations() {
	s.mockCore.EXPECT().ListOperations(s.hasRequestID()).Return([]entities.OperationInfo{}).Times(1)

	s.routed(httptest.NewRequest(http.MethodGet, testOperationsPath, nil))

	s.Equal(http.StatusOK, s.recorder.Code)
	s.JSONEq(wantEmptyOpsJS, s.recorder.Body.String())
}

func (s *ServerTestSuite) TestRegisterRoutesRejectsGetOnCalculateWith405() {
	s.routed(httptest.NewRequest(http.MethodGet, testCalcPath, nil))

	s.requireError(http.StatusMethodNotAllowed, apperror.CodeMethodNotAllowed)
}

func (s *ServerTestSuite) TestRegisterRoutesRejectsPostOnOperationsWith405() {
	s.routed(httptest.NewRequest(http.MethodPost, testOperationsPath, nil))

	s.requireError(http.StatusMethodNotAllowed, apperror.CodeMethodNotAllowed)
}
