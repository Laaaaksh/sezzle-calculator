// Package boot builds the application in a fixed order and runs it.
package boot

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/config"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/logger"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/health"
)

// App is the assembled service: configuration, HTTP server and the modules
// whose state boot controls.
type App struct {
	cfg    *config.Config
	server *http.Server
	health health.IModule
}

// Initialize wires the service in a fixed order: configuration, logger,
// modules, router. The order is a dependency chain: the logger needs the
// environment from the config, modules log, and the router needs modules.
// It returns early on the first failure, tagged with ErrConfigLoad or
// ErrLoggerInit so the caller can pick an exit code.
func Initialize(ctx context.Context) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, &stageError{stage: ErrConfigLoad, cause: err}
	}
	if err := logger.Initialize(cfg.App.Env); err != nil {
		return nil, &stageError{stage: ErrLoggerInit, cause: err}
	}
	logger.Ctx(ctx).Infow(constants.LogMsgConfigLoaded,
		constants.LogKeyEnv, string(cfg.App.Env),
		constants.LogKeyPort, cfg.Server.Port,
	)

	modules := Modules{Health: health.NewModule(ctx)}
	router := NewRouter(cfg, modules)

	app := &App{cfg: cfg, server: newServer(cfg, router), health: modules.Health}
	// Ready only once the router exists, so a probe never sees a half-built app.
	app.health.GetCore().SetReady(true)
	return app, nil
}

// newServer applies the configured timeouts. ReadHeaderTimeout reuses the read
// timeout so a client dribbling headers (slowloris) cannot hold a connection.
func newServer(cfg *config.Config, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf(constants.ListenAddrFmt, cfg.Server.Port),
		Handler:           handler,
		ReadTimeout:       cfg.Server.ReadTimeout,
		ReadHeaderTimeout: cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
	}
}

// Handler returns the root HTTP handler, for tests and embedding.
func (a *App) Handler() http.Handler { return a.server.Handler }

// Run binds the configured address and serves until ctx is cancelled or
// SIGINT/SIGTERM arrives, then shuts the server down within the configured
// grace period. A clean shutdown returns nil.
func (a *App) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	// Restore default signal handling as soon as shutdown begins, so a second
	// SIGINT/SIGTERM during the drain or grace period exits immediately.
	context.AfterFunc(ctx, stop)

	// Binding first, separately from serving, makes "port already in use" a
	// synchronous startup error instead of a goroutine failure.
	listener, err := net.Listen("tcp", a.server.Addr)
	if err != nil {
		return fmt.Errorf(constants.ErrFmtListen, a.server.Addr, err)
	}
	return a.Serve(ctx, listener)
}

// Serve runs the server on listener until ctx is done, then shuts down
// gracefully. It is split from Run so tests can supply an in-memory listener
// and never need a real port.
func (a *App) Serve(ctx context.Context, listener net.Listener) error {
	logger.Ctx(ctx).Infow(constants.LogMsgServerStarting, constants.LogKeyAddr, listener.Addr().String())

	serveErr := make(chan error, 1)
	go func() { serveErr <- a.server.Serve(listener) }()

	select {
	case err := <-serveErr:
		return a.serveFailed(ctx, err)
	case <-ctx.Done():
	}
	return a.shutdown(ctx, serveErr)
}

// serveFailed handles Serve returning before any shutdown was requested.
func (a *App) serveFailed(ctx context.Context, err error) error {
	a.health.GetCore().SetReady(false)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	logger.Ctx(ctx).Errorw(constants.LogMsgServerStopped, constants.LogKeyError, err.Error())
	return fmt.Errorf(constants.ErrFmtServe, err)
}

// shutdown flips readiness first and keeps serving for the drain delay, so
// load balancers polling readiness stop sending traffic before the listener
// closes. It then lets in-flight requests finish within the grace period. If
// they do not, connections are closed so the process never hangs past the
// deadline.
func (a *App) shutdown(ctx context.Context, serveErr <-chan error) error {
	log := logger.Ctx(ctx)
	log.Infow(constants.LogMsgShutdownStarted)
	a.health.GetCore().SetReady(false)
	time.Sleep(a.cfg.Server.ShutdownDrainDelay)

	// The run context is already cancelled; the grace period needs its own.
	graceCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), a.cfg.Server.ShutdownGracePeriod)
	defer cancel()

	if err := a.server.Shutdown(graceCtx); err != nil {
		log.Errorw(constants.LogMsgShutdownFailed, constants.LogKeyError, err.Error())
		// Close is best effort after a failed graceful stop.
		_ = a.server.Close()
		return fmt.Errorf(constants.ErrFmtShutdown, err)
	}
	// Serve returns http.ErrServerClosed once Shutdown has been called; any
	// other value means it failed while draining.
	if err := <-serveErr; !errors.Is(err, http.ErrServerClosed) {
		log.Errorw(constants.LogMsgShutdownFailed, constants.LogKeyError, err.Error())
		return fmt.Errorf(constants.ErrFmtServe, err)
	}
	log.Infow(constants.LogMsgShutdownCompleted)
	log.Infow(constants.LogMsgServerStopped)
	return nil
}
