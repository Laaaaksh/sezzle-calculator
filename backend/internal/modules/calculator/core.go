// Package calculator is the calculator domain: operations, the registry that
// resolves them, the parser for operand strings and the core service that
// validates and normalises results.
//
// Operands cross the API as strings ("0.1"), never as JSON numbers (0.1).
// Most JSON clients, JavaScript included, parse a JSON number into a float64
// before the service sees it, and a float64 cannot hold 0.1 or a 20-digit
// integer exactly: the digits are already lost on the client. A string reaches
// the service untouched, so ParseOperands can turn exactly what the user typed
// into a decimal and 0.1 + 0.2 can be exactly 0.3.
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

// Calculate runs the pipeline: look up the operation, check the operand count,
// check each operand against the limits, apply, check the result size, then
// normalise. The size check is here, not in each operation, so every operation
// shares one digit budget and a new operation cannot forget it. The checks run
// before Apply so an operation never sees input it was not written for.
//
// ctx is honoured at three points: before any work, inside Apply (the loops in
// power and square root check it every iteration) and after Apply, so a result
// finished past the deadline is dropped and the client gets REQUEST_TIMEOUT,
// never a late answer it has already given up on.
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

// ListOperations describes each registered operation, in registration order.
// The list lets a client discover the supported operations and their arity; the
// web keypad itself is static and uses the list only to disable keys the
// server does not support.
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
