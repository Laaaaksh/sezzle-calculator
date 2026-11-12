package operations_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/config"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/operations"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

type DefaultRegistryTestSuite struct {
	suite.Suite
	registry *operations.Registry
}

func TestDefaultRegistryTestSuite(t *testing.T) {
	suite.Run(t, new(DefaultRegistryTestSuite))
}

func (s *DefaultRegistryTestSuite) SetupTest() {
	s.registry = operations.NewDefaultRegistry(config.CalculatorConfig{
		MaxOperandLength: testMaxOperandLength,
		MaxExponent:      testMaxExponent,
		DecimalPrecision: testConfiguredPrecision,
	})
}

func (s *DefaultRegistryTestSuite) lookupName(name entities.OperationType) entities.OperationType {
	op, err := s.registry.Lookup(name)
	s.Require().Nil(err)
	return op.Name()
}

func (s *DefaultRegistryTestSuite) TestRegistersAdd() {
	s.Equal(entities.OpAdd, s.lookupName(entities.OpAdd))
}

func (s *DefaultRegistryTestSuite) TestRegistersSubtract() {
	s.Equal(entities.OpSubtract, s.lookupName(entities.OpSubtract))
}

func (s *DefaultRegistryTestSuite) TestRegistersMultiply() {
	s.Equal(entities.OpMultiply, s.lookupName(entities.OpMultiply))
}

func (s *DefaultRegistryTestSuite) TestRegistersDivide() {
	s.Equal(entities.OpDivide, s.lookupName(entities.OpDivide))
}

func (s *DefaultRegistryTestSuite) TestDividePicksUpThePrecisionFromConfig() {
	op, err := s.registry.Lookup(entities.OpDivide)
	s.Require().Nil(err)

	got, applyErr := applyText(op, "1", "3")

	s.Nil(applyErr)
	s.Equal("0.33333333333333333333", got)
}

func (s *DefaultRegistryTestSuite) TestRegistersPower() {
	s.Equal(entities.OpPower, s.lookupName(entities.OpPower))
}

func (s *DefaultRegistryTestSuite) TestRegistersSqrt() {
	s.Equal(entities.OpSqrt, s.lookupName(entities.OpSqrt))
}

func (s *DefaultRegistryTestSuite) TestRegistersPercentage() {
	s.Equal(entities.OpPercentage, s.lookupName(entities.OpPercentage))
}

func (s *DefaultRegistryTestSuite) TestPowerPicksUpTheExponentLimitFromConfig() {
	op, err := s.registry.Lookup(entities.OpPower)
	s.Require().Nil(err)

	_, applyErr := applyText(op, "1", "1001")

	s.Require().NotNil(applyErr)
	s.Equal(apperror.CodeExponentTooLarge, applyErr.Code())
}

func (s *DefaultRegistryTestSuite) TestSqrtPicksUpThePrecisionFromConfig() {
	op, err := s.registry.Lookup(entities.OpSqrt)
	s.Require().Nil(err)

	got, applyErr := applyUnary(op, "2")

	s.Nil(applyErr)
	s.Equal(testSqrtTwoAt20, got)
}
