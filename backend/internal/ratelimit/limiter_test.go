package ratelimit_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/ratelimit"
)

const (
	testRate        = 2.0
	testBurst       = 3
	testClientA     = "203.0.113.7"
	testClientB     = "203.0.113.8"
	testHalfSecond  = 500 * time.Millisecond
	testOneSecond   = time.Second
	testPastRefill  = 3 * time.Second
	testStart       = "2026-01-01T00:00:00Z"
	testTrackedNone = 0
	testTrackedOne  = 1
)

type LimiterTestSuite struct {
	suite.Suite
	now     time.Time
	limiter *ratelimit.Limiter
}

func TestLimiterTestSuite(t *testing.T) {
	suite.Run(t, new(LimiterTestSuite))
}

func (s *LimiterTestSuite) SetupTest() {
	start, err := time.Parse(time.RFC3339, testStart)
	s.Require().NoError(err)
	s.now = start
	s.limiter = ratelimit.New(testRate, testBurst, func() time.Time { return s.now })
}

func (s *LimiterTestSuite) spendBurst(key string) {
	for range testBurst {
		allowed, _ := s.limiter.Allow(key)
		s.Require().True(allowed)
	}
}

func (s *LimiterTestSuite) TestAllowWhenWithinBurstReturnsTrue() {
	s.spendBurst(testClientA)
}

func (s *LimiterTestSuite) TestAllowWhenBurstIsSpentRejectsWithWaitForNextToken() {
	s.spendBurst(testClientA)

	allowed, wait := s.limiter.Allow(testClientA)

	s.False(allowed)
	s.Equal(testHalfSecond, wait)
}

func (s *LimiterTestSuite) TestAllowWhenTimePassesRefillsTokens() {
	s.spendBurst(testClientA)
	s.now = s.now.Add(testOneSecond)

	first, _ := s.limiter.Allow(testClientA)
	second, _ := s.limiter.Allow(testClientA)
	third, _ := s.limiter.Allow(testClientA)

	s.True(first)
	s.True(second)
	s.False(third)
}

func (s *LimiterTestSuite) TestAllowWhenIdleLongNeverExceedsBurst() {
	s.spendBurst(testClientA)
	s.now = s.now.Add(testPastRefill * 10)

	s.spendBurst(testClientA)
	allowed, _ := s.limiter.Allow(testClientA)

	s.False(allowed)
}

func (s *LimiterTestSuite) TestAllowWhenOtherClientIsLimitedStaysIndependent() {
	s.spendBurst(testClientA)

	s.spendBurst(testClientB)

	allowed, _ := s.limiter.Allow(testClientA)
	s.False(allowed)
}

func (s *LimiterTestSuite) TestAllowWhenBucketsRefilledSweepsIdleKeys() {
	s.spendBurst(testClientA)
	s.Equal(testTrackedOne, s.limiter.Tracked())
	s.now = s.now.Add(testPastRefill)

	allowed, _ := s.limiter.Allow(testClientB)

	s.True(allowed)
	s.Equal(testTrackedOne, s.limiter.Tracked())
}

func (s *LimiterTestSuite) TestAllowWhenBucketNotYetFullKeepsKey() {
	s.spendBurst(testClientA)
	s.now = s.now.Add(testOneSecond)

	_, _ = s.limiter.Allow(testClientB)

	s.Equal(testTrackedOne+1, s.limiter.Tracked())
}

func (s *LimiterTestSuite) TestTrackedWhenNothingSeenReturnsZero() {
	s.Equal(testTrackedNone, s.limiter.Tracked())
}
