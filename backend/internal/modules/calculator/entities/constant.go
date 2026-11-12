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
)
