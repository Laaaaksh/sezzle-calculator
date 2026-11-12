package boot_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/boot"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
)

const runConfiguredAddress = ":8080"

// runProbeClient has a timeout so a probe against a stuck server fails the
// poll instead of blocking it.
var runProbeClient = &http.Client{Timeout: runWaitLimit}

var errTestBind = errors.New("bind exploded")

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

	s.app, err = boot.Initialize(context.Background())
	s.Require().NoError(err)
}

func (s *RunTestSuite) TearDownTest() {
}

func (s *RunTestSuite) url(path string) string {
	return fmt.Sprintf(runURLFmt, s.port, path)
}

func (s *RunTestSuite) waitUntilServing() {
	s.Require().Eventually(func() bool {
		resp, err := runProbeClient.Get(s.url(testLivePath))
		if err != nil {
			return false
		}
		defer func() { _ = resp.Body.Close() }()
		return resp.StatusCode == http.StatusOK
	}, runWaitLimit, runPollInterval)
}

func (s *RunTestSuite) TestRunServesUntilContextCancelledThenStopsCleanly() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*runWaitLimit)
	defer cancel()
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
	// The deadline makes a Run that wrongly bound the port stop and fail the
	// assertion below, instead of serving until the test binary times out.
	ctx, cancel := context.WithTimeout(context.Background(), runWaitLimit)
	defer cancel()

	err = s.app.Run(ctx)

	s.Error(err)
}

// RunInjectedListenerTestSuite drives Run without a real port: the listener
// comes from a test double, so the suite runs wherever binding is forbidden.
type RunInjectedListenerTestSuite struct {
	suite.Suite
	app      *boot.App
	network  string
	address  string
	listener *pipeListener
}

func TestRunInjectedListenerTestSuite(t *testing.T) {
	suite.Run(t, new(RunInjectedListenerTestSuite))
}

func (s *RunInjectedListenerTestSuite) SetupTest() {
	dir := s.T().TempDir()
	s.Require().NoError(os.WriteFile(filepath.Join(dir, fileDefault), []byte(fixtureDefaultToml), 0o600))
	s.Require().NoError(os.WriteFile(filepath.Join(dir, fileTest), []byte(fixtureTestToml), 0o600))
	s.T().Setenv(envVarEnv, envNameTest)
	s.T().Setenv(envVarConfigDir, dir)
	s.T().Setenv(envVarPort, envValueEmpty)
	s.listener = newPipeListener()

	app, err := boot.Initialize(context.Background())
	s.Require().NoError(err)
	app.UseListen(func(network, address string) (net.Listener, error) {
		s.network, s.address = network, address
		return s.listener, nil
	})
	s.app = app
}

func (s *RunInjectedListenerTestSuite) TestRunWithACancelledContextShutsDownCleanly() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	s.NoError(s.app.Run(ctx))
}

func (s *RunInjectedListenerTestSuite) TestRunBindsTheConfiguredAddressOverTCP() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	s.Require().NoError(s.app.Run(ctx))

	s.Equal(constants.ListenNetwork, s.network)
	s.Equal(runConfiguredAddress, s.address)
}

func (s *RunInjectedListenerTestSuite) TestRunClosesTheListenerOnShutdown() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	s.Require().NoError(s.app.Run(ctx))

	_, err := s.listener.Accept()
	s.ErrorIs(err, net.ErrClosed)
}

func (s *RunInjectedListenerTestSuite) TestRunReturnsTheBindFailure() {
	s.app.UseListen(func(string, string) (net.Listener, error) { return nil, errTestBind })

	err := s.app.Run(context.Background())

	s.ErrorIs(err, errTestBind)
}
