package calculator_test

import (
	"context"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/operations"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

type ResultSizeTestSuite struct {
	suite.Suite
	ctx context.Context
}

func TestResultSizeTestSuite(t *testing.T) {
	suite.Run(t, new(ResultSizeTestSuite))
}

func (s *ResultSizeTestSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *ResultSizeTestSuite) limits(budget int) entities.Limits {
	return entities.Limits{
		MaxOperandLength: testDefaultMaxLength,
		MaxExponent:      testMaxExponent,
		DecimalPrecision: testPrecision,
		MaxResultDigits:  budget,
	}
}

func (s *ResultSizeTestSuite) realCore(budget int) *calculator.Core {
	limits := s.limits(budget)
	return calculator.NewCore(operations.NewDefaultRegistry(limits), limits)
}

func (s *ResultSizeTestSuite) fakeCore(result string) *calculator.Core {
	fake := &fakeOperation{name: testOpFake, arity: entities.ArityBinary, result: mustDecimal(result)}
	return calculator.NewCore(operations.NewRegistry(fake), s.limits(testSmallResultDigits))
}

func (s *ResultSizeTestSuite) apply(core *calculator.Core, op entities.OperationType, operands ...string) (entities.Result, apperror.IError) {
	parsed := make([]decimal.Decimal, len(operands))
	for i, text := range operands {
		parsed[i] = mustDecimal(text)
	}
	return core.Calculate(s.ctx, op, parsed)
}

func (s *ResultSizeTestSuite) requireTooLarge(err apperror.IError, budget int) {
	s.Require().NotNil(err)
	s.Equal(entities.CodeResultTooLarge, err.Code())
	s.Equal(int64(budget), err.Fields()[entities.FieldLimit])
}

func (s *ResultSizeTestSuite) TestCalculateWhenResultHasExactlyTheBudgetDigitsReturnsIt() {
	_, err := s.apply(s.fakeCore(sizeAtLimit), testOpFake, "1", "2")

	s.Nil(err)
}

func (s *ResultSizeTestSuite) TestCalculateWhenResultHasOneDigitOverTheBudgetReturnsResultTooLarge() {
	_, err := s.apply(s.fakeCore(sizeOverLimit), testOpFake, "1", "2")

	s.requireTooLarge(err, testSmallResultDigits)
}

func (s *ResultSizeTestSuite) TestCalculateWhenNegativeResultHasExactlyTheBudgetDigitsReturnsIt() {
	_, err := s.apply(s.fakeCore(sizeNegAtLimit), testOpFake, "1", "2")

	s.Nil(err)
}

func (s *ResultSizeTestSuite) TestCalculateWhenFractionHasExactlyTheBudgetDigitsReturnsIt() {
	_, err := s.apply(s.fakeCore(sizeFractionAt), testOpFake, "1", "2")

	s.Nil(err)
}

func (s *ResultSizeTestSuite) TestCalculateWhenFractionHasOneDigitOverTheBudgetReturnsResultTooLarge() {
	_, err := s.apply(s.fakeCore(sizeFractionOver), testOpFake, "1", "2")

	s.requireTooLarge(err, testSmallResultDigits)
}

func (s *ResultSizeTestSuite) TestCalculateDoesNotCountTrailingFractionZeros() {
	_, err := s.apply(s.fakeCore(sizeTrailingZeros), testOpFake, "1", "2")

	s.Nil(err)
}

func (s *ResultSizeTestSuite) TestCalculateWhenResultIsTooLargeReturnsNoResult() {
	result, _ := s.apply(s.fakeCore(sizeOverLimit), testOpFake, "1", "2")

	s.Equal(entities.Result{}, result)
}

func (s *ResultSizeTestSuite) TestCalculateAddOverTheBudgetReturnsResultTooLarge() {
	_, err := s.apply(s.realCore(sizeTinyBudget), entities.OpAdd, sizeAddLeft, sizeAddRight)

	s.requireTooLarge(err, sizeTinyBudget)
}

func (s *ResultSizeTestSuite) TestCalculateSubtractOverTheBudgetReturnsResultTooLarge() {
	_, err := s.apply(s.realCore(sizeTinyBudget), entities.OpSubtract, sizeSubtractLeft, sizeSubtractRight)

	s.requireTooLarge(err, sizeTinyBudget)
}

func (s *ResultSizeTestSuite) TestCalculateMultiplyOverTheBudgetReturnsResultTooLarge() {
	_, err := s.apply(s.realCore(sizeTinyBudget), entities.OpMultiply, sizeMultiplyLeft, sizeMultiplyRight)

	s.requireTooLarge(err, sizeTinyBudget)
}

func (s *ResultSizeTestSuite) TestCalculateDivideOverTheBudgetReturnsResultTooLarge() {
	_, err := s.apply(s.realCore(sizeTinyBudget), entities.OpDivide, sizeDivideLeft, sizeDivideRight)

	s.requireTooLarge(err, sizeTinyBudget)
}

func (s *ResultSizeTestSuite) TestCalculatePercentageOverTheBudgetReturnsResultTooLarge() {
	_, err := s.apply(s.realCore(sizeTinyBudget), entities.OpPercentage, sizePercentLeft, sizePercentRight)

	s.requireTooLarge(err, sizeTinyBudget)
}

func (s *ResultSizeTestSuite) TestCalculatePowerOverTheBudgetReturnsResultTooLarge() {
	_, err := s.apply(s.realCore(sizeTinyBudget), entities.OpPower, sizePowerBase, sizePowerExponent)

	s.requireTooLarge(err, sizeTinyBudget)
}

func (s *ResultSizeTestSuite) TestCalculateSqrtOverTheBudgetReturnsResultTooLarge() {
	_, err := s.apply(s.realCore(sizeTinyBudget), entities.OpSqrt, sizeSqrtOperand)

	s.requireTooLarge(err, sizeTinyBudget)
}

func (s *ResultSizeTestSuite) maxOperand() string {
	return strings.Repeat(sizeMaxOperandNine, testDefaultMaxLength)
}

func (s *ResultSizeTestSuite) maxOperandSquare() string {
	return strings.Repeat(sizeMaxOperandNine, testDefaultMaxLength-1) + sizeSquareEight +
		strings.Repeat(sizeSquareZero, testDefaultMaxLength-1) + sizeSquareOne
}

func (s *ResultSizeTestSuite) TestCalculateMultiplyOfTwoMaxLengthOperandsFitsTheDefaultBudget() {
	result, err := s.apply(s.realCore(testResultDigits), entities.OpMultiply, s.maxOperand(), s.maxOperand())

	s.Require().Nil(err)
	s.Equal(s.maxOperandSquare(), result.Text)
}

func (s *ResultSizeTestSuite) TestCalculateMultiplyOfTwoMaxLengthOperandsAtTheProductLengthReturnsIt() {
	result, err := s.apply(s.realCore(testMaxProductDigits), entities.OpMultiply, s.maxOperand(), s.maxOperand())

	s.Require().Nil(err)
	s.Equal(s.maxOperandSquare(), result.Text)
}

func (s *ResultSizeTestSuite) TestCalculateMultiplyOfTwoMaxLengthOperandsOneDigitUnderTheProductReturnsResultTooLarge() {
	_, err := s.apply(s.realCore(testBelowProductLimit), entities.OpMultiply, s.maxOperand(), s.maxOperand())

	s.requireTooLarge(err, testBelowProductLimit)
}
