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
	s.Equal(sqrtNum0, s.sqrt(sqrtNum0))
}

func (s *SqrtTestSuite) TestSqrtOfNegativeZeroIsZero() {
	s.Equal(sqrtNum0, s.sqrt(sqrtNumNeg0Pt00))
}

func (s *SqrtTestSuite) TestSqrtOfOneIsOne() {
	s.Equal(sqrtNum1, s.sqrt(sqrtNum1))
}

func (s *SqrtTestSuite) TestSqrtOfAPerfectSquareIsExact() {
	s.Equal(sqrtNum12, s.sqrt(sqrtNum144))
}

func (s *SqrtTestSuite) TestSqrtOfAFractionalPerfectSquareIsExact() {
	s.Equal(sqrtNum0Pt5, s.sqrt(sqrtNum0Pt25))
}

func (s *SqrtTestSuite) TestSqrtOfAPerfectSquareWithTrailingZerosIsExact() {
	s.Equal(sqrtNum1Pt5, s.sqrt(sqrtNum2Pt2500))
}

func (s *SqrtTestSuite) TestSqrtOfATinyPerfectSquareIsExact() {
	s.Equal(sqrtNum0Pt00001, s.sqrt(sqrtNum0Pt0000000001))
}

func (s *SqrtTestSuite) TestSqrtOfALargePerfectSquareIsExact() {
	s.Equal(testBigSquareRoot, s.sqrt(testBigSquare))
}

func (s *SqrtTestSuite) TestSqrtOfAPowerOfTenIsExact() {
	s.Equal(sqrtLong1, s.sqrt(sqrtLong2))
}

func (s *SqrtTestSuite) TestSqrtOfTwoMatchesTheKnownValueAtTheConfiguredPrecision() {
	s.Equal(testSqrtTwoAt20, s.sqrt(sqrtNum2))
}

func (s *SqrtTestSuite) TestSqrtOfTwoMatchesTheKnownValueAtThirtyPlaces() {
	s.Equal(testSqrtTwoAt30, s.sqrtAt(30, sqrtNum2))
}

func (s *SqrtTestSuite) TestSqrtOfThreeRoundsTheLastDigitUp() {
	s.Equal(testSqrtThreeAt20, s.sqrt(sqrtNum3))
}

func (s *SqrtTestSuite) TestSqrtOfANumberBelowOneKeepsTheLeadingDigitsOfTheRoot() {
	s.Equal(sqrtLong3, s.sqrt(sqrtNum0Pt5))
}

func (s *SqrtTestSuite) TestSqrtOfATinyNonSquareRoundsAtThePrecision() {
	s.Equal(testSqrtTinyNonSquare, s.sqrt(sqrtNum0Pt+strings.Repeat(sqrtNum0, 29)+sqrtNum2))
}

func (s *SqrtTestSuite) TestSqrtOfAValueWhoseRootIsBelowHalfTheLastPlaceIsZero() {
	s.Equal(sqrtNum0, s.sqrt(sqrtNum0Pt+strings.Repeat(sqrtNum0, 49)+sqrtNum1))
}

func (s *SqrtTestSuite) TestSqrtOfAnExponentNotationDecimalFarBelowThePrecisionIsZero() {
	s.Equal(sqrtNum0, s.sqrt("1e-1000"))
}

func (s *SqrtTestSuite) TestSqrtOfAVeryLargePowerOfTenConvergesWithinTheBound() {
	s.Equal(sqrtNum1+strings.Repeat(sqrtNum0, 150), s.sqrt(sqrtNum1+strings.Repeat(sqrtNum0, 300)))
}

func (s *SqrtTestSuite) TestSqrtOfALargeNonSquareJustAboveAPowerOfTenRoundsToThePower() {
	got := s.sqrt(sqrtNum1 + strings.Repeat(sqrtNum0, 299) + sqrtNum1)

	s.Equal(sqrtNum1+strings.Repeat(sqrtNum0, 149), got[:150])
}

func (s *SqrtTestSuite) TestSqrtWhenOperandIsNegativeReturnsNegativeSquareRoot() {
	_, err := applyUnary(s.op, sqrtNumNeg4)

	s.Require().NotNil(err)
	s.Equal(entities.CodeNegativeSquareRoot, err.Code())
	s.ErrorIs(err, operations.ErrNegativeSquareRoot)
}

func (s *SqrtTestSuite) TestSqrtWhenOperandIsSlightlyNegativeReturnsNegativeSquareRoot() {
	_, err := applyUnary(s.op, sqrtNumNeg0Pt0000000001)

	s.Require().NotNil(err)
	s.Equal(entities.CodeNegativeSquareRoot, err.Code())
}

func (s *SqrtTestSuite) TestSqrtNegativeSquareRootUsesTheDocumentedMessage() {
	_, err := applyUnary(s.op, sqrtNumNeg1)

	s.Require().NotNil(err)
	s.Equal(entities.MsgNegativeSquareRoot, err.PublicMessage())
}

func (s *SqrtTestSuite) TestSqrtWhenExactTieFollowsAnEvenDigitRoundsDown() {
	s.Equal(sqrtNum2, s.sqrtAt(0, sqrtNum6Pt25))
}

func (s *SqrtTestSuite) TestSqrtWhenExactTieFollowsAnOddDigitRoundsUp() {
	s.Equal(sqrtNum2, s.sqrtAt(0, sqrtNum2Pt25))
}

func (s *SqrtTestSuite) TestSqrtWhenExactTieFollowsZeroRoundsToZero() {
	s.Equal(sqrtNum0, s.sqrtAt(0, sqrtNum0Pt25))
}

func (s *SqrtTestSuite) TestSqrtWhenRootIsJustAboveATieRoundsUpEvenWithGuardDigitsShowingATie() {
	// The root is 2.500000000001: past the ten guard digits it looks like an
	// exact 2.5 tie (which would go down to the even 2), but it is above it.
	s.Equal(sqrtNum3, s.sqrtAt(0, testNearTieSquare))
}

func (s *SqrtTestSuite) TestSqrtWhenIterationCapIsReachedReturnsInternalInsteadOfHanging() {
	capped := operations.NewBoundedSqrt(testConfiguredPrecision, testSqrtIterations)

	_, err := applyUnary(capped, sqrtNum2)

	s.Require().NotNil(err)
	s.Equal(apperror.CodeInternal, err.Code())
	s.ErrorIs(err, operations.ErrSqrtDidNotConverge)
}

func (s *SqrtTestSuite) TestSqrtWhenContextEndedReturnsRequestTimeout() {
	_, err := s.op.Apply(cancelled(), []decimal.Decimal{mustDecimal(sqrtNum2)})

	s.Require().NotNil(err)
	s.Equal(apperror.CodeRequestTimeout, err.Code())
	s.ErrorIs(err, context.Canceled)
}

func (s *SqrtTestSuite) TestSqrtOfTheLargestOperandPrintsFarBelowThePowerBudget() {
	largest := strings.Repeat(digitNine, testFullOperandDigits)

	got := s.sqrt(largest)

	s.Less(len(got), testResultBudget)
}
