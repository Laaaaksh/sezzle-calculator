package calculator

import (
	"context"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/config"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/operations"
)

// NewModule builds a fresh module on every call from the cfg it is given.
// There is deliberately no package-level instance: a cached one would make a
// later call silently ignore its cfg, and tests would share state. Boot calls
// it once and owns the result.
func NewModule(_ context.Context, cfg config.CalculatorConfig) IModule {
	limits := limitsFrom(cfg)
	core := NewCore(operations.NewDefaultRegistry(limits), limits)
	return &Module{Core: core, Handler: NewHTTPHandler(core, limits)}
}

// limitsFrom is the one place the module's config type becomes the domain
// type, so nothing below this package imports internal/config.
func limitsFrom(cfg config.CalculatorConfig) entities.Limits {
	return entities.Limits{
		MaxOperandLength: cfg.MaxOperandLength,
		MaxExponent:      cfg.MaxExponent,
		DecimalPrecision: cfg.DecimalPrecision,
	}
}

// IModule exposes the module's parts to boot.
type IModule interface {
	GetCore() ICore
	GetHandler() *HTTPHandler
}

// Module is the concrete IModule.
type Module struct {
	Core    ICore
	Handler *HTTPHandler
}

var _ IModule = (*Module)(nil)

// GetCore returns the business logic.
func (m *Module) GetCore() ICore { return m.Core }

// GetHandler returns the HTTP handler.
func (m *Module) GetHandler() *HTTPHandler { return m.Handler }
