package config

import "time"

// CORS configures the cross-origin middleware.
type CORS struct {
	Enabled          bool          `mapstructure:"enabled"`
	AllowOrigins     []string      `mapstructure:"allow-origins"`     // e.g. ["https://example.com"]
	AllowMethods     []string      `mapstructure:"allow-methods"`     // e.g. ["GET", "POST"]
	AllowHeaders     []string      `mapstructure:"allow-headers"`     // e.g. ["Content-Type", "Authorization"]
	ExposeHeaders    []string      `mapstructure:"expose-headers"`    // headers readable by the browser
	AllowCredentials bool          `mapstructure:"allow-credentials"` // send cookies / auth headers
	MaxAge           time.Duration `mapstructure:"max-age"`           // preflight cache duration
}
