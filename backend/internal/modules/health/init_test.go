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

// SetupTest and TearDownTest reset the package singleton so no test sees the
// module another one built.
func (s *InitTestSuite) SetupTest() {
	s.ctx = context.Background()
	health.HealthModule = nil
}

func (s *InitTestSuite) TearDownTest() {
	health.HealthModule = nil
}

func (s *InitTestSuite) TestNewModuleWiresCoreAndHandler() {
	module := health.NewModule(s.ctx)

	s.NotNil(module.GetCore())
	s.NotNil(module.GetHandler())
}

func (s *InitTestSuite) TestNewModuleReturnsSameInstanceOnSecondCall() {
	first := health.NewModule(s.ctx)

	second := health.NewModule(s.ctx)

	s.Same(first, second)
}

func (s *InitTestSuite) TestNewModuleStartsNotReady() {
	module := health.NewModule(s.ctx)

	s.False(module.GetCore().IsReady(s.ctx))
}
