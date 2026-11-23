package config

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Log is a reserved section; the logger is currently mode-driven (dev console,
// prod JSON + rotation). Uncomment the schema in cmd/etc/config.example.yaml
// and consume these fields here when per-key log tuning is needed.
type Log struct {
	Level            string                 `mapstructure:"level"`              // debug, info, warn, error, dpanic, panic, fatal
	Encoding         string                 `mapstructure:"encoding"`           // json or console
	OutputPaths      []string               `mapstructure:"output-paths"`       // stdout, stderr or file paths
	ErrorOutputPaths []string               `mapstructure:"error-output-paths"` // error output paths
	InitialFields    map[string]interface{} `mapstructure:"initial-fields"`     // fields attached to every entry
	EnableSampling   *bool                  `mapstructure:"enable-sampling"`    // sampling is recommended in production
	Sampling         *SamplingConfig        `mapstructure:"sampling"`
	EnableRotate     bool                   `mapstructure:"enable-rotate"` // rotate via lumberjack
	RotateConfig     *RotateConfig          `mapstructure:"rotate-config"`
}

type SamplingConfig struct {
	Initial    int `mapstructure:"initial"`    // first N entries per second logged
	Thereafter int `mapstructure:"thereafter"` // then 1 per N
}

type RotateConfig struct {
	Filename   string `mapstructure:"filename"`    // log file path
	MaxSize    int    `mapstructure:"max-size"`    // MB per file
	MaxBackups int    `mapstructure:"max-backups"` // rotated files kept
	MaxAge     int    `mapstructure:"max-age"`     // days kept
	Compress   bool   `mapstructure:"compress"`    // gzip rotated files
}

// NewLogger builds a zap logger for the given mode. The service field is set
// from the app name so log aggregators can filter by it.
func NewLogger(mode Mode, service string) (*zap.Logger, error) {
	var (
		logger *zap.Logger
		err    error
	)
	switch mode {
	case ModeProd:
		logger, err = newProductionLogger()
	default:
		logger, err = newDevelopmentLogger()
	}
	if err != nil {
		return nil, err
	}
	if service != "" {
		logger = logger.With(zap.String("service", service))
	}
	return logger, nil
}

// newDevelopmentLogger: human-readable console output on stdout, debug level.
func newDevelopmentLogger() (*zap.Logger, error) {
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		zap.NewAtomicLevelAt(zapcore.DebugLevel),
	)
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)), nil
}

// newProductionLogger: JSON to stdout and to a rotating file.
func newProductionLogger() (*zap.Logger, error) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	consoleWriter := zapcore.AddSync(os.Stdout)
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "./logs/app.log",
		MaxSize:    100, // MB
		MaxBackups: 5,
		MaxAge:     30, // days
		Compress:   true,
	})

	level := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, consoleWriter, level),
		zapcore.NewCore(encoder, fileWriter, level),
	)
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)), nil
}
