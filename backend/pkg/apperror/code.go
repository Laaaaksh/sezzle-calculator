// Package apperror is the single error model of the service: a typed code, a
// safe public message, an HTTP status and structured fields for logs.
package apperror

import "net/http"

// Code identifies a class of failure. A named type stops arbitrary strings
// being passed where a known code is expected.
type Code string

// Error codes. Calculator domain codes are 422: the request was well formed
// but the calculation it asks for cannot be performed.
const (
	CodeBadRequest           Code = "BAD_REQUEST"
	CodeValidationFailed     Code = "VALIDATION_FAILED"
	CodeUnsupportedMediaType Code = "UNSUPPORTED_MEDIA_TYPE"
	CodePayloadTooLarge      Code = "PAYLOAD_TOO_LARGE"
	CodeNotFound             Code = "NOT_FOUND"
	CodeMethodNotAllowed     Code = "METHOD_NOT_ALLOWED"
	CodeInternal             Code = "INTERNAL"
	CodeServiceUnavailable   Code = "SERVICE_UNAVAILABLE"
	CodeUnknownOperation     Code = "UNKNOWN_OPERATION"
	CodeInvalidOperandCount  Code = "INVALID_OPERAND_COUNT"
	CodeOperandTooLarge      Code = "OPERAND_TOO_LARGE"
	CodeDivisionByZero       Code = "DIVISION_BY_ZERO"
)

// httpStatusByCode is the only place a code is tied to a status.
var httpStatusByCode = map[Code]int{
	CodeBadRequest:           http.StatusBadRequest,
	CodeValidationFailed:     http.StatusUnprocessableEntity,
	CodeUnsupportedMediaType: http.StatusUnsupportedMediaType,
	CodePayloadTooLarge:      http.StatusRequestEntityTooLarge,
	CodeNotFound:             http.StatusNotFound,
	CodeMethodNotAllowed:     http.StatusMethodNotAllowed,
	CodeInternal:             http.StatusInternalServerError,
	CodeServiceUnavailable:   http.StatusServiceUnavailable,
	CodeUnknownOperation:     http.StatusUnprocessableEntity,
	CodeInvalidOperandCount:  http.StatusUnprocessableEntity,
	CodeOperandTooLarge:      http.StatusUnprocessableEntity,
	CodeDivisionByZero:       http.StatusUnprocessableEntity,
}

// publicMessageByCode holds the default client-facing text per code.
var publicMessageByCode = map[Code]string{
	CodeBadRequest:           MsgBadRequest,
	CodeValidationFailed:     MsgValidationFailed,
	CodeUnsupportedMediaType: MsgUnsupportedMediaType,
	CodePayloadTooLarge:      MsgPayloadTooLarge,
	CodeNotFound:             MsgNotFound,
	CodeMethodNotAllowed:     MsgMethodNotAllowed,
	CodeInternal:             MsgInternal,
	CodeServiceUnavailable:   MsgServiceUnavailable,
	CodeUnknownOperation:     MsgUnknownOperation,
	CodeInvalidOperandCount:  MsgInvalidOperandCount,
	CodeOperandTooLarge:      MsgOperandTooLarge,
	CodeDivisionByZero:       MsgDivisionByZero,
}

// HTTPStatus returns the status for c. An unknown code maps to 500 so a
// forgotten mapping fails closed as a server error, never as a success.
func (c Code) HTTPStatus() int {
	if status, ok := httpStatusByCode[c]; ok {
		return status
	}
	return http.StatusInternalServerError
}

// DefaultMessage returns the public text for c, falling back to the INTERNAL
// message for an unknown code.
func (c Code) DefaultMessage() string {
	if msg, ok := publicMessageByCode[c]; ok {
		return msg
	}
	return MsgInternal
}
