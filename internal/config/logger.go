package config

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Log 暂时采用硬编码模式 不启用
type Log struct {
	Level            string                 `mapstructure:"level"`              // 日志级别：debug, info, warn, error, dpanic, panic, fatal
	Encoding         string                 `mapstructure:"encoding"`           // 输出格式：json 或 console
	OutputPaths      []string               `mapstructure:"output-paths"`       // 输出路径，如 stdout, stderr 或文件路径
	ErrorOutputPaths []string               `mapstructure:"error-output-paths"` // 错误输出路径
	InitialFields    map[string]interface{} `mapstructure:"initial-fields"`     // 每条日志都会包含的固定字段
	EnableSampling   *bool                  `mapstructure:"enable-sampling"`    // 是否开启采样（生产环境建议开启）
	Sampling         *SamplingConfig        `mapstructure:"sampling"`           // 采样配置
	EnableRotate     bool                   `mapstructure:"enable-rotate"`      // 是否启用日志轮转（需配合 lumberjack）
	RotateConfig     *RotateConfig          `mapstructure:"rotate-config"`      // 轮转配置
}

type SamplingConfig struct {
	Initial    int `mapstructure:"initial"`    // 初始采样数
	Thereafter int `mapstructure:"thereafter"` // 后续每 N 条采样一条
}

type RotateConfig struct {
	Filename   string `mapstructure:"filename"`    // 日志文件路径
	MaxSize    int    `mapstructure:"max-size"`    // 每个日志文件最大尺寸（MB）
	MaxBackups int    `mapstructure:"max-backups"` // 保留的旧日志文件最大数量
	MaxAge     int    `mapstructure:"max-age"`     // 保留旧文件的最大天数
	Compress   bool   `mapstructure:"compress"`    // 是否压缩
}

// NewLogger 根据环境创建 logger
func NewLogger(mode Mode) (*zap.Logger, error) {
	var logger *zap.Logger
	var err error

	switch mode {
	case ModeRelease:
		logger, err = newProductionLogger()
	default:
		logger, err = newDevelopmentLogger()
	}

	if err != nil {
		return nil, err
	}

	// TODO 添加服务名等全局字段
	logger = logger.With(zap.String("service", "my-api"))

	return logger, nil
}

// newDevelopmentLogger 构建开发环境 logger
func newDevelopmentLogger() (*zap.Logger, error) {
	// 1. 编码器：console 格式，方便阅读
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // 更可读的时间格式
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	// 2. 输出目标：stdout
	writeSyncer := zapcore.AddSync(os.Stdout)

	// 3. 日志级别：Debug
	level := zap.NewAtomicLevelAt(zapcore.DebugLevel)

	// 4. 创建 Core
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// 5. 构建 Logger，并启用调用者信息（文件名:行号）
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger, nil
}

// newProductionLogger 构建生产环境 logger
func newProductionLogger() (*zap.Logger, error) {
	// 1. 编码器：JSON 格式
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // 使用 ISO8601 时间格式
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// 2. 输出目标：同时输出到控制台和文件（带切割）
	// 控制台
	consoleWriter := zapcore.AddSync(os.Stdout)
	// 文件（使用 lumberjack）
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "./logs/app.log",
		MaxSize:    100,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	})

	// 3. 为不同输出设置不同级别（可选）
	// 比如：控制台只输出 Warn 及以上，文件输出 Info 及以上
	// 这里为了简单，统一使用 Info 级别
	level := zap.NewAtomicLevelAt(zapcore.InfoLevel)

	// 4. 创建多个 Core，并用 Tee 组合
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, consoleWriter, level),
		zapcore.NewCore(encoder, fileWriter, level),
	)

	// 5. 构建 Logger，启用调用者和堆栈
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger, nil
}
