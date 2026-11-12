package operations

import (
	"context"
	"math"

	"github.com/shopspring/decimal"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

// roundingFactor is 2 in 2*|remainder| versus |divisor| (the half-way test).
const roundingFactor = 2

var two = decimal.NewFromInt(roundingFactor)

// Divide divides the first operand by the second.
//
// A quotient that does not terminate (1/3) is rounded to the configured number
// of decimal places using half-even rounding, also called banker's rounding:
// an exact tie goes to the nearest even digit, so 0.125 becomes 0.12 and 0.135
// becomes 0.14. This is the standard for financial rounding because ties round
// up half the time and down half the time, so totals are not biased in one
// direction the way always-round-half-up is.
type Divide struct {
	descriptor
	precision int32
}

var _ Operation = (*Divide)(nil)

// NewDivide returns the divide operation rounding to precision decimal places.
func NewDivide(precision int) *Divide {
	return &Divide{
		descriptor: descriptor{
			name:        entities.OpDivide,
			arity:       entities.ArityBinary,
			symbol:      entities.SymbolDivide,
			description: entities.DescriptionDivide,
		},
		precision: scaleOf(precision),
	}
}

// Apply returns operands[0] / operands[1] or DIVISION_BY_ZERO.
func (d *Divide) Apply(_ context.Context, operands []decimal.Decimal) (decimal.Decimal, apperror.IError) {
	dividend, divisor := operands[0], operands[1]
	if divisor.IsZero() {
		return decimal.Zero, apperror.New(apperror.CodeDivisionByZero, ErrDivisionByZero)
	}
	return divideHalfEven(dividend, divisor, d.precision), nil
}

// divideHalfEven rounds the exact quotient to precision places, ties to even.
//
// decimal's own DivRound is half-up, and rounding an already rounded number
// (divide to extra digits, then round) can turn a non-tie into a false tie.
// QuoRem avoids both: it gives the truncated quotient and the exact remainder,
// so the half-way test compares 2*|remainder| with |divisor| in whole numbers.
func divideHalfEven(dividend, divisor decimal.Decimal, precision int32) decimal.Decimal {
	quotient, remainder := dividend.QuoRem(divisor, precision)
	if remainder.IsZero() {
		return quotient
	}
	unit := decimal.New(1, -precision)
	half := remainder.Abs().Mul(two).Cmp(divisor.Abs().Mul(unit))
	if half < 0 || (half == 0 && isEven(quotient, precision)) {
		return quotient
	}
	if dividend.Sign()*divisor.Sign() < 0 {
		unit = unit.Neg()
	}
	return quotient.Add(unit)
}

// isEven reports whether the last kept digit of q (at precision places) is even.
func isEven(q decimal.Decimal, precision int32) bool {
	return q.Shift(precision).Mod(two).IsZero()
}

// scaleOf converts the configured precision to decimal's int32 scale. Config
// validation caps precision far below MaxInt32; the clamp keeps the conversion
// safe for any other caller.
func scaleOf(precision int) int32 {
	if precision > math.MaxInt32 {
		return math.MaxInt32
	}
	if precision < 0 {
		return 0
	}
	return int32(precision)
}
