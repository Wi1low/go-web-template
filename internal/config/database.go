package config

import (
	"fmt"
	"time"

	"github.com/Wi1low/go-web-template/internal/metrics"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type Database struct {
	Driver     string `mapstructure:"driver"` // mysql, postgres, sqlite
	DSN        string `mapstructure:"dsn"`    // 主数据源名称
	Dbname     string `mapstructure:"dbname"` // 数据库名称
	Prometheus bool   `mapstructure:"prometheus"`
	//ReadDSN         []string      `mapstructure:"read-dsn"`           // 只读实例 DSN 列表（读写分离用）
	MaxOpenConns    int           `mapstructure:"max-open-conns"`     // 最大打开连接数
	MaxIdleConns    int           `mapstructure:"max-idle-conns"`     // 最大空闲连接数
	ConnMaxLifetime time.Duration `mapstructure:"conn-max-lifetime"`  // 连接最大存活时间
	ConnMaxIdleTime time.Duration `mapstructure:"conn-max-idle-time"` // 连接最大空闲时间
	LogLevel        string        `mapstructure:"log-level"`          // GORM 日志级别：silent, error, warn, info
	SlowThreshold   time.Duration `mapstructure:"slow-threshold"`     // 慢查询阈值（如 200ms）

	// 高级参数（可选，用于更细粒度的调优）
	TablePrefix   string `mapstructure:"table-prefix"`    // 表前缀
	SingularTable bool   `mapstructure:"singular-table"`  // 是否使用单数表名
	PrepareStmt   bool   `mapstructure:"prepare-stmt"`    // 是否缓存预编译语句
	SkipDefaultTx bool   `mapstructure:"skip-default-tx"` // 是否跳过默认事务
}

func NewGORM(cfg Database, zapLogger *zap.Logger) (*gorm.DB, error) {
	// 配置 GORM Logger（对接 Zap）
	var logLevel logger.LogLevel
	switch cfg.LogLevel {
	case "silent":
		logLevel = logger.Silent
	case "error":
		logLevel = logger.Error
	case "warn":
		logLevel = logger.Warn
	default:
		logLevel = logger.Info
	}

	gormLogger := logger.New(
		&ZapWriter{logger: zapLogger}, // 自定义 Writer，将日志写入 Zap
		logger.Config{
			SlowThreshold:             time.Duration(cfg.SlowThreshold) * time.Millisecond,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// 打开数据库连接
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: gormLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: cfg.SingularTable, // 使用单数表名
		},
		SkipDefaultTransaction: cfg.SkipDefaultTx, // 关闭默认事务提升性能
		PrepareStmt:            cfg.PrepareStmt,   // 启用预编译缓存
	})
	if err != nil {
		return nil, fmt.Errorf("Connecting database failed and err: %w", err)
	}

	// 如果启用了 Prometheus，注册插件
	if cfg.Prometheus {
		if err := db.Use(metrics.GormPrometheusPlugin(cfg.Dbname)); err != nil {
			zapLogger.Error("register gorm prometheus plugin failed", zap.Error(err))
		}
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("Getting sql.DB failed and err: %w", err)
	}
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	return db, nil
}

// ZapWriter 实现 gorm.io/gorm/logger.Writer 接口
type ZapWriter struct {
	logger *zap.Logger
}

func (w *ZapWriter) Printf(format string, args ...interface{}) {
	w.logger.Sugar().Infof(format, args...)
}
