package operations

import (
	"context"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/config"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

const (
	// resultDigitsFactor sets the digit budget of a power result as a multiple
	// of the configured operand length: a result may be ten times longer than
	// any operand the service accepts, and no longer.
	resultDigitsFactor = 10
	// exponentOperandIndex is where the exponent sits in the operand list; it
	// is reported in INVALID_EXPONENT so a client knows which operand failed.
	exponentOperandIndex = 1
	// digitMargin is added to the digit bound so float rounding in log10 can
	// never make it an under-estimate.
	digitMargin = 1
	// leadingDigits is how many leading digits of the coefficient are read
	// as a float64 (it holds 15 decimal digits exactly).
	leadingDigits = 15
)

var one = decimal.NewFromInt(1)

// Power raises the first operand (the base) to the second (the exponent).
//
// Only whole-number exponents are accepted. A fractional exponent such as 0.5
// has no exact decimal answer in general (2^0.5 is irrational), and computing
// it would need a float or an approximation the caller never asked for, so it
// fails with INVALID_EXPONENT instead of returning a quietly inexact number.
//
// The exponent must lie between -MaxExponent and +MaxExponent (EXPONENT_TOO_LARGE
// otherwise). Anything to the power 0 is 1, including 0^0. Zero to a negative
// power is a division by zero. A negative exponent is 1 / base^|exponent| using
// the same half-even division as Divide, at the configured precision.
//
// Size guard: the result may hold at most a digit budget of digits, counting
// integer and fraction digits (10^1000 has 1001, 0.1^1000 has 1000). Before any
// multiply, an upper bound on those digits is checked against the budget, so a
// request like 9^1000, 10^1000 or 0.1^-1000 fails fast with RESULT_TOO_LARGE
// instead of allocating a number with thousands of digits. 2^1000 (302 digits)
// fits the default budget of 640. The computed result is then counted exactly
// and refused if the estimate let it slip past the budget.
type Power struct {
	descriptor
	precision       int32
	maxExponent     decimal.Decimal
	maxResultDigits int64
}

var _ Operation = (*Power)(nil)

// NewPower returns the power operation bounded by cfg.
func NewPower(cfg config.CalculatorConfig) *Power {
	return &Power{
		descriptor: descriptor{
			name:        entities.OpPower,
			arity:       entities.ArityBinary,
			symbol:      entities.SymbolPower,
			description: entities.DescriptionPower,
		},
		precision:       scaleOf(cfg.DecimalPrecision),
		maxExponent:     decimal.NewFromInt(int64(cfg.MaxExponent)),
		maxResultDigits: int64(cfg.MaxOperandLength) * resultDigitsFactor,
	}
}

// Apply returns operands[0] ^ operands[1]. The exponent is validated before
// the base so a bad exponent is reported even for a zero base.
func (p *Power) Apply(ctx context.Context, operands []decimal.Decimal) (decimal.Decimal, apperror.IError) {
	base := operands[0]
	exponent, aerr := p.wholeExponent(operands[1])
	if aerr != nil {
		return decimal.Zero, aerr
	}
	if exponent == 0 {
		return one, nil
	}
	if base.IsZero() {
		return zeroToPower(exponent)
	}
	if base.Abs().Equal(one) {
		return unitToPower(base, exponent), nil
	}
	base = stripTrailingZeros(base)
	if aerr = p.checkResultSize(base, exponent); aerr != nil {
		return decimal.Zero, aerr
	}
	result, aerr := integerPower(ctx, base, abs64(exponent))
	if aerr != nil {
		return decimal.Zero, aerr
	}
	if exponent < 0 {
		result = divideHalfEven(one, result, p.precision)
	}
	if digitsOf(result) > p.maxResultDigits {
		return decimal.Zero, p.tooLarge()
	}
	return result, nil
}

// wholeExponent returns the exponent as an int64 after rejecting fractions and
// values beyond the configured range. 2.0 is accepted: it equals 2 exactly.
func (p *Power) wholeExponent(exponent decimal.Decimal) (int64, apperror.IError) {
	if !exponent.Equal(exponent.Truncate(0)) {
		return 0, apperror.New(apperror.CodeInvalidExponent, ErrFractionalExponent).
			WithField(entities.FieldOperandIndex, exponentOperandIndex)
	}
	if exponent.Abs().GreaterThan(p.maxExponent) {
		return 0, apperror.New(apperror.CodeExponentTooLarge, ErrExponentBeyondLimit).
			WithField(entities.FieldOperandIndex, exponentOperandIndex).
			WithField(entities.FieldLimit, p.maxExponent.IntPart())
	}
	return exponent.IntPart(), nil
}

// checkResultSize rejects a power whose result could exceed the digit budget.
// base must have no trailing zeros in its coefficient, so |base| = c * 10^x
// with c not divisible by 10 and c^n has no trailing zeros either.
//
// c^n has at most floor(n * log10(c)) + 1 digits. That coefficient is built for
// both signs of the exponent (as the divisor when it is negative), so it is
// bounded first. Then the result itself:
//   - base^n is c^n * 10^(n*x): with n*x >= 0 the zeros extend the integer
//     part, with n*x < 0 the result has -n*x fraction digits.
//   - base^-n is 1 / base^n rounded to the precision: its integer part has
//     floor(-n * log10|base|) + 1 digits when |base| < 1. Its fraction digits
//     are bounded by the precision and counted exactly after the division.
//
// One extra digit of margin covers float rounding, so each bound can only
// over-estimate. The log10 is a float64 on purpose: it only sizes this guard
// and never produces part of a result.
func (p *Power) checkResultSize(base decimal.Decimal, exponent int64) apperror.IError {
	n := abs64(exponent)
	nLogC := float64(n) * log10Coefficient(base)
	coefficientDigits := int64(math.Floor(nLogC)) + 1 + digitMargin
	if coefficientDigits > p.maxResultDigits {
		return p.tooLarge()
	}
	shift := n * int64(base.Exponent())
	var estimate int64
	switch {
	case exponent < 0:
		estimate = max(int64(math.Floor(-nLogC-float64(shift)))+1+digitMargin, 0)
	case shift >= 0:
		estimate = coefficientDigits + shift
	default:
		estimate = max(coefficientDigits, -shift)
	}
	if estimate > p.maxResultDigits {
		return p.tooLarge()
	}
	return nil
}

func (p *Power) tooLarge() apperror.IError {
	return apperror.New(apperror.CodeResultTooLarge, ErrResultBeyondBudget).
		WithField(entities.FieldLimit, p.maxResultDigits)
}

// log10Coefficient is log10 of the coefficient of |d|. Only the leading
// leadingDigits digits go through float64, so a coefficient far beyond
// float64's range cannot overflow.
func log10Coefficient(d decimal.Decimal) float64 {
	digits := d.Abs().Coefficient().String()
	head := digits
	if len(head) > leadingDigits {
		head = head[:leadingDigits]
	}
	leading, _ := strconv.ParseFloat(head, 64)
	return math.Log10(leading) + float64(len(digits)-len(head))
}

// stripTrailingZeros moves the trailing zeros of the coefficient into the
// exponent. decimal keeps "10" as coefficient 10, exponent 0; this returns
// coefficient 1, exponent 1, the same value. d must not be zero.
func stripTrailingZeros(d decimal.Decimal) decimal.Decimal {
	digits := d.Coefficient().String()
	// Count zeros as int32 directly so no int -> int32 conversion is needed.
	var zeros int32
	for strings.HasSuffix(digits, "0") {
		digits = digits[:len(digits)-1]
		zeros++
	}
	coefficient, _ := new(big.Int).SetString(digits, 10)
	return decimal.NewFromBigInt(coefficient, d.Exponent()+zeros)
}

// digitsOf counts the digits of d as written without trailing fraction zeros
// or a lone leading zero: integer digits plus fraction digits, so 1000 has 4,
// 0.001 has 3 and 12.5 has 3. Zero has 1.
func digitsOf(d decimal.Decimal) int64 {
	if d.IsZero() {
		return 1
	}
	d = stripTrailingZeros(d)
	coefficientDigits := int64(len(d.Abs().Coefficient().String()))
	exp := int64(d.Exponent())
	if exp >= 0 {
		return coefficientDigits + exp
	}
	return max(coefficientDigits, -exp)
}

// zeroToPower handles a zero base with a non-zero exponent: 0^n is 0 and
// 0^-n is a division by zero.
func zeroToPower(exponent int64) (decimal.Decimal, apperror.IError) {
	if exponent < 0 {
		return decimal.Zero, apperror.New(apperror.CodeDivisionByZero, ErrZeroToNegativePower)
	}
	return decimal.Zero, nil
}

// unitToPower handles base 1 and -1 directly: the answer is the base or 1 by
// parity. Without this, 1^1000 would be refused by the digit estimate even
// though its result has one digit.
func unitToPower(base decimal.Decimal, exponent int64) decimal.Decimal {
	if base.Sign() > 0 || exponent%2 == 0 {
		return one
	}
	return one.Neg()
}

// integerPower is exponentiation by squaring: O(log e) exact multiplications.
// The final squaring is skipped so the base is never squared past what the
// result needs.
func integerPower(ctx context.Context, base decimal.Decimal, exponent int64) (decimal.Decimal, apperror.IError) {
	result := one
	for exponent > 0 {
		if aerr := CheckContext(ctx); aerr != nil {
			return decimal.Zero, aerr
		}
		if exponent&1 == 1 {
			result = result.Mul(base)
		}
		exponent >>= 1
		if exponent > 0 {
			base = base.Mul(base)
		}
	}
	return result, nil
}

func abs64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
