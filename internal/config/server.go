package config

import "time"

type Mode string

const ModeDev Mode = "dev"
const ModeRelease Mode = "release"
const ModeTest Mode = "test"

type Server struct {
	Addr         string        `mapstructure:"addr"`
	ReadTimeout  time.Duration `mapstructure:"read-timeout"`
	WriteTimeout time.Duration `mapstructure:"write-timeout"`
	Mode         Mode          `mapstructure:"mode"` // debug, release, test
}
