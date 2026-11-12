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
