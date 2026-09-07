package app

import (
	"log"

	"github.com/Wi1low/go-web-template/internal/config"
	"go.uber.org/zap"
)

func initLogger(mode config.Mode) {
	logger, err := config.NewLogger(mode)
	if err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	// 替换全局 logger
	zap.ReplaceGlobals(logger)
	// 可选：将标准库 log 也重定向到 zap（方便第三方包）
	// zap.RedirectStdLog(logger)
}
