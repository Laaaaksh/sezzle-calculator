package operations_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/config"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/operations"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

type PowerTestSuite struct {
	suite.Suite
	op *operations.Power
}

func TestPowerTestSuite(t *testing.T) {
	suite.Run(t, new(PowerTestSuite))
}

func (s *PowerTestSuite) SetupTest() {
	s.op = operations.NewPower(config.CalculatorConfig{
		MaxOperandLength: testMaxOperandLength,
		MaxExponent:      testMaxExponent,
		DecimalPrecision: testPowerPrecision,
	})
}

func (s *PowerTestSuite) power(base, exponent string) string {
	got, err := applyText(s.op, base, exponent)
	s.Require().Nil(err)
	return got
}

func (s *PowerTestSuite) failure(base, exponent string) apperror.IError {
	_, err := applyText(s.op, base, exponent)
	s.Require().NotNil(err)
	return err
}

func (s *PowerTestSuite) TestPowerNameIsPower() {
	s.Equal(entities.OpPower, s.op.Name())
}

func (s *PowerTestSuite) TestPowerArityIsTwo() {
	s.Equal(entities.ArityBinary, s.op.Arity())
}

func (s *PowerTestSuite) TestPowerWhenExponentIsPositiveReturnsTheProduct() {
	s.Equal("1024", s.power("2", "10"))
}

func (s *PowerTestSuite) TestPowerWhenExponentIsOneReturnsTheBase() {
	s.Equal("7.25", s.power("7.25", "1"))
}

func (s *PowerTestSuite) TestPowerWhenExponentIsZeroReturnsOne() {
	s.Equal("1", s.power("123.456", "0"))
}

func (s *PowerTestSuite) TestPowerWhenBaseAndExponentAreZeroReturnsOne() {
	s.Equal("1", s.power("0", "0"))
}

func (s *PowerTestSuite) TestPowerWhenExponentIsNegativeZeroReturnsOne() {
	s.Equal("1", s.power("9", "-0"))
}

func (s *PowerTestSuite) TestPowerWhenNegativeBaseHasOddExponentReturnsNegative() {
	s.Equal("-8", s.power("-2", "3"))
}

func (s *PowerTestSuite) TestPowerWhenNegativeBaseHasEvenExponentReturnsPositive() {
	s.Equal("16", s.power("-2", "4"))
}

func (s *PowerTestSuite) TestPowerWhenBaseIsFractionalMultipliesExactly() {
	s.Equal("2.25", s.power("1.5", "2"))
}

func (s *PowerTestSuite) TestPowerWhenBaseIsPointOneCubedIsExactlyPointZeroZeroOne() {
	s.Equal("0.001", s.power("0.1", "3"))
}

func (s *PowerTestSuite) TestPowerWhenExponentIsWholeWithDecimalPointIsAccepted() {
	s.Equal("4", s.power("2", "2.0"))
}

func (s *PowerTestSuite) TestPowerWhenExponentIsNegativeReturnsTheReciprocal() {
	s.Equal("0.25", s.power("2", "-2"))
}

func (s *PowerTestSuite) TestPowerWhenReciprocalDoesNotTerminateRoundsToThePrecision() {
	s.Equal("0.33333333333333333333", s.power("3", "-1"))
}

func (s *PowerTestSuite) TestPowerWhenNegativeBaseHasNegativeOddExponentReturnsNegativeFraction() {
	s.Equal("-0.125", s.power("-2", "-3"))
}

func (s *PowerTestSuite) TestPowerWhenNegativeExponentTieRoundsHalfEven() {
	twoPlaces := operations.NewPower(config.CalculatorConfig{
		MaxOperandLength: testMaxOperandLength,
		MaxExponent:      testMaxExponent,
		DecimalPrecision: 2,
	})

	got, err := applyText(twoPlaces, "2", "-3")

	s.Nil(err)
	s.Equal("0.12", got)
}

func (s *PowerTestSuite) TestPowerWhenResultIsLargeKeepsEveryDigit() {
	s.Equal(testTwoToTheHundred, s.power("2", "100"))
}

func (s *PowerTestSuite) TestPowerWhenBaseIsTenReturnsOneFollowedByZeros() {
	s.Equal(testTenToTheThirty, s.power("10", "30"))
}

func (s *PowerTestSuite) TestPowerWhenBaseHasTrailingZerosKeepsThemInTheResult() {
	got := s.power("10", "500")

	s.Len(got, 501)
}

func (s *PowerTestSuite) TestPowerWhenResultIsExactlyTheSmallestRepresentableKeepsIt() {
	s.Equal(testTinyUnit, s.power("10", "-20"))
}

func (s *PowerTestSuite) TestPowerWhenResultIsBelowHalfTheLastPlaceRoundsToZero() {
	s.Equal("0", s.power("10", "-21"))
}

func (s *PowerTestSuite) TestPowerWhenBaseIsOneAtTheMaxExponentReturnsOne() {
	s.Equal("1", s.power("1", "1000"))
}

func (s *PowerTestSuite) TestPowerWhenBaseIsMinusOneWithEvenExponentReturnsOne() {
	s.Equal("1", s.power("-1", "1000"))
}

func (s *PowerTestSuite) TestPowerWhenBaseIsMinusOneWithOddExponentReturnsMinusOne() {
	s.Equal("-1", s.power("-1", "-999"))
}

func (s *PowerTestSuite) TestPowerWhenBaseIsZeroAndExponentPositiveReturnsZero() {
	s.Equal("0", s.power("0", "5"))
}

func (s *PowerTestSuite) TestPowerWhenBaseIsZeroAndExponentNegativeReturnsDivisionByZero() {
	err := s.failure("0", "-1")

	s.Equal(apperror.CodeDivisionByZero, err.Code())
	s.Equal("Cannot divide by zero.", err.PublicMessage())
}

func (s *PowerTestSuite) TestPowerWhenBaseIsNegativeZeroAndExponentNegativeReturnsDivisionByZero() {
	err := s.failure("-0.0", "-3")

	s.Equal(apperror.CodeDivisionByZero, err.Code())
}

func (s *PowerTestSuite) TestPowerWhenExponentIsFractionalReturnsInvalidExponent() {
	err := s.failure("2", "0.5")

	s.Equal(apperror.CodeInvalidExponent, err.Code())
	s.ErrorIs(err, operations.ErrFractionalExponent)
	s.Equal(testExponentIndex, err.Fields()[entities.FieldOperandIndex])
}

func (s *PowerTestSuite) TestPowerWhenNegativeExponentIsFractionalReturnsInvalidExponent() {
	s.Equal(apperror.CodeInvalidExponent, s.failure("2", "-0.5").Code())
}

func (s *PowerTestSuite) TestPowerWhenExponentIsBarelyFractionalReturnsInvalidExponent() {
	s.Equal(apperror.CodeInvalidExponent, s.failure("2", "3.0000000000000000001").Code())
}

func (s *PowerTestSuite) TestPowerWhenBaseIsZeroAndExponentIsFractionalReturnsInvalidExponentFirst() {
	s.Equal(apperror.CodeInvalidExponent, s.failure("0", "-0.5").Code())
}

func (s *PowerTestSuite) TestPowerWhenExponentIsAboveTheMaximumReturnsExponentTooLarge() {
	err := s.failure("1", "1001")

	s.Equal(apperror.CodeExponentTooLarge, err.Code())
	s.Equal(int64(testMaxExponent), err.Fields()[entities.FieldLimit])
}

func (s *PowerTestSuite) TestPowerWhenExponentIsBelowTheMinimumReturnsExponentTooLarge() {
	s.Equal(apperror.CodeExponentTooLarge, s.failure("1", "-1001").Code())
}

func (s *PowerTestSuite) TestPowerWhenExponentIsAHugeNumberReturnsExponentTooLarge() {
	s.Equal(apperror.CodeExponentTooLarge, s.failure("2", "99999999999999999999").Code())
}

func (s *PowerTestSuite) TestPowerWhenResultWouldExceedTheDigitBudgetReturnsResultTooLarge() {
	err := s.failure("9", "1000")

	s.Equal(apperror.CodeResultTooLarge, err.Code())
	s.Equal(int64(testResultBudget), err.Fields()[entities.FieldLimit])
}

func (s *PowerTestSuite) TestPowerWhenNegativeExponentWouldExceedTheDigitBudgetReturnsResultTooLarge() {
	s.Equal(apperror.CodeResultTooLarge, s.failure("9", "-1000").Code())
}

func (s *PowerTestSuite) TestPowerWhenBaseIsLongAndExponentModestReturnsResultTooLarge() {
	s.Equal(apperror.CodeResultTooLarge, s.failure("99999999999999999999", "50").Code())
}

func (s *PowerTestSuite) TestPowerWhenTwoToTheThousandFitsTheBudgetReturnsTheExactValue() {
	got := s.power("2", "1000")

	s.Len(got, 302)
	s.True(strings.HasPrefix(got, testTwoToTheThousandPrefix))
}

func (s *PowerTestSuite) TestPowerWhenBoundEqualsTheBudgetComputesTheResult() {
	got := s.power("99", "320")

	s.Len(got, 639)
}

func (s *PowerTestSuite) TestPowerWhenBoundIsOneOverTheBudgetReturnsResultTooLarge() {
	s.Equal(apperror.CodeResultTooLarge, s.failure("99", "321").Code())
}

func (s *PowerTestSuite) TestPowerDigitBudgetScalesWithTheConfiguredOperandLength() {
	small := operations.NewPower(config.CalculatorConfig{
		MaxOperandLength: testSmallOperandLen,
		MaxExponent:      testMaxExponent,
		DecimalPrecision: testPowerPrecision,
	})

	allowed, errAllowed := applyText(small, "2", "63")
	_, errRefused := applyText(small, "2", "64")

	s.Nil(errAllowed)
	s.Equal("9223372036854775808", allowed)
	s.Require().NotNil(errRefused)
	s.Equal(int64(testSmallBudget), errRefused.Fields()[entities.FieldLimit])
}

func (s *PowerTestSuite) TestPowerWhenTrailingZerosPushPastTheBudgetReturnsResultTooLarge() {
	s.Equal(apperror.CodeResultTooLarge, s.failure("10", "1000").Code())
}

func (s *PowerTestSuite) TestPowerWhenBaseIsAMillionAndResultIsTooLongReturnsResultTooLarge() {
	s.Equal(apperror.CodeResultTooLarge, s.failure("1000000", "1000").Code())
}

func (s *PowerTestSuite) TestPowerWhenBaseIsTenAndResultFitsReturnsEveryDigit() {
	got := s.power("10", "638")

	s.Len(got, 639)
}

func (s *PowerTestSuite) TestPowerWhenFractionDigitsExceedTheBudgetReturnsResultTooLarge() {
	s.Equal(apperror.CodeResultTooLarge, s.failure("0.1", "1000").Code())
}

func (s *PowerTestSuite) TestPowerWhenFractionDigitsEqualTheBudgetComputesTheResult() {
	got := s.power("0.1", "640")

	s.Equal("0."+strings.Repeat("0", 639)+"1", got)
}

func (s *PowerTestSuite) TestPowerWhenFractionDigitsAreOneOverTheBudgetReturnsResultTooLarge() {
	s.Equal(apperror.CodeResultTooLarge, s.failure("0.1", "641").Code())
}

func (s *PowerTestSuite) TestPowerWhenSmallBaseHasNegativeExponentReturnsResultTooLarge() {
	s.Equal(apperror.CodeResultTooLarge, s.failure("0.1", "-1000").Code())
}

func (s *PowerTestSuite) TestPowerWhenTinyBaseHasNegativeExponentReturnsResultTooLarge() {
	base := "0." + strings.Repeat("0", 61) + "2"

	s.Equal(apperror.CodeResultTooLarge, s.failure(base, "-1000").Code())
}

func (s *PowerTestSuite) TestPowerWhenSmallBaseHasNegativeExponentThatFitsReturnsTheExactValue() {
	s.Equal("1"+strings.Repeat("0", 600), s.power("0.1", "-600"))
}
