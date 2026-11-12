package operations

import "github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"

// NewDefaultRegistry is the single registration point. Adding an operation
// means adding its file and one line to the list below; nothing else changes.
func NewDefaultRegistry(limits entities.Limits) *Registry {
	return NewRegistry(
		NewAdd(),
		NewSubtract(),
		NewMultiply(),
		NewDivide(limits.DecimalPrecision),
		NewPower(limits),
		NewSqrt(limits.DecimalPrecision),
		NewPercentage(),
	)
}
