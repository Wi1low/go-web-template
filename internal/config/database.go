package config

import "time"

// Database configures the GORM connection. Connection construction lives in
// internal/module/database, not here.
type Database struct {
	Enabled bool   `mapstructure:"enabled"`
	Driver  string `mapstructure:"driver"` // mysql (postgres, sqlite: add a case in the module)
	DSN     string `mapstructure:"dsn"`    // user:pass@tcp(host:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	Dbname  string `mapstructure:"dbname"` // used as a label on the prometheus metrics
	// Prometheus exports dbstats gauges when the metrics module is enabled.
	Prometheus bool `mapstructure:"prometheus"`

	MaxOpenConns    int           `mapstructure:"max-open-conns"`
	MaxIdleConns    int           `mapstructure:"max-idle-conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn-max-lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn-max-idle-time"`

	LogLevel      string        `mapstructure:"log-level"`      // silent, error, warn, info
	SlowThreshold time.Duration `mapstructure:"slow-threshold"` // e.g. 200ms

	TablePrefix   string `mapstructure:"table-prefix"`
	SingularTable bool   `mapstructure:"singular-table"`
	PrepareStmt   bool   `mapstructure:"prepare-stmt"`    // cache prepared statements
	SkipDefaultTx bool   `mapstructure:"skip-default-tx"` // drop the per-write default transaction
}
