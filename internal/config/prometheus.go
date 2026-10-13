package config

// Prometheus configures the standalone metrics server. All application
// metrics (gin requests, gorm dbstats, Go runtime) land on the default
// registry and are exposed from a single /metrics endpoint on this port.
type Prometheus struct {
	Enabled   bool   `mapstructure:"enabled"`
	Port      int    `mapstructure:"port"`      // admin port; keep off the app port
	Path      string `mapstructure:"path"`      // e.g. /metrics
	Namespace string `mapstructure:"namespace"` // metric name prefix, e.g. myapp
}
