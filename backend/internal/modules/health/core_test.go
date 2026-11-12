package health_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/health"
)

type CoreTestSuite struct {
	suite.Suite
	ctx  context.Context
	core *health.Core
}

func TestCoreTestSuite(t *testing.T) {
	suite.Run(t, new(CoreTestSuite))
}

func (s *CoreTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.core = health.NewCore()
}

func (s *CoreTestSuite) TestNewCoreStartsNotReady() {
	s.False(s.core.IsReady(s.ctx))
}

func (s *CoreTestSuite) TestSetReadyTrueMakesCoreReady() {
	s.core.SetReady(true)

	s.True(s.core.IsReady(s.ctx))
}

func (s *CoreTestSuite) TestSetReadyFalseDrainsCore() {
	s.core.SetReady(true)

	s.core.SetReady(false)

	s.False(s.core.IsReady(s.ctx))
}

func (s *CoreTestSuite) TestErrNotReadyHasStableText() {
	s.EqualError(health.ErrNotReady, testNotReadyText)
}
