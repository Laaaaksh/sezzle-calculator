package operations

import (
	"context"

	"github.com/shopspring/decimal"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

const (
	// sqrtGuardDigits are extra decimal places carried through Newton's method
	// so the final rounding to the configured precision sees the true tail.
	sqrtGuardDigits = 10
	// sqrtMaxIterations caps the loop. Newton's method needs about 20 steps for the widest inputs, so 200 is never reached
	// in practice and exists only so the loop cannot hang.
	sqrtMaxIterations = 200
	// halfUnitDigit is the digit 5 that, one place past the precision, marks a
	// rounding tie.
	halfUnitDigit = 5
	// rootDigitDivisor is 2: a square root has about half the digits of its
	// operand, which is how the starting guess is sized.
	rootDigitDivisor = 2
)

// Sqrt is the square root of one operand. Newton's method on decimals, g' = (g + x/g) / 2 with divisions truncated at the
// working scale, is the integer-square-root iteration on numbers scaled by 10^scale. It ends exactly at
// floor(sqrt(x) * 10^scale), so a perfect square is exact (sqrt(144) is 12, not 11.99999...). The root is carried with
// guard digits, then rounded half-even: a tie is real only when the root is exact, and one that merely looks like
// ...5000000000 within the guard digits is slightly above it and rounds up.
type Sqrt struct {
	descriptor
	precision     int32
	maxIterations int
}

var _ Operation = (*Sqrt)(nil)

// NewSqrt returns the square root operation rounding to precision places.
func NewSqrt(precision int) *Sqrt {
	return NewBoundedSqrt(precision, sqrtMaxIterations)
}

// NewBoundedSqrt is NewSqrt with an explicit iteration cap, so a test can
// prove that hitting the cap fails cleanly instead of hanging.
func NewBoundedSqrt(precision, maxIterations int) *Sqrt {
	return &Sqrt{
		descriptor: descriptor{
			name:        entities.OpSqrt,
			arity:       entities.ArityUnary,
			symbol:      entities.SymbolSqrt,
			description: entities.DescriptionSqrt,
		},
		precision:     scaleOf(precision),
		maxIterations: maxIterations,
	}
}

// Apply returns the square root of operands[0], NEGATIVE_SQUARE_ROOT for a negative operand, REQUEST_TIMEOUT if ctx ends
// mid-loop, or INTERNAL if the iteration cap is hit. The root is far shorter than the power result budget (see the
// sqrt result size test).
func (s *Sqrt) Apply(ctx context.Context, operands []decimal.Decimal) (decimal.Decimal, apperror.IError) {
	x := operands[0]
	if x.IsNegative() {
		return decimal.Zero, apperror.New(entities.CodeNegativeSquareRoot, ErrNegativeSquareRoot)
	}
	if x.IsZero() {
		return decimal.Zero, nil
	}
	root, aerr := s.newton(ctx, x)
	if aerr != nil {
		return decimal.Zero, aerr
	}
	return roundRoot(root, s.precision, root.Mul(root).Equal(x)), nil
}

// newton returns floor(sqrt(x)) at the working scale.
func (s *Sqrt) newton(ctx context.Context, x decimal.Decimal) (decimal.Decimal, apperror.IError) {
	scale := s.precision + sqrtGuardDigits
	guess := initialGuess(x, scale)
	for range s.maxIterations {
		if aerr := CheckContext(ctx); aerr != nil {
			return decimal.Zero, aerr
		}
		quotient, _ := x.QuoRem(guess, scale)
		next, _ := guess.Add(quotient).QuoRem(two, scale)
		if next.GreaterThanOrEqual(guess) {
			return guess, nil
		}
		guess = next
		if guess.IsZero() {
			return guess, nil
		}
	}
	return decimal.Zero, apperror.New(apperror.CodeInternal, ErrSqrtDidNotConverge)
}

// initialGuess is 10^ceil(n/2), where n is the number of digits before the
// decimal point of x (negative for x below 0.1). Since x < 10^n, the guess is
// at least sqrt(x) and within a factor of 10 of it, which is what bounds the
// iteration count. It is floored at one unit of the working scale so the first
// division is never by zero.
func initialGuess(x decimal.Decimal, scale int32) decimal.Decimal {
	digitsBeforePoint := scaleOf(len(x.Coefficient().String())) + x.Exponent()
	halved := digitsBeforePoint / rootDigitDivisor
	if digitsBeforePoint > 0 && digitsBeforePoint%rootDigitDivisor != 0 {
		halved++
	}
	return decimal.Max(decimal.New(1, halved), decimal.New(1, -scale))
}

// roundRoot rounds a non-negative root to precision places, half-even for an
// exact tie. exact says whether root is the true root rather than a truncation.
func roundRoot(root decimal.Decimal, precision int32, exact bool) decimal.Decimal {
	kept := root.Truncate(precision)
	tail := root.Sub(kept)
	if tail.IsZero() {
		return kept
	}
	half := decimal.New(halfUnitDigit, -precision-1)
	order := tail.Cmp(half)
	if order < 0 || (order == 0 && exact && isEven(kept, precision)) {
		return kept
	}
	return kept.Add(decimal.New(1, -precision))
}
