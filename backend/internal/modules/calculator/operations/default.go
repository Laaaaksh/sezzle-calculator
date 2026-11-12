package operations

import "github.com/Laaaaksh/sezzle-calculator/backend/internal/config"

// NewDefaultRegistry is the single registration point. Adding an operation
// means adding its file and one line to the list below; nothing else changes.
func NewDefaultRegistry(cfg config.CalculatorConfig) *Registry {
	return NewRegistry(
		NewAdd(),
		NewSubtract(),
		NewMultiply(),
		NewDivide(cfg.DecimalPrecision),
	)
}
