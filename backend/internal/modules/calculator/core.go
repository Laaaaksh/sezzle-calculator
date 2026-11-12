// Package calculator is the calculator domain: operations, the registry that
// resolves them and the core service that validates and normalises results.
package calculator

import (
	"context"

	"github.com/shopspring/decimal"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/config"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/operations"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

//go:generate go tool mockgen -source=core.go -destination=mock/mock_core.go -package=mock

// ICore is the calculator business logic.
type ICore interface {
	// Calculate applies op to operands and returns the normalised result.
	Calculate(ctx context.Context, op entities.OperationType, operands []decimal.Decimal) (entities.Result, apperror.IError)
}

// Core resolves, validates and runs operations.
type Core struct {
	registry operations.IRegistry
	limits   config.CalculatorConfig
}

var _ ICore = (*Core)(nil)

// NewCore returns a Core that resolves operations from registry and enforces
// the operand limits in limits.
func NewCore(registry operations.IRegistry, limits config.CalculatorConfig) *Core {
	return &Core{registry: registry, limits: limits}
}

// Calculate runs the pipeline: look up the operation, check the operand count,
// check each operand against the limits, apply, then normalise. The checks run
// before Apply so an operation never sees input it was not written for.
func (c *Core) Calculate(_ context.Context, op entities.OperationType, operands []decimal.Decimal) (entities.Result, apperror.IError) {
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
	value, aerr := operation.Apply(operands)
	if aerr != nil {
		return entities.Result{}, aerr
	}
	return normalise(value), nil
}
