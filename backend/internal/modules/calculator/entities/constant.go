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

// Display symbols, as listed by GET /api/v1/operations. Clients always send the operation name, never a symbol.
// Multiply and divide use plain ASCII so they survive any font; a frontend may draw "*" as a multiplication sign.
const (
	SymbolAdd        = "+"
	SymbolSubtract   = "-"
	SymbolMultiply   = "*"
	SymbolDivide     = "/"
	SymbolPower      = "^"
	SymbolSqrt       = "\u221a"
	SymbolPercentage = "%"
)

// One-line descriptions, as listed by GET /api/v1/operations.
const (
	DescriptionAdd        = "Adds the two operands."
	DescriptionSubtract   = "Subtracts the second operand from the first."
	DescriptionMultiply   = "Multiplies the two operands."
	DescriptionDivide     = "Divides the first operand by the second."
	DescriptionPower      = "Raises the first operand to the whole-number power given by the second."
	DescriptionSqrt       = "Returns the square root of the operand."
	DescriptionPercentage = "Returns the first operand percent of the second."
)

// Route paths, relative to the versioned API root (/api/v1).
const (
	PathCalculate  = "/calculate"
	PathOperations = "/operations"
)

// Public field keys in a VALIDATION_FAILED body, and the problem text for each.
const (
	FieldNameOperation = "operation"
	FieldNameOperands  = "operands"
	ProblemRequired    = "is required"
	ProblemTooLong     = "is too long"
	ProblemTooMany     = "too many"
)

// MaxOperandCount is a cheap pre-parse bound, well above any arity, so a request cannot make the service parse
// thousands of strings only for the arity check to reject them. A list within it but of the wrong size still reaches
// the core and gets INVALID_OPERAND_COUNT. A test keeps it at or above the widest arity.
const MaxOperandCount = 16

// MaxOperationNameLength is the longest operation name, in bytes, the API accepts or echoes back (the longest real
// name is "percentage", 10). Without a bound, every error response and warn log line could be as large as the body cap.
const MaxOperationNameLength = 32

// Error field keys attached to calculator errors, so logs and clients can say
// which operation or operand failed.
const (
	FieldOperation    = "operation"
	FieldOperandIndex = "operand_index"
	FieldExpected     = "expected"
	FieldActual       = "actual"
	FieldLimit        = "limit"
	FieldReason       = "reason"
	// FieldKey names the offending JSON key in a BAD_REQUEST about the body's keys.
	FieldKey = "key"
)

// Reasons a JSON key is rejected, reported in the "reason" field of
// BAD_REQUEST next to the "key" field.
const (
	ReasonDuplicateKey = "duplicate"
	ReasonKeyWrongCase = "wrong_case"
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
