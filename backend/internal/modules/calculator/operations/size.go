package operations

import (
	"github.com/shopspring/decimal"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

// CheckResultSize returns RESULT_TOO_LARGE when value has more than maxDigits digits. The core applies it to every
// operation's result so all share one budget; power also uses it after its cheaper pre-checks.
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

// DigitsOf counts integer plus fraction digits without trailing fraction zeros or a lone leading zero: 1000 has 4,
// 0.001 has 3, 12.5 has 3, zero has 1. It counts on the coefficient and exponent, so no float is involved.
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

func coefficientDigits(d decimal.Decimal) int64 {
	return int64(len(d.Abs().Coefficient().String()))
}
