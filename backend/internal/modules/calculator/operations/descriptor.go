package operations

import "github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"

// descriptor holds what every operation reports about itself. Operations embed
// it, so a new operation states its identity once, as data, and the Operation
// methods that only return those values are not written seven times.
type descriptor struct {
	name        entities.OperationType
	arity       int
	symbol      string
	description string
}

// Name is the registry key and the name clients send.
func (d descriptor) Name() entities.OperationType { return d.name }

// Arity is the exact number of operands the operation needs.
func (d descriptor) Arity() int { return d.arity }

// Symbol is the short display symbol.
func (d descriptor) Symbol() string { return d.symbol }

// Description is the one-line help text.
func (d descriptor) Description() string { return d.description }
