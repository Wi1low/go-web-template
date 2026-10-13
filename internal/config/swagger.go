package config

// Swagger configures the OpenAPI UI served from the generated docs package.
type Swagger struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"` // e.g. /swagger/*any
}
