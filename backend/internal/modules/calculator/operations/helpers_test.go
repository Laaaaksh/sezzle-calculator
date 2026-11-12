package operations_test

import (
	"context"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/operations"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

func mustDecimal(text string) decimal.Decimal {
	return decimal.RequireFromString(text)
}

// applyText runs op on two operands given as text and returns the result
// string, so assertions compare exact decimal text, never a float.
func applyText(op operations.Operation, left, right string) (string, apperror.IError) {
	result, err := op.Apply(context.Background(), []decimal.Decimal{mustDecimal(left), mustDecimal(right)})
	return result.String(), err
}

// applyUnary runs a one-operand op and returns the result text.
func applyUnary(op operations.Operation, operand string) (string, apperror.IError) {
	result, err := op.Apply(context.Background(), []decimal.Decimal{mustDecimal(operand)})
	return result.String(), err
}

// roundBase is a 1 followed by zeros: one significant digit, many printed ones.
func roundBase(zeros int) string {
	return digitOne + strings.Repeat(digitZero, zeros)
}

// tinyBase is 0.00...01 with the given number of zeros after the point.
func tinyBase(zeros int) string {
	return "0." + strings.Repeat(digitZero, zeros) + digitOne
}

// cancelled returns a context that has already ended.
func cancelled() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}
