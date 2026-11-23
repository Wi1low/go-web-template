// Package metrics provides the standalone Prometheus admin server. All
// metrics (gin HTTP counters, gorm dbstats, Go runtime) share the default
// registry and are exposed from exactly one endpoint — never from the main
// application engine.
package metrics

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Wi1low/go-web-template/internal/config"
	"github.com/Wi1low/go-web-template/internal/kernel"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type module struct {
	srv *http.Server
}

// New returns the metrics module.
func New() kernel.Module { return &module{} }

func (*module) Name() string { return "metrics" }

func (*module) Enabled(c *config.Config) bool { return c.Prometheus.Enabled }

func (m *module) Init(_ context.Context, a *kernel.App) error {
	cfg := a.Cfg.Prometheus
	path := cfg.Path
	if path == "" {
		path = "/metrics"
	}

	engine := gin.New()
	engine.GET(path, gin.WrapH(promhttp.Handler()))

	m.srv = &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		zap.L().Info("metrics server listening", zap.String("addr", m.srv.Addr), zap.String("path", path))
		if err := m.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.Fail(fmt.Errorf("metrics server: %w", err))
		}
	}()
	return nil
}

func (m *module) Stop(ctx context.Context) error {
	if m.srv == nil {
		return nil
	}
	return m.srv.Shutdown(ctx)
}
