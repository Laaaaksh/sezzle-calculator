package operations_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
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

func (s *SubtractTestSuite) TestSubtractNameIsSubtract() {
	s.Equal(entities.OpSubtract, s.op.Name())
}

func (s *SubtractTestSuite) TestSubtractArityIsTwo() {
	s.Equal(entities.ArityBinary, s.op.Arity())
}

func (s *SubtractTestSuite) TestSubtractWhenBothPositiveReturnsTheDifference() {
	s.Equal("7", s.subtract("10", "3"))
}

func (s *SubtractTestSuite) TestSubtractWhenSecondIsLargerReturnsANegative() {
	s.Equal("-7", s.subtract("3", "10"))
}

func (s *SubtractTestSuite) TestSubtractWhenBothNegativeReturnsTheDifference() {
	s.Equal("1", s.subtract("-2", "-3"))
}

func (s *SubtractTestSuite) TestSubtractWhenSubtrahendIsZeroReturnsTheFirst() {
	s.Equal("-4.5", s.subtract("-4.5", "0"))
}

func (s *SubtractTestSuite) TestSubtractWhenFirstIsZeroReturnsTheNegatedSecond() {
	s.Equal("-4.5", s.subtract("0", "4.5"))
}

func (s *SubtractTestSuite) TestSubtractWhenOperandsAreEqualReturnsZero() {
	s.Equal("0", s.subtract("1.25", "1.25"))
}

func (s *SubtractTestSuite) TestSubtractPointThreeMinusPointOneIsExactlyPointTwo() {
	s.Equal("0.2", s.subtract("0.3", "0.1"))
}

func (s *SubtractTestSuite) TestSubtractWhenOperandsAreVeryLargeKeepsEveryDigit() {
	s.Equal("99999999999999999999999999999999", s.subtract("100000000000000000000000000000000", "1"))
}

func (s *SubtractTestSuite) TestSubtractWhenOperandsAreVerySmallKeepsEveryDigit() {
	s.Equal("0.000000000000000000000000001", s.subtract("0.000000000000000000000000003", "0.000000000000000000000000002"))
}

func (s *SubtractTestSuite) TestSubtractWhenOperandsHaveManyDecimalPlacesKeepsThemAll() {
	s.Equal("0.1234567890123456788", s.subtract("0.1234567890123456789", "0.0000000000000000001"))
}

func (s *SubtractTestSuite) TestSubtractWhenZeroMinusZeroReturnsZeroWithoutASign() {
	s.Equal("0", s.subtract("-0", "0"))
}
