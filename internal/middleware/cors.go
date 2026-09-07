// Package middleware can be used in gin
package middleware

import (
	"time"

	"github.com/Wi1low/go-web-template/internal/config"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// NewCorsMiddleware 根据配置创建 CORS 中间件
func Cors(cfg *config.CORS) gin.HandlerFunc {
	// 构建 cors.Config
	c := cors.Config{
		AllowOrigins:     cfg.AllowOrigins,
		AllowMethods:     cfg.AllowMethods,
		AllowHeaders:     cfg.AllowHeaders,
		ExposeHeaders:    cfg.ExposeHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           cfg.MaxAge,
	}

	// 如果未配置 AllowOrigins，可以设置默认值或直接使用默认中间件
	if len(c.AllowOrigins) == 0 {
		logger.Info("CORS allowOrigins not configured, using default allow all origins")
		c.AllowOrigins = []string{"*"}
	}
	if len(c.AllowMethods) == 0 {
		// 默认常见方法
		c.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	}
	if len(c.AllowHeaders) == 0 {
		// 默认常见头
		c.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Requested-With"}
	}
	if c.MaxAge == 0 {
		c.MaxAge = 12 * time.Hour
	}

	// 使用配置创建中间件
	return cors.New(c)
}
