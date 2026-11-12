package health_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/health"
)

type InitTestSuite struct {
	suite.Suite
	ctx context.Context
}

func TestInitTestSuite(t *testing.T) {
	suite.Run(t, new(InitTestSuite))
}

func (s *InitTestSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *InitTestSuite) TestNewModuleWiresCoreAndHandler() {
	module := health.NewModule(s.ctx)

	s.NotNil(module.GetCore())
	s.NotNil(module.GetHandler())
}

func (s *InitTestSuite) TestNewModuleReturnsADistinctInstanceOnEachCall() {
	first := health.NewModule(s.ctx)

	second := health.NewModule(s.ctx)

	s.NotSame(first, second)
}

func (s *InitTestSuite) TestNewModuleStartsNotReady() {
	module := health.NewModule(s.ctx)

	s.False(module.GetCore().IsReady(s.ctx))
}

func (s *InitTestSuite) TestNewModuleDoesNotShareReadinessBetweenInstances() {
	first := health.NewModule(s.ctx)
	first.GetCore().SetReady(true)

	second := health.NewModule(s.ctx)

	s.False(second.GetCore().IsReady(s.ctx))
}
