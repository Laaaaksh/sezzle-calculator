package operations

import (
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
// Cost guard: base^e has at most e times as many significant digits as the
// base, so that product is checked against a digit budget before any multiply
// happens. A request like 99999999999999999999^1000 fails fast with
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
// budget. Trailing zeros are ignored (10 counts as 1 digit) because decimal
// stores them in the exponent, so they cost no memory. The estimate is an
// upper bound, so it can refuse a result that would have fitted, never the
// reverse.
func (p *Power) checkResultSize(base decimal.Decimal, exponent int64) apperror.IError {
	if significantDigits(base)*exponent <= p.maxResultDigits {
		return nil
	}
	return apperror.New(apperror.CodeResultTooLarge, ErrResultBeyondBudget).
		WithField(entities.FieldLimit, p.maxResultDigits)
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

// significantDigits counts the coefficient digits of d without trailing zeros.
func significantDigits(d decimal.Decimal) int64 {
	digits := strings.TrimRight(d.Abs().Coefficient().String(), "0")
	return int64(len(digits))
}

func abs64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
