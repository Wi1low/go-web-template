package kernel

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/Wi1low/go-web-template/internal/config"
	"go.uber.org/zap"
)

// App owns the registered modules, the service container and the shutdown
// orchestration. Create it with New and drive it with Run.
type App struct {
	// Cfg is filled by the config module's Init. It is only valid after that
	// module has started; Enabled/Init of earlier modules must not read it.
	Cfg *config.Config

	modules []Module
	started []Module

	mu       sync.RWMutex
	services map[reflect.Type]any

	failCh chan error
}

// New registers modules in initialization order. The first module must be the
// config module (Name() == "config"). Module names must be unique.
func New(modules ...Module) *App {
	if len(modules) == 0 {
		panic("kernel: no modules registered")
	}
	if modules[0].Name() != "config" {
		panic(`kernel: the first module must be "config"`)
	}
	seen := make(map[string]bool, len(modules))
	for _, m := range modules {
		if seen[m.Name()] {
			panic(fmt.Sprintf("kernel: duplicate module %q", m.Name()))
		}
		seen[m.Name()] = true
	}
	return &App{
		modules:  modules,
		services: make(map[reflect.Type]any),
		failCh:   make(chan error, 1),
	}
}

// Started returns the modules whose Init succeeded, in order. The httpserver
// module uses it to discover route providers among live modules only.
func (a *App) Started() []Module { return a.started }

// Fail reports an asynchronous fatal error (e.g. a listener that died) and
// makes Run begin shutdown. Non-blocking: later reports are dropped.
func (a *App) Fail(err error) {
	select {
	case a.failCh <- err:
	default:
	}
}

// Run initializes modules, blocks until ctx is cancelled or Fail fires, then
// stops everything in reverse order. An Init failure rolls back the modules
// that already started.
func (a *App) Run(ctx context.Context) error {
	for _, m := range a.modules {
		if !m.Enabled(a.Cfg) {
			zap.L().Info("module disabled, skipping", zap.String("module", m.Name()))
			continue
		}
		if err := m.Init(ctx, a); err != nil {
			rollbackErr := a.stopAll(10 * time.Second)
			return fmt.Errorf("kernel: module %q: %w (rollback: %v)", m.Name(), err, rollbackErr)
		}
		a.started = append(a.started, m)
	}

	zap.L().Info("all modules running, waiting for signal or failure")
	var failure error
	select {
	case <-ctx.Done():
		zap.L().Info("shutdown signal received")
	case failure = <-a.failCh:
		zap.L().Error("component exited with error", zap.Error(failure))
	}

	if err := a.stopAll(a.shutdownTimeout()); err != nil {
		return errors.Join(failure, fmt.Errorf("kernel: shutdown: %w", err))
	}
	if failure != nil {
		// A dead component is an abnormal exit: surface it so main exits
		// non-zero and supervisors restart or alert.
		return failure
	}
	zap.L().Info("server exiting")
	return nil
}

func (a *App) shutdownTimeout() time.Duration {
	if a.Cfg != nil && a.Cfg.Server.ShutdownTimeout > 0 {
		return a.Cfg.Server.ShutdownTimeout
	}
	return 10 * time.Second
}

// stopAll stops started modules in reverse order within the total timeout and
// resets the started list. The per-module context is derived from a fresh
// background context so an already-cancelled signal ctx never aborts cleanup.
func (a *App) stopAll(timeout time.Duration) error {
	stopCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var errs []error
	for i := len(a.started) - 1; i >= 0; i-- {
		m := a.started[i]
		if err := m.Stop(stopCtx); err != nil {
			errs = append(errs, fmt.Errorf("module %q: %w", m.Name(), err))
		}
	}
	a.started = nil
	return errors.Join(errs...)
}
