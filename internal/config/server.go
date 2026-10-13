package config

import "time"

// Mode selects the deployment environment. One value drives the config file
// name (config.<env>.yaml), the logger encoding and the gin mode.
type Mode string

const (
	ModeDev  Mode = "dev"
	ModeTest Mode = "test"
	ModeProd Mode = "prod"
)

// Server configures the main HTTP listener.
type Server struct {
	Addr            string        `mapstructure:"addr"`
	Mode            Mode          `mapstructure:"mode"`
	ReadTimeout     time.Duration `mapstructure:"read-timeout"`
	WriteTimeout    time.Duration `mapstructure:"write-timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown-timeout"` // graceful shutdown budget
}
