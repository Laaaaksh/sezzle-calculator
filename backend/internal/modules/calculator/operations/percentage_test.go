package operations_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/operations"
)

type PercentageTestSuite struct {
	suite.Suite
	op *operations.Percentage
}

func TestPercentageTestSuite(t *testing.T) {
	suite.Run(t, new(PercentageTestSuite))
}

func (s *PercentageTestSuite) SetupTest() {
	s.op = operations.NewPercentage()
}

func (s *PercentageTestSuite) percentage(percent, amount string) string {
	got, err := applyText(s.op, percent, amount)
	s.Require().Nil(err)
	return got
}

func (s *PercentageTestSuite) TestPercentageNameIsPercentage() {
	s.Equal(entities.OpPercentage, s.op.Name())
}

func (s *PercentageTestSuite) TestPercentageArityIsTwo() {
	s.Equal(entities.ArityBinary, s.op.Arity())
}

func (s *PercentageTestSuite) TestPercentageTwentyOfFiftyIsTen() {
	s.Equal("10", s.percentage("20", "50"))
}

func (s *PercentageTestSuite) TestPercentageIsPercentFirstAndAmountSecond() {
	s.Equal("10", s.percentage("50", "20"))
	s.Equal("0.2", s.percentage("20", "1"))
}

func (s *PercentageTestSuite) TestPercentageOfAHundredReturnsTheAmount() {
	s.Equal("7.5", s.percentage("100", "7.5"))
}

func (s *PercentageTestSuite) TestPercentageOverAHundredReturnsMoreThanTheAmount() {
	s.Equal("15", s.percentage("150", "10"))
}

func (s *PercentageTestSuite) TestPercentageOfZeroPercentIsZero() {
	s.Equal("0", s.percentage("0", "100"))
}

func (s *PercentageTestSuite) TestPercentageOfAZeroAmountIsZero() {
	s.Equal("0", s.percentage("12.5", "0"))
}

func (s *PercentageTestSuite) TestPercentageWhenPercentIsFractionalIsExact() {
	s.Equal("10", s.percentage("12.5", "80"))
}

func (s *PercentageTestSuite) TestPercentageWhenPercentIsNegativeReturnsNegative() {
	s.Equal("-5", s.percentage("-10", "50"))
}

func (s *PercentageTestSuite) TestPercentageWhenAmountIsNegativeReturnsNegative() {
	s.Equal("-5", s.percentage("10", "-50"))
}

func (s *PercentageTestSuite) TestPercentageWhenBothAreNegativeReturnsPositive() {
	s.Equal("5", s.percentage("-10", "-50"))
}

func (s *PercentageTestSuite) TestPercentageOfOnePercentOfThreeIsExactlyPointZeroThree() {
	s.Equal("0.03", s.percentage("1", "3"))
}

func (s *PercentageTestSuite) TestPercentageDoesNotRoundManyDecimalPlaces() {
	s.Equal("0.99999", s.percentage("33.333", "3"))
}

func (s *PercentageTestSuite) TestPercentageWhenResultIsLargeKeepsEveryDigit() {
	s.Equal(testPercentSquareWant, s.percentage(testPercentSquare, testPercentSquare))
}

func (s *PercentageTestSuite) TestPercentageWhenResultIsTinyKeepsEveryDigit() {
	s.Equal(testTinyUnitSquareWnt, s.percentage(testTinyUnit, testTinyUnit))
}
