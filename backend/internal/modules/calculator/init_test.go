package calculator_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/config"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

type InitTestSuite struct {
	suite.Suite
	ctx context.Context
	cfg config.CalculatorConfig
}

func TestInitTestSuite(t *testing.T) {
	suite.Run(t, new(InitTestSuite))
}

// SetupTest and TearDownTest reset the package singleton so no test sees the
// module another one built.
func (s *InitTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.cfg = config.CalculatorConfig{MaxOperandLength: testMaxOperandLength, MaxExponent: testMaxExponent, DecimalPrecision: testPrecision}
	calculator.CalculatorModule = nil
}

func (s *InitTestSuite) TearDownTest() {
	calculator.CalculatorModule = nil
}

func (s *InitTestSuite) TestNewModuleWiresACore() {
	module := calculator.NewModule(s.ctx, s.cfg)

	s.NotNil(module.GetCore())
}

func (s *InitTestSuite) TestNewModuleReturnsSameInstanceOnSecondCall() {
	first := calculator.NewModule(s.ctx, s.cfg)

	second := calculator.NewModule(s.ctx, s.cfg)

	s.Same(first, second)
}

func (s *InitTestSuite) TestNewModuleCoreRejectsOperationsThatAreNotRegistered() {
	module := calculator.NewModule(s.ctx, s.cfg)

	_, err := module.GetCore().Calculate(s.ctx, testOpMissing, nil)

	s.Require().NotNil(err)
	s.Equal(apperror.CodeUnknownOperation, err.Code())
}

func (s *InitTestSuite) TestNewModuleWiresAHandler() {
	module := calculator.NewModule(s.ctx, s.cfg)

	s.NotNil(module.GetHandler())
}
