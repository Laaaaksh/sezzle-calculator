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

// Internal causes behind the HTTP handler's errors. They stay in the logs;
// clients see only the public message of the error code.
var (
	ErrContentTypeMissing  = errors.New("content type is missing or unreadable")
	ErrContentTypeNotJSON  = errors.New("content type is not application/json")
	ErrCharsetNotUTF8      = errors.New("content type charset is not utf-8")
	ErrEmptyBody           = errors.New("request body is empty")
	ErrTrailingData        = errors.New("request body has data after the JSON object")
	ErrMissingRequestField = errors.New("operation or operands is missing")
)
