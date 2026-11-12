package operations

import (
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

// IRegistry resolves an operation by name. The core depends on this interface
// so tests can supply their own registry.
type IRegistry interface {
	// Lookup returns the operation registered under name, or UNKNOWN_OPERATION.
	Lookup(name entities.OperationType) (Operation, apperror.IError)
}

// Registry maps operation names to operations. It is built once at boot and
// only read afterwards, so it needs no lock.
type Registry struct {
	byName map[entities.OperationType]Operation
}

var _ IRegistry = (*Registry)(nil)

// NewRegistry indexes ops by Name. It panics on a duplicate name: two
// operations fighting for one name is a wiring bug that must stop boot, not
// let the second silently replace the first.
func NewRegistry(ops ...Operation) *Registry {
	byName := make(map[entities.OperationType]Operation, len(ops))
	for _, op := range ops {
		if _, exists := byName[op.Name()]; exists {
			panic(errDuplicateOperation)
		}
		byName[op.Name()] = op
	}
	return &Registry{byName: byName}
}

// Lookup returns the operation for name. Anything unregistered, including the
// empty name, is UNKNOWN_OPERATION.
func (r *Registry) Lookup(name entities.OperationType) (Operation, apperror.IError) {
	op, ok := r.byName[name]
	if !ok {
		return nil, apperror.New(apperror.CodeUnknownOperation, ErrUnknownOperation).
			WithField(entities.FieldOperation, string(name))
	}
	return op, nil
}
