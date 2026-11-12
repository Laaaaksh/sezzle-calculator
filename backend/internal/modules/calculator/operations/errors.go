package operations

import "errors"

// ErrUnknownOperation is the internal cause behind UNKNOWN_OPERATION. It stays
// in the logs; clients see only the generic public message.
var ErrUnknownOperation = errors.New("operation is not registered")

// ErrDivisionByZero is the internal cause behind DIVISION_BY_ZERO.
var ErrDivisionByZero = errors.New("divisor is zero")

// errDuplicateOperation is the panic value for registering one name twice.
var errDuplicateOperation = errors.New("operation registered twice")
