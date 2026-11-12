package operations

import (
	"math"
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
// Cost guard: base^e has at most floor(e*log10(base))+1 digits, so that bound
// is checked against a digit budget before any multiply happens (2^1000 has 302
// digits and fits the default budget of 640). A request like 9^1000 or
// 99999999999999999999^1000 fails fast with
// RESULT_TOO_LARGE instead of allocating a number with thousands of digits.
type Power struct {
	precision       int32
	maxExponent     decimal.Decimal
	maxResultDigits int64
}

var _ Operation = (*Power)(nil)

// NewPower returns the power operation bounded by cfg.
func NewPower(cfg config.CalculatorConfig) *Power {
	return &Power{
		precision:       scaleOf(cfg.DecimalPrecision),
		maxExponent:     decimal.NewFromInt(int64(cfg.MaxExponent)),
		maxResultDigits: int64(cfg.MaxOperandLength) * resultDigitsFactor,
	}
}

// Name returns the registry key.
func (*Power) Name() entities.OperationType { return entities.OpPower }

// Arity is two: base and exponent.
func (*Power) Arity() int { return entities.ArityBinary }

// Apply returns operands[0] ^ operands[1]. The exponent is validated before
// the base so a bad exponent is reported even for a zero base.
func (p *Power) Apply(operands []decimal.Decimal) (decimal.Decimal, apperror.IError) {
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
	if aerr = p.checkResultSize(base, abs64(exponent)); aerr != nil {
		return decimal.Zero, aerr
	}
	magnitude := integerPower(base, abs64(exponent))
	if exponent < 0 {
		return divideHalfEven(one, magnitude, p.precision), nil
	}
	return magnitude, nil
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

// checkResultSize rejects a power whose coefficient could exceed the digit
// budget. base^e has at most floor(e * log10(c)) + 1 digits, where c is the
// base coefficient without trailing zeros (decimal keeps those, and the scale,
// in the exponent, so they add no coefficient digits). One extra digit of
// margin covers float rounding, so the bound can only over-estimate.
//
// The log10 is a float64 on purpose: it only sizes this guard and never
// produces part of a result, so a float error cannot change any answer, and
// the margin keeps it from letting an oversized result through.
func (p *Power) checkResultSize(base decimal.Decimal, exponent int64) apperror.IError {
	bound := int64(math.Floor(float64(exponent)*log10Coefficient(base))) + 1 + digitMargin
	if bound <= p.maxResultDigits {
		return nil
	}
	return apperror.New(apperror.CodeResultTooLarge, ErrResultBeyondBudget).
		WithField(entities.FieldLimit, p.maxResultDigits)
}

// log10Coefficient is log10 of |base| with trailing zeros of the coefficient
// removed. Only the leading leadingDigits digits go through float64, so a
// coefficient far beyond float64's range cannot overflow.
func log10Coefficient(d decimal.Decimal) float64 {
	digits := strings.TrimRight(d.Abs().Coefficient().String(), "0")
	head := digits
	if len(head) > leadingDigits {
		head = head[:leadingDigits]
	}
	leading, _ := strconv.ParseFloat(head, 64)
	return math.Log10(leading) + float64(len(digits)-len(head))
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
func integerPower(base decimal.Decimal, exponent int64) decimal.Decimal {
	result := one
	for exponent > 0 {
		if exponent&1 == 1 {
			result = result.Mul(base)
		}
		exponent >>= 1
		if exponent > 0 {
			base = base.Mul(base)
		}
	}
	return result
}

func abs64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
