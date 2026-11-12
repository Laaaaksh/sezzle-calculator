package calculator_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/operations"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

type CoreTestSuite struct {
	suite.Suite
	ctx  context.Context
	fake *fakeOperation
	core *calculator.Core
}

func TestCoreTestSuite(t *testing.T) {
	suite.Run(t, new(CoreTestSuite))
}

func (s *CoreTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.fake = &fakeOperation{name: testOpFake, arity: entities.ArityBinary, result: mustDecimal(testFakeResult)}
	limits := entities.Limits{
		MaxOperandLength: testMaxOperandLength,
		MaxExponent:      testMaxExponent,
		DecimalPrecision: testPrecision,
	}
	s.core = calculator.NewCore(operations.NewRegistry(s.fake), limits)
}

func (s *CoreTestSuite) binary(first, second string) []decimal.Decimal {
	return []decimal.Decimal{mustDecimal(first), mustDecimal(second)}
}

func (s *CoreTestSuite) TestCalculateWhenOperationIsUnknownReturnsUnknownOperation() {
	_, err := s.core.Calculate(s.ctx, testOpMissing, s.binary("1", "2"))

	s.Require().NotNil(err)
	s.Equal(entities.CodeUnknownOperation, err.Code())
	s.Equal(testOpMissing, err.Fields()[entities.FieldOperation])
}

func (s *CoreTestSuite) TestCalculateWhenOperationIsUnknownDoesNotCallApply() {
	_, _ = s.core.Calculate(s.ctx, testOpMissing, s.binary("1", "2"))

	s.Zero(s.fake.calls)
}

func (s *CoreTestSuite) TestCalculateWhenOperationNameIsEmptyReturnsUnknownOperation() {
	_, err := s.core.Calculate(s.ctx, testOpBlank, s.binary("1", "2"))

	s.Require().NotNil(err)
	s.Equal(entities.CodeUnknownOperation, err.Code())
}

func (s *CoreTestSuite) TestCalculateWhenTooFewOperandsReturnsInvalidOperandCount() {
	_, err := s.core.Calculate(s.ctx, testOpFake, []decimal.Decimal{mustDecimal("1")})

	s.Require().NotNil(err)
	s.Equal(entities.CodeInvalidOperandCount, err.Code())
	s.Equal(entities.ArityBinary, err.Fields()[entities.FieldExpected])
	s.Equal(1, err.Fields()[entities.FieldActual])
}

func (s *CoreTestSuite) TestCalculateWhenTooManyOperandsReturnsInvalidOperandCount() {
	operands := append(s.binary("1", "2"), mustDecimal("3"))

	_, err := s.core.Calculate(s.ctx, testOpFake, operands)

	s.Require().NotNil(err)
	s.Equal(entities.CodeInvalidOperandCount, err.Code())
	s.Equal(3, err.Fields()[entities.FieldActual])
}

func (s *CoreTestSuite) TestCalculateWhenNoOperandsReturnsInvalidOperandCount() {
	_, err := s.core.Calculate(s.ctx, testOpFake, nil)

	s.Require().NotNil(err)
	s.Equal(entities.CodeInvalidOperandCount, err.Code())
}

func (s *CoreTestSuite) TestCalculateWhenOperandCountIsWrongDoesNotCallApply() {
	_, _ = s.core.Calculate(s.ctx, testOpFake, nil)

	s.Zero(s.fake.calls)
}

func (s *CoreTestSuite) TestCalculateHonoursAUnaryArity() {
	unary := &fakeOperation{name: testOpUnary, arity: entities.ArityUnary, result: mustDecimal("3")}
	core := calculator.NewCore(operations.NewRegistry(unary), entities.Limits{MaxOperandLength: testMaxOperandLength})

	result, err := core.Calculate(s.ctx, testOpUnary, []decimal.Decimal{mustDecimal("9")})

	s.Require().Nil(err)
	s.Equal("3", result.Text)
}

func (s *CoreTestSuite) TestCalculateWhenOperandAtDigitLimitIsAccepted() {
	_, err := s.core.Calculate(s.ctx, testOpFake, s.binary(testOperandOk, "1"))

	s.Nil(err)
}

func (s *CoreTestSuite) TestCalculateWhenOperandExceedsDigitLimitReturnsOperandTooLarge() {
	_, err := s.core.Calculate(s.ctx, testOpFake, s.binary("1", testOperandTooLong))

	s.Require().NotNil(err)
	s.Equal(entities.CodeOperandTooLarge, err.Code())
	s.Equal(testIndexSecond, err.Fields()[entities.FieldOperandIndex])
	s.Equal(testMaxOperandLength, err.Fields()[entities.FieldLimit])
}

func (s *CoreTestSuite) TestCalculateWhenNegativeOperandExceedsDigitLimitReturnsOperandTooLarge() {
	_, err := s.core.Calculate(s.ctx, testOpFake, s.binary(testOperandNegLong, "1"))

	s.Require().NotNil(err)
	s.Equal(entities.CodeOperandTooLarge, err.Code())
	s.Equal(testIndexFirst, err.Fields()[entities.FieldOperandIndex])
}

func (s *CoreTestSuite) TestCalculateCountsDigitsNotTheDecimalPoint() {
	_, err := s.core.Calculate(s.ctx, testOpFake, s.binary(testOperandDecimal, "1"))

	s.Nil(err)
}

func (s *CoreTestSuite) TestCalculateWhenFractionExceedsDigitLimitReturnsOperandTooLarge() {
	_, err := s.core.Calculate(s.ctx, testOpFake, s.binary(testOperandLongFrac, "1"))

	s.Require().NotNil(err)
	s.Equal(entities.CodeOperandTooLarge, err.Code())
}

func (s *CoreTestSuite) TestCalculateWhenTinyExponentExceedsLimitReturnsOperandTooLarge() {
	_, err := s.core.Calculate(s.ctx, testOpFake, s.binary(testTinyExponent, "1"))

	s.Require().NotNil(err)
	s.Equal(entities.CodeOperandTooLarge, err.Code())
}

func (s *CoreTestSuite) TestCalculateWhenHugeExponentExceedsLimitReturnsOperandTooLarge() {
	_, err := s.core.Calculate(s.ctx, testOpFake, s.binary("1", testHugeExponent))

	s.Require().NotNil(err)
	s.Equal(entities.CodeOperandTooLarge, err.Code())
}

func (s *CoreTestSuite) TestCalculateWhenOperandIsTooLargeDoesNotCallApply() {
	_, _ = s.core.Calculate(s.ctx, testOpFake, s.binary(testOperandTooLong, "1"))

	s.Zero(s.fake.calls)
}

func (s *CoreTestSuite) TestCalculatePassesOperandsToApplyUnchanged() {
	operands := s.binary("1.5", "-2")

	_, err := s.core.Calculate(s.ctx, testOpFake, operands)

	s.Require().Nil(err)
	s.Equal(1, s.fake.calls)
	s.True(operands[0].Equal(s.fake.seen[0]))
	s.True(operands[1].Equal(s.fake.seen[1]))
}

func (s *CoreTestSuite) TestCalculateReturnsOperationErrorUnchanged() {
	boom := apperror.New(apperror.CodeValidationFailed, errFakeCause)
	failing := &fakeOperation{name: testOpBoom, arity: entities.ArityBinary, err: boom}
	core := calculator.NewCore(operations.NewRegistry(failing), entities.Limits{MaxOperandLength: testMaxOperandLength})

	_, err := core.Calculate(s.ctx, testOpBoom, s.binary("1", "2"))

	s.Equal(boom, err)
}

func (s *CoreTestSuite) TestCalculateTrimsTrailingZerosFromTheText() {
	result, err := s.core.Calculate(s.ctx, testOpFake, s.binary("1", "2"))

	s.Require().Nil(err)
	s.Equal(testFakeText, result.Text)
}

func (s *CoreTestSuite) TestCalculateWritesWholeNumbersWithoutAFraction() {
	s.fake.result = mustDecimal(testBigText)

	result, err := s.core.Calculate(s.ctx, testOpFake, s.binary("1", "2"))

	s.Require().Nil(err)
	s.Equal(testBigWant, result.Text)
}

func (s *CoreTestSuite) TestCalculateTurnsNegativeZeroIntoZero() {
	s.fake.result = mustDecimal(testNegZero).Neg()

	result, err := s.core.Calculate(s.ctx, testOpFake, s.binary("1", "2"))

	s.Require().Nil(err)
	s.Equal(testZeroText, result.Text)
}

func (s *CoreTestSuite) TestCalculateKeepsTheSignOfNegativeResults() {
	s.fake.result = mustDecimal("-0.250")

	result, err := s.core.Calculate(s.ctx, testOpFake, s.binary("1", "2"))

	s.Require().Nil(err)
	s.Equal("-0.25", result.Text)
}

type DefaultOperationsTestSuite struct {
	suite.Suite
	ctx  context.Context
	core *calculator.Core
}

func TestDefaultOperationsTestSuite(t *testing.T) {
	suite.Run(t, new(DefaultOperationsTestSuite))
}

func (s *DefaultOperationsTestSuite) SetupTest() {
	s.ctx = context.Background()
	cfg := entities.Limits{MaxOperandLength: testDefaultMaxLength, MaxExponent: testMaxExponent, DecimalPrecision: testPrecision}
	s.core = calculator.NewCore(operations.NewDefaultRegistry(cfg), cfg)
}

func (s *DefaultOperationsTestSuite) calculate(op entities.OperationType, left, right string) (entities.Result, apperror.IError) {
	return s.core.Calculate(s.ctx, op, []decimal.Decimal{mustDecimal(left), mustDecimal(right)})
}

func (s *DefaultOperationsTestSuite) TestCalculateAddPointOnePlusPointTwoIsExactlyPointThree() {
	result, err := s.calculate(entities.OpAdd, "0.1", "0.2")

	s.Require().Nil(err)
	s.Equal("0.3", result.Text)
}

func (s *DefaultOperationsTestSuite) TestCalculateMultiplyNegativeByZeroReturnsPlainZero() {
	result, err := s.calculate(entities.OpMultiply, "-5", "0")

	s.Require().Nil(err)
	s.Equal(testZeroText, result.Text)
}

func (s *DefaultOperationsTestSuite) TestCalculateDivideOneByThreeReturnsTheConfiguredPrecision() {
	result, err := s.calculate(entities.OpDivide, "1", "3")

	s.Require().Nil(err)
	s.Equal("0.33333333333333333333", result.Text)
}

func (s *DefaultOperationsTestSuite) TestCalculateDivideByZeroReturnsDivisionByZeroWithItsMessage() {
	_, err := s.calculate(entities.OpDivide, "1", "0")

	s.Require().NotNil(err)
	s.Equal(entities.CodeDivisionByZero, err.Code())
	s.Equal(entities.MsgDivisionByZero, err.PublicMessage())
}

func (s *DefaultOperationsTestSuite) TestCalculateSubtractTrimsTrailingZeros() {
	result, err := s.calculate(entities.OpSubtract, "1.50", "0.25")

	s.Require().Nil(err)
	s.Equal("1.25", result.Text)
}

func (s *DefaultOperationsTestSuite) TestCalculateWhenNameIsNotRegisteredReturnsUnknownOperation() {
	_, err := s.calculate(testOpMissing, "1", "2")

	s.Require().NotNil(err)
	s.Equal(entities.CodeUnknownOperation, err.Code())
}

func (s *CoreTestSuite) TestListOperationsDescribesEachRegisteredOperationInOrder() {
	unary := &fakeOperation{name: testOpUnary, arity: entities.ArityUnary}
	core := calculator.NewCore(operations.NewRegistry(s.fake, unary), entities.Limits{})

	got := core.ListOperations(s.ctx)

	s.Equal([]entities.OperationInfo{
		{Name: testOpFake, Arity: entities.ArityBinary, Symbol: testFakeSymbol, Description: testFakeDescription},
		{Name: testOpUnary, Arity: entities.ArityUnary, Symbol: testFakeSymbol, Description: testFakeDescription},
	}, got)
}

func (s *CoreTestSuite) TestListOperationsWithNothingRegisteredReturnsAnEmptyList() {
	core := calculator.NewCore(operations.NewRegistry(), entities.Limits{})

	got := core.ListOperations(s.ctx)

	s.NotNil(got)
	s.Empty(got)
}

func (s *CoreTestSuite) endedContext() context.Context {
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()
	return ctx
}

func (s *CoreTestSuite) TestCalculateWhenContextEndedReturnsRequestTimeout() {
	_, err := s.core.Calculate(s.endedContext(), testOpFake, s.binary("1", "2"))

	s.Require().NotNil(err)
	s.Equal(apperror.CodeRequestTimeout, err.Code())
	s.ErrorIs(err, context.Canceled)
}

func (s *CoreTestSuite) TestCalculateWhenContextEndedDoesNotCallApply() {
	_, _ = s.core.Calculate(s.endedContext(), testOpFake, s.binary("1", "2"))

	s.Zero(s.fake.calls)
}

func (s *CoreTestSuite) TestCalculateWhenDeadlinePassedKeepsTheDeadlineCause() {
	ctx, cancel := context.WithDeadline(s.ctx, time.Now().Add(-time.Second))
	defer cancel()

	_, err := s.core.Calculate(ctx, testOpFake, s.binary("1", "2"))

	s.Require().NotNil(err)
	s.True(errors.Is(err, context.DeadlineExceeded))
}

func (s *CoreTestSuite) TestCalculateWhenContextEndsDuringApplyDropsTheLateResult() {
	ctx, cancel := context.WithCancel(s.ctx)
	s.fake.onApply = cancel

	result, err := s.core.Calculate(ctx, testOpFake, s.binary("1", "2"))

	s.Require().NotNil(err)
	s.Equal(apperror.CodeRequestTimeout, err.Code())
	s.Empty(result.Text)
}

func (s *CoreTestSuite) TestCalculatePassesItsContextToApply() {
	_, _ = s.core.Calculate(s.ctx, testOpFake, s.binary("1", "2"))

	s.Equal(s.ctx, s.fake.seenCtx)
}
