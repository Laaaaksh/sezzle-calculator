package calculator_test

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/config"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
)

type InitTestSuite struct {
	suite.Suite
	ctx context.Context
	cfg config.CalculatorConfig
}

func TestInitTestSuite(t *testing.T) {
	suite.Run(t, new(InitTestSuite))
}

func (s *InitTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.cfg = config.CalculatorConfig{MaxOperandLength: testMaxOperandLength, MaxExponent: testMaxExponent, DecimalPrecision: testPrecision}
}

func (s *InitTestSuite) TestNewModuleWiresACore() {
	module := calculator.NewModule(s.ctx, s.cfg)

	s.NotNil(module.GetCore())
}

func (s *InitTestSuite) TestNewModuleReturnsADistinctInstanceOnEachCall() {
	first := calculator.NewModule(s.ctx, s.cfg)

	second := calculator.NewModule(s.ctx, s.cfg)

	s.NotSame(first, second)
}

func (s *InitTestSuite) TestNewModuleCoreRejectsOperationsThatAreNotRegistered() {
	module := calculator.NewModule(s.ctx, s.cfg)

	_, err := module.GetCore().Calculate(s.ctx, testOpMissing, nil)

	s.Require().NotNil(err)
	s.Equal(entities.CodeUnknownOperation, err.Code())
}

func (s *InitTestSuite) TestNewModuleWiresAHandler() {
	module := calculator.NewModule(s.ctx, s.cfg)

	s.NotNil(module.GetHandler())
}

func (s *InitTestSuite) TestNewModuleUsesTheConfigOfEachCall() {
	tight := s.cfg
	tight.MaxExponent = testTightExponent
	loose := s.cfg
	loose.MaxExponent = testMaxExponent
	operands := []decimal.Decimal{mustDecimal("2"), mustDecimal(testExponentAboveTight)}

	_, tightErr := calculator.NewModule(s.ctx, tight).GetCore().Calculate(s.ctx, entities.OpPower, operands)
	_, looseErr := calculator.NewModule(s.ctx, loose).GetCore().Calculate(s.ctx, entities.OpPower, operands)

	s.Require().NotNil(tightErr)
	s.Equal(entities.CodeExponentTooLarge, tightErr.Code())
	s.Nil(looseErr)
}
