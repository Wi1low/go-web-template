// Package kernel provides a minimal application kernel that manages the
// lifecycle of pluggable modules.
//
// A module is any component that implements the Module interface. Modules are
// initialized in registration order and stopped in reverse order. Use
// Provide/Require to publish and consume services between modules.
//
// Adding a component to an application means implementing Module and appending
// one entry to kernel.New in cmd/api/main.go; removing it means deleting that
// line (or setting its config section enabled: false).
package kernel

import (
	"context"

	"github.com/Wi1low/go-web-template/internal/config"
)

// Module is a lifecycle component managed by the App.
//
// Contract:
//   - Enabled is called with the fully loaded configuration; a false result
//     skips Init and Stop entirely for that module.
//   - Init runs in registration order and may call Provide to publish
//     services for modules initialized later. It should return quickly and
//     run long-lived servers in goroutines, reporting async failures with
//     App.Fail.
//   - Stop is called only for modules whose Init succeeded, in reverse
//     registration order, with a fresh timeout context that is never already
//     cancelled.
type Module interface {
	Name() string
	Enabled(cfg *config.Config) bool
	Init(ctx context.Context, app *App) error
	Stop(ctx context.Context) error
}
