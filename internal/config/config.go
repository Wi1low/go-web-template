// Package config defines the application configuration schema and its loading
// rules (YAML file per environment + .env + APP_-prefixed environment
// overrides).
package config

import "fmt"

// Config is the root configuration object.
type Config struct {
	App        App        `mapstructure:"app"`
	Server     Server     `mapstructure:"server"`
	Database   Database   `mapstructure:"database"`
	Redis      Redis      `mapstructure:"redis"`
	Log        Log        `mapstructure:"log"`
	Jwt        JWT        `mapstructure:"jwt"`
	Cors       CORS       `mapstructure:"cors"`
	Prometheus Prometheus `mapstructure:"prometheus"`
	Swagger    Swagger    `mapstructure:"swagger"`
}

// Validate checks invariants that modules rely on unconditionally.
func (c *Config) Validate() error {
	switch c.Server.Mode {
	case ModeDev, ModeTest, ModeProd:
	default:
		return fmt.Errorf("config: invalid server.mode %q (want dev, test or prod)", c.Server.Mode)
	}
	if c.Server.Addr == "" {
		return fmt.Errorf("config: server.addr must not be empty")
	}
	if c.Prometheus.Enabled && c.Prometheus.Port == 0 {
		return fmt.Errorf("config: prometheus.port must be set when prometheus is enabled")
	}
	if c.Database.Enabled && c.Database.DSN == "" {
		return fmt.Errorf("config: database.dsn must not be empty when the database is enabled")
	}
	if c.Redis.Enabled && c.Redis.Addr == "" {
		return fmt.Errorf("config: redis.addr must not be empty when redis is enabled")
	}
	return nil
}
