package operations

import "errors"

// ErrUnknownOperation is the internal cause behind UNKNOWN_OPERATION. It stays
// in the logs; clients see only the generic public message.
var ErrUnknownOperation = errors.New("operation is not registered")

// errDuplicateOperation is the panic value for registering one name twice.
var errDuplicateOperation = errors.New("operation registered twice")
