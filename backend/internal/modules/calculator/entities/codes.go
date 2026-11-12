package entities

import (
	"net/http"

	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

// Calculator domain error codes. They are 422: the request was well formed but
// the calculation it asks for cannot be performed. pkg/apperror owns only the
// generic codes; these are added to its registry by init below.
const (
	CodeUnknownOperation    apperror.Code = "UNKNOWN_OPERATION"
	CodeInvalidOperandCount apperror.Code = "INVALID_OPERAND_COUNT"
	CodeOperandTooLarge     apperror.Code = "OPERAND_TOO_LARGE"
	CodeInvalidOperand      apperror.Code = "INVALID_OPERAND"
	CodeDivisionByZero      apperror.Code = "DIVISION_BY_ZERO"
	CodeInvalidExponent     apperror.Code = "INVALID_EXPONENT"
	CodeExponentTooLarge    apperror.Code = "EXPONENT_TOO_LARGE"
	CodeResultTooLarge      apperror.Code = "RESULT_TOO_LARGE"
	CodeNegativeSquareRoot  apperror.Code = "NEGATIVE_SQUARE_ROOT"
)

// Default public messages for the calculator codes. Like the generic ones they
// are deliberately generic: anything specific stays in the cause and the logs.
const (
	MsgUnknownOperation    = "The requested operation is not supported."
	MsgInvalidOperandCount = "The operation received the wrong number of operands."
	MsgOperandTooLarge     = "An operand is larger than the allowed limit."
	MsgInvalidOperand      = "An operand is not a valid number."
	MsgDivisionByZero      = "Cannot divide by zero."
	MsgInvalidExponent     = "The exponent must be a whole number."
	MsgExponentTooLarge    = "The exponent is outside the allowed range."
	MsgResultTooLarge      = "The result would be too large to calculate."
	MsgNegativeSquareRoot  = "Cannot take the square root of a negative number."
)

// init registers the codes in package init, not in NewModule: NewModule runs
// many times in tests and Register rejects a duplicate, while init runs exactly
// once and guarantees any importer of this package, the docs-sync test
// included, sees every calculator code in apperror.Codes.
func init() {
	for code, def := range map[apperror.Code]struct {
		status  int
		message string
	}{
		CodeUnknownOperation:    {http.StatusUnprocessableEntity, MsgUnknownOperation},
		CodeInvalidOperandCount: {http.StatusUnprocessableEntity, MsgInvalidOperandCount},
		CodeOperandTooLarge:     {http.StatusUnprocessableEntity, MsgOperandTooLarge},
		CodeInvalidOperand:      {http.StatusUnprocessableEntity, MsgInvalidOperand},
		CodeDivisionByZero:      {http.StatusUnprocessableEntity, MsgDivisionByZero},
		CodeInvalidExponent:     {http.StatusUnprocessableEntity, MsgInvalidExponent},
		CodeExponentTooLarge:    {http.StatusUnprocessableEntity, MsgExponentTooLarge},
		CodeResultTooLarge:      {http.StatusUnprocessableEntity, MsgResultTooLarge},
		CodeNegativeSquareRoot:  {http.StatusUnprocessableEntity, MsgNegativeSquareRoot},
	} {
		apperror.Register(code, def.status, def.message)
	}
}
