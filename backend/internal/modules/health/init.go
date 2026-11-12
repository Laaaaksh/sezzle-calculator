// Package health serves the liveness and readiness probes.
package health

import "context"

// HealthModule is the process-wide module instance, built once by NewModule.
var HealthModule IModule

// NewModule builds the module on first call and returns the same instance
// afterwards. It is a var so tests can replace it.
var NewModule = func(_ context.Context) IModule {
	if HealthModule == nil {
		core := NewCore()
		HealthModule = &Module{Core: core, Handler: NewHTTPHandler(core)}
	}
	return HealthModule
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
