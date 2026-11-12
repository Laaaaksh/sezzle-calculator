package apperror

// Default public messages. They are deliberately generic: anything specific
// about the cause stays in the internal error and the logs.
const (
	MsgBadRequest           = "The request could not be understood."
	MsgValidationFailed     = "The request failed validation."
	MsgUnsupportedMediaType = "The content type is not supported."
	MsgPayloadTooLarge      = "The request body is too large."
	MsgNotFound             = "The requested resource was not found."
	MsgMethodNotAllowed     = "The method is not allowed for this resource."
	MsgInternal             = "An internal error occurred."
	MsgServiceUnavailable   = "The service is temporarily unavailable."
	MsgRequestTimeout       = "The request took too long to process."
	MsgUnknownOperation     = "The requested operation is not supported."
	MsgInvalidOperandCount  = "The operation received the wrong number of operands."
	MsgOperandTooLarge      = "An operand is larger than the allowed limit."
	MsgInvalidOperand       = "An operand is not a valid number."
	MsgDivisionByZero       = "Cannot divide by zero."
	MsgInvalidExponent      = "The exponent must be a whole number."
	MsgExponentTooLarge     = "The exponent is outside the allowed range."
	MsgResultTooLarge       = "The result would be too large to calculate."
	MsgNegativeSquareRoot   = "Cannot take the square root of a negative number."
)

// Formats for the internal Error() string, which is for logs only.
const (
	errFmtNoCause   = "%s: %s"
	errFmtWithCause = "%s: %s: %v"
)
