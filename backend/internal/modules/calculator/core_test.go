package calculator_test

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/config"
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
	limits := config.CalculatorConfig{
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
	s.Equal(apperror.CodeUnknownOperation, err.Code())
	s.Equal(testOpMissing, err.Fields()[entities.FieldOperation])
}

func (s *CoreTestSuite) TestCalculateWhenOperationIsUnknownDoesNotCallApply() {
	_, _ = s.core.Calculate(s.ctx, testOpMissing, s.binary("1", "2"))

	s.Zero(s.fake.calls)
}

func (s *CoreTestSuite) TestCalculateWhenOperationNameIsEmptyReturnsUnknownOperation() {
	_, err := s.core.Calculate(s.ctx, testOpBlank, s.binary("1", "2"))

	s.Require().NotNil(err)
	s.Equal(apperror.CodeUnknownOperation, err.Code())
}

func (s *CoreTestSuite) TestCalculateWhenTooFewOperandsReturnsInvalidOperandCount() {
	_, err := s.core.Calculate(s.ctx, testOpFake, []decimal.Decimal{mustDecimal("1")})

	s.Require().NotNil(err)
	s.Equal(apperror.CodeInvalidOperandCount, err.Code())
	s.Equal(entities.ArityBinary, err.Fields()[entities.FieldExpected])
	s.Equal(1, err.Fields()[entities.FieldActual])
}

func (s *CoreTestSuite) TestCalculateWhenTooManyOperandsReturnsInvalidOperandCount() {
	operands := append(s.binary("1", "2"), mustDecimal("3"))

	_, err := s.core.Calculate(s.ctx, testOpFake, operands)

	s.Require().NotNil(err)
	s.Equal(apperror.CodeInvalidOperandCount, err.Code())
	s.Equal(3, err.Fields()[entities.FieldActual])
}

func (s *CoreTestSuite) TestCalculateWhenNoOperandsReturnsInvalidOperandCount() {
	_, err := s.core.Calculate(s.ctx, testOpFake, nil)

	s.Require().NotNil(err)
	s.Equal(apperror.CodeInvalidOperandCount, err.Code())
}

func (s *CoreTestSuite) TestCalculateWhenOperandCountIsWrongDoesNotCallApply() {
	_, _ = s.core.Calculate(s.ctx, testOpFake, nil)

	s.Zero(s.fake.calls)
}

func (s *CoreTestSuite) TestCalculateHonoursAUnaryArity() {
	unary := &fakeOperation{name: testOpUnary, arity: entities.ArityUnary, result: mustDecimal("3")}
	core := calculator.NewCore(operations.NewRegistry(unary), config.CalculatorConfig{MaxOperandLength: testMaxOperandLength})

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
	s.Equal(apperror.CodeOperandTooLarge, err.Code())
	s.Equal(testIndexSecond, err.Fields()[entities.FieldOperandIndex])
	s.Equal(testMaxOperandLength, err.Fields()[entities.FieldLimit])
}

func (s *CoreTestSuite) TestCalculateWhenNegativeOperandExceedsDigitLimitReturnsOperandTooLarge() {
	_, err := s.core.Calculate(s.ctx, testOpFake, s.binary(testOperandNegLong, "1"))

	s.Require().NotNil(err)
	s.Equal(apperror.CodeOperandTooLarge, err.Code())
	s.Equal(testIndexFirst, err.Fields()[entities.FieldOperandIndex])
}

func (s *CoreTestSuite) TestCalculateCountsDigitsNotTheDecimalPoint() {
	_, err := s.core.Calculate(s.ctx, testOpFake, s.binary(testOperandDecimal, "1"))

	s.Nil(err)
}

func (s *CoreTestSuite) TestCalculateWhenFractionExceedsDigitLimitReturnsOperandTooLarge() {
	_, err := s.core.Calculate(s.ctx, testOpFake, s.binary(testOperandLongFrac, "1"))

	s.Require().NotNil(err)
	s.Equal(apperror.CodeOperandTooLarge, err.Code())
}

func (s *CoreTestSuite) TestCalculateWhenTinyExponentExceedsLimitReturnsOperandTooLarge() {
	_, err := s.core.Calculate(s.ctx, testOpFake, s.binary(testTinyExponent, "1"))

	s.Require().NotNil(err)
	s.Equal(apperror.CodeOperandTooLarge, err.Code())
}

func (s *CoreTestSuite) TestCalculateWhenHugeExponentExceedsLimitReturnsOperandTooLarge() {
	_, err := s.core.Calculate(s.ctx, testOpFake, s.binary("1", testHugeExponent))

	s.Require().NotNil(err)
	s.Equal(apperror.CodeOperandTooLarge, err.Code())
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
	core := calculator.NewCore(operations.NewRegistry(failing), config.CalculatorConfig{MaxOperandLength: testMaxOperandLength})

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
	s.True(result.Value.IsZero())
}

func (s *CoreTestSuite) TestCalculateKeepsTheSignOfNegativeResults() {
	s.fake.result = mustDecimal("-0.250")

	result, err := s.core.Calculate(s.ctx, testOpFake, s.binary("1", "2"))

	s.Require().Nil(err)
	s.Equal("-0.25", result.Text)
}

func (s *CoreTestSuite) TestCalculateKeepsTheNumericValueOfTheResult() {
	result, err := s.core.Calculate(s.ctx, testOpFake, s.binary("1", "2"))

	s.Require().Nil(err)
	s.True(mustDecimal(testFakeResult).Equal(result.Value))
}
