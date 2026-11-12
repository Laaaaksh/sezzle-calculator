package operations

import (
	"github.com/shopspring/decimal"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

// CheckResultSize is the one result-size rule. It returns RESULT_TOO_LARGE when
// value has more than maxDigits digits and nil otherwise. The core applies it
// to the result of every operation, so add, subtract, multiply, divide,
// percentage, power and sqrt all share the same budget. Power also uses it on
// its own result, after its cheaper pre-checks.
func CheckResultSize(value decimal.Decimal, maxDigits int64) apperror.IError {
	if DigitsOf(value) > maxDigits {
		return ResultTooLarge(maxDigits)
	}
	return nil
}

// ResultTooLarge builds the RESULT_TOO_LARGE error with the digit limit as its
// public field, so every refusal reports the same limit the same way.
func ResultTooLarge(maxDigits int64) apperror.IError {
	return apperror.New(entities.CodeResultTooLarge, ErrResultBeyondBudget).
		WithField(entities.FieldLimit, maxDigits)
}

// DigitsOf counts the digits of d as written without trailing fraction zeros
// or a lone leading zero: integer digits plus fraction digits, so 1000 has 4,
// 0.001 has 3 and 12.5 has 3. Zero has 1. It is exact integer counting on the
// coefficient and the exponent; no float is involved.
func DigitsOf(d decimal.Decimal) int64 {
	if d.IsZero() {
		return 1
	}
	d = stripTrailingZeros(d)
	digits := coefficientDigits(d)
	exp := int64(d.Exponent())
	if exp >= 0 {
		return digits + exp
	}
	return max(digits, -exp)
}

// coefficientDigits is the number of digits in the coefficient of |d|.
func coefficientDigits(d decimal.Decimal) int64 {
	return int64(len(d.Abs().Coefficient().String()))
}
