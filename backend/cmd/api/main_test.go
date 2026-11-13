package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/boot"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
)

const (
	testConfigDir    = "../../config"
	testBadConfigDir = "does-not-exist"
	testBadEnv       = "staging-typo"
	testEnvPort      = "APP_SERVER_PORT"
	testListenHost   = "127.0.0.1:0"
	skipNoListen     = "sandbox does not allow binding a local port"
	testCauseText    = "something broke"
	// testRunLimit is the hard deadline of any test that calls run with a live
	// listener. run stops on its context, so a test that never cancels it would
	// serve forever; every such test passes a context with this deadline.
	testRunLimit = 5 * time.Second
	// testWildcardAddr is the same form of address the app binds (":port"), so
	// a blocker on it conflicts on every OS. A blocker on 127.0.0.1 does not
	// always conflict with a wildcard bind (macOS allows both).
	testWildcardAddr = ":0"
)

var errTestCause = errors.New(testCauseText)

type RunTestSuite struct {
	suite.Suite
	stderr *bytes.Buffer
}

func TestRunTestSuite(t *testing.T) {
	suite.Run(t, new(RunTestSuite))
}

func (s *RunTestSuite) SetupTest() {
	s.stderr = &bytes.Buffer{}
	s.T().Setenv(constants.EnvVarEnvironment, string(constants.EnvTest))
	s.T().Setenv(constants.EnvVarConfigDir, testConfigDir)
}

// freePort reserves a port and releases it for the app; the test is skipped where binding is not permitted.
func (s *RunTestSuite) freePort() int {
	listener, err := net.Listen(constants.ListenNetwork, testListenHost)
	if err != nil {
		s.T().Skip(skipNoListen)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	s.Require().NoError(listener.Close())
	return port
}

// runBounded fails the test, instead of hanging it, if run does not return shortly after testRunLimit.
func (s *RunTestSuite) runBounded(ctx context.Context) int {
	ctx, cancel := context.WithTimeout(ctx, testRunLimit)
	defer cancel()
	done := make(chan int, 1)
	go func() { done <- run(ctx, s.stderr) }()
	select {
	case code := <-done:
		return code
	case <-time.After(2 * testRunLimit):
		s.FailNow("run did not return within the deadline")
		return 0
	}
}

func (s *RunTestSuite) TestRunWithAnUnknownEnvironmentReturnsConfigCode() {
	s.T().Setenv(constants.EnvVarEnvironment, testBadEnv)

	code := run(context.Background(), s.stderr)

	s.Equal(constants.ExitCodeConfigInvalid, code)
}

func (s *RunTestSuite) TestRunWithAnUnknownEnvironmentExplainsOnStderr() {
	s.T().Setenv(constants.EnvVarEnvironment, testBadEnv)

	run(context.Background(), s.stderr)

	s.Contains(s.stderr.String(), "failed to load configuration")
}

func (s *RunTestSuite) TestRunWithAMissingConfigDirReturnsConfigCode() {
	s.T().Setenv(constants.EnvVarConfigDir, testBadConfigDir)

	code := run(context.Background(), s.stderr)

	s.Equal(constants.ExitCodeConfigInvalid, code)
}

func (s *RunTestSuite) TestRunWithACancelledContextServesAndExitsCleanly() {
	s.T().Setenv(testEnvPort, fmt.Sprint(s.freePort()))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	code := s.runBounded(ctx)

	s.Equal(constants.ExitCodeOK, code)
	s.Empty(s.stderr.String())
}

func (s *RunTestSuite) TestRunWhenThePortIsTakenReturnsRuntimeCode() {
	// The wildcard bind is the point: it must conflict with the app's own wildcard bind.
	blocker, err := net.Listen(constants.ListenNetwork, testWildcardAddr) //nolint:gosec // G102
	if err != nil {
		s.T().Skip(skipNoListen)
	}
	defer func() { _ = blocker.Close() }()
	s.T().Setenv(testEnvPort, fmt.Sprint(blocker.Addr().(*net.TCPAddr).Port))

	code := s.runBounded(context.Background())

	s.Equal(constants.ExitCodeRuntime, code)
	s.Contains(s.stderr.String(), "server failed")
}

type StartupFailureTestSuite struct {
	suite.Suite
	stderr *bytes.Buffer
}

func TestStartupFailureTestSuite(t *testing.T) {
	suite.Run(t, new(StartupFailureTestSuite))
}

func (s *StartupFailureTestSuite) SetupTest() {
	s.stderr = &bytes.Buffer{}
}

// The logger stage cannot be reached through run: config validation rejects an
// environment the logger would refuse. The mapping is still checked here, so
// the stage keeps its own exit code if the two ever drift apart.
func (s *StartupFailureTestSuite) TestLoggerStageFailureReturnsLoggerCode() {
	err := fmt.Errorf("%w: %w", boot.ErrLoggerInit, errTestCause)

	code := reportStartupFailure(s.stderr, err)

	s.Equal(constants.ExitCodeLoggerInvalid, code)
	s.Contains(s.stderr.String(), "failed to initialise logger")
}

func (s *StartupFailureTestSuite) TestConfigStageFailureReturnsConfigCode() {
	err := fmt.Errorf("%w: %w", boot.ErrConfigLoad, errTestCause)

	code := reportStartupFailure(s.stderr, err)

	s.Equal(constants.ExitCodeConfigInvalid, code)
}

func (s *StartupFailureTestSuite) TestUntaggedFailureReturnsRuntimeCode() {
	code := reportStartupFailure(s.stderr, errTestCause)

	s.Equal(constants.ExitCodeRuntime, code)
	s.Contains(s.stderr.String(), testCauseText)
}
