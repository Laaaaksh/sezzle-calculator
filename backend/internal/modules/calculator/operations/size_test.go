package operations_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/operations"
)

type SizeTestSuite struct {
	suite.Suite
}

func TestSizeTestSuite(t *testing.T) {
	suite.Run(t, new(SizeTestSuite))
}

func (s *SizeTestSuite) digits(text string) int64 {
	return operations.DigitsOf(mustDecimal(text))
}

func (s *SizeTestSuite) TestDigitsOfZeroIsOne() {
	s.Equal(int64(1), s.digits(sizeZero))
}

func (s *SizeTestSuite) TestDigitsOfAnIntegerCountsEveryDigit() {
	s.Equal(int64(4), s.digits(sizeThousand))
}

func (s *SizeTestSuite) TestDigitsOfAFractionCountsIntegerAndFractionDigits() {
	s.Equal(int64(3), s.digits(sizeTwelvePointFive))
}

func (s *SizeTestSuite) TestDigitsOfAFractionBelowOneIgnoresTheLeadingZero() {
	s.Equal(int64(3), s.digits(sizePointZeroZeroOne))
}

func (s *SizeTestSuite) TestDigitsOfIgnoresTrailingFractionZeros() {
	s.Equal(int64(2), s.digits(sizeOnePointFiveZeros))
}

func (s *SizeTestSuite) TestDigitsOfIgnoresTheSign() {
	s.Equal(int64(4), s.digits(sizeNegThousand))
}

func (s *SizeTestSuite) TestDigitsOfAValueWithManyDigitsCountsThemExactly() {
	s.Equal(int64(sizeLongDigits), s.digits(strings.Repeat(digitNine, sizeLongDigits)))
}

func (s *SizeTestSuite) TestCheckResultSizeWhenDigitsEqualTheLimitReturnsNil() {
	s.Nil(operations.CheckResultSize(mustDecimal(sizeThousand), sizeFourDigits))
}

func (s *SizeTestSuite) TestCheckResultSizeWhenOneDigitOverTheLimitReturnsResultTooLarge() {
	err := operations.CheckResultSize(mustDecimal(sizeTenThousand), sizeFourDigits)

	s.Require().NotNil(err)
	s.Equal(entities.CodeResultTooLarge, err.Code())
	s.Equal(int64(sizeFourDigits), err.Fields()[entities.FieldLimit])
	s.ErrorIs(err, operations.ErrResultBeyondBudget)
}

func (s *SizeTestSuite) TestCheckResultSizeCountsFractionDigits() {
	err := operations.CheckResultSize(mustDecimal(sizeFiveFractionDigits), sizeFourDigits)

	s.Require().NotNil(err)
	s.Equal(entities.CodeResultTooLarge, err.Code())
}

func (s *SizeTestSuite) TestResultTooLargeCarriesTheLimitAsAField() {
	err := operations.ResultTooLarge(sizeFourDigits)

	s.Equal(entities.CodeResultTooLarge, err.Code())
	s.Equal(int64(sizeFourDigits), err.Fields()[entities.FieldLimit])
}
