package app

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func initPrometheus() {
	zap.L().Info("Initializing Prometheus")
	// 启动 Prometheus 指标服务器 (独立端口)
	if AppConfig.Prometheus.Enabled {
		go func() {
			metricsEngine := gin.New()
			// 使用 promhttp.Handler() 暴露指标
			metricsEngine.GET(AppConfig.Prometheus.Path, gin.WrapH(promhttp.Handler()))
			zap.L().Info("starting metrics server", zap.Int("port", AppConfig.Prometheus.Port))
			if err := metricsEngine.Run(fmt.Sprintf(":%d", AppConfig.Prometheus.Port)); err != nil {
				zap.L().Error("metrics server failed", zap.Error(err))
			}
		}()
	}
}
