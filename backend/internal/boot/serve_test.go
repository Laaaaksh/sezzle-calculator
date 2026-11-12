package boot_test

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/boot"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/health"
)

const (
	serveWaitLimit     = 5 * time.Second
	serveGraceDefault  = `shutdown_grace_period = "2s"`
	serveGraceOverride = `shutdown_grace_period = "5s"`
	serveGraceShort    = `shutdown_grace_period = "50ms"`
	serveGraceDrain    = "shutdown_grace_period = \"5s\"\nshutdown_drain_delay = \"2s\""
	serveDrainPoll     = 10 * time.Millisecond
	serveURL           = "http://pipe"
	// serveSettleDelay lets the server register the accepted connection; the
	// net/http server exposes no hook to observe that.
	serveSettleDelay = 20 * time.Millisecond
)

var errTestAccept = errors.New("accept exploded")

// pipeAddr is the address of an in-memory listener.
type pipeAddr struct{}

func (pipeAddr) Network() string { return "pipe" }
func (pipeAddr) String() string  { return "pipe" }

// pipeListener serves connections created by dial over net.Pipe, so the HTTP
// server runs without binding a real port.
type pipeListener struct {
	conns chan net.Conn
	done  chan struct{}
	once  sync.Once
}

func newPipeListener() *pipeListener {
	return &pipeListener{conns: make(chan net.Conn), done: make(chan struct{})}
}

func (l *pipeListener) Accept() (net.Conn, error) {
	select {
	case conn := <-l.conns:
		return conn, nil
	case <-l.done:
		return nil, net.ErrClosed
	}
}

func (l *pipeListener) Close() error {
	l.once.Do(func() { close(l.done) })
	return nil
}

func (l *pipeListener) Addr() net.Addr { return pipeAddr{} }

func (l *pipeListener) dial(ctx context.Context, _, _ string) (net.Conn, error) {
	client, server := net.Pipe()
	select {
	case l.conns <- server:
		return client, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// failingListener returns a permanent error from Accept.
type failingListener struct{ *pipeListener }

func (l *failingListener) Accept() (net.Conn, error) { return nil, errTestAccept }

type ServeTestSuite struct {
	suite.Suite
	dir      string
	listener *pipeListener
	client   *http.Client
}

func TestServeTestSuite(t *testing.T) {
	suite.Run(t, new(ServeTestSuite))
}

func (s *ServeTestSuite) SetupTest() {
	s.dir = s.T().TempDir()
	s.writeConfig(serveGraceOverride)
	s.T().Setenv(envVarEnv, envNameTest)
	s.T().Setenv(envVarConfigDir, s.dir)
	s.T().Setenv(envVarPort, envValueEmpty)
	health.HealthModule = nil
	s.listener = newPipeListener()
	s.client = &http.Client{Transport: &http.Transport{DialContext: s.listener.dial}}
}

func (s *ServeTestSuite) TearDownTest() {
	s.client.CloseIdleConnections()
	health.HealthModule = nil
}

// writeConfig writes the default file with the given grace period line.
func (s *ServeTestSuite) writeConfig(graceLine string) {
	body := strings.Replace(fixtureDefaultToml, serveGraceDefault, graceLine, 1)
	s.Require().NoError(os.WriteFile(filepath.Join(s.dir, fileDefault), []byte(body), 0o600))
	s.Require().NoError(os.WriteFile(filepath.Join(s.dir, fileTest), []byte(fixtureTestToml), 0o600))
}

func (s *ServeTestSuite) initialize() *boot.App {
	app, err := boot.Initialize(context.Background())
	s.Require().NoError(err)
	return app
}

func (s *ServeTestSuite) start(ctx context.Context, app *boot.App, ln net.Listener) <-chan error {
	done := make(chan error, 1)
	go func() { done <- app.Serve(ctx, ln) }()
	return done
}

func (s *ServeTestSuite) status(path string) int {
	resp, err := s.client.Get(serveURL + path)
	s.Require().NoError(err)
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode
}

func (s *ServeTestSuite) wait(done <-chan error) error {
	select {
	case err := <-done:
		return err
	case <-time.After(serveWaitLimit):
		s.FailNow("Serve did not return in time")
		return nil
	}
}

func (s *ServeTestSuite) TestServeAnswersRequestsOverTheListener() {
	app := s.initialize()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := s.start(ctx, app, s.listener)

	s.Equal(http.StatusOK, s.status(testLivePath))
	s.Equal(http.StatusOK, s.status(testReadyPath))

	cancel()
	s.NoError(s.wait(done))
}

func (s *ServeTestSuite) TestServeFlipsReadinessOffWhenShutdownStarts() {
	app := s.initialize()
	ctx, cancel := context.WithCancel(context.Background())
	done := s.start(ctx, app, s.listener)
	s.Equal(http.StatusOK, s.status(testReadyPath))

	cancel()
	s.Require().NoError(s.wait(done))

	s.False(health.HealthModule.GetCore().IsReady(context.Background()))
}

func (s *ServeTestSuite) TestServeKeepsAnsweringNotReadyDuringDrainDelay() {
	s.writeConfig(serveGraceDrain)
	app := s.initialize()
	ctx, cancel := context.WithCancel(context.Background())
	done := s.start(ctx, app, s.listener)
	s.Equal(http.StatusOK, s.status(testReadyPath))

	cancel()

	s.Eventually(func() bool {
		return s.status(testReadyPath) == http.StatusServiceUnavailable
	}, serveWaitLimit, serveDrainPoll)
	s.NoError(s.wait(done))
}

func (s *ServeTestSuite) TestServeReturnsErrorAndDrainsWhenListenerFails() {
	app := s.initialize()

	err := s.wait(s.start(context.Background(), app, &failingListener{newPipeListener()}))

	s.ErrorIs(err, errTestAccept)
	s.False(health.HealthModule.GetCore().IsReady(context.Background()))
}

func (s *ServeTestSuite) TestServeForcesCloseWhenGracePeriodExpires() {
	s.writeConfig(serveGraceShort)
	app := s.initialize()
	ctx, cancel := context.WithCancel(context.Background())
	done := s.start(ctx, app, s.listener)
	s.Equal(http.StatusOK, s.status(testLivePath))
	// A connection that never sends a request is not idle yet, so graceful
	// shutdown has to wait for it until the grace period runs out.
	stuck, err := s.listener.dial(context.Background(), "", "")
	s.Require().NoError(err)
	defer func() { _ = stuck.Close() }()
	time.Sleep(serveSettleDelay)

	cancel()

	s.ErrorIs(s.wait(done), context.DeadlineExceeded)
}
