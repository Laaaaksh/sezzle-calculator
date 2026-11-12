package calculator

import (
	"context"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/config"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/operations"
)

// CalculatorModule is the process-wide module instance, built once by NewModule.
var CalculatorModule IModule

// NewModule builds the module on first call and returns the same instance
// afterwards. It is a var so tests can replace it.
var NewModule = func(_ context.Context, cfg config.CalculatorConfig) IModule {
	if CalculatorModule == nil {
		core := NewCore(operations.NewDefaultRegistry(cfg), cfg)
		CalculatorModule = &Module{Core: core}
	}
	return CalculatorModule
}

// IModule exposes the module's parts to boot. A handler is added with the
// HTTP step.
type IModule interface {
	GetCore() ICore
}

// Module is the concrete IModule.
type Module struct {
	Core ICore
}

var _ IModule = (*Module)(nil)

// GetCore returns the business logic.
func (m *Module) GetCore() ICore { return m.Core }
