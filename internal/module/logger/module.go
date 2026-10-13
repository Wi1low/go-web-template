// Package logger provides the zap logger module. It replaces the global
// sugared logger so third-party code can use zap.L(), and publishes the
// instance for explicit dependency consumption.
package logger

import (
	"context"

	"github.com/Wi1low/go-web-template/internal/config"
	"github.com/Wi1low/go-web-template/internal/kernel"
	"go.uber.org/zap"
)

type module struct{}

// New returns the logger module.
func New() kernel.Module { return module{} }

func (module) Name() string { return "logger" }

// Enabled always true.
func (module) Enabled(*config.Config) bool { return true }

func (module) Init(_ context.Context, a *kernel.App) error {
	log, err := config.NewLogger(a.Cfg.Server.Mode, a.Cfg.App.Name)
	if err != nil {
		return err
	}
	zap.ReplaceGlobals(log)
	kernel.Provide(a, log)
	return nil
}

func (module) Stop(context.Context) error {
	// Sync flushes buffered entries; ENOSYS on some stdio targets is ignored
	// deliberately.
	_ = zap.L().Sync()
	return nil
}
