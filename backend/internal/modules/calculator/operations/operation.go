// Package operations holds the calculator's operations and the registry that
// resolves them by name. Each operation lives in its own file.
package operations

import (
	"context"

	"github.com/shopspring/decimal"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

// Operation is one arithmetic operation. Implementations are stateless apart
// from configuration injected at construction, so they are safe to share
// across requests.
type Operation interface {
	// Name is the registry key and the name clients send.
	Name() entities.OperationType
	// Arity is the exact number of operands Apply needs. The core checks it
	// before calling Apply, so Apply may index operands without a length check.
	Arity() int
	// Symbol is a short display symbol such as "+". It is for showing an
	// expression and is never accepted as input.
	Symbol() string
	// Description is one line of plain text for help text and tooltips.
	Description() string
	// Apply computes the result. It uses decimal maths only: a float64 would
	// make 0.1 + 0.2 differ from 0.3. Operations that loop (power, square
	// root) check ctx on every iteration and return REQUEST_TIMEOUT once it is
	// done, so a request deadline stops the work and not just the reply.
	Apply(ctx context.Context, operands []decimal.Decimal) (decimal.Decimal, apperror.IError)
}
