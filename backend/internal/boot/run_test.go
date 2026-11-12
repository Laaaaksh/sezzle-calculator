package boot_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/boot"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/health"
)

const (
	runPollInterval = 20 * time.Millisecond
	runWaitLimit    = 5 * time.Second
	runURLFmt       = "http://127.0.0.1:%d%s"
	skipNoListen    = "sandbox does not allow binding a local port"
)

type RunTestSuite struct {
	suite.Suite
	port int
	app  *boot.App
}

func TestRunTestSuite(t *testing.T) {
	suite.Run(t, new(RunTestSuite))
}

func (s *RunTestSuite) SetupTest() {
	// Reserve a free port, then release it for the app to take. Where binding
	// is not permitted the whole suite is skipped rather than failed.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		s.T().Skip(skipNoListen)
	}
	s.port = listener.Addr().(*net.TCPAddr).Port
	s.Require().NoError(listener.Close())

	dir := s.T().TempDir()
	s.Require().NoError(os.WriteFile(filepath.Join(dir, fileDefault), []byte(fixtureDefaultToml), 0o600))
	s.Require().NoError(os.WriteFile(filepath.Join(dir, fileTest), []byte(fixtureTestToml), 0o600))
	s.T().Setenv(envVarEnv, envNameTest)
	s.T().Setenv(envVarConfigDir, dir)
	s.T().Setenv(envVarPort, fmt.Sprint(s.port))
	health.HealthModule = nil
	calculator.CalculatorModule = nil

	s.app, err = boot.Initialize(context.Background())
	s.Require().NoError(err)
}

func (s *RunTestSuite) TearDownTest() {
	health.HealthModule = nil
	calculator.CalculatorModule = nil
}

func (s *RunTestSuite) url(path string) string {
	return fmt.Sprintf(runURLFmt, s.port, path)
}

func (s *RunTestSuite) waitUntilServing() {
	s.Require().Eventually(func() bool {
		resp, err := http.Get(s.url(testLivePath))
		if err != nil {
			return false
		}
		defer func() { _ = resp.Body.Close() }()
		return resp.StatusCode == http.StatusOK
	}, runWaitLimit, runPollInterval)
}

func (s *RunTestSuite) TestRunServesUntilContextCancelledThenStopsCleanly() {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.app.Run(ctx) }()
	s.waitUntilServing()

	cancel()

	select {
	case err := <-done:
		s.NoError(err)
	case <-time.After(runWaitLimit):
		s.Fail("Run did not return after cancellation")
	}
}

func (s *RunTestSuite) TestRunFailsFastWhenPortIsTaken() {
	blocker, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	s.Require().NoError(err)
	defer func() { _ = blocker.Close() }()

	err = s.app.Run(context.Background())

	s.Error(err)
}
