package apperror_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

type CalculatorCodesTestSuite struct {
	suite.Suite
}

func TestCalculatorCodesTestSuite(t *testing.T) {
	suite.Run(t, new(CalculatorCodesTestSuite))
}

func (s *CalculatorCodesTestSuite) TestUnknownOperationIs422WithItsMessage() {
	err := apperror.New(apperror.CodeUnknownOperation, nil)

	s.Equal(http.StatusUnprocessableEntity, err.HTTPStatus())
	s.Equal(apperror.MsgUnknownOperation, err.PublicMessage())
}

func (s *CalculatorCodesTestSuite) TestInvalidOperandCountIs422WithItsMessage() {
	err := apperror.New(apperror.CodeInvalidOperandCount, nil)

	s.Equal(http.StatusUnprocessableEntity, err.HTTPStatus())
	s.Equal(apperror.MsgInvalidOperandCount, err.PublicMessage())
}

func (s *CalculatorCodesTestSuite) TestOperandTooLargeIs422WithItsMessage() {
	err := apperror.New(apperror.CodeOperandTooLarge, nil)

	s.Equal(http.StatusUnprocessableEntity, err.HTTPStatus())
	s.Equal(apperror.MsgOperandTooLarge, err.PublicMessage())
}

func (s *CalculatorCodesTestSuite) TestDivisionByZeroIs422WithItsMessage() {
	err := apperror.New(apperror.CodeDivisionByZero, nil)

	s.Equal(http.StatusUnprocessableEntity, err.HTTPStatus())
	s.Equal(apperror.MsgDivisionByZero, err.PublicMessage())
}

func (s *CalculatorCodesTestSuite) TestDivisionByZeroMessageIsTheDocumentedText() {
	s.Equal("Cannot divide by zero.", apperror.MsgDivisionByZero)
}

func (s *CalculatorCodesTestSuite) TestInvalidExponentIs422WithItsMessage() {
	err := apperror.New(apperror.CodeInvalidExponent, nil)

	s.Equal(http.StatusUnprocessableEntity, err.HTTPStatus())
	s.Equal(apperror.MsgInvalidExponent, err.PublicMessage())
}

func (s *CalculatorCodesTestSuite) TestExponentTooLargeIs422WithItsMessage() {
	err := apperror.New(apperror.CodeExponentTooLarge, nil)

	s.Equal(http.StatusUnprocessableEntity, err.HTTPStatus())
	s.Equal(apperror.MsgExponentTooLarge, err.PublicMessage())
}

func (s *CalculatorCodesTestSuite) TestResultTooLargeIs422WithItsMessage() {
	err := apperror.New(apperror.CodeResultTooLarge, nil)

	s.Equal(http.StatusUnprocessableEntity, err.HTTPStatus())
	s.Equal(apperror.MsgResultTooLarge, err.PublicMessage())
}

func (s *CalculatorCodesTestSuite) TestNegativeSquareRootIs422WithItsMessage() {
	err := apperror.New(apperror.CodeNegativeSquareRoot, nil)

	s.Equal(http.StatusUnprocessableEntity, err.HTTPStatus())
	s.Equal(apperror.MsgNegativeSquareRoot, err.PublicMessage())
}
