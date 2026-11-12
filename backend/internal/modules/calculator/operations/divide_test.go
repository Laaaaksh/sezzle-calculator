package operations_test

import (
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/operations"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
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
	s.Equal("4", s.divide("12", "3"))
}

func (s *DivideTestSuite) TestDivideWhenQuotientTerminatesKeepsItWithoutPadding() {
	s.Equal("2.5", s.divide("10", "4"))
}

func (s *DivideTestSuite) TestDivideWhenSignsDifferReturnsANegativeQuotient() {
	s.Equal("-4", s.divide("12", "-3"))
}

func (s *DivideTestSuite) TestDivideWhenBothNegativeReturnsAPositiveQuotient() {
	s.Equal("4", s.divide("-12", "-3"))
}

func (s *DivideTestSuite) TestDivideWhenDividendIsZeroReturnsZero() {
	s.Equal("0", s.divide("0", "7"))
}

func (s *DivideTestSuite) TestDivideWhenNegativeZeroIsDividedReturnsZeroWithoutASign() {
	s.Equal("0", s.divide("-0", "-7"))
}

func (s *DivideTestSuite) TestDivideWhenDivisorIsZeroReturnsDivisionByZero() {
	_, err := applyText(s.op, "5", "0")

	s.Require().NotNil(err)
	s.Equal(apperror.CodeDivisionByZero, err.Code())
	s.ErrorIs(err, operations.ErrDivisionByZero)
}

func (s *DivideTestSuite) TestDivideWhenDivisorIsZeroUsesTheDocumentedMessage() {
	_, err := applyText(s.op, "5", "0")

	s.Require().NotNil(err)
	s.Equal(testDivideByZeroText, err.PublicMessage())
}

func (s *DivideTestSuite) TestDivideWhenBothAreZeroReturnsDivisionByZero() {
	_, err := applyText(s.op, "0", "0")

	s.Require().NotNil(err)
	s.Equal(apperror.CodeDivisionByZero, err.Code())
}

func (s *DivideTestSuite) TestDivideWhenDivisorIsNegativeZeroReturnsDivisionByZero() {
	_, err := applyText(s.op, "5", "-0.00")

	s.Require().NotNil(err)
	s.Equal(apperror.CodeDivisionByZero, err.Code())
}

func (s *DivideTestSuite) TestDivideOneByThreeRoundsToTheConfiguredPrecision() {
	s.Equal("0.33333333333333333333", s.divide("1", "3"))
}

func (s *DivideTestSuite) TestDivideNegativeOneByThreeRoundsToTheConfiguredPrecision() {
	s.Equal("-0.33333333333333333333", s.divide("-1", "3"))
}

func (s *DivideTestSuite) TestDivideTwoByThreeRoundsTheLastDigitUp() {
	s.Equal("0.66666666666666666667", s.divide("2", "3"))
}

func (s *DivideTestSuite) TestDivideNegativeTwoByThreeRoundsTheLastDigitAwayFromZero() {
	s.Equal("-0.66666666666666666667", s.divide("-2", "3"))
}

func (s *DivideTestSuite) TestDivideWhenQuotientIsBelowHalfAUnitRoundsDown() {
	s.Equal("0.14", s.divideTwoPlaces("1", "7"))
}

func (s *DivideTestSuite) TestDivideWhenTieFollowsAnEvenDigitRoundsDown() {
	s.Equal("0.12", s.divideTwoPlaces("1", "8"))
}

func (s *DivideTestSuite) TestDivideWhenTieFollowsAnOddDigitRoundsUp() {
	s.Equal("0.38", s.divideTwoPlaces("3", "8"))
}

func (s *DivideTestSuite) TestDivideWhenNegativeTieFollowsAnEvenDigitRoundsTowardZero() {
	s.Equal("-0.12", s.divideTwoPlaces("-1", "8"))
}

func (s *DivideTestSuite) TestDivideWhenNegativeTieFollowsAnOddDigitRoundsAwayFromZero() {
	s.Equal("-0.38", s.divideTwoPlaces("-3", "8"))
}

func (s *DivideTestSuite) TestDivideWhenDivisorIsNegativeAndTieFollowsAnOddDigitRoundsAwayFromZero() {
	s.Equal("-0.38", s.divideTwoPlaces("3", "-8"))
}

func (s *DivideTestSuite) TestDivideWhenOnlyJustAboveATieRoundsUp() {
	s.Equal("0.13", s.divideTwoPlaces("1.0001", "8"))
}

func (s *DivideTestSuite) TestDivideWhenOnlyJustBelowATieRoundsDown() {
	s.Equal("0.12", s.divideTwoPlaces("0.9999", "8"))
}

func (s *DivideTestSuite) TestDivideHalfEvenDoesNotBiasTheTotalOfTies() {
	// 0.5 and 1.5 are both ties at zero places. Half-up gives 1 + 2 = 3, a
	// total pushed up by 0.5; half-even gives 0 + 2 = 2, matching the true
	// total of 2.
	whole := operations.NewDivide(testZeroPlaces)
	first, errFirst := applyText(whole, "1", "2")
	second, errSecond := applyText(whole, "3", "2")

	s.Nil(errFirst)
	s.Nil(errSecond)
	s.Equal("0", first)
	s.Equal("2", second)
}

func (s *DivideTestSuite) TestDivideWhenOperandsAreVeryLargeKeepsEveryDigit() {
	s.Equal("99999999999999999999999999999999", s.divide("99999999999999999999999999999999", "1"))
}

func (s *DivideTestSuite) TestDivideWhenResultIsVerySmallKeepsSignificantDigits() {
	s.Equal("0.00000000000000000001", s.divide("1", "100000000000000000000"))
}

func (s *DivideTestSuite) TestDivideWhenResultIsBelowHalfTheLastPlaceRoundsToZero() {
	s.Equal("0", s.divide("1", "300000000000000000000"))
}

func (s *DivideTestSuite) TestDivideWhenOperandsHaveManyDecimalPlacesDividesExactly() {
	s.Equal("0.5", s.divide("0.1234567890123456789", "0.2469135780246913578"))
}

func (s *DivideTestSuite) TestDivideZeroPointOneByTenIsExactlyPointZeroOne() {
	s.Equal("0.01", s.divide("0.1", "10"))
}

func (s *DivideTestSuite) TestNewDivideWhenPrecisionIsNegativeRoundsToWholeNumbers() {
	got, err := applyText(operations.NewDivide(testNegativePlaces), "5", "2")

	s.Nil(err)
	s.Equal("2", got)
}

func (s *DivideTestSuite) TestNewDivideWhenPrecisionExceedsInt32ClampsInsteadOfWrapping() {
	// A wrapped int32 would turn this precision negative and silently round to
	// whole numbers. Dividing by zero returns before any rounding, so the test
	// proves construction and Apply stay safe without allocating 2^31 digits.
	clamped := operations.NewDivide(testPrecisionBeyondInt32)

	_, err := applyText(clamped, "1", "0")

	s.Require().NotNil(err)
	s.Equal(apperror.CodeDivisionByZero, err.Code())
}

func (s *DivideTestSuite) TestDivideOfTheLargestByTheSmallestOperandPrintsFarBelowThePowerBudget() {
	dividend := strings.Repeat(digitNine, testNarrowOperandDigits)

	got := s.divide(dividend, tinyBase(testNarrowFractionZeros))

	s.Less(len(got), testResultBudget)
}
