// Package entities holds the health module's request and response types and
// constants.
package entities

// Route paths, relative to the server root. Health endpoints sit outside the
// versioned API so load balancers and orchestrators do not depend on the
// API version.
const (
	PathLive  = "/health/live"
	PathReady = "/health/ready"
)

// Status values reported in response bodies.
const (
	StatusOK    = "ok"
	StatusReady = "ready"
)
