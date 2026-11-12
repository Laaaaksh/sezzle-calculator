package operations_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/operations"
)

type AddTestSuite struct {
	suite.Suite
	op *operations.Add
}

func TestAddTestSuite(t *testing.T) {
	suite.Run(t, new(AddTestSuite))
}

func (s *AddTestSuite) SetupTest() {
	s.op = operations.NewAdd()
}

func (s *AddTestSuite) add(left, right string) string {
	got, err := applyText(s.op, left, right)
	s.Require().Nil(err)
	return got
}

func (s *AddTestSuite) TestAddNameIsAdd() {
	s.Equal(entities.OpAdd, s.op.Name())
}

func (s *AddTestSuite) TestAddArityIsTwo() {
	s.Equal(entities.ArityBinary, s.op.Arity())
}

func (s *AddTestSuite) TestAddWhenBothPositiveReturnsTheSum() {
	s.Equal("5", s.add("2", "3"))
}

func (s *AddTestSuite) TestAddWhenBothNegativeReturnsANegativeSum() {
	s.Equal("-5", s.add("-2", "-3"))
}

func (s *AddTestSuite) TestAddWhenSignsDifferReturnsTheDifference() {
	s.Equal("-1", s.add("2", "-3"))
}

func (s *AddTestSuite) TestAddWhenOneOperandIsZeroReturnsTheOther() {
	s.Equal("7.5", s.add("7.5", "0"))
}

func (s *AddTestSuite) TestAddWhenOperandsCancelReturnsZero() {
	s.Equal("0", s.add("4.2", "-4.2"))
}

func (s *AddTestSuite) TestAddPointOneAndPointTwoIsExactlyPointThree() {
	s.Equal("0.3", s.add("0.1", "0.2"))
}

func (s *AddTestSuite) TestAddWhenOperandsAreVeryLargeKeepsEveryDigit() {
	s.Equal("100000000000000000000000000000000", s.add("99999999999999999999999999999999", "1"))
}

func (s *AddTestSuite) TestAddWhenOperandsAreVerySmallKeepsEveryDigit() {
	s.Equal("0.000000000000000000000000002", s.add("0.000000000000000000000000001", "0.000000000000000000000000001"))
}

func (s *AddTestSuite) TestAddWhenOperandsHaveManyDecimalPlacesKeepsThemAll() {
	s.Equal("0.123456789012345679", s.add("0.1234567890123456789", "0.0000000000000000001"))
}

func (s *AddTestSuite) TestAddWhenResultIsNegativeZeroReturnsZero() {
	s.Equal("0", s.add("-0", "-0"))
}
