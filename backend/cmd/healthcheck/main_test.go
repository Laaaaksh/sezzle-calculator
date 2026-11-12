package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/health/entities"
)

const (
	testConfigDir    = "../../config"
	testBadConfigDir = "does-not-exist"
	testEnvPort      = "APP_SERVER_PORT"
	testPort         = "9191"
	testProbeURL     = "http://127.0.0.1:9191/health/live"
	testMalformedURL = "http://bad host/"
)

var errTestDial = errors.New("connection refused")

// roundTripFunc stubs the HTTP transport so the tests open no sockets.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type HealthcheckTestSuite struct {
	suite.Suite
	ctx       context.Context
	stderr    *bytes.Buffer
	requested string
}

func TestHealthcheckTestSuite(t *testing.T) {
	suite.Run(t, new(HealthcheckTestSuite))
}

func (s *HealthcheckTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.stderr = &bytes.Buffer{}
	s.requested = ""
}

// clientAnswering returns a client whose transport replies with status and
// records the URL it was asked for.
func (s *HealthcheckTestSuite) clientAnswering(status int) *http.Client {
	return &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		s.requested = r.URL.String()
		return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(""))}, nil
	})}
}

func (s *HealthcheckTestSuite) clientFailing() *http.Client {
	return &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errTestDial
	})}
}

func (s *HealthcheckTestSuite) useTestConfig(dir string) {
	s.T().Setenv(constants.EnvVarEnvironment, string(constants.EnvTest))
	s.T().Setenv(constants.EnvVarConfigDir, dir)
	s.T().Setenv(testEnvPort, testPort)
}

func (s *HealthcheckTestSuite) TestRunProbesConfiguredPortAndReturnsOK() {
	s.useTestConfig(testConfigDir)

	code := run(s.ctx, s.clientAnswering(http.StatusOK), s.stderr)

	s.Equal(constants.ExitCodeOK, code)
	s.Equal("http://127.0.0.1:"+testPort+entities.PathLive, s.requested)
	s.Empty(s.stderr.String())
}

func (s *HealthcheckTestSuite) TestRunReturnsRuntimeCodeWhenEndpointAnswers503() {
	s.useTestConfig(testConfigDir)

	code := run(s.ctx, s.clientAnswering(http.StatusServiceUnavailable), s.stderr)

	s.Equal(constants.ExitCodeRuntime, code)
	s.Contains(s.stderr.String(), "unexpected status 503")
}

func (s *HealthcheckTestSuite) TestRunReturnsConfigCodeWhenConfigCannotLoad() {
	s.useTestConfig(testBadConfigDir)

	code := run(s.ctx, s.clientAnswering(http.StatusOK), s.stderr)

	s.Equal(constants.ExitCodeConfigInvalid, code)
	s.Contains(s.stderr.String(), "load configuration")
	s.Empty(s.requested)
}

func (s *HealthcheckTestSuite) TestProbeFailsWhenTransportFails() {
	err := probe(s.ctx, s.clientFailing(), testProbeURL)

	s.ErrorIs(err, errTestDial)
}

func (s *HealthcheckTestSuite) TestProbeFailsOnMalformedURL() {
	err := probe(s.ctx, s.clientAnswering(http.StatusOK), testMalformedURL)

	s.Error(err)
	s.Empty(s.requested)
}
