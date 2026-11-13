package operations_test

import (
	"math"
	"testing"
)

// Native fuzz targets. Plain `go test` runs only the seed corpus below, so CI
// stays fast and deterministic. To search for new failures run, for example:
//
//	go test ./internal/modules/calculator/operations -run '^$' -fuzz FuzzDivide -fuzztime 20s
//
// Each target compares the service with the math/big reference in reference_test.go.

func FuzzDivide(f *testing.F) {
	f.Add(int64(1), uint8(0), int64(3), uint8(0))
	f.Add(int64(125), uint8(3), int64(1), uint8(0))
	f.Add(int64(-1), uint8(0), int64(8), uint8(0))
	f.Add(int64(1), uint8(20), int64(2), uint8(0))
	f.Add(int64(math.MaxInt64), uint8(20), int64(-7), uint8(13))
	f.Fuzz(func(t *testing.T, a int64, aScale uint8, b int64, bScale uint8) {
		checkDivide(t, a, int(aScale), b, int(bScale))
	})
}

func FuzzMultiply(f *testing.F) {
	f.Add(int64(1), uint8(1), int64(3), uint8(0))
	f.Add(int64(-25), uint8(2), int64(4), uint8(1))
	f.Add(int64(math.MaxInt64), uint8(20), int64(math.MinInt64), uint8(20))
	f.Fuzz(func(t *testing.T, a int64, aScale uint8, b int64, bScale uint8) {
		checkMultiply(t, a, int(aScale), b, int(bScale))
	})
}

func FuzzSqrt(f *testing.F) {
	f.Add(int64(144), uint8(0))
	f.Add(int64(2), uint8(0))
	f.Add(int64(1), uint8(20))
	f.Add(int64(math.MaxInt64), uint8(0))
	f.Fuzz(func(t *testing.T, a int64, aScale uint8) {
		checkSqrt(t, a, int(aScale))
	})
}

func FuzzPower(f *testing.F) {
	f.Add(int64(2), uint8(0), int8(10))
	f.Add(int64(15), uint8(1), int8(-3))
	f.Add(int64(-3), uint8(0), int8(7))
	f.Add(int64(math.MaxInt64), uint8(20), int8(20))
	f.Fuzz(func(t *testing.T, a int64, aScale uint8, exponent int8) {
		checkPower(t, a, int(aScale), int(exponent))
	})
}
