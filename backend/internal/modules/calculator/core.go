// Package calculator is the calculator domain: operations, registry, operand parser and the core service.
package calculator

import (
	"context"

	"github.com/shopspring/decimal"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/operations"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

//go:generate go tool mockgen -source=core.go -destination=mock/mock_core.go -package=mock

// ICore is the calculator business logic.
type ICore interface {
	// Calculate applies op to operands and returns the normalised result.
	Calculate(ctx context.Context, op entities.OperationType, operands []decimal.Decimal) (entities.Result, apperror.IError)
	// ListOperations describes every registered operation in a fixed order.
	ListOperations(ctx context.Context) []entities.OperationInfo
}

// Core resolves, validates and runs operations.
type Core struct {
	registry operations.IRegistry
	limits   entities.Limits
}

var _ ICore = (*Core)(nil)

// NewCore returns a Core that resolves operations from registry and enforces
// the operand limits in limits.
func NewCore(registry operations.IRegistry, limits entities.Limits) *Core {
	return &Core{registry: registry, limits: limits}
}

// Calculate checks the operand count and limits, applies the operation, checks the result size, then normalises, so an
// operation never sees input it was not written for. The size check lives here, not in each operation, so all share one
// digit budget. ctx is checked before work, inside Apply and after it: a result finished past the deadline is dropped
// and the client gets REQUEST_TIMEOUT, never a late answer.
func (c *Core) Calculate(ctx context.Context, op entities.OperationType, operands []decimal.Decimal) (entities.Result, apperror.IError) {
	if aerr := operations.CheckContext(ctx); aerr != nil {
		return entities.Result{}, aerr
	}
	operation, aerr := c.registry.Lookup(op)
	if aerr != nil {
		return entities.Result{}, aerr
	}
	if aerr = checkArity(operation, operands); aerr != nil {
		return entities.Result{}, aerr
	}
	if aerr = c.checkLimits(operands); aerr != nil {
		return entities.Result{}, aerr
	}
	value, aerr := operation.Apply(ctx, operands)
	if aerr != nil {
		return entities.Result{}, aerr
	}
	if aerr = operations.CheckResultSize(value, int64(c.limits.MaxResultDigits)); aerr != nil {
		return entities.Result{}, aerr
	}
	if aerr = operations.CheckContext(ctx); aerr != nil {
		return entities.Result{}, aerr
	}
	return normalise(value), nil
}

// ListOperations lists registered operations in registration order. The web keypad is static and uses the list only to
// disable keys the server does not support.
func (c *Core) ListOperations(_ context.Context) []entities.OperationInfo {
	registered := c.registry.List()
	infos := make([]entities.OperationInfo, len(registered))
	for i, op := range registered {
		infos[i] = entities.OperationInfo{
			Name:        op.Name(),
			Arity:       op.Arity(),
			Symbol:      op.Symbol(),
			Description: op.Description(),
		}
	}
	return infos
}
