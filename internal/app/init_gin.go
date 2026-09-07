package app

import (
	"net/http"

	"github.com/Wi1low/go-web-template/internal/config"
	"github.com/Wi1low/go-web-template/internal/middleware"
	"github.com/gin-gonic/gin"
	ginprom "github.com/logocomune/gin-prometheus"
	"go.uber.org/zap"
)

func initGin() *gin.Engine {
	if AppConfig.Server.Mode == config.ModeDev {
		gin.SetMode(gin.DebugMode)
	}
	gin.SetMode(gin.ReleaseMode)

	e := gin.New()
	
	// register global middlewares
	e.Use(middleware.ZapLogger(zap.L()))
	e.Use(middleware.RecoveryMiddleware())

	// checking the gin health
	e.GET("/ping", Ping)

	// metrics
	if AppConfig.Prometheus.Enabled {
		e.Use(ginprom.Middleware(
			ginprom.WithFilterRoutes([]string{ // 过滤不需要监控的路由
				"/health",
				"/metrics",
			})))
		// 暴露 /metrics 端点供 Prometheus 拉取
		e.GET("/metrics", gin.WrapH(ginprom.GetMetricHandler()))

	}
	return e
}

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
