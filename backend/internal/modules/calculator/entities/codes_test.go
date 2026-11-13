package entities_test

import (
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

func (s *CodesTestSuite) TestEveryCalculatorCodeIsListedByApperrorCodes() {
	listed := apperror.Codes()

	for _, code := range calculatorCodes {
		s.Contains(listed, code)
	}
}
