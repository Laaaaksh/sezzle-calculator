package apperror

import (
	"errors"
	"fmt"
	"maps"
)

// IError is the contract every service error satisfies.
type IError interface {
	error
	Code() Code
	// PublicMessage is safe to show a client; it never contains the cause.
	PublicMessage() string
	HTTPStatus() int
	// Fields returns a copy of the structured context for logging.
	Fields() map[string]any
	Unwrap() error
	// WithField returns a new error with the extra field; the receiver is unchanged.
	WithField(key string, value any) IError
}

// Error is the concrete IError.
type Error struct {
	code    Code
	message string
	cause   error
	fields  map[string]any
}

var _ IError = (*Error)(nil)

// New builds an error with the default public message for code. cause may be nil.
func New(code Code, cause error) IError {
	return NewWithMessage(code, cause, code.DefaultMessage())
}

// NewWithMessage builds an error with a custom public message. The message
// reaches clients, so it must never embed cause text. cause may be nil.
func NewWithMessage(code Code, cause error, msg string) IError {
	return &Error{code: code, message: msg, cause: cause}
}

// Error is the internal description (code, message and cause) for logs only.
// Never write it to a response.
func (e *Error) Error() string {
	if e.cause == nil {
		return fmt.Sprintf(errFmtNoCause, e.code, e.message)
	}
	return fmt.Sprintf(errFmtWithCause, e.code, e.message, e.cause)
}

// Code returns the error class.
func (e *Error) Code() Code { return e.code }

// PublicMessage returns the client-safe message.
func (e *Error) PublicMessage() string { return e.message }

// HTTPStatus returns the status mapped to the code.
func (e *Error) HTTPStatus() int { return e.code.HTTPStatus() }

// Unwrap returns the cause, or nil.
func (e *Error) Unwrap() error { return e.cause }

// Fields returns a copy, so a caller mutating it cannot alter a shared error.
func (e *Error) Fields() map[string]any {
	out := make(map[string]any, len(e.fields))
	maps.Copy(out, e.fields)
	return out
}

// WithField copies the field map instead of mutating it: sentinel errors are
// shared across requests, and writing to them would be a data race.
func (e *Error) WithField(key string, value any) IError {
	fields := make(map[string]any, len(e.fields)+1)
	maps.Copy(fields, e.fields)
	fields[key] = value
	return &Error{code: e.code, message: e.message, cause: e.cause, fields: fields}
}

// From returns the IError inside err, or wraps err as INTERNAL so a handler
// can treat every error uniformly. It returns nil for a nil err.
func From(err error) IError {
	if err == nil {
		return nil
	}
	var ae *Error
	if errors.As(err, &ae) {
		return ae
	}
	return New(CodeInternal, err)
}
