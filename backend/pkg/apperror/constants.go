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
	MsgUnknownOperation     = "The requested operation is not supported."
	MsgInvalidOperandCount  = "The operation received the wrong number of operands."
	MsgOperandTooLarge      = "An operand is larger than the allowed limit."
	MsgDivisionByZero       = "Cannot divide by zero."
)

// Formats for the internal Error() string, which is for logs only.
const (
	errFmtNoCause   = "%s: %s"
	errFmtWithCause = "%s: %s: %v"
)
