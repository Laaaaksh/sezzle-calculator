package operations_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/operations"
)

// The reference side of the correctness checks. It computes the exact answer
// with math/big rationals, which have no rounding of their own, and applies
// half-even rounding to refPlaces by hand. The code under test uses
// shopspring/decimal, so the two sides share no arithmetic.
const (
	refPlaces        = 20
	refMaxScale      = 20
	refMaxExponent   = 20
	refSqrtScaleSum  = 2 * refPlaces
	refRadix         = 10
	refBudgetOperand = 64
	refMaxPowerCap   = 1000
)

var bigTen = big.NewInt(refRadix)

func pow10(n int) *big.Int {
	return new(big.Int).Exp(bigTen, big.NewInt(int64(n)), nil)
}

func operandOf(coefficient int64, scale int) (decimal.Decimal, *big.Rat) {
	places := scale % (refMaxScale + 1)
	exact := new(big.Rat).SetFrac(big.NewInt(coefficient), pow10(places))
	return decimal.New(coefficient, -int32(places)), exact
}

// ratOfResult reads a result back as an exact rational through its text, so
// a formatting slip in the result shows up as a mismatch.
func ratOfResult(t testing.TB, result decimal.Decimal) *big.Rat {
	t.Helper()
	exact, ok := new(big.Rat).SetString(result.String())
	require.True(t, ok, "result %q is not a number", result.String())
	return exact
}

func roundHalfEven(r *big.Rat, places int) *big.Rat {
	factor := pow10(places)
	scaled := new(big.Rat).Mul(r, new(big.Rat).SetInt(factor))
	quotient, remainder := new(big.Int).QuoRem(scaled.Num(), scaled.Denom(), new(big.Int))
	twiceRemainder := new(big.Int).Lsh(new(big.Int).Abs(remainder), 1)
	order := twiceRemainder.Cmp(scaled.Denom())
	if order > 0 || (order == 0 && quotient.Bit(0) == 1) {
		quotient.Add(quotient, big.NewInt(int64(scaled.Sign())))
	}
	return new(big.Rat).SetFrac(quotient, factor)
}

func referenceDivide(dividend, divisor *big.Rat) *big.Rat {
	return roundHalfEven(new(big.Rat).Quo(dividend, divisor), refPlaces)
}

func referenceMultiply(left, right *big.Rat) *big.Rat {
	return new(big.Rat).Mul(left, right)
}

// referenceSqrt returns sqrt(coefficient * 10^-scale) rounded half-even to refPlaces.
// With at most refMaxScale input places, radicand = x * 10^(2*refPlaces) is a
// whole number, so a tie (sqrt = m + 1/2) cannot occur and the rounding
// direction is decided by comparing radicand with m*m+m.
func referenceSqrt(coefficient int64, scale int) *big.Rat {
	places := scale % (refMaxScale + 1)
	radicand := new(big.Int).Mul(big.NewInt(coefficient), pow10(refSqrtScaleSum-places))
	root := new(big.Int).Sqrt(radicand)
	rootSquared := new(big.Int).Mul(root, root)
	if new(big.Int).Sub(radicand, rootSquared).Cmp(root) > 0 {
		root.Add(root, big.NewInt(1))
	}
	return new(big.Rat).SetFrac(root, pow10(refPlaces))
}

// referencePower raises base to a whole exponent. A positive exponent is exact;
// a negative one is 1 / base^n rounded to refPlaces, as the service does.
func referencePower(base *big.Rat, exponent int) *big.Rat {
	count := exponent
	if count < 0 {
		count = -count
	}
	product := big.NewRat(1, 1)
	for range count {
		product.Mul(product, base)
	}
	if exponent >= 0 {
		return product
	}
	return roundHalfEven(new(big.Rat).Inv(product), refPlaces)
}

func newPower() *operations.Power {
	return operations.NewPower(entities.Limits{
		MaxOperandLength: refBudgetOperand,
		MaxExponent:      refMaxPowerCap,
		DecimalPrecision: refPlaces,
		MaxResultDigits:  entities.ResultDigitBudget(refBudgetOperand),
	})
}

func requireSameValue(t testing.TB, want, got *big.Rat, label string) {
	t.Helper()
	require.Zero(t, want.Cmp(got), "%s: want %s, got %s", label, want.FloatString(refPlaces+1), got.FloatString(refPlaces+1))
}

func checkDivide(t testing.TB, dividend int64, dividendScale int, divisor int64, divisorScale int) {
	t.Helper()
	left, leftExact := operandOf(dividend, dividendScale)
	right, rightExact := operandOf(divisor, divisorScale)
	if right.IsZero() {
		return
	}
	got, aerr := operations.NewDivide(refPlaces).Apply(context.Background(), []decimal.Decimal{left, right})
	require.Nil(t, aerr)
	requireSameValue(t, referenceDivide(leftExact, rightExact), ratOfResult(t, got), "divide "+left.String()+" by "+right.String())
}

func checkMultiply(t testing.TB, left int64, leftScale int, right int64, rightScale int) {
	t.Helper()
	a, aExact := operandOf(left, leftScale)
	b, bExact := operandOf(right, rightScale)
	got, aerr := operations.NewMultiply().Apply(context.Background(), []decimal.Decimal{a, b})
	require.Nil(t, aerr)
	requireSameValue(t, referenceMultiply(aExact, bExact), ratOfResult(t, got), "multiply "+a.String()+" by "+b.String())
}

func checkSqrt(t testing.TB, coefficient int64, scale int) {
	t.Helper()
	if coefficient < 0 {
		return
	}
	operand, _ := operandOf(coefficient, scale)
	got, aerr := operations.NewSqrt(refPlaces).Apply(context.Background(), []decimal.Decimal{operand})
	require.Nil(t, aerr)
	requireSameValue(t, referenceSqrt(coefficient, scale), ratOfResult(t, got), "sqrt "+operand.String())
}

// checkPower folds the exponent into -refMaxExponent..refMaxExponent, which
// keeps every result inside the default digit budget, so an error here is a bug.
func checkPower(t testing.TB, coefficient int64, scale int, exponent int) {
	t.Helper()
	whole := exponent % (refMaxExponent + 1)
	base, baseExact := operandOf(coefficient, scale)
	if base.IsZero() && whole < 0 {
		return
	}
	got, aerr := newPower().Apply(context.Background(), []decimal.Decimal{base, decimal.NewFromInt(int64(whole))})
	require.Nil(t, aerr)
	requireSameValue(t, referencePower(baseExact, whole), ratOfResult(t, got), "power "+base.String())
}
