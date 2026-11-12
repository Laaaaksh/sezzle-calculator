// Command healthcheck probes the API's liveness endpoint and exits 0 when it
// answers 200, 1 otherwise. It exists because the distroless runtime image has
// no shell or curl, so a Docker HEALTHCHECK needs a binary to run. It reads the
// port through the same config loader as the API, so the two can never disagree.
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/config"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/health/entities"
)

func main() {
	client := &http.Client{Timeout: constants.HealthcheckTimeoutSeconds * time.Second}
	os.Exit(run(context.Background(), client, os.Stderr))
}

// run maps the probe outcome to an exit code. It returns the code instead of
// calling os.Exit so the logic stays testable. The client is injected so tests
// can stub the transport without opening a socket.
func run(ctx context.Context, client *http.Client, stderr io.Writer) int {
	cfg, err := config.Load()
	if err != nil {
		report(stderr, constants.MsgHealthcheckConfigFailed, err)
		return constants.ExitCodeConfigInvalid
	}
	url := fmt.Sprintf(constants.HealthcheckURLFmt,
		constants.HealthcheckScheme, constants.HealthcheckHost, cfg.Server.Port, entities.PathLive)
	if err := probe(ctx, client, url); err != nil {
		report(stderr, constants.MsgHealthcheckFailed, err)
		return constants.ExitCodeRuntime
	}
	return constants.ExitCodeOK
}

// probe sends one GET and succeeds only on 200. The client timeout set in main
// guarantees a hung server fails the check instead of hanging it.
func probe(ctx context.Context, client *http.Client, url string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf(constants.ErrFmtHealthcheckRequest, err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf(constants.ErrFmtHealthcheckDo, url, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(constants.ErrFmtHealthcheckStatus, url, resp.StatusCode)
	}
	return nil
}

func report(stderr io.Writer, format string, err error) {
	// A failed stderr write leaves nothing else to report to, so the error is dropped.
	_, _ = fmt.Fprintf(stderr, format, err)
}
