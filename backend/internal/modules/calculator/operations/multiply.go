package operations

import (
	"github.com/shopspring/decimal"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

// Multiply multiplies the two operands. The result is exact: decimal keeps every
// digit, so 0.1 + 0.2 is 0.3 and not the 0.30000000000000004 a float64 gives.
type Multiply struct{ descriptor }

var _ Operation = (*Multiply)(nil)

// NewMultiply returns the multiply operation.
func NewMultiply() *Multiply {
	return &Multiply{descriptor: descriptor{
		name:        entities.OpMultiply,
		arity:       entities.ArityBinary,
		symbol:      entities.SymbolMultiply,
		description: entities.DescriptionMultiply,
	}}
}

// Apply returns operands[0] * operands[1].
func (*Multiply) Apply(operands []decimal.Decimal) (decimal.Decimal, apperror.IError) {
	return operands[0].Mul(operands[1]), nil
}
