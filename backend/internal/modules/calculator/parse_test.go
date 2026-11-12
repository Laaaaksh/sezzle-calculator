package calculator_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

type ParseOperandsTestSuite struct {
	suite.Suite
}

func TestParseOperandsTestSuite(t *testing.T) {
	suite.Run(t, new(ParseOperandsTestSuite))
}

// parsed parses one operand and returns its canonical text.
func (s *ParseOperandsTestSuite) parsed(text string) string {
	got, err := calculator.ParseOperands([]string{text}, testDefaultMaxLength)
	s.Require().Nil(err)
	s.Require().Len(got, 1)
	return got[0].String()
}

// rejected parses one operand that must be refused as INVALID_OPERAND and
// returns the reason.
func (s *ParseOperandsTestSuite) rejected(text string) any {
	_, err := calculator.ParseOperands([]string{text}, testDefaultMaxLength)
	s.Require().NotNil(err)
	s.Require().Equal(apperror.CodeInvalidOperand, err.Code())
	s.Equal(testIndexFirst, err.Fields()[entities.FieldOperandIndex])
	return err.Fields()[entities.FieldReason]
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandIsAnIntegerReturnsIt() {
	s.Equal("42", s.parsed("42"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandIsZeroReturnsZero() {
	s.Equal("0", s.parsed("0"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandIsNegativeKeepsTheSign() {
	s.Equal("-42", s.parsed("-42"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandHasADecimalPointKeepsTheFraction() {
	s.Equal("1.5", s.parsed("1.5"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandIsANegativeDecimalKeepsBoth() {
	s.Equal("-0.001", s.parsed("-0.001"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandIsNegativeZeroReadsAsZero() {
	s.Equal("0", s.parsed("-0"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandHasLeadingZerosDropsThem() {
	s.Equal("7", s.parsed("007"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandEndsWithAPointAcceptsIt() {
	s.Equal("1", s.parsed("1."))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandStartsWithAPointAcceptsIt() {
	s.Equal("0.5", s.parsed(".5"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandIsANegativeLeadingPointAcceptsIt() {
	s.Equal("-0.5", s.parsed("-.5"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandHasManyDecimalPlacesKeepsEveryDigit() {
	s.Equal("0.1234567890123456789", s.parsed("0.1234567890123456789"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandIsPointOneKeepsItExact() {
	got, err := calculator.ParseOperands([]string{"0.1", "0.2"}, testDefaultMaxLength)

	s.Require().Nil(err)
	s.Equal("0.3", got[0].Add(got[1]).String())
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandHasTwentyDigitsKeepsEveryDigit() {
	s.Equal("12345678901234567890", s.parsed("12345678901234567890"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandIsExactlyTheMaxLengthAcceptsIt() {
	got, err := calculator.ParseOperands([]string{testOperandOk}, testParseMaxLen)

	s.Nil(err)
	s.Equal(testOperandOk, got[0].String())
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenMinusSignCountsTowardTheMaxLengthRejectsIt() {
	_, err := calculator.ParseOperands([]string{"-" + testOperandOk}, testParseMaxLen)

	s.Require().NotNil(err)
	s.Equal(apperror.CodeOperandTooLarge, err.Code())
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenThereAreNoOperandsReturnsAnEmptySlice() {
	got, err := calculator.ParseOperands(nil, testParseMaxLen)

	s.Nil(err)
	s.Empty(got)
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenAllOperandsAreValidKeepsTheirOrder() {
	got, err := calculator.ParseOperands([]string{"3", "-1.5", "0"}, testDefaultMaxLength)

	s.Require().Nil(err)
	s.Equal("3", got[0].String())
	s.Equal("-1.5", got[1].String())
	s.Equal("0", got[2].String())
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandIsEmptyReturnsEmptyReason() {
	s.Equal(entities.ReasonEmpty, s.rejected(""))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandIsOnlySpacesReturnsWhitespaceReason() {
	s.Equal(entities.ReasonWhitespace, s.rejected("   "))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandHasALeadingSpaceReturnsWhitespaceReason() {
	s.Equal(entities.ReasonWhitespace, s.rejected(" 1"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandHasATrailingSpaceReturnsWhitespaceReason() {
	s.Equal(entities.ReasonWhitespace, s.rejected("1 "))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandHasATrailingNewlineReturnsWhitespaceReason() {
	s.Equal(entities.ReasonWhitespace, s.rejected("1\n"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandHasATabReturnsWhitespaceReason() {
	s.Equal(entities.ReasonWhitespace, s.rejected("\t1"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandHasASpaceInsideReturnsWhitespaceReason() {
	s.Equal(entities.ReasonWhitespace, s.rejected("1 000"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandHasASpaceAfterTheMinusReturnsWhitespaceReason() {
	s.Equal(entities.ReasonWhitespace, s.rejected("- 1"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandHasALeadingPlusReturnsPlusSignReason() {
	s.Equal(entities.ReasonPlusSign, s.rejected("+1"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandIsOnlyAPlusReturnsPlusSignReason() {
	s.Equal(entities.ReasonPlusSign, s.rejected("+"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandUsesLowercaseExponentReturnsExponentReason() {
	s.Equal(entities.ReasonExponent, s.rejected("1e5"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandUsesUppercaseExponentReturnsExponentReason() {
	s.Equal(entities.ReasonExponent, s.rejected("1E5"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandUsesNegativeExponentReturnsExponentReason() {
	s.Equal(entities.ReasonExponent, s.rejected("1.5e-3"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandUsesSignedExponentReturnsExponentReason() {
	s.Equal(entities.ReasonExponent, s.rejected("2e+3"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandIsHexReturnsHexReason() {
	s.Equal(entities.ReasonHex, s.rejected("0x1F"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandIsUppercaseHexReturnsHexReason() {
	s.Equal(entities.ReasonHex, s.rejected("0XFF"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandIsNegativeHexReturnsHexReason() {
	s.Equal(entities.ReasonHex, s.rejected("-0x10"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenHexDigitsLookLikeAnExponentReturnsHexReason() {
	s.Equal(entities.ReasonHex, s.rejected("0x1e5"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandIsNaNReturnsNaNReason() {
	s.Equal(entities.ReasonNaN, s.rejected("NaN"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandIsLowercaseNaNReturnsNaNReason() {
	s.Equal(entities.ReasonNaN, s.rejected("nan"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandIsNegativeNaNReturnsNaNReason() {
	s.Equal(entities.ReasonNaN, s.rejected("-NaN"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandIsInfinityReturnsInfinityReason() {
	s.Equal(entities.ReasonInfinity, s.rejected("Infinity"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandIsNegativeInfinityReturnsInfinityReason() {
	s.Equal(entities.ReasonInfinity, s.rejected("-Infinity"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandIsInfReturnsInfinityReason() {
	s.Equal(entities.ReasonInfinity, s.rejected("inf"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandHasACommaSeparatorReturnsSeparatorReason() {
	s.Equal(entities.ReasonSeparator, s.rejected("1,000"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandHasACommaAndDecimalsReturnsSeparatorReason() {
	s.Equal(entities.ReasonSeparator, s.rejected("1,000.50"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandHasAnUnderscoreSeparatorReturnsSeparatorReason() {
	s.Equal(entities.ReasonSeparator, s.rejected("1_000"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandHasTwoDecimalPointsReturnsMultiplePointsReason() {
	s.Equal(entities.ReasonMultiplePoints, s.rejected("1.2.3"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandHasAdjacentDecimalPointsReturnsMultiplePointsReason() {
	s.Equal(entities.ReasonMultiplePoints, s.rejected("1..2"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandIsOnlyTwoPointsReturnsMultiplePointsReason() {
	s.Equal(entities.ReasonMultiplePoints, s.rejected(".."))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandIsOnlyAPointReturnsInvalidFormatReason() {
	s.Equal(entities.ReasonInvalidFormat, s.rejected("."))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandIsOnlyAMinusReturnsInvalidFormatReason() {
	s.Equal(entities.ReasonInvalidFormat, s.rejected("-"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandIsMinusAndPointReturnsInvalidFormatReason() {
	s.Equal(entities.ReasonInvalidFormat, s.rejected("-."))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandHasTwoMinusSignsReturnsInvalidFormatReason() {
	s.Equal(entities.ReasonInvalidFormat, s.rejected("--1"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandHasATrailingMinusReturnsInvalidFormatReason() {
	s.Equal(entities.ReasonInvalidFormat, s.rejected("1-"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandIsWordsReturnsInvalidFormatReason() {
	s.Equal(entities.ReasonInvalidFormat, s.rejected(testWordsText))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandIsBinaryReturnsInvalidFormatReason() {
	s.Equal(entities.ReasonInvalidFormat, s.rejected("0b101"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandIsAFractionReturnsInvalidFormatReason() {
	s.Equal(entities.ReasonInvalidFormat, s.rejected("1/2"))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandUsesFullwidthDigitsReturnsInvalidFormatReason() {
	s.Equal(entities.ReasonInvalidFormat, s.rejected(testFullwidth))
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandIsLongerThanTheLimitReturnsOperandTooLarge() {
	_, err := calculator.ParseOperands([]string{testOperandTooLong}, testParseMaxLen)

	s.Require().NotNil(err)
	s.Equal(apperror.CodeOperandTooLarge, err.Code())
	s.Equal(testIndexFirst, err.Fields()[entities.FieldOperandIndex])
	s.Equal(testParseMaxLen, err.Fields()[entities.FieldLimit])
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenALongOperandIsNotANumberChecksLengthBeforeSyntax() {
	_, err := calculator.ParseOperands([]string{testParseLongText}, testParseMaxLen)

	s.Require().NotNil(err)
	s.Equal(apperror.CodeOperandTooLarge, err.Code())
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenOperandIsHugeReturnsOperandTooLargeWithoutParsingIt() {
	_, err := calculator.ParseOperands([]string{strings.Repeat("9", 10_000_000)}, testDefaultMaxLength)

	s.Require().NotNil(err)
	s.Equal(apperror.CodeOperandTooLarge, err.Code())
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenMultibyteCharactersExceedTheLimitCountsBytes() {
	_, err := calculator.ParseOperands([]string{testFullwidth}, testParseMaxLen-1)

	s.Require().NotNil(err)
	s.Equal(apperror.CodeOperandTooLarge, err.Code())
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenTheSecondOperandIsInvalidReportsIndexOne() {
	_, err := calculator.ParseOperands([]string{"1", testWordsText}, testDefaultMaxLength)

	s.Require().NotNil(err)
	s.Equal(apperror.CodeInvalidOperand, err.Code())
	s.Equal(testIndexSecond, err.Fields()[entities.FieldOperandIndex])
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenTheSecondOperandIsTooLongReportsIndexOne() {
	_, err := calculator.ParseOperands([]string{"1", testOperandTooLong}, testParseMaxLen)

	s.Require().NotNil(err)
	s.Equal(apperror.CodeOperandTooLarge, err.Code())
	s.Equal(testIndexSecond, err.Fields()[entities.FieldOperandIndex])
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenSeveralOperandsAreInvalidReportsTheFirst() {
	_, err := calculator.ParseOperands([]string{"1", "+2", testWordsText}, testDefaultMaxLength)

	s.Require().NotNil(err)
	s.Equal(testIndexSecond, err.Fields()[entities.FieldOperandIndex])
	s.Equal(entities.ReasonPlusSign, err.Fields()[entities.FieldReason])
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenAnOperandIsInvalidReturnsNoOperands() {
	got, err := calculator.ParseOperands([]string{"1", testWordsText}, testDefaultMaxLength)

	s.NotNil(err)
	s.Nil(got)
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenAnOperandIsInvalidUsesTheDocumentedMessage() {
	_, err := calculator.ParseOperands([]string{testWordsText}, testDefaultMaxLength)

	s.Require().NotNil(err)
	s.Equal(apperror.MsgInvalidOperand, err.PublicMessage())
	s.ErrorIs(err, calculator.ErrOperandSyntax)
}

func (s *ParseOperandsTestSuite) TestParseOperandsWhenAnOperandIsInvalidDoesNotEchoItInTheErrorFields() {
	_, err := calculator.ParseOperands([]string{testNoLeakText}, testDefaultMaxLength)

	s.Require().NotNil(err)
	s.NotContains(fmt.Sprint(err.Fields()), testNoLeakText)
	s.NotContains(err.PublicMessage(), testNoLeakText)
}
