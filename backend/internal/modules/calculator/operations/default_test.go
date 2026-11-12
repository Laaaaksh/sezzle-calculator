package operations_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/operations"
)

type DefaultRegistryTestSuite struct {
	suite.Suite
	registry *operations.Registry
}

func TestDefaultRegistryTestSuite(t *testing.T) {
	suite.Run(t, new(DefaultRegistryTestSuite))
}

func (s *DefaultRegistryTestSuite) SetupTest() {
	s.registry = operations.NewDefaultRegistry(entities.Limits{
		MaxOperandLength: testMaxOperandLength,
		MaxExponent:      testMaxExponent,
		DecimalPrecision: testConfiguredPrecision,
		MaxResultDigits:  testResultBudget,
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
	s.Equal(entities.CodeExponentTooLarge, applyErr.Code())
}

func (s *DefaultRegistryTestSuite) TestSqrtPicksUpThePrecisionFromConfig() {
	op, err := s.registry.Lookup(entities.OpSqrt)
	s.Require().Nil(err)

	got, applyErr := applyUnary(op, "2")

	s.Nil(applyErr)
	s.Equal(testSqrtTwoAt20, got)
}

func (s *DefaultRegistryTestSuite) TestAddIsListedWithItsSymbolDescriptionAndArity() {
	op := s.listed(entities.OpAdd)

	s.Equal(wantAddSymbol, op.Symbol())
	s.Equal(wantAddDescription, op.Description())
	s.Equal(2, op.Arity())
}

func (s *DefaultRegistryTestSuite) TestSubtractIsListedWithItsSymbolDescriptionAndArity() {
	op := s.listed(entities.OpSubtract)

	s.Equal(wantSubtractSymbol, op.Symbol())
	s.Equal(wantSubtractDescription, op.Description())
	s.Equal(2, op.Arity())
}

func (s *DefaultRegistryTestSuite) TestMultiplyIsListedWithItsSymbolDescriptionAndArity() {
	op := s.listed(entities.OpMultiply)

	s.Equal(wantMultiplySymbol, op.Symbol())
	s.Equal(wantMultiplyDescription, op.Description())
	s.Equal(2, op.Arity())
}

func (s *DefaultRegistryTestSuite) TestDivideIsListedWithItsSymbolDescriptionAndArity() {
	op := s.listed(entities.OpDivide)

	s.Equal(wantDivideSymbol, op.Symbol())
	s.Equal(wantDivideDescription, op.Description())
	s.Equal(2, op.Arity())
}

func (s *DefaultRegistryTestSuite) TestPowerIsListedWithItsSymbolDescriptionAndArity() {
	op := s.listed(entities.OpPower)

	s.Equal(wantPowerSymbol, op.Symbol())
	s.Equal(wantPowerDescription, op.Description())
	s.Equal(2, op.Arity())
}

func (s *DefaultRegistryTestSuite) TestSqrtIsListedWithItsSymbolDescriptionAndArity() {
	op := s.listed(entities.OpSqrt)

	s.Equal(wantSqrtSymbol, op.Symbol())
	s.Equal(wantSqrtDescription, op.Description())
	s.Equal(1, op.Arity())
}

func (s *DefaultRegistryTestSuite) TestPercentageIsListedWithItsSymbolDescriptionAndArity() {
	op := s.listed(entities.OpPercentage)

	s.Equal(wantPercentageSymbol, op.Symbol())
	s.Equal(wantPercentageDescription, op.Description())
	s.Equal(2, op.Arity())
}

func (s *DefaultRegistryTestSuite) TestListReturnsEveryOperationInRegistrationOrder() {
	var names []entities.OperationType
	for _, op := range s.registry.List() {
		names = append(names, op.Name())
	}

	s.Equal([]entities.OperationType{
		entities.OpAdd, entities.OpSubtract, entities.OpMultiply, entities.OpDivide,
		entities.OpPower, entities.OpSqrt, entities.OpPercentage,
	}, names)
}

func (s *DefaultRegistryTestSuite) listed(name entities.OperationType) operations.Operation {
	for _, op := range s.registry.List() {
		if op.Name() == name {
			return op
		}
	}
	s.FailNow("operation is not listed", "name %s", name)
	return nil
}

// maxArity is the widest arity among the registered operations.
func (s *DefaultRegistryTestSuite) maxArity() int {
	widest := 0
	for _, op := range s.registry.List() {
		widest = max(widest, op.Arity())
	}
	return widest
}

// The handler refuses more than entities.MaxOperandCount operands before it
// looks at the operation, so a wider operation would be unreachable.
func (s *DefaultRegistryTestSuite) TestMaxOperandCountIsAtLeastTheWidestArity() {
	s.LessOrEqual(s.maxArity(), entities.MaxOperandCount)
}
