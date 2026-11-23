package config

import "time"

// Redis configures the go-redis v9 client.
type Redis struct {
	Enabled  bool   `mapstructure:"enabled"`
	Addr     string `mapstructure:"addr"` // host:port
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`

	PoolSize     int           `mapstructure:"pool-size"`
	MinIdleConns int           `mapstructure:"min-idle-conns"`
	MaxIdleConns int           `mapstructure:"max-idle-conns"`
	IdleTimeout  time.Duration `mapstructure:"idle-timeout"`
	DialTimeout  time.Duration `mapstructure:"dial-timeout"`
	ReadTimeout  time.Duration `mapstructure:"read-timeout"`
	WriteTimeout time.Duration `mapstructure:"write-timeout"`
	MaxRetries   int           `mapstructure:"max-retries"`

	MasterName    string   `mapstructure:"master-name"`    // sentinel primary name
	SentinelAddrs []string `mapstructure:"sentinel-addrs"` // sentinel addresses
	ClusterMode   bool     `mapstructure:"cluster-mode"`   // redis cluster client
}
