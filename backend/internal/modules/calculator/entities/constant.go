// Package entities holds the calculator module's types and constants.
package entities

// OperationType names an operation. A named type stops an arbitrary string
// being used where a registered operation name is expected.
type OperationType string

// Operation names, as clients send them.
const (
	OpAdd      OperationType = "add"
	OpSubtract OperationType = "subtract"
	OpMultiply OperationType = "multiply"
	OpDivide   OperationType = "divide"
	OpPower    OperationType = "power"
	OpSqrt     OperationType = "sqrt"

	// OpPercentage means "x percent of y": x / 100 * y.
	OpPercentage OperationType = "percentage"
)

// Arity values: how many operands an operation consumes.
const (
	ArityUnary  = 1
	ArityBinary = 2
)

// Error field keys attached to calculator errors, so logs and clients can say
// which operation or operand failed.
const (
	FieldOperation    = "operation"
	FieldOperandIndex = "operand_index"
	FieldExpected     = "expected"
	FieldActual       = "actual"
	FieldLimit        = "limit"
	FieldReason       = "reason"
)

// Reasons an operand string is rejected, reported in the "reason" field of
// INVALID_OPERAND so a client can tell the user what to fix.
const (
	ReasonEmpty          = "empty"
	ReasonWhitespace     = "whitespace"
	ReasonPlusSign       = "plus_sign"
	ReasonExponent       = "exponent"
	ReasonHex            = "hex"
	ReasonNaN            = "nan"
	ReasonInfinity       = "infinity"
	ReasonSeparator      = "separator"
	ReasonMultiplePoints = "multiple_points"
	ReasonInvalidFormat  = "invalid_format"
)
