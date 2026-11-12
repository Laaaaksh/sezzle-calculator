package operations

import "github.com/Laaaaksh/sezzle-calculator/backend/internal/config"

// NewDefaultRegistry is the single registration point. Adding an operation
// means adding its file and one line to the list below; nothing else changes.
func NewDefaultRegistry(_ config.CalculatorConfig) *Registry {
	return NewRegistry()
}
