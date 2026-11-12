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
	s.Equal(addNum5, s.add(addNum2, addNum3))
}

func (s *AddTestSuite) TestAddWhenBothNegativeReturnsANegativeSum() {
	s.Equal(addNumNeg5, s.add(addNumNeg2, addNumNeg3))
}

func (s *AddTestSuite) TestAddWhenSignsDifferReturnsTheDifference() {
	s.Equal(addNumNeg1, s.add(addNum2, addNumNeg3))
}

func (s *AddTestSuite) TestAddWhenOneOperandIsZeroReturnsTheOther() {
	s.Equal(addNum7Pt5, s.add(addNum7Pt5, addNum0))
}

func (s *AddTestSuite) TestAddWhenOperandsCancelReturnsZero() {
	s.Equal(addNum0, s.add(addNum4Pt2, addNumNeg4Pt2))
}

func (s *AddTestSuite) TestAddPointOneAndPointTwoIsExactlyPointThree() {
	s.Equal(addNum0Pt3, s.add(addNum0Pt1, addNum0Pt2))
}

func (s *AddTestSuite) TestAddWhenOperandsAreVeryLargeKeepsEveryDigit() {
	s.Equal(addLong1, s.add(addLong2, addNum1))
}

func (s *AddTestSuite) TestAddWhenOperandsAreVerySmallKeepsEveryDigit() {
	s.Equal(addLong3, s.add(addLong4, addLong4))
}

func (s *AddTestSuite) TestAddWhenOperandsHaveManyDecimalPlacesKeepsThemAll() {
	s.Equal(addLong5, s.add(addLong6, addLong7))
}

func (s *AddTestSuite) TestAddWhenResultIsNegativeZeroReturnsZero() {
	s.Equal(addNum0, s.add(addNumNeg0, addNumNeg0))
}
