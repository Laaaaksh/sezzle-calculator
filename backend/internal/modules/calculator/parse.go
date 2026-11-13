package calculator

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/shopspring/decimal"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

const (
	plusSign        = "+"
	minusSign       = "-"
	decimalPoint    = "."
	hexPrefix       = "0x"
	nanText         = "nan"
	infText         = "inf"
	infinityText    = "infinity"
	exponentMarkers = "eE"
	separatorMarks  = ",_"
)

// operandPattern is the whole operand grammar: an optional minus, then digits with at most one decimal point. "1.", ".5",
// "-0" and "007" are accepted. \d matches ASCII digits only in Go, so other scripts' digits fail, and $ matches only at the
// very end, so a trailing newline does not slip through.
var operandPattern = regexp.MustCompile(`^-?(\d+(\.\d*)?|\.\d+)$`)

type rejection struct {
	reason  string
	matches func(text string) bool
}

// rejections are tried in order against text that failed operandPattern; the
// first match names the reason. Order matters: "0x1e5" is hex, not an
// exponent, and "1 000" is whitespace, not a separator. Anything no rule
// recognises is invalid_format.
var rejections = []rejection{
	{entities.ReasonEmpty, func(t string) bool { return t == "" }},
	{entities.ReasonWhitespace, func(t string) bool { return strings.ContainsFunc(t, unicode.IsSpace) }},
	{entities.ReasonPlusSign, func(t string) bool { return strings.HasPrefix(t, plusSign) }},
	{entities.ReasonHex, func(t string) bool { return hasPrefixFold(unsigned(t), hexPrefix) }},
	{entities.ReasonNaN, func(t string) bool { return strings.EqualFold(unsigned(t), nanText) }},
	{entities.ReasonInfinity, isInfinity},
	{entities.ReasonExponent, func(t string) bool { return strings.ContainsAny(t, exponentMarkers) }},
	{entities.ReasonSeparator, func(t string) bool { return strings.ContainsAny(t, separatorMarks) }},
	{entities.ReasonMultiplePoints, func(t string) bool { return strings.Count(t, decimalPoint) > 1 }},
}

// ParseOperands turns operand strings into decimals under strict rules; nothing is guessed or repaired. Anything
// outside the grammar fails with INVALID_OPERAND carrying operand_index and a reason, never the offending text. The
// length check comes first and counts bytes, so a huge string costs one comparison, not a scan or an allocation.
// Operands are checked in order and the first failure is returned. See CalculateRequest for why operands are strings.
func ParseOperands(raw []string, maxLen int) ([]decimal.Decimal, apperror.IError) {
	operands := make([]decimal.Decimal, len(raw))
	for i, text := range raw {
		value, aerr := parseOperand(text, i, maxLen)
		if aerr != nil {
			return nil, aerr
		}
		operands[i] = value
	}
	return operands, nil
}

func parseOperand(text string, index, maxLen int) (decimal.Decimal, apperror.IError) {
	if len(text) > maxLen {
		return decimal.Zero, apperror.New(entities.CodeOperandTooLarge, ErrOperandTooLong).
			WithField(entities.FieldOperandIndex, index).
			WithField(entities.FieldLimit, maxLen)
	}
	if !operandPattern.MatchString(text) {
		return decimal.Zero, invalidOperand(index, rejectionReason(text))
	}
	value, err := decimal.NewFromString(text)
	if err != nil {
		// Unreachable while operandPattern is a subset of what decimal reads;
		// kept so a future library change fails as a clean 422, not a panic.
		return decimal.Zero, invalidOperand(index, entities.ReasonInvalidFormat)
	}
	return value, nil
}

func invalidOperand(index int, reason string) apperror.IError {
	return apperror.New(entities.CodeInvalidOperand, ErrOperandSyntax).
		WithField(entities.FieldOperandIndex, index).
		WithField(entities.FieldReason, reason)
}

func rejectionReason(text string) string {
	for _, rule := range rejections {
		if rule.matches(text) {
			return rule.reason
		}
	}
	return entities.ReasonInvalidFormat
}

// unsigned drops one leading minus so "-NaN" and "-0x1" are named like their
// unsigned forms.
func unsigned(text string) string {
	return strings.TrimPrefix(text, minusSign)
}

func hasPrefixFold(text, prefix string) bool {
	return len(text) >= len(prefix) && strings.EqualFold(text[:len(prefix)], prefix)
}

func isInfinity(text string) bool {
	body := unsigned(text)
	return strings.EqualFold(body, infText) || strings.EqualFold(body, infinityText)
}
