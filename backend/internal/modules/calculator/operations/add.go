package operations

import (
	"github.com/shopspring/decimal"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

// Add adds the two operands. The result is exact: decimal keeps every
// digit, so 0.1 + 0.2 is 0.3 and not the 0.30000000000000004 a float64 gives.
type Add struct{}

var _ Operation = (*Add)(nil)

// NewAdd returns the add operation.
func NewAdd() *Add { return &Add{} }

// Name returns the registry key.
func (*Add) Name() entities.OperationType { return entities.OpAdd }

// Arity is two: left and right.
func (*Add) Arity() int { return entities.ArityBinary }

// Apply returns operands[0] + operands[1].
func (*Add) Apply(operands []decimal.Decimal) (decimal.Decimal, apperror.IError) {
	return operands[0].Add(operands[1]), nil
}
