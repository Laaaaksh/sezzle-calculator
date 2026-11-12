package operations

import (
	"context"
	"errors"

	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

// CheckContext returns REQUEST_TIMEOUT once ctx has ended (deadline passed or
// request cancelled), and nil while it is live. Loops call it on every
// iteration and the core calls it around Apply, so a request deadline stops
// the work itself and not just the reply. The ctx error is kept as the cause
// for the logs.
func CheckContext(ctx context.Context) apperror.IError {
	if err := ctx.Err(); err != nil {
		return apperror.New(apperror.CodeRequestTimeout, errors.Join(ErrContextDone, err))
	}
	return nil
}
