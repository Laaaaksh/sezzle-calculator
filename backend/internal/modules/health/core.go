package health

import (
	"context"
	"errors"
	"sync/atomic"
)

//go:generate go tool mockgen -source=core.go -destination=mock/mock_core.go -package=mock

// ErrNotReady is the internal cause behind a 503 from the readiness probe. It
// stays in the logs; clients only see the generic unavailable message.
var ErrNotReady = errors.New("service is not ready")

// ICore is the health business logic.
type ICore interface {
	// IsReady reports whether the service should receive traffic.
	IsReady(ctx context.Context) bool
	// SetReady flips readiness. Boot sets it true once the router is built and
	// false when shutdown begins, so load balancers drain the instance first.
	SetReady(ready bool)
}

// Core holds readiness state. It starts not ready, so an instance never
// reports ready before boot says it is.
type Core struct {
	ready atomic.Bool
}

var _ ICore = (*Core)(nil)

// NewCore returns a Core that is not yet ready.
func NewCore() *Core {
	return &Core{}
}

// IsReady reports the current readiness. The atomic makes it safe to read from
// request goroutines while shutdown flips it.
func (c *Core) IsReady(_ context.Context) bool {
	return c.ready.Load()
}

// SetReady stores the readiness flag.
func (c *Core) SetReady(ready bool) {
	c.ready.Store(ready)
}
