package calculator

import (
	"github.com/shopspring/decimal"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/operations"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

// checkArity rejects a wrong operand count with INVALID_OPERAND_COUNT.
func checkArity(operation operations.Operation, operands []decimal.Decimal) apperror.IError {
	if len(operands) == operation.Arity() {
		return nil
	}
	return apperror.New(apperror.CodeInvalidOperandCount, ErrWrongOperandCount).
		WithField(entities.FieldOperation, string(operation.Name())).
		WithField(entities.FieldExpected, operation.Arity()).
		WithField(entities.FieldActual, len(operands))
}

// checkLimits rejects the first operand over the configured limits. The parser
// already bounds the input string (a later step); this is the second line of
// defence for any caller that builds decimals another way.
func (c *Core) checkLimits(operands []decimal.Decimal) apperror.IError {
	for i, operand := range operands {
		if digitCount(operand) > c.limits.MaxOperandLength {
			return c.tooLarge(ErrOperandDigits, i)
		}
		if scale(operand) > c.limits.MaxOperandLength {
			return c.tooLarge(ErrOperandScale, i)
		}
	}
	return nil
}

func (c *Core) tooLarge(cause error, index int) apperror.IError {
	return apperror.New(apperror.CodeOperandTooLarge, cause).
		WithField(entities.FieldOperandIndex, index).
		WithField(entities.FieldLimit, c.limits.MaxOperandLength)
}

// digitCount is the number of significant digits (the coefficient), ignoring
// sign and decimal point, so "-1.50" counts the same as "150".
func digitCount(d decimal.Decimal) int {
	return len(d.Abs().Coefficient().String())
}

// scale is how far the decimal point is shifted from the coefficient. A value
// like 1e-900000000 has one digit but forces huge allocations when it is
// added to another number, so the exponent needs its own bound.
func scale(d decimal.Decimal) int {
	exp := int(d.Exponent())
	if exp < 0 {
		return -exp
	}
	return exp
}
