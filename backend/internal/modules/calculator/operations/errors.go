package operations

import "errors"

// ErrUnknownOperation is the internal cause behind UNKNOWN_OPERATION. It stays
// in the logs; clients see only the generic public message.
var ErrUnknownOperation = errors.New("operation is not registered")

// ErrDivisionByZero is the internal cause behind DIVISION_BY_ZERO.
var ErrDivisionByZero = errors.New("divisor is zero")

// errDuplicateOperation is the panic value for registering one name twice.
var errDuplicateOperation = errors.New("operation registered twice")

// Internal causes behind the advanced operations' errors. Clients see only the
// public message of the error code.
var (
	ErrFractionalExponent  = errors.New("exponent is not a whole number")
	ErrExponentBeyondLimit = errors.New("exponent is outside the configured range")
	ErrZeroToNegativePower = errors.New("zero raised to a negative power")
	ErrResultBeyondBudget  = errors.New("printed result length exceeds the budget")
	ErrNegativeSquareRoot  = errors.New("square root of a negative number")
	ErrSqrtDidNotConverge  = errors.New("square root iteration hit its bound")
)

// ErrContextDone is the internal cause behind REQUEST_TIMEOUT when an
// operation notices that its context ended mid-calculation.
var ErrContextDone = errors.New("request context ended during the calculation")
