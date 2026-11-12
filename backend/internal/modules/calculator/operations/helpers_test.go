package operations_test

import (
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
	result, err := op.Apply([]decimal.Decimal{mustDecimal(left), mustDecimal(right)})
	return result.String(), err
}

// applyUnary runs a one-operand op and returns the result text.
func applyUnary(op operations.Operation, operand string) (string, apperror.IError) {
	result, err := op.Apply([]decimal.Decimal{mustDecimal(operand)})
	return result.String(), err
}
