package operations_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

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

func (s *PercentageTestSuite) TestPercentageTwentyOfFiftyIsTen() {
	s.Equal(percentageNum10, s.percentage(percentageNum20, percentageNum50))
}

func (s *PercentageTestSuite) TestPercentageIsPercentFirstAndAmountSecond() {
	s.Equal(percentageNum10, s.percentage(percentageNum50, percentageNum20))
	s.Equal(percentageNum0Pt2, s.percentage(percentageNum20, percentageNum1))
}

func (s *PercentageTestSuite) TestPercentageOfAHundredReturnsTheAmount() {
	s.Equal(percentageNum7Pt5, s.percentage(percentageNum100, percentageNum7Pt5))
}

func (s *PercentageTestSuite) TestPercentageOverAHundredReturnsMoreThanTheAmount() {
	s.Equal(percentageNum15, s.percentage(percentageNum150, percentageNum10))
}

func (s *PercentageTestSuite) TestPercentageOfZeroPercentIsZero() {
	s.Equal(percentageNum0, s.percentage(percentageNum0, percentageNum100))
}

func (s *PercentageTestSuite) TestPercentageOfAZeroAmountIsZero() {
	s.Equal(percentageNum0, s.percentage(percentageNum12Pt5, percentageNum0))
}

func (s *PercentageTestSuite) TestPercentageWhenPercentIsFractionalIsExact() {
	s.Equal(percentageNum10, s.percentage(percentageNum12Pt5, percentageNum80))
}

func (s *PercentageTestSuite) TestPercentageWhenPercentIsNegativeReturnsNegative() {
	s.Equal(percentageNumNeg5, s.percentage(percentageNumNeg10, percentageNum50))
}

func (s *PercentageTestSuite) TestPercentageWhenAmountIsNegativeReturnsNegative() {
	s.Equal(percentageNumNeg5, s.percentage(percentageNum10, percentageNumNeg50))
}

func (s *PercentageTestSuite) TestPercentageWhenBothAreNegativeReturnsPositive() {
	s.Equal(percentageNum5, s.percentage(percentageNumNeg10, percentageNumNeg50))
}

func (s *PercentageTestSuite) TestPercentageOfOnePercentOfThreeIsExactlyPointZeroThree() {
	s.Equal(percentageNum0Pt03, s.percentage(percentageNum1, percentageNum3))
}

func (s *PercentageTestSuite) TestPercentageDoesNotRoundManyDecimalPlaces() {
	s.Equal(percentageNum0Pt99999, s.percentage(percentageNum33Pt333, percentageNum3))
}

func (s *PercentageTestSuite) TestPercentageWhenResultIsLargeKeepsEveryDigit() {
	s.Equal(testPercentSquareWant, s.percentage(testPercentSquare, testPercentSquare))
}

func (s *PercentageTestSuite) TestPercentageWhenResultIsTinyKeepsEveryDigit() {
	s.Equal(testTinyUnitSquareWnt, s.percentage(testTinyUnit, testTinyUnit))
}
