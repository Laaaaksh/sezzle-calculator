package operations

import (
	"context"

	"github.com/shopspring/decimal"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

// percentShift moves the decimal point two places left: dividing by 100.
const percentShift = -2

// Percentage is "x percent of y": x / 100 * y, so percentage(20, 50) is 10. The other readings (what percent x is of y,
// x increased by y percent) are not supported. The percent is the first operand, the base the second.
// The result is exact: dividing by 100 only moves the decimal point, so no rounding or precision applies.
// docs/API.md carries this definition for clients.
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
