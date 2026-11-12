package operations

import (
	"context"

	"github.com/shopspring/decimal"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

// percentShift moves the decimal point two places left: dividing by 100.
const percentShift = -2

// Percentage is "x percent of y": percentage(x, y) = x / 100 * y.
//
// So percentage(20, 50) is 10, the same as "20% of 50". "Percentage" is
// ambiguous in everyday use (it can also mean "what percent is x of y", or "x
// increased by y percent"); this operation is only the first meaning, and the
// operand order is the percent first, the base second.
//
// The result is exact and needs no rounding: dividing by 100 only moves the
// decimal point, so it always terminates, and multiplying decimals is exact.
// No precision setting is involved. The API docs (docs/API.md) carry this
// definition for clients.
type Percentage struct{ descriptor }

var _ Operation = (*Percentage)(nil)

// NewPercentage returns the percentage operation.
func NewPercentage() *Percentage {
	return &Percentage{descriptor: descriptor{
		name:        entities.OpPercentage,
		arity:       entities.ArityBinary,
		symbol:      entities.SymbolPercentage,
		description: entities.DescriptionPercentage,
	}}
}

// Apply returns operands[0] / 100 * operands[1].
func (*Percentage) Apply(_ context.Context, operands []decimal.Decimal) (decimal.Decimal, apperror.IError) {
	return operands[0].Shift(percentShift).Mul(operands[1]), nil
}
