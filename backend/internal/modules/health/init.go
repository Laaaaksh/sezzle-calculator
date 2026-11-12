// Package health serves the liveness and readiness probes.
package health

import "context"

// NewModule builds a fresh module on every call. There is deliberately no
// package-level instance: a cached one would leak readiness state between
// callers and tests. Boot calls it once and owns the result.
func NewModule(_ context.Context) IModule {
	core := NewCore()
	return &Module{Core: core, Handler: NewHTTPHandler(core)}
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
