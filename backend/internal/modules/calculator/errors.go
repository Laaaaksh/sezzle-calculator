package calculator

import "errors"

// Internal causes behind the core's errors. They stay in the logs; clients see
// only the public message of the error code.
var (
	ErrWrongOperandCount = errors.New("operand count does not match the operation arity")
	ErrOperandDigits     = errors.New("operand has more digits than the configured limit")
	ErrOperandScale      = errors.New("operand exponent is beyond the configured limit")
	ErrOperandTooLong    = errors.New("operand string is longer than the configured limit")
	ErrOperandSyntax     = errors.New("operand is not a plain decimal number")
)
