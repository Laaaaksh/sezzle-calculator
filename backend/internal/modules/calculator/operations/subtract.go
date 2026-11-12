package operations

import (
	"context"

	"github.com/shopspring/decimal"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

// Subtract subtracts the second operand from the first. The result is exact:
// decimal keeps every digit, so 0.3 - 0.1 is 0.2 and not the 0.19999999999999998
// a float64 gives.
type Subtract struct{ descriptor }

var _ Operation = (*Subtract)(nil)

// NewSubtract returns the subtract operation.
func NewSubtract() *Subtract {
	return &Subtract{descriptor: descriptor{
		name:        entities.OpSubtract,
		arity:       entities.ArityBinary,
		symbol:      entities.SymbolSubtract,
		description: entities.DescriptionSubtract,
	}}
}

// Apply returns operands[0] - operands[1].
func (*Subtract) Apply(_ context.Context, operands []decimal.Decimal) (decimal.Decimal, apperror.IError) {
	return operands[0].Sub(operands[1]), nil
}
