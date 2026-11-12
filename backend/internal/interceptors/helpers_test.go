package interceptors_test

import (
	"context"
	"net"
	"net/http"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/logger"
)

// captureLogs swaps the base logger for an in-memory observer and returns the
// observed entries plus a function that restores the previous logger.
func captureLogs() (*observer.ObservedLogs, func()) {
	core, observed := observer.New(zapcore.DebugLevel)
	return observed, logger.SetForTest(zap.New(core).Sugar())
}

// okHandler answers 200 with no body.
var okHandler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
})

type pipeAddr struct{}

func (pipeAddr) Network() string { return "pipe" }
func (pipeAddr) String() string  { return "pipe" }

// pipeListener serves connections created by dial over net.Pipe, so a real
// HTTP server runs without binding a port.
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

// startPipeServer runs h on a real net/http server over a pipeListener and
// returns a client wired to it plus a stop function.
func startPipeServer(h http.Handler) (*http.Client, func()) {
	ln := newPipeListener()
	server := &http.Server{Handler: h}
	go func() { _ = server.Serve(ln) }()
	client := &http.Client{Transport: &http.Transport{DialContext: ln.dial}}
	return client, func() { _ = server.Close() }
}
