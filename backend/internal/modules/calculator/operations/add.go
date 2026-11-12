package operations

import (
	"context"

	"github.com/shopspring/decimal"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

// Add adds the two operands. The result is exact: decimal keeps every
// digit, so 0.1 + 0.2 is 0.3 and not the 0.30000000000000004 a float64 gives.
type Add struct{ descriptor }

var _ Operation = (*Add)(nil)

// NewAdd returns the add operation.
func NewAdd() *Add {
	return &Add{descriptor: descriptor{
		name:        entities.OpAdd,
		arity:       entities.ArityBinary,
		symbol:      entities.SymbolAdd,
		description: entities.DescriptionAdd,
	}}
}

// Apply returns operands[0] + operands[1].
func (*Add) Apply(_ context.Context, operands []decimal.Decimal) (decimal.Decimal, apperror.IError) {
	return operands[0].Add(operands[1]), nil
}
