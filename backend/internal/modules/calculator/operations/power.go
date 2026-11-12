package operations

import (
	"context"
	"math/big"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

const (
	// exponentOperandIndex is where the exponent sits in the operand list; it
	// is reported in INVALID_EXPONENT so a client knows which operand failed.
	exponentOperandIndex = 1
	// parityDivisor is 2: an exponent is even when it divides by 2 without remainder.
	parityDivisor = 2
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
// Size guard: the result may hold at most the shared digit budget of digits,
// counting integer and fraction digits (10^1000 has 1001, 0.1^1000 has 1000).
// The checks use whole-number arithmetic only, and run before the multiplication
// they would make expensive:
//   - the digits the shift of the decimal point adds (n*x for |base| = c*10^x)
//     are known without multiplying, so 10^1000 and 0.1^1000 fail at once;
//   - while c^n is built, each product of two numbers a and b has at least
//     digits(a)+digits(b)-1 digits, so a product that must exceed the budget is
//     refused before it is computed, and no multiplication ever sees an operand
//     longer than the budget;
//   - the finished result is counted exactly by the shared CheckResultSize.
//
// 2^1000 (302 digits) fits the default budget of 640.
type Power struct {
	descriptor
	precision       int32
	maxExponent     decimal.Decimal
	maxResultDigits int64
}

var _ Operation = (*Power)(nil)

// NewPower returns the power operation bounded by limits: the exponent range,
// the rounding precision and the result digit budget.
func NewPower(limits entities.Limits) *Power {
	return &Power{
		descriptor: descriptor{
			name:        entities.OpPower,
			arity:       entities.ArityBinary,
			symbol:      entities.SymbolPower,
			description: entities.DescriptionPower,
		},
		precision:       scaleOf(limits.DecimalPrecision),
		maxExponent:     decimal.NewFromInt(int64(limits.MaxExponent)),
		maxResultDigits: int64(limits.MaxResultDigits),
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
	return p.compute(ctx, stripTrailingZeros(base), exponent)
}

// compute raises base, which has no trailing zeros in its coefficient and is
// neither 0 nor +-1, to a non-zero whole exponent inside the digit budget.
func (p *Power) compute(ctx context.Context, base decimal.Decimal, exponent int64) (decimal.Decimal, apperror.IError) {
	n := abs64(exponent)
	if aerr := p.checkShift(base, n, exponent < 0); aerr != nil {
		return decimal.Zero, aerr
	}
	result, aerr := p.boundedPower(ctx, base, n)
	if aerr != nil {
		return decimal.Zero, aerr
	}
	if exponent < 0 {
		if result, aerr = p.reciprocal(result); aerr != nil {
			return decimal.Zero, aerr
		}
	}
	if aerr = CheckResultSize(result, p.maxResultDigits); aerr != nil {
		return decimal.Zero, aerr
	}
	return result, nil
}

// wholeExponent returns the exponent as an int64 after rejecting fractions and
// values beyond the configured range. 2.0 is accepted: it equals 2 exactly.
func (p *Power) wholeExponent(exponent decimal.Decimal) (int64, apperror.IError) {
	if !exponent.Equal(exponent.Truncate(0)) {
		return 0, apperror.New(entities.CodeInvalidExponent, ErrFractionalExponent).
			WithField(entities.FieldOperandIndex, exponentOperandIndex)
	}
	if exponent.Abs().GreaterThan(p.maxExponent) {
		return 0, apperror.New(entities.CodeExponentTooLarge, ErrExponentBeyondLimit).
			WithField(entities.FieldOperandIndex, exponentOperandIndex).
			WithField(entities.FieldLimit, p.maxExponent.IntPart())
	}
	return exponent.IntPart(), nil
}

// checkShift rejects a power whose result must exceed the digit budget just
// from where the decimal point lands. base has no trailing zeros in its
// coefficient, so |base| = c * 10^x and base^n = c^n * 10^(n*x). With shift =
// n*x and c^n at least one digit long:
//   - shift >= 0: the zeros extend the integer part, so there are at least
//     shift+1 digits (10^1000 has 1001);
//   - shift < 0: the result has -shift fraction digits (0.1^1000 has 1000).
//
// A negative exponent divides 1 by that number and rounds to the precision, so
// its size is only known after the division. reciprocal bounds the integer part
// before dividing and CheckResultSize counts the whole result afterwards.
func (p *Power) checkShift(base decimal.Decimal, n int64, negativeExponent bool) apperror.IError {
	if negativeExponent {
		return nil
	}
	shift := n * int64(base.Exponent())
	minimum := -shift
	if shift >= 0 {
		minimum = shift + 1
	}
	if minimum > p.maxResultDigits {
		return ResultTooLarge(p.maxResultDigits)
	}
	return nil
}

// reciprocal returns 1 / divisor rounded half-even to the precision. divisor is
// c * 10^shift with c < 10^digits(c), so the quotient exceeds 10^(-shift-digits)
// and has at least -shift-digits+1 integer digits. If that lower bound is over
// the budget the division is skipped, so a tiny base such as 0.1 never has its
// huge reciprocal built just to be refused.
func (p *Power) reciprocal(divisor decimal.Decimal) (decimal.Decimal, apperror.IError) {
	integerDigits := -int64(divisor.Exponent()) - coefficientDigits(divisor) + 1
	if integerDigits > p.maxResultDigits {
		return decimal.Zero, ResultTooLarge(p.maxResultDigits)
	}
	return divideHalfEven(one, divisor, p.precision), nil
}

// checkProduct refuses a product of a and b before it is computed. A product
// has at least digits(a)+digits(b)-1 digits, so if even that lower bound is over
// the budget the final power certainly is: every later step only multiplies
// this number in again. Because each factor passed the same check earlier, no
// multiplication ever starts from an operand longer than the budget, which
// keeps the work bounded.
func (p *Power) checkProduct(a, b decimal.Decimal) apperror.IError {
	if coefficientDigits(a)+coefficientDigits(b)-1 > p.maxResultDigits {
		return ResultTooLarge(p.maxResultDigits)
	}
	return nil
}

// boundedPower is exponentiation by squaring, O(log n) exact multiplications,
// with checkProduct run before each one. The final squaring is skipped so the
// base is never squared past what the result needs. It returns base^n for n > 0
// and RESULT_TOO_LARGE as soon as the coefficient cannot fit the budget; this
// applies to a negative exponent too, whose divisor is this same number.
func (p *Power) boundedPower(ctx context.Context, base decimal.Decimal, n int64) (decimal.Decimal, apperror.IError) {
	result := one
	for n > 0 {
		if aerr := CheckContext(ctx); aerr != nil {
			return decimal.Zero, aerr
		}
		if n&1 == 1 {
			if aerr := p.checkProduct(result, base); aerr != nil {
				return decimal.Zero, aerr
			}
			result = result.Mul(base)
		}
		n >>= 1
		if n > 0 {
			if aerr := p.checkProduct(base, base); aerr != nil {
				return decimal.Zero, aerr
			}
			base = base.Mul(base)
		}
	}
	return result, nil
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

// zeroToPower handles a zero base with a non-zero exponent: 0^n is 0 and
// 0^-n is a division by zero.
func zeroToPower(exponent int64) (decimal.Decimal, apperror.IError) {
	if exponent < 0 {
		return decimal.Zero, apperror.New(entities.CodeDivisionByZero, ErrZeroToNegativePower)
	}
	return decimal.Zero, nil
}

// unitToPower handles base 1 and -1 directly: the answer is the base or 1 by
// parity. Without this, 1^1000 would be refused by the digit estimate even
// though its result has one digit.
func unitToPower(base decimal.Decimal, exponent int64) decimal.Decimal {
	if base.Sign() > 0 || exponent%parityDivisor == 0 {
		return one
	}
	return one.Neg()
}

func abs64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
