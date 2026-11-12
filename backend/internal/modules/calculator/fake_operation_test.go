package calculator_test

import (
	"errors"

	"github.com/shopspring/decimal"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

var errFakeCause = errors.New("fake operation failed")

// fakeOperation is a configurable stand-in for a real operation. It records the
// operands it was given so tests can prove Apply was (or was not) reached.
type fakeOperation struct {
	name   entities.OperationType
	arity  int
	result decimal.Decimal
	err    apperror.IError
	calls  int
	seen   []decimal.Decimal
}

func (f *fakeOperation) Name() entities.OperationType { return f.name }

func (f *fakeOperation) Arity() int { return f.arity }

func (f *fakeOperation) Symbol() string { return testFakeSymbol }

func (f *fakeOperation) Description() string { return testFakeDescription }

func (f *fakeOperation) Apply(operands []decimal.Decimal) (decimal.Decimal, apperror.IError) {
	f.calls++
	f.seen = operands
	return f.result, f.err
}

func mustDecimal(text string) decimal.Decimal {
	return decimal.RequireFromString(text)
}
