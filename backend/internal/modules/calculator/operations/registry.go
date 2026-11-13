package operations

import (
	"slices"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

// IRegistry resolves an operation by name. The core depends on this interface
// so tests can supply their own registry.
type IRegistry interface {
	// Lookup returns the operation registered under name, or UNKNOWN_OPERATION.
	Lookup(name entities.OperationType) (Operation, apperror.IError)
	// List returns every operation in registration order. The order is fixed
	// so the operations endpoint answers the same way on every call.
	List() []Operation
}

// Registry maps operation names to operations. It is built once at boot and
// only read afterwards, so it needs no lock.
type Registry struct {
	byName map[entities.OperationType]Operation
	// ordered keeps registration order; a map alone would list operations in a
	// random order that changes between calls.
	ordered []Operation
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
	return &Registry{byName: byName, ordered: slices.Clone(ops)}
}

// Lookup returns the operation for name; anything unregistered, including the empty name, is UNKNOWN_OPERATION.
func (r *Registry) Lookup(name entities.OperationType) (Operation, apperror.IError) {
	op, ok := r.byName[name]
	if !ok {
		aerr := apperror.New(entities.CodeUnknownOperation, ErrUnknownOperation)
		// Echo the name only when it is short. Fields reach the client and the
		// warn log, so an unbounded name would flood both. Omitting beats
		// truncating: a cut name could look like a real one.
		if len(name) <= entities.MaxOperationNameLength {
			aerr = aerr.WithField(entities.FieldOperation, string(name))
		}
		return nil, aerr
	}
	return op, nil
}

// List returns a copy of the operations in registration order, so a caller
// cannot reorder or replace the registry's own slice.
func (r *Registry) List() []Operation {
	return slices.Clone(r.ordered)
}
