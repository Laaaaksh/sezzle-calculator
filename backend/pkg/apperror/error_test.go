package apperror_test

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

const (
	testCauseText     = "pq: password authentication failed for user admin"
	testCustomMessage = "operand is too long"
	testUnknownCode   = apperror.Code("MADE_UP")
	testFieldKey      = "operand"
	testFieldKeyTwo   = "limit"
	testFieldValue    = "123"
	testFieldValueTwo = 1000
	testWrapText      = "handler"
	testPathText      = "/missing"
)

var errTestCause = errors.New(testCauseText)

type ErrorTestSuite struct {
	suite.Suite
}

func TestErrorTestSuite(t *testing.T) {
	suite.Run(t, new(ErrorTestSuite))
}

func (s *ErrorTestSuite) TestStatusBadRequestIs400() {
	s.Equal(http.StatusBadRequest, apperror.New(apperror.CodeBadRequest, nil).HTTPStatus())
}

func (s *ErrorTestSuite) TestStatusValidationFailedIs422() {
	s.Equal(http.StatusUnprocessableEntity, apperror.New(apperror.CodeValidationFailed, nil).HTTPStatus())
}

func (s *ErrorTestSuite) TestStatusUnsupportedMediaTypeIs415() {
	s.Equal(http.StatusUnsupportedMediaType, apperror.New(apperror.CodeUnsupportedMediaType, nil).HTTPStatus())
}

func (s *ErrorTestSuite) TestStatusPayloadTooLargeIs413() {
	s.Equal(http.StatusRequestEntityTooLarge, apperror.New(apperror.CodePayloadTooLarge, nil).HTTPStatus())
}

func (s *ErrorTestSuite) TestStatusNotFoundIs404() {
	s.Equal(http.StatusNotFound, apperror.New(apperror.CodeNotFound, nil).HTTPStatus())
}

func (s *ErrorTestSuite) TestStatusMethodNotAllowedIs405() {
	s.Equal(http.StatusMethodNotAllowed, apperror.New(apperror.CodeMethodNotAllowed, nil).HTTPStatus())
}

func (s *ErrorTestSuite) TestStatusInternalIs500() {
	s.Equal(http.StatusInternalServerError, apperror.New(apperror.CodeInternal, nil).HTTPStatus())
}

func (s *ErrorTestSuite) TestStatusUnknownCodeIs500() {
	s.Equal(http.StatusInternalServerError, apperror.New(testUnknownCode, nil).HTTPStatus())
}

func (s *ErrorTestSuite) TestNewUnknownCodeUsesInternalMessage() {
	err := apperror.New(testUnknownCode, nil)

	s.Equal(testUnknownCode, err.Code())
	s.Equal(apperror.MsgInternal, err.PublicMessage())
}

func (s *ErrorTestSuite) TestNewUsesDefaultMessage() {
	err := apperror.New(apperror.CodeNotFound, nil)

	s.Equal(apperror.CodeNotFound, err.Code())
	s.Equal(apperror.MsgNotFound, err.PublicMessage())
}

func (s *ErrorTestSuite) TestNewWithMessageOverridesDefault() {
	err := apperror.NewWithMessage(apperror.CodeValidationFailed, nil, testCustomMessage)

	s.Equal(testCustomMessage, err.PublicMessage())
	s.Equal(apperror.CodeValidationFailed, err.Code())
}

func (s *ErrorTestSuite) TestPublicMessageExcludesCause() {
	err := apperror.New(apperror.CodeInternal, errTestCause)

	s.NotContains(err.PublicMessage(), testCauseText)
}

func (s *ErrorTestSuite) TestErrorStringIncludesCodeMessageAndCause() {
	err := apperror.New(apperror.CodeInternal, errTestCause)

	s.Contains(err.Error(), string(apperror.CodeInternal))
	s.Contains(err.Error(), apperror.MsgInternal)
	s.Contains(err.Error(), testCauseText)
}

func (s *ErrorTestSuite) TestErrorStringWithNilCauseOmitsCause() {
	err := apperror.New(apperror.CodeNotFound, nil)

	s.Contains(err.Error(), string(apperror.CodeNotFound))
	s.Contains(err.Error(), apperror.MsgNotFound)
}

func (s *ErrorTestSuite) TestUnwrapWithNilCauseReturnsNil() {
	s.Nil(apperror.New(apperror.CodeBadRequest, nil).Unwrap())
}

func (s *ErrorTestSuite) TestUnwrapReturnsCause() {
	s.Equal(errTestCause, apperror.New(apperror.CodeInternal, errTestCause).Unwrap())
}

func (s *ErrorTestSuite) TestErrorsIsReachesCause() {
	err := apperror.New(apperror.CodeInternal, errTestCause)

	s.True(errors.Is(err, errTestCause))
}

func (s *ErrorTestSuite) TestErrorsAsFindsCauseType() {
	cause := &fs.PathError{Op: testWrapText, Path: testPathText, Err: fs.ErrNotExist}
	err := apperror.New(apperror.CodeInternal, cause)

	var pathErr *fs.PathError

	s.True(errors.As(err, &pathErr))
	s.Equal(testPathText, pathErr.Path)
}

func (s *ErrorTestSuite) TestErrorsAsFindsAppErrorThroughFmtWrap() {
	wrapped := fmt.Errorf("%s: %w", testWrapText, apperror.New(apperror.CodeNotFound, nil))

	var target *apperror.Error

	s.True(errors.As(wrapped, &target))
	s.Equal(apperror.CodeNotFound, target.Code())
}

func (s *ErrorTestSuite) TestWithFieldAddsField() {
	err := apperror.New(apperror.CodeBadRequest, nil).WithField(testFieldKey, testFieldValue)

	s.Equal(map[string]any{testFieldKey: testFieldValue}, err.Fields())
}

func (s *ErrorTestSuite) TestWithFieldKeepsEarlierFields() {
	err := apperror.New(apperror.CodeBadRequest, nil).
		WithField(testFieldKey, testFieldValue).
		WithField(testFieldKeyTwo, testFieldValueTwo)

	s.Equal(map[string]any{testFieldKey: testFieldValue, testFieldKeyTwo: testFieldValueTwo}, err.Fields())
}

func (s *ErrorTestSuite) TestWithFieldDoesNotMutateReceiver() {
	original := apperror.New(apperror.CodeBadRequest, nil)

	derived := original.WithField(testFieldKey, testFieldValue)

	s.Empty(original.Fields())
	s.Len(derived.Fields(), 1)
}

func (s *ErrorTestSuite) TestWithFieldPreservesCodeMessageAndCause() {
	original := apperror.NewWithMessage(apperror.CodeValidationFailed, errTestCause, testCustomMessage)

	derived := original.WithField(testFieldKey, testFieldValue)

	s.Equal(original.Code(), derived.Code())
	s.Equal(original.PublicMessage(), derived.PublicMessage())
	s.True(errors.Is(derived, errTestCause))
}

func (s *ErrorTestSuite) TestFieldsReturnsCopy() {
	err := apperror.New(apperror.CodeBadRequest, nil).WithField(testFieldKey, testFieldValue)

	err.Fields()[testFieldKey] = testCustomMessage

	s.Equal(testFieldValue, err.Fields()[testFieldKey])
}

func (s *ErrorTestSuite) TestFieldsWithoutFieldsIsEmptyNotNil() {
	fields := apperror.New(apperror.CodeBadRequest, nil).Fields()

	s.NotNil(fields)
	s.Empty(fields)
}

func (s *ErrorTestSuite) TestFromNilReturnsNil() {
	s.Nil(apperror.From(nil))
}

func (s *ErrorTestSuite) TestFromAppErrorReturnsItUnchanged() {
	original := apperror.New(apperror.CodeNotFound, nil)

	s.Equal(original, apperror.From(fmt.Errorf("%s: %w", testWrapText, original)))
}

func (s *ErrorTestSuite) TestFromPlainErrorWrapsAsInternal() {
	err := apperror.From(errTestCause)

	s.Equal(apperror.CodeInternal, err.Code())
	s.True(errors.Is(err, errTestCause))
	s.NotContains(err.PublicMessage(), testCauseText)
}

func (s *ErrorTestSuite) TestToPublicCarriesCodeAndMessageOnly() {
	err := apperror.New(apperror.CodeNotFound, errTestCause)

	pub := apperror.ToPublic(err)

	s.Equal(apperror.Public{Code: apperror.CodeNotFound, Message: apperror.MsgNotFound}, pub)
}
