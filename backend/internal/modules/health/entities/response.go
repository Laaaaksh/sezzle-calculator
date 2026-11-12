package entities

// LiveResponse is the body of GET /health/live.
type LiveResponse struct {
	Status string `json:"status"`
}

// ReadyResponse is the body of GET /health/ready when the service is ready.
// When it is not, the standard JSON error shape is returned instead.
type ReadyResponse struct {
	Status string `json:"status"`
}
