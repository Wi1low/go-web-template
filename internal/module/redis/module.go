// Package redis provides the go-redis v9 client module.
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/Wi1low/go-web-template/internal/config"
	"github.com/Wi1low/go-web-template/internal/kernel"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type module struct {
	client *goredis.Client
}

// New returns the redis module.
func New() kernel.Module { return &module{} }

func (*module) Name() string { return "redis" }

func (*module) Enabled(c *config.Config) bool { return c.Redis.Enabled }

func (m *module) Init(ctx context.Context, a *kernel.App) error {
	cfg := a.Cfg.Redis

	// Sentinel and cluster need different constructors; the template keeps the
	// single-node client and surfaces the intent in config. Swap the line
	// below for goredis.NewFailoverClient / goredis.NewClusterClient when you
	// graduate to a topology — the rest of the module is unchanged.
	client := goredis.NewClient(options(cfg))
	m.client = client

	timeout := cfg.DialTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	pingCtx, cancel := context.WithTimeout(ctx, timeout+2*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return fmt.Errorf("redis unreachable at %s: %w", cfg.Addr, err)
	}

	kernel.Provide(a, client)
	zap.L().Info("redis connected", zap.String("addr", cfg.Addr))
	return nil
}

func (m *module) Stop(context.Context) error {
	if m.client == nil {
		return nil
	}
	return m.client.Close()
}

func options(cfg config.Redis) *goredis.Options {
	return &goredis.Options{
		Addr:            cfg.Addr,
		Password:        cfg.Password,
		DB:              cfg.DB,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		MaxIdleConns:    cfg.MaxIdleConns,
		ConnMaxIdleTime: cfg.IdleTimeout, // idle connections are closed after this
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		MaxRetries:      cfg.MaxRetries,
	}
}
