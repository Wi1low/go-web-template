package config

type Config struct {
	Server     Server     `mapstructure:"server"`
	Database   Database   `mapstructure:"database"`
	Redis      Redis      `mapstructure:"redis"`
	Log        Log        `mapstructure:"log"`
	App        App        `mapstructure:"app"`
	Jwt        JWT        `mapstructure:"jwt"`
	Cors       CORS       `mapstructure:"cors"`
	Prometheus Prometheus `mapstructure:"prometheus"`
}
