// Command api boots the application from a list of kernel modules.
//
// Plugging components in or out is exactly one line here: append or remove a
// module constructor. Per-environment on/off switches live in the config
// files, so an unregistered-but-present module also disappears by setting its
// enabled: false — no code changes needed at runtime either.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Wi1low/go-web-template/internal/kernel"
	configmod "github.com/Wi1low/go-web-template/internal/module/config"
	databasemod "github.com/Wi1low/go-web-template/internal/module/database"
	httpservermod "github.com/Wi1low/go-web-template/internal/module/httpserver"
	loggermod "github.com/Wi1low/go-web-template/internal/module/logger"
	metricsmod "github.com/Wi1low/go-web-template/internal/module/metrics"
	redismod "github.com/Wi1low/go-web-template/internal/module/redis"
	usermod "github.com/Wi1low/go-web-template/internal/module/user"
)

//	@title			go-web-template API
//	@version		1.0
//	@description	Gin application template with a pluggable module kernel.
//	@BasePath		/

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Bearer <token> from /api/v1/login

func main() {
	// Install the shutdown signal before any server starts; Run blocks until
	// this ctx is cancelled or a module reports a fatal failure.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app := kernel.New(
		configmod.New(),     // must be first: loads cmd/etc/config.<APP_ENV>.yaml + env overrides
		loggermod.New(),     // zap global logger
		databasemod.New(),   // gorm (mysql); database.enabled
		redismod.New(),      // go-redis v9; redis.enabled
		metricsmod.New(),    // prometheus admin server; prometheus.enabled
		usermod.New(),       // demo feature module (CRUD + jwt); database.enabled
		httpservermod.New(), // main gin server; always last
	)

	if err := app.Run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}
