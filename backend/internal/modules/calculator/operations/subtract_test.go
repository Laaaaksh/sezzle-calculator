package operations_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/operations"
)

type SubtractTestSuite struct {
	suite.Suite
	op *operations.Subtract
}

func TestSubtractTestSuite(t *testing.T) {
	suite.Run(t, new(SubtractTestSuite))
}

func (s *SubtractTestSuite) SetupTest() {
	s.op = operations.NewSubtract()
}

func (s *SubtractTestSuite) subtract(left, right string) string {
	got, err := applyText(s.op, left, right)
	s.Require().Nil(err)
	return got
}

func (s *SubtractTestSuite) TestSubtractWhenBothPositiveReturnsTheDifference() {
	s.Equal(subtractNum7, s.subtract(subtractNum10, subtractNum3))
}

func (s *SubtractTestSuite) TestSubtractWhenSecondIsLargerReturnsANegative() {
	s.Equal(subtractNumNeg7, s.subtract(subtractNum3, subtractNum10))
}

func (s *SubtractTestSuite) TestSubtractWhenBothNegativeReturnsTheDifference() {
	s.Equal(subtractNum1, s.subtract(subtractNumNeg2, subtractNumNeg3))
}

func (s *SubtractTestSuite) TestSubtractWhenSubtrahendIsZeroReturnsTheFirst() {
	s.Equal(subtractNumNeg4Pt5, s.subtract(subtractNumNeg4Pt5, subtractNum0))
}

func (s *SubtractTestSuite) TestSubtractWhenFirstIsZeroReturnsTheNegatedSecond() {
	s.Equal(subtractNumNeg4Pt5, s.subtract(subtractNum0, subtractNum4Pt5))
}

func (s *SubtractTestSuite) TestSubtractWhenOperandsAreEqualReturnsZero() {
	s.Equal(subtractNum0, s.subtract(subtractNum1Pt25, subtractNum1Pt25))
}

func (s *SubtractTestSuite) TestSubtractPointThreeMinusPointOneIsExactlyPointTwo() {
	s.Equal(subtractNum0Pt2, s.subtract(subtractNum0Pt3, subtractNum0Pt1))
}

func (s *SubtractTestSuite) TestSubtractWhenOperandsAreVeryLargeKeepsEveryDigit() {
	s.Equal(subtractLong1, s.subtract(subtractLong2, subtractNum1))
}

func (s *SubtractTestSuite) TestSubtractWhenOperandsAreVerySmallKeepsEveryDigit() {
	s.Equal(subtractLong3, s.subtract(subtractLong4, subtractLong5))
}

func (s *SubtractTestSuite) TestSubtractWhenOperandsHaveManyDecimalPlacesKeepsThemAll() {
	s.Equal(subtractLong6, s.subtract(subtractLong7, subtractLong8))
}

func (s *SubtractTestSuite) TestSubtractWhenZeroMinusZeroReturnsZeroWithoutASign() {
	s.Equal(subtractNum0, s.subtract(subtractNumNeg0, subtractNum0))
}
