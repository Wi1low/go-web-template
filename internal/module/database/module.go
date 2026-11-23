// Package database provides the GORM module. Adding another SQL engine means
// one dialector case below and a driver dependency — nothing else changes.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/Wi1low/go-web-template/internal/config"
	"github.com/Wi1low/go-web-template/internal/kernel"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/prometheus"
)

type module struct {
	db *gorm.DB
}

// New returns the database module.
func New() kernel.Module { return &module{} }

func (*module) Name() string { return "database" }

func (*module) Enabled(c *config.Config) bool { return c.Database.Enabled }

func (m *module) Init(ctx context.Context, a *kernel.App) error {
	cfg := a.Cfg.Database

	dial, err := dialector(cfg)
	if err != nil {
		return err
	}
	db, err := gorm.Open(dial, &gorm.Config{
		Logger: logger.New(&zapWriter{log: zap.L()}, logger.Config{
			SlowThreshold:             cfg.SlowThreshold,
			LogLevel:                  gormLogLevel(cfg.LogLevel),
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		}),
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   cfg.TablePrefix,
			SingularTable: cfg.SingularTable,
		},
		SkipDefaultTransaction: cfg.SkipDefaultTx,
		PrepareStmt:            cfg.PrepareStmt,
	})
	if err != nil {
		return fmt.Errorf("open %s: %w", cfg.Driver, err)
	}

	if cfg.Prometheus && a.Cfg.Prometheus.Enabled {
		// Registers dbstats gauges on the default prometheus registry with
		// Register (not MustRegister), so repeated boots never panic.
		if err := db.Use(prometheus.New(prometheus.Config{
			DBName:          cfg.Dbname,
			RefreshInterval: 15,
		})); err != nil {
			zap.L().Error("register gorm prometheus plugin failed", zap.Error(err))
		}
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		return fmt.Errorf("database unreachable (%s driver): %w", cfg.Driver, err)
	}

	m.db = db
	kernel.Provide(a, db)
	zap.L().Info("database connected", zap.String("driver", cfg.Driver))
	return nil
}

func (m *module) Stop(context.Context) error {
	if m.db == nil {
		return nil
	}
	sqlDB, err := m.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// dialector maps config.driver to a GORM dialector. To support another
// engine, add a case and the matching gorm.io/driver dependency.
func dialector(cfg config.Database) (gorm.Dialector, error) {
	switch cfg.Driver {
	case "mysql", "":
		return mysql.Open(cfg.DSN), nil
	// case "postgres":
	// 	return postgres.Open(cfg.DSN), nil
	// case "sqlite":
	// 	return sqlite.Open(cfg.DSN), nil
	default:
		return nil, fmt.Errorf("unsupported database.driver %q", cfg.Driver)
	}
}

func gormLogLevel(level string) logger.LogLevel {
	switch level {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	default:
		return logger.Info
	}
}

// zapWriter adapts zap to gorm's logger.Writer interface.
type zapWriter struct {
	log *zap.Logger
}

func (w *zapWriter) Printf(format string, args ...any) {
	w.log.Sugar().Infof(format, args...)
}
