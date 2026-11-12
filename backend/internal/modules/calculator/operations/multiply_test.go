package operations_test

import (
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
	s.Equal("12", s.multiply("3", "4"))
}

func (s *MultiplyTestSuite) TestMultiplyWhenSignsDifferReturnsANegativeProduct() {
	s.Equal("-12", s.multiply("3", "-4"))
}

func (s *MultiplyTestSuite) TestMultiplyWhenBothNegativeReturnsAPositiveProduct() {
	s.Equal("12", s.multiply("-3", "-4"))
}

func (s *MultiplyTestSuite) TestMultiplyWhenOneOperandIsZeroReturnsZero() {
	s.Equal("0", s.multiply("123.456", "0"))
}

func (s *MultiplyTestSuite) TestMultiplyWhenNegativeTimesZeroReturnsZeroWithoutASign() {
	s.Equal("0", s.multiply("-5", "0"))
}

func (s *MultiplyTestSuite) TestMultiplyWhenOneOperandIsOneReturnsTheOther() {
	s.Equal("-9.75", s.multiply("-9.75", "1"))
}

func (s *MultiplyTestSuite) TestMultiplyPointOneTimesPointThreeIsExactlyPointZeroThree() {
	s.Equal("0.03", s.multiply("0.1", "0.3"))
}

func (s *MultiplyTestSuite) TestMultiplyWhenOperandsAreVeryLargeKeepsEveryDigit() {
	s.Equal("10000000000000000000000000000000000000000000000000000000000000000", s.multiply("100000000000000000000000000000000", "100000000000000000000000000000000"))
}

func (s *MultiplyTestSuite) TestMultiplyWhenOperandsAreVerySmallKeepsEveryDigit() {
	s.Equal("0.000000000000000000000000000000001", s.multiply("0.000000000001", "0.000000000000000000001"))
}

func (s *MultiplyTestSuite) TestMultiplyWhenOperandsHaveManyDecimalPlacesKeepsThemAll() {
	s.Equal("12.193263113702179522374638011112635269", s.multiply("1.234567890123456789", "9.876543210987654321"))
}
