// Package config provides the kernel module that loads application
// configuration. It must be the first module registered.
package config

import (
	"context"

	appcfg "github.com/Wi1low/go-web-template/internal/config"
	"github.com/Wi1low/go-web-template/internal/kernel"
)

type module struct{}

// New returns the config module.
func New() kernel.Module { return module{} }

func (module) Name() string { return "config" }

// Enabled always true: every application needs configuration.
func (module) Enabled(*appcfg.Config) bool { return true }

func (module) Init(_ context.Context, a *kernel.App) error {
	cfg, err := appcfg.Load()
	if err != nil {
		return err
	}
	a.Cfg = cfg
	kernel.Provide(a, cfg)
	return nil
}

func (module) Stop(context.Context) error { return nil }
