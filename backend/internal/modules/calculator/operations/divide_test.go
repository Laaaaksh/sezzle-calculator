package operations_test

import (
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/operations"
)

const (
	testConfiguredPrecision  = 20
	testTwoPlaces            = 2
	testZeroPlaces           = 0
	testNegativePlaces       = -3
	testPrecisionBeyondInt32 = math.MaxInt32 + 1
	testDivideByZeroText     = "Cannot divide by zero."
)

type DivideTestSuite struct {
	suite.Suite
	op      *operations.Divide
	twoDecs *operations.Divide
}

func TestDivideTestSuite(t *testing.T) {
	suite.Run(t, new(DivideTestSuite))
}

func (s *DivideTestSuite) SetupTest() {
	s.op = operations.NewDivide(testConfiguredPrecision)
	s.twoDecs = operations.NewDivide(testTwoPlaces)
}

func (s *DivideTestSuite) divide(left, right string) string {
	got, err := applyText(s.op, left, right)
	s.Require().Nil(err)
	return got
}

func (s *DivideTestSuite) divideTwoPlaces(left, right string) string {
	got, err := applyText(s.twoDecs, left, right)
	s.Require().Nil(err)
	return got
}

func (s *DivideTestSuite) TestDivideNameIsDivide() {
	s.Equal(entities.OpDivide, s.op.Name())
}

func (s *DivideTestSuite) TestDivideArityIsTwo() {
	s.Equal(entities.ArityBinary, s.op.Arity())
}

func (s *DivideTestSuite) TestDivideWhenDivisionIsExactReturnsTheQuotient() {
	s.Equal(divideNum4, s.divide(divideNum12, divideNum3))
}

func (s *DivideTestSuite) TestDivideWhenQuotientTerminatesKeepsItWithoutPadding() {
	s.Equal(divideNum2Pt5, s.divide(divideNum10, divideNum4))
}

func (s *DivideTestSuite) TestDivideWhenSignsDifferReturnsANegativeQuotient() {
	s.Equal(divideNumNeg4, s.divide(divideNum12, divideNumNeg3))
}

func (s *DivideTestSuite) TestDivideWhenBothNegativeReturnsAPositiveQuotient() {
	s.Equal(divideNum4, s.divide(divideNumNeg12, divideNumNeg3))
}

func (s *DivideTestSuite) TestDivideWhenDividendIsZeroReturnsZero() {
	s.Equal(divideNum0, s.divide(divideNum0, divideNum7))
}

func (s *DivideTestSuite) TestDivideWhenNegativeZeroIsDividedReturnsZeroWithoutASign() {
	s.Equal(divideNum0, s.divide(divideNumNeg0, divideNumNeg7))
}

func (s *DivideTestSuite) TestDivideWhenDivisorIsZeroReturnsDivisionByZero() {
	_, err := applyText(s.op, divideNum5, divideNum0)

	s.Require().NotNil(err)
	s.Equal(entities.CodeDivisionByZero, err.Code())
	s.ErrorIs(err, operations.ErrDivisionByZero)
}

func (s *DivideTestSuite) TestDivideWhenDivisorIsZeroUsesTheDocumentedMessage() {
	_, err := applyText(s.op, divideNum5, divideNum0)

	s.Require().NotNil(err)
	s.Equal(testDivideByZeroText, err.PublicMessage())
}

func (s *DivideTestSuite) TestDivideWhenBothAreZeroReturnsDivisionByZero() {
	_, err := applyText(s.op, divideNum0, divideNum0)

	s.Require().NotNil(err)
	s.Equal(entities.CodeDivisionByZero, err.Code())
}

func (s *DivideTestSuite) TestDivideWhenDivisorIsNegativeZeroReturnsDivisionByZero() {
	_, err := applyText(s.op, divideNum5, divideNumNeg0Pt00)

	s.Require().NotNil(err)
	s.Equal(entities.CodeDivisionByZero, err.Code())
}

func (s *DivideTestSuite) TestDivideOneByThreeRoundsToTheConfiguredPrecision() {
	s.Equal(divideLong1, s.divide(divideNum1, divideNum3))
}

func (s *DivideTestSuite) TestDivideNegativeOneByThreeRoundsToTheConfiguredPrecision() {
	s.Equal(divideLong2, s.divide(divideNumNeg1, divideNum3))
}

func (s *DivideTestSuite) TestDivideTwoByThreeRoundsTheLastDigitUp() {
	s.Equal(divideLong3, s.divide(divideNum2, divideNum3))
}

func (s *DivideTestSuite) TestDivideNegativeTwoByThreeRoundsTheLastDigitAwayFromZero() {
	s.Equal(divideLong4, s.divide(divideNumNeg2, divideNum3))
}

func (s *DivideTestSuite) TestDivideWhenQuotientIsBelowHalfAUnitRoundsDown() {
	s.Equal(divideNum0Pt14, s.divideTwoPlaces(divideNum1, divideNum7))
}

func (s *DivideTestSuite) TestDivideWhenTieFollowsAnEvenDigitRoundsDown() {
	s.Equal(divideNum0Pt12, s.divideTwoPlaces(divideNum1, divideNum8))
}

func (s *DivideTestSuite) TestDivideWhenTieFollowsAnOddDigitRoundsUp() {
	s.Equal(divideNum0Pt38, s.divideTwoPlaces(divideNum3, divideNum8))
}

func (s *DivideTestSuite) TestDivideWhenNegativeTieFollowsAnEvenDigitRoundsTowardZero() {
	s.Equal(divideNumNeg0Pt12, s.divideTwoPlaces(divideNumNeg1, divideNum8))
}

func (s *DivideTestSuite) TestDivideWhenNegativeTieFollowsAnOddDigitRoundsAwayFromZero() {
	s.Equal(divideNumNeg0Pt38, s.divideTwoPlaces(divideNumNeg3, divideNum8))
}

func (s *DivideTestSuite) TestDivideWhenDivisorIsNegativeAndTieFollowsAnOddDigitRoundsAwayFromZero() {
	s.Equal(divideNumNeg0Pt38, s.divideTwoPlaces(divideNum3, divideNumNeg8))
}

func (s *DivideTestSuite) TestDivideWhenOnlyJustAboveATieRoundsUp() {
	s.Equal(divideNum0Pt13, s.divideTwoPlaces(divideNum1Pt0001, divideNum8))
}

func (s *DivideTestSuite) TestDivideWhenOnlyJustBelowATieRoundsDown() {
	s.Equal(divideNum0Pt12, s.divideTwoPlaces(divideNum0Pt9999, divideNum8))
}

func (s *DivideTestSuite) TestDivideHalfEvenDoesNotBiasTheTotalOfTies() {
	// 0.5 and 1.5 are both ties at zero places. Half-up gives 1 + 2 = 3, a
	// total pushed up by 0.5; half-even gives 0 + 2 = 2, matching the true
	// total of 2.
	whole := operations.NewDivide(testZeroPlaces)
	first, errFirst := applyText(whole, divideNum1, divideNum2)
	second, errSecond := applyText(whole, divideNum3, divideNum2)

	s.Nil(errFirst)
	s.Nil(errSecond)
	s.Equal(divideNum0, first)
	s.Equal(divideNum2, second)
}

func (s *DivideTestSuite) TestDivideWhenOperandsAreVeryLargeKeepsEveryDigit() {
	s.Equal(divideLong5, s.divide(divideLong5, divideNum1))
}

func (s *DivideTestSuite) TestDivideWhenResultIsVerySmallKeepsSignificantDigits() {
	s.Equal(divideLong6, s.divide(divideNum1, divideLong7))
}

func (s *DivideTestSuite) TestDivideWhenResultIsBelowHalfTheLastPlaceRoundsToZero() {
	s.Equal(divideNum0, s.divide(divideNum1, divideLong8))
}

func (s *DivideTestSuite) TestDivideWhenOperandsHaveManyDecimalPlacesDividesExactly() {
	s.Equal(divideNum0Pt5, s.divide(divideLong9, divideLong10))
}

func (s *DivideTestSuite) TestDivideZeroPointOneByTenIsExactlyPointZeroOne() {
	s.Equal(divideNum0Pt01, s.divide(divideNum0Pt1, divideNum10))
}

func (s *DivideTestSuite) TestNewDivideWhenPrecisionIsNegativeRoundsToWholeNumbers() {
	got, err := applyText(operations.NewDivide(testNegativePlaces), divideNum5, divideNum2)

	s.Nil(err)
	s.Equal(divideNum2, got)
}

func (s *DivideTestSuite) TestNewDivideWhenPrecisionExceedsInt32ClampsInsteadOfWrapping() {
	// A wrapped int32 would turn this precision negative and silently round to
	// whole numbers. Dividing by zero returns before any rounding, so the test
	// proves construction and Apply stay safe without allocating 2^31 digits.
	clamped := operations.NewDivide(testPrecisionBeyondInt32)

	_, err := applyText(clamped, divideNum1, divideNum0)

	s.Require().NotNil(err)
	s.Equal(entities.CodeDivisionByZero, err.Code())
}

func (s *DivideTestSuite) TestDivideOfTheLargestByTheSmallestOperandPrintsFarBelowThePowerBudget() {
	dividend := strings.Repeat(digitNine, testNarrowOperandDigits)

	got := s.divide(dividend, tinyBase(testNarrowFractionZeros))

	s.Less(len(got), testResultBudget)
}
