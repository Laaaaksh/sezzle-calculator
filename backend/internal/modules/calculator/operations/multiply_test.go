package operations_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/operations"
)

type MultiplyTestSuite struct {
	suite.Suite
	op *operations.Multiply
}

func TestMultiplyTestSuite(t *testing.T) {
	suite.Run(t, new(MultiplyTestSuite))
}

func (s *MultiplyTestSuite) SetupTest() {
	s.op = operations.NewMultiply()
}

func (s *MultiplyTestSuite) multiply(left, right string) string {
	got, err := applyText(s.op, left, right)
	s.Require().Nil(err)
	return got
}

func (s *MultiplyTestSuite) TestMultiplyNameIsMultiply() {
	s.Equal(entities.OpMultiply, s.op.Name())
}

func (s *MultiplyTestSuite) TestMultiplyArityIsTwo() {
	s.Equal(entities.ArityBinary, s.op.Arity())
}

func (s *MultiplyTestSuite) TestMultiplyWhenBothPositiveReturnsTheProduct() {
	s.Equal(multiplyNum12, s.multiply(multiplyNum3, multiplyNum4))
}

func (s *MultiplyTestSuite) TestMultiplyWhenSignsDifferReturnsANegativeProduct() {
	s.Equal(multiplyNumNeg12, s.multiply(multiplyNum3, multiplyNumNeg4))
}

func (s *MultiplyTestSuite) TestMultiplyWhenBothNegativeReturnsAPositiveProduct() {
	s.Equal(multiplyNum12, s.multiply(multiplyNumNeg3, multiplyNumNeg4))
}

func (s *MultiplyTestSuite) TestMultiplyWhenOneOperandIsZeroReturnsZero() {
	s.Equal(multiplyNum0, s.multiply(multiplyNum123Pt456, multiplyNum0))
}

func (s *MultiplyTestSuite) TestMultiplyWhenNegativeTimesZeroReturnsZeroWithoutASign() {
	s.Equal(multiplyNum0, s.multiply(multiplyNumNeg5, multiplyNum0))
}

func (s *MultiplyTestSuite) TestMultiplyWhenOneOperandIsOneReturnsTheOther() {
	s.Equal(multiplyNumNeg9Pt75, s.multiply(multiplyNumNeg9Pt75, multiplyNum1))
}

func (s *MultiplyTestSuite) TestMultiplyPointOneTimesPointThreeIsExactlyPointZeroThree() {
	s.Equal(multiplyNum0Pt03, s.multiply(multiplyNum0Pt1, multiplyNum0Pt3))
}

func (s *MultiplyTestSuite) TestMultiplyWhenOperandsAreVeryLargeKeepsEveryDigit() {
	s.Equal(multiplyLong1, s.multiply(multiplyLong2, multiplyLong2))
}

func (s *MultiplyTestSuite) TestMultiplyWhenOperandsAreVerySmallKeepsEveryDigit() {
	s.Equal(multiplyLong3, s.multiply(multiplyNum0Pt000000000001, multiplyLong4))
}

func (s *MultiplyTestSuite) TestMultiplyWhenOperandsHaveManyDecimalPlacesKeepsThemAll() {
	s.Equal(multiplyLong5, s.multiply(multiplyLong6, multiplyLong7))
}

// (10^64 - 1)^2 = 10^128 - 2*10^64 + 1: 63 nines, an 8, 63 zeros, a 1.
func (s *MultiplyTestSuite) maxOperandSquare() string {
	return strings.Repeat(digitNine, testMaxOperandLength-1) + digitEight +
		strings.Repeat(digitZero, testMaxOperandLength-1) + digitOne
}

func (s *MultiplyTestSuite) TestMultiplyOfTwoMaxLengthOperandsReturnsEveryDigit() {
	maxOperand := strings.Repeat(digitNine, testMaxOperandLength)

	s.Equal(s.maxOperandSquare(), s.multiply(maxOperand, maxOperand))
}

func (s *MultiplyTestSuite) TestMultiplyOfTwoMaxLengthOperandsFitsTheResultBudget() {
	maxOperand := strings.Repeat(digitNine, testMaxOperandLength)
	product := mustDecimal(s.multiply(maxOperand, maxOperand))

	s.Equal(int64(2*testMaxOperandLength), operations.DigitsOf(product))
	s.Nil(operations.CheckResultSize(product, testResultBudget))
}
