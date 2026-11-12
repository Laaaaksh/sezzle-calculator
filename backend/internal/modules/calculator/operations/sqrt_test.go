package operations_test

import (
	"context"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/operations"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

type SqrtTestSuite struct {
	suite.Suite
	op *operations.Sqrt
}

func TestSqrtTestSuite(t *testing.T) {
	suite.Run(t, new(SqrtTestSuite))
}

func (s *SqrtTestSuite) SetupTest() {
	s.op = operations.NewSqrt(testConfiguredPrecision)
}

func (s *SqrtTestSuite) sqrt(operand string) string {
	got, err := applyUnary(s.op, operand)
	s.Require().Nil(err)
	return got
}

func (s *SqrtTestSuite) sqrtAt(precision int, operand string) string {
	got, err := applyUnary(operations.NewSqrt(precision), operand)
	s.Require().Nil(err)
	return got
}

func (s *SqrtTestSuite) TestSqrtNameIsSqrt() {
	s.Equal(entities.OpSqrt, s.op.Name())
}

func (s *SqrtTestSuite) TestSqrtArityIsOne() {
	s.Equal(entities.ArityUnary, s.op.Arity())
}

func (s *SqrtTestSuite) TestSqrtOfZeroIsZero() {
	s.Equal("0", s.sqrt("0"))
}

func (s *SqrtTestSuite) TestSqrtOfNegativeZeroIsZero() {
	s.Equal("0", s.sqrt("-0.00"))
}

func (s *SqrtTestSuite) TestSqrtOfOneIsOne() {
	s.Equal("1", s.sqrt("1"))
}

func (s *SqrtTestSuite) TestSqrtOfAPerfectSquareIsExact() {
	s.Equal("12", s.sqrt("144"))
}

func (s *SqrtTestSuite) TestSqrtOfAFractionalPerfectSquareIsExact() {
	s.Equal("0.5", s.sqrt("0.25"))
}

func (s *SqrtTestSuite) TestSqrtOfAPerfectSquareWithTrailingZerosIsExact() {
	s.Equal("1.5", s.sqrt("2.2500"))
}

func (s *SqrtTestSuite) TestSqrtOfATinyPerfectSquareIsExact() {
	s.Equal("0.00001", s.sqrt("0.0000000001"))
}

func (s *SqrtTestSuite) TestSqrtOfALargePerfectSquareIsExact() {
	s.Equal(testBigSquareRoot, s.sqrt(testBigSquare))
}

func (s *SqrtTestSuite) TestSqrtOfAPowerOfTenIsExact() {
	s.Equal("100000000000000000000", s.sqrt("10000000000000000000000000000000000000000"))
}

func (s *SqrtTestSuite) TestSqrtOfTwoMatchesTheKnownValueAtTheConfiguredPrecision() {
	s.Equal(testSqrtTwoAt20, s.sqrt("2"))
}

func (s *SqrtTestSuite) TestSqrtOfTwoMatchesTheKnownValueAtThirtyPlaces() {
	s.Equal(testSqrtTwoAt30, s.sqrtAt(30, "2"))
}

func (s *SqrtTestSuite) TestSqrtOfThreeRoundsTheLastDigitUp() {
	s.Equal(testSqrtThreeAt20, s.sqrt("3"))
}

func (s *SqrtTestSuite) TestSqrtOfANumberBelowOneKeepsTheLeadingDigitsOfTheRoot() {
	s.Equal("0.7071067811865475244", s.sqrt("0.5"))
}

func (s *SqrtTestSuite) TestSqrtOfATinyNonSquareRoundsAtThePrecision() {
	s.Equal(testSqrtTinyNonSquare, s.sqrt("0."+strings.Repeat("0", 29)+"2"))
}

func (s *SqrtTestSuite) TestSqrtOfAValueWhoseRootIsBelowHalfTheLastPlaceIsZero() {
	s.Equal("0", s.sqrt("0."+strings.Repeat("0", 49)+"1"))
}

func (s *SqrtTestSuite) TestSqrtOfAnExponentNotationDecimalFarBelowThePrecisionIsZero() {
	s.Equal("0", s.sqrt("1e-1000"))
}

func (s *SqrtTestSuite) TestSqrtOfAVeryLargePowerOfTenConvergesWithinTheBound() {
	s.Equal("1"+strings.Repeat("0", 150), s.sqrt("1"+strings.Repeat("0", 300)))
}

func (s *SqrtTestSuite) TestSqrtOfALargeNonSquareJustAboveAPowerOfTenRoundsToThePower() {
	got := s.sqrt("1" + strings.Repeat("0", 299) + "1")

	s.Equal("1"+strings.Repeat("0", 149), got[:150])
}

func (s *SqrtTestSuite) TestSqrtWhenOperandIsNegativeReturnsNegativeSquareRoot() {
	_, err := applyUnary(s.op, "-4")

	s.Require().NotNil(err)
	s.Equal(apperror.CodeNegativeSquareRoot, err.Code())
	s.ErrorIs(err, operations.ErrNegativeSquareRoot)
}

func (s *SqrtTestSuite) TestSqrtWhenOperandIsSlightlyNegativeReturnsNegativeSquareRoot() {
	_, err := applyUnary(s.op, "-0.0000000001")

	s.Require().NotNil(err)
	s.Equal(apperror.CodeNegativeSquareRoot, err.Code())
}

func (s *SqrtTestSuite) TestSqrtNegativeSquareRootUsesTheDocumentedMessage() {
	_, err := applyUnary(s.op, "-1")

	s.Require().NotNil(err)
	s.Equal(apperror.MsgNegativeSquareRoot, err.PublicMessage())
}

func (s *SqrtTestSuite) TestSqrtWhenExactTieFollowsAnEvenDigitRoundsDown() {
	s.Equal("2", s.sqrtAt(0, "6.25"))
}

func (s *SqrtTestSuite) TestSqrtWhenExactTieFollowsAnOddDigitRoundsUp() {
	s.Equal("2", s.sqrtAt(0, "2.25"))
}

func (s *SqrtTestSuite) TestSqrtWhenExactTieFollowsZeroRoundsToZero() {
	s.Equal("0", s.sqrtAt(0, "0.25"))
}

func (s *SqrtTestSuite) TestSqrtWhenRootIsJustAboveATieRoundsUpEvenWithGuardDigitsShowingATie() {
	// The root is 2.500000000001: past the ten guard digits it looks like an
	// exact 2.5 tie (which would go down to the even 2), but it is above it.
	s.Equal("3", s.sqrtAt(0, testNearTieSquare))
}

func (s *SqrtTestSuite) TestSqrtWhenIterationCapIsReachedReturnsInternalInsteadOfHanging() {
	capped := operations.NewBoundedSqrt(testConfiguredPrecision, testSqrtIterations)

	_, err := applyUnary(capped, "2")

	s.Require().NotNil(err)
	s.Equal(apperror.CodeInternal, err.Code())
	s.ErrorIs(err, operations.ErrSqrtDidNotConverge)
}

func (s *SqrtTestSuite) TestSqrtWhenContextEndedReturnsRequestTimeout() {
	_, err := s.op.Apply(cancelled(), []decimal.Decimal{mustDecimal("2")})

	s.Require().NotNil(err)
	s.Equal(apperror.CodeRequestTimeout, err.Code())
	s.ErrorIs(err, context.Canceled)
}

func (s *SqrtTestSuite) TestSqrtOfTheLargestOperandPrintsFarBelowThePowerBudget() {
	largest := strings.Repeat(digitNine, testFullOperandDigits)

	got := s.sqrt(largest)

	s.Less(len(got), testResultBudget)
}
