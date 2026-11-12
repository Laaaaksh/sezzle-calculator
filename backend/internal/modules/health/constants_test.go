package health_test

const (
	testRequestID    = "req-health-1"
	wantLiveBody     = `{"status":"ok"}`
	wantReadyBody    = `{"status":"ready"}`
	wantContentType  = "application/json"
	testUnknownPath  = "/health/unknown"
	testNotReadyText = "service is not ready"
)
