package operations_test

import (
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	seededCases      = 400
	seededSeedHigh   = 20260101
	seededSeedLow    = 7
	seededScaleRange = 21
	seededExpRange   = 41
	seededExpShift   = 20
	seededSmallRange = 2000
	seededSmallShift = 1000
)

// ReferenceTestSuite runs the fuzz targets' checks on a fixed pseudo-random sample, so CI covers thousands of cases and a
// failure repeats. Operands mix full-range int64 values with small ones, where ties and exact quotients occur.
type ReferenceTestSuite struct {
	suite.Suite
	rng *rand.Rand
}

func TestReferenceTestSuite(t *testing.T) {
	suite.Run(t, new(ReferenceTestSuite))
}

func (s *ReferenceTestSuite) SetupTest() {
	// A fixed seed is the point: the same cases run every time, so a failure repeats.
	s.rng = rand.New(rand.NewPCG(seededSeedHigh, seededSeedLow)) //nolint:gosec // not security related
}

func (s *ReferenceTestSuite) coefficient() int64 {
	if s.rng.IntN(2) == 0 {
		return s.rng.Int64() - s.rng.Int64()
	}
	return int64(s.rng.IntN(seededSmallRange) - seededSmallShift)
}

func (s *ReferenceTestSuite) scale() int {
	return s.rng.IntN(seededScaleRange)
}

func (s *ReferenceTestSuite) TestDivideMatchesExactReference() {
	for range seededCases {
		checkDivide(s.T(), s.coefficient(), s.scale(), s.coefficient(), s.scale())
	}
}

func (s *ReferenceTestSuite) TestMultiplyMatchesExactReference() {
	for range seededCases {
		checkMultiply(s.T(), s.coefficient(), s.scale(), s.coefficient(), s.scale())
	}
}

func (s *ReferenceTestSuite) TestSqrtMatchesExactReference() {
	for range seededCases {
		checkSqrt(s.T(), s.coefficient(), s.scale())
	}
}

func (s *ReferenceTestSuite) TestPowerMatchesExactReference() {
	for range seededCases {
		exponent := s.rng.IntN(seededExpRange) - seededExpShift
		checkPower(s.T(), s.coefficient(), s.scale(), exponent)
	}
}
