package contextkeys_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/contextkeys"
)

const (
	testRequestID  = "req-123"
	testStringKey  = "request_id"
	testOtherValue = "plain-string-key-value"
)

type ContextKeysTestSuite struct {
	suite.Suite
	ctx context.Context
}

func TestContextKeysTestSuite(t *testing.T) {
	suite.Run(t, new(ContextKeysTestSuite))
}

func (s *ContextKeysTestSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *ContextKeysTestSuite) TestRequestIDRoundTrip() {
	ctx := contextkeys.WithRequestID(s.ctx, testRequestID)

	id, ok := contextkeys.RequestID(ctx)

	s.True(ok)
	s.Equal(testRequestID, id)
}

func (s *ContextKeysTestSuite) TestRequestIDMissingReturnsFalse() {
	id, ok := contextkeys.RequestID(s.ctx)

	s.False(ok)
	s.Empty(id)
}

func (s *ContextKeysTestSuite) TestRequestIDIgnoresPlainStringKey() {
	//nolint:staticcheck // proves a string key cannot collide with the typed key
	ctx := context.WithValue(s.ctx, testStringKey, testOtherValue)

	_, ok := contextkeys.RequestID(ctx)

	s.False(ok)
}
