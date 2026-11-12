package entities_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

var calculatorCodes = []apperror.Code{
	entities.CodeUnknownOperation,
	entities.CodeInvalidOperandCount,
	entities.CodeOperandTooLarge,
	entities.CodeInvalidOperand,
	entities.CodeDivisionByZero,
	entities.CodeInvalidExponent,
	entities.CodeExponentTooLarge,
	entities.CodeResultTooLarge,
	entities.CodeNegativeSquareRoot,
}

type CodesTestSuite struct {
	suite.Suite
}

func TestCodesTestSuite(t *testing.T) {
	suite.Run(t, new(CodesTestSuite))
}

func (s *CodesTestSuite) TestUnknownOperationIs422WithItsMessage() {
	err := apperror.New(entities.CodeUnknownOperation, nil)

	s.Equal(http.StatusUnprocessableEntity, err.HTTPStatus())
	s.Equal(entities.MsgUnknownOperation, err.PublicMessage())
}

func (s *CodesTestSuite) TestInvalidOperandCountIs422WithItsMessage() {
	err := apperror.New(entities.CodeInvalidOperandCount, nil)

	s.Equal(http.StatusUnprocessableEntity, err.HTTPStatus())
	s.Equal(entities.MsgInvalidOperandCount, err.PublicMessage())
}

func (s *CodesTestSuite) TestOperandTooLargeIs422WithItsMessage() {
	err := apperror.New(entities.CodeOperandTooLarge, nil)

	s.Equal(http.StatusUnprocessableEntity, err.HTTPStatus())
	s.Equal(entities.MsgOperandTooLarge, err.PublicMessage())
}

func (s *CodesTestSuite) TestDivisionByZeroIs422WithItsMessage() {
	err := apperror.New(entities.CodeDivisionByZero, nil)

	s.Equal(http.StatusUnprocessableEntity, err.HTTPStatus())
	s.Equal(entities.MsgDivisionByZero, err.PublicMessage())
}

func (s *CodesTestSuite) TestDivisionByZeroMessageIsTheDocumentedText() {
	s.Equal("Cannot divide by zero.", entities.MsgDivisionByZero)
}

func (s *CodesTestSuite) TestInvalidExponentIs422WithItsMessage() {
	err := apperror.New(entities.CodeInvalidExponent, nil)

	s.Equal(http.StatusUnprocessableEntity, err.HTTPStatus())
	s.Equal(entities.MsgInvalidExponent, err.PublicMessage())
}

func (s *CodesTestSuite) TestExponentTooLargeIs422WithItsMessage() {
	err := apperror.New(entities.CodeExponentTooLarge, nil)

	s.Equal(http.StatusUnprocessableEntity, err.HTTPStatus())
	s.Equal(entities.MsgExponentTooLarge, err.PublicMessage())
}

func (s *CodesTestSuite) TestResultTooLargeIs422WithItsMessage() {
	err := apperror.New(entities.CodeResultTooLarge, nil)

	s.Equal(http.StatusUnprocessableEntity, err.HTTPStatus())
	s.Equal(entities.MsgResultTooLarge, err.PublicMessage())
}

func (s *CodesTestSuite) TestNegativeSquareRootIs422WithItsMessage() {
	err := apperror.New(entities.CodeNegativeSquareRoot, nil)

	s.Equal(http.StatusUnprocessableEntity, err.HTTPStatus())
	s.Equal(entities.MsgNegativeSquareRoot, err.PublicMessage())
}

func (s *CodesTestSuite) TestInvalidOperandIs422WithItsMessage() {
	err := apperror.New(entities.CodeInvalidOperand, nil)

	s.Equal(http.StatusUnprocessableEntity, err.HTTPStatus())
	s.Equal(entities.MsgInvalidOperand, err.PublicMessage())
}

func (s *CodesTestSuite) TestEveryCalculatorCodeIsListedByApperrorCodes() {
	listed := apperror.Codes()

	for _, code := range calculatorCodes {
		s.Contains(listed, code)
	}
}
