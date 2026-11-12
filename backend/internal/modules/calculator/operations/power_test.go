package operations_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"

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
	s.op = operations.NewPower(entities.Limits{
		MaxOperandLength: testMaxOperandLength,
		MaxExponent:      testMaxExponent,
		DecimalPrecision: testPowerPrecision,
		MaxResultDigits:  testResultBudget,
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
	s.Equal(powerNum1024, s.power(powerNum2, powerNum10))
}

func (s *PowerTestSuite) TestPowerWhenExponentIsOneReturnsTheBase() {
	s.Equal(powerNum7Pt25, s.power(powerNum7Pt25, powerNum1))
}

func (s *PowerTestSuite) TestPowerWhenExponentIsZeroReturnsOne() {
	s.Equal(powerNum1, s.power(powerNum123Pt456, powerNum0))
}

func (s *PowerTestSuite) TestPowerWhenBaseAndExponentAreZeroReturnsOne() {
	s.Equal(powerNum1, s.power(powerNum0, powerNum0))
}

func (s *PowerTestSuite) TestPowerWhenExponentIsNegativeZeroReturnsOne() {
	s.Equal(powerNum1, s.power(powerNum9, powerNumNeg0))
}

func (s *PowerTestSuite) TestPowerWhenNegativeBaseHasOddExponentReturnsNegative() {
	s.Equal(powerNumNeg8, s.power(powerNumNeg2, powerNum3))
}

func (s *PowerTestSuite) TestPowerWhenNegativeBaseHasEvenExponentReturnsPositive() {
	s.Equal(powerNum16, s.power(powerNumNeg2, powerNum4))
}

func (s *PowerTestSuite) TestPowerWhenBaseIsFractionalMultipliesExactly() {
	s.Equal(powerNum2Pt25, s.power(powerNum1Pt5, powerNum2))
}

func (s *PowerTestSuite) TestPowerWhenBaseIsPointOneCubedIsExactlyPointZeroZeroOne() {
	s.Equal(powerNum0Pt001, s.power(powerNum0Pt1, powerNum3))
}

func (s *PowerTestSuite) TestPowerWhenExponentIsWholeWithDecimalPointIsAccepted() {
	s.Equal(powerNum4, s.power(powerNum2, powerNum2Pt0))
}

func (s *PowerTestSuite) TestPowerWhenExponentIsNegativeReturnsTheReciprocal() {
	s.Equal(powerNum0Pt25, s.power(powerNum2, powerNumNeg2))
}

func (s *PowerTestSuite) TestPowerWhenReciprocalDoesNotTerminateRoundsToThePrecision() {
	s.Equal(powerLong1, s.power(powerNum3, powerNumNeg1))
}

func (s *PowerTestSuite) TestPowerWhenNegativeBaseHasNegativeOddExponentReturnsNegativeFraction() {
	s.Equal(powerNumNeg0Pt125, s.power(powerNumNeg2, powerNumNeg3))
}

func (s *PowerTestSuite) TestPowerWhenNegativeExponentTieRoundsHalfEven() {
	twoPlaces := operations.NewPower(entities.Limits{
		MaxOperandLength: testMaxOperandLength,
		MaxExponent:      testMaxExponent,
		DecimalPrecision: 2,
		MaxResultDigits:  testResultBudget,
	})

	got, err := applyText(twoPlaces, powerNum2, powerNumNeg3)

	s.Nil(err)
	s.Equal(powerNum0Pt12, got)
}

func (s *PowerTestSuite) TestPowerWhenResultIsLargeKeepsEveryDigit() {
	s.Equal(testTwoToTheHundred, s.power(powerNum2, powerNum100))
}

func (s *PowerTestSuite) TestPowerWhenBaseIsTenReturnsOneFollowedByZeros() {
	s.Equal(testTenToTheThirty, s.power(powerNum10, powerNum30))
}

func (s *PowerTestSuite) TestPowerWhenBaseHasTrailingZerosKeepsThemInTheResult() {
	got := s.power(powerNum10, powerNum500)

	s.Len(got, 501)
}

func (s *PowerTestSuite) TestPowerWhenResultIsExactlyTheSmallestRepresentableKeepsIt() {
	s.Equal(testTinyUnit, s.power(powerNum10, powerNumNeg20))
}

func (s *PowerTestSuite) TestPowerWhenResultIsBelowHalfTheLastPlaceRoundsToZero() {
	s.Equal(powerNum0, s.power(powerNum10, powerNumNeg21))
}

func (s *PowerTestSuite) TestPowerWhenBaseIsOneAtTheMaxExponentReturnsOne() {
	s.Equal(powerNum1, s.power(powerNum1, powerNum1000))
}

func (s *PowerTestSuite) TestPowerWhenBaseIsMinusOneWithEvenExponentReturnsOne() {
	s.Equal(powerNum1, s.power(powerNumNeg1, powerNum1000))
}

func (s *PowerTestSuite) TestPowerWhenBaseIsMinusOneWithOddExponentReturnsMinusOne() {
	s.Equal(powerNumNeg1, s.power(powerNumNeg1, powerNumNeg999))
}

func (s *PowerTestSuite) TestPowerWhenBaseIsZeroAndExponentPositiveReturnsZero() {
	s.Equal(powerNum0, s.power(powerNum0, powerNum5))
}

func (s *PowerTestSuite) TestPowerWhenBaseIsZeroAndExponentNegativeReturnsDivisionByZero() {
	err := s.failure(powerNum0, powerNumNeg1)

	s.Equal(entities.CodeDivisionByZero, err.Code())
	s.Equal("Cannot divide by zero.", err.PublicMessage())
}

func (s *PowerTestSuite) TestPowerWhenBaseIsNegativeZeroAndExponentNegativeReturnsDivisionByZero() {
	err := s.failure(powerNumNeg0Pt0, powerNumNeg3)

	s.Equal(entities.CodeDivisionByZero, err.Code())
}

func (s *PowerTestSuite) TestPowerWhenExponentIsFractionalReturnsInvalidExponent() {
	err := s.failure(powerNum2, powerNum0Pt5)

	s.Equal(entities.CodeInvalidExponent, err.Code())
	s.ErrorIs(err, operations.ErrFractionalExponent)
	s.Equal(testExponentIndex, err.Fields()[entities.FieldOperandIndex])
}

func (s *PowerTestSuite) TestPowerWhenNegativeExponentIsFractionalReturnsInvalidExponent() {
	s.Equal(entities.CodeInvalidExponent, s.failure(powerNum2, powerNumNeg0Pt5).Code())
}

func (s *PowerTestSuite) TestPowerWhenExponentIsBarelyFractionalReturnsInvalidExponent() {
	s.Equal(entities.CodeInvalidExponent, s.failure(powerNum2, powerLong2).Code())
}

func (s *PowerTestSuite) TestPowerWhenBaseIsZeroAndExponentIsFractionalReturnsInvalidExponentFirst() {
	s.Equal(entities.CodeInvalidExponent, s.failure(powerNum0, powerNumNeg0Pt5).Code())
}

func (s *PowerTestSuite) TestPowerWhenExponentIsAboveTheMaximumReturnsExponentTooLarge() {
	err := s.failure(powerNum1, powerNum1001)

	s.Equal(entities.CodeExponentTooLarge, err.Code())
	s.Equal(int64(testMaxExponent), err.Fields()[entities.FieldLimit])
}

func (s *PowerTestSuite) TestPowerWhenExponentIsBelowTheMinimumReturnsExponentTooLarge() {
	s.Equal(entities.CodeExponentTooLarge, s.failure(powerNum1, powerNumNeg1001).Code())
}

func (s *PowerTestSuite) TestPowerWhenExponentIsAHugeNumberReturnsExponentTooLarge() {
	s.Equal(entities.CodeExponentTooLarge, s.failure(powerNum2, powerLong3).Code())
}

func (s *PowerTestSuite) TestPowerWhenResultWouldExceedTheDigitBudgetReturnsResultTooLarge() {
	err := s.failure(powerNum9, powerNum1000)

	s.Equal(entities.CodeResultTooLarge, err.Code())
	s.Equal(int64(testResultBudget), err.Fields()[entities.FieldLimit])
}

func (s *PowerTestSuite) TestPowerWhenNegativeExponentWouldExceedTheDigitBudgetReturnsResultTooLarge() {
	s.Equal(entities.CodeResultTooLarge, s.failure(powerNum9, powerNumNeg1000).Code())
}

func (s *PowerTestSuite) TestPowerWhenBaseIsLongAndExponentModestReturnsResultTooLarge() {
	s.Equal(entities.CodeResultTooLarge, s.failure(powerLong3, powerNum50).Code())
}

func (s *PowerTestSuite) TestPowerWhenTwoToTheThousandFitsTheBudgetReturnsTheExactValue() {
	got := s.power(powerNum2, powerNum1000)

	s.Len(got, 302)
	s.True(strings.HasPrefix(got, testTwoToTheThousandPrefix))
}

// 7^757 has 640 digits and 7^758 has 641. A float log10 estimate lands within
// rounding distance of these, so they pin the exact integer counting.
func (s *PowerTestSuite) TestPowerWhenNonRoundBaseHasExactlyTheBudgetDigitsReturnsIt() {
	s.Len(s.power(powerNum7, powerNum757), testResultBudget)
}

func (s *PowerTestSuite) TestPowerWhenNonRoundBaseHasOneDigitOverTheBudgetReturnsResultTooLarge() {
	s.Equal(entities.CodeResultTooLarge, s.failure(powerNum7, powerNum758).Code())
}

func (s *PowerTestSuite) TestPowerWhenBoundEqualsTheBudgetComputesTheResult() {
	got := s.power(powerNum99, powerNum320)

	s.Len(got, 639)
}

func (s *PowerTestSuite) TestPowerWhenBoundIsOneOverTheBudgetReturnsResultTooLarge() {
	s.Equal(entities.CodeResultTooLarge, s.failure(powerNum99, powerNum321).Code())
}

func (s *PowerTestSuite) TestPowerDigitBudgetScalesWithTheConfiguredOperandLength() {
	small := operations.NewPower(entities.Limits{
		MaxOperandLength: testSmallOperandLen,
		MaxExponent:      testMaxExponent,
		DecimalPrecision: testPowerPrecision,
		MaxResultDigits:  testSmallBudget,
	})

	allowed, errAllowed := applyText(small, powerNum2, powerNum63)
	_, errRefused := applyText(small, powerNum2, powerNum67)

	s.Nil(errAllowed)
	s.Equal(powerLong4, allowed)
	s.Require().NotNil(errRefused)
	s.Equal(int64(testSmallBudget), errRefused.Fields()[entities.FieldLimit])
}

func (s *PowerTestSuite) TestPowerWhenTrailingZerosPushPastTheBudgetReturnsResultTooLarge() {
	s.Equal(entities.CodeResultTooLarge, s.failure(powerNum10, powerNum1000).Code())
}

func (s *PowerTestSuite) TestPowerWhenBaseIsAMillionAndResultIsTooLongReturnsResultTooLarge() {
	s.Equal(entities.CodeResultTooLarge, s.failure(powerNum1000000, powerNum1000).Code())
}

func (s *PowerTestSuite) TestPowerWhenBaseIsTenAndResultFitsReturnsEveryDigit() {
	got := s.power(powerNum10, powerNum638)

	s.Len(got, 639)
}

func (s *PowerTestSuite) TestPowerWhenFractionDigitsExceedTheBudgetReturnsResultTooLarge() {
	s.Equal(entities.CodeResultTooLarge, s.failure(powerNum0Pt1, powerNum1000).Code())
}

func (s *PowerTestSuite) TestPowerWhenFractionDigitsEqualTheBudgetComputesTheResult() {
	got := s.power(powerNum0Pt1, powerNum640)

	s.Equal(powerNum0Pt+strings.Repeat(powerNum0, 639)+powerNum1, got)
}

func (s *PowerTestSuite) TestPowerWhenFractionDigitsAreOneOverTheBudgetReturnsResultTooLarge() {
	s.Equal(entities.CodeResultTooLarge, s.failure(powerNum0Pt1, powerNum641).Code())
}

func (s *PowerTestSuite) TestPowerWhenSmallBaseHasNegativeExponentReturnsResultTooLarge() {
	s.Equal(entities.CodeResultTooLarge, s.failure(powerNum0Pt1, powerNumNeg1000).Code())
}

func (s *PowerTestSuite) TestPowerWhenTinyBaseHasNegativeExponentReturnsResultTooLarge() {
	base := powerNum0Pt + strings.Repeat(powerNum0, 61) + powerNum2

	s.Equal(entities.CodeResultTooLarge, s.failure(base, powerNumNeg1000).Code())
}

func (s *PowerTestSuite) TestPowerWhenSmallBaseHasNegativeExponentThatFitsReturnsTheExactValue() {
	s.Equal(powerNum1+strings.Repeat(powerNum0, 600), s.power(powerNum0Pt1, powerNumNeg600))
}

func (s *PowerTestSuite) TestPowerWhenBaseIsTenToTheMaxExponentReturnsResultTooLarge() {
	err := s.failure(powerNum10, testMaxPowerExponent)

	s.Equal(entities.CodeResultTooLarge, err.Code())
	s.Equal(int64(testResultBudget), err.Fields()[entities.FieldLimit])
}

func (s *PowerTestSuite) TestPowerWhenTenToTheSixHundredFitsTheBudgetReturnsIt() {
	got := s.power(powerNum10, testSixHundred)

	s.Len(got, testTenToTheSixHundredLen)
}

// The digit count is exact, with no safety margin: 10^639 has exactly 640
// digits and is the last power of ten allowed, 10^640 has 641 and is refused.
func (s *PowerTestSuite) TestPowerWhenTenHasExactlyTheBudgetDigitsReturnsIt() {
	s.Len(s.power(powerNum10, testTenToTheLastAllowed), testResultBudget)
}

func (s *PowerTestSuite) TestPowerWhenTenHasOneDigitOverTheBudgetReturnsResultTooLarge() {
	s.Equal(entities.CodeResultTooLarge, s.failure(powerNum10, testTenToTheFirstRefused).Code())
}

func (s *PowerTestSuite) TestPowerWhenBaseIsAOneWithManyZerosReturnsResultTooLarge() {
	s.Equal(entities.CodeResultTooLarge, s.failure(roundBase(testRoundBaseZeros), testMaxPowerExponent).Code())
}

func (s *PowerTestSuite) TestPowerWhenBaseIsTinyAndExponentPositiveReturnsResultTooLarge() {
	s.Equal(entities.CodeResultTooLarge, s.failure(tinyBase(testNarrowFractionZeros), testMaxPowerExponent).Code())
}

func (s *PowerTestSuite) TestPowerWhenBaseIsTinyAndExponentNegativeReturnsResultTooLarge() {
	s.Equal(entities.CodeResultTooLarge, s.failure(tinyBase(testNarrowFractionZeros), testNegMaxPowerExponent).Code())
}

func (s *PowerTestSuite) TestPowerWhenPointOneToTheMaxExponentReturnsResultTooLarge() {
	s.Equal(entities.CodeResultTooLarge, s.failure(powerNum0Pt1, testMaxPowerExponent).Code())
}

func (s *PowerTestSuite) TestPowerWhenPointOneToTheSixHundredFitsTheBudgetReturnsIt() {
	s.Len(s.power(powerNum0Pt1, testSixHundred), testPointOneToTheSixHundred)
}

func (s *PowerTestSuite) TestPowerWhenFractionHasManyPlacesButFitsTheBudgetReturnsIt() {
	s.LessOrEqual(len(s.power(powerNum0Pt5, testHalfToTheFiveHundred)), testResultBudget)
}

func (s *PowerTestSuite) TestPowerWhenNegativeBaseTenHasOneDigitOverTheBudgetReturnsResultTooLarge() {
	err := s.failure(testSignedBoundaryBase, testSignedBoundaryExponent)

	s.Equal(entities.CodeResultTooLarge, err.Code())
}

// The reciprocal of a huge round base rounds to zero at the configured precision,
// so the result is tiny and the guard lets it through.
func (s *PowerTestSuite) TestPowerWhenNegativeExponentBaseIsRoundAndHugeReturnsZero() {
	s.Equal(powerNum0, s.power(roundBase(testRoundBaseZeros), testNegMaxPowerExponent))
}

func (s *PowerTestSuite) TestPowerWhenContextEndedBeforeTheLoopReturnsRequestTimeout() {
	_, err := s.op.Apply(cancelled(), []decimal.Decimal{mustDecimal(powerNum2), mustDecimal(powerNum10)})

	s.Require().NotNil(err)
	s.Equal(apperror.CodeRequestTimeout, err.Code())
	s.ErrorIs(err, context.Canceled)
}

func (s *PowerTestSuite) TestPowerWhenDeadlineHasPassedKeepsTheDeadlineCause() {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	_, err := s.op.Apply(ctx, []decimal.Decimal{mustDecimal(powerNum2), mustDecimal(powerNum10)})

	s.Require().NotNil(err)
	s.True(errors.Is(err, context.DeadlineExceeded))
}

func (s *PowerTestSuite) TestPowerWhenResultIsOneIgnoresAnEndedContext() {
	got, err := s.op.Apply(cancelled(), []decimal.Decimal{mustDecimal(powerNum5), mustDecimal(powerNum0)})

	s.Nil(err)
	s.Equal(powerNum1, got.String())
}
