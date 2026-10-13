package middleware

import (
	"time"

	"github.com/Wi1low/go-web-template/internal/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Cors builds the CORS middleware from configuration. Empty lists fall back
// to sensible defaults; with credentials enabled a wildcard origin is refused
// by the browser anyway, so set explicit allow-origins in that case.
func Cors(cfg *config.CORS) gin.HandlerFunc {
	c := cors.Config{
		AllowOrigins:     cfg.AllowOrigins,
		AllowMethods:     cfg.AllowMethods,
		AllowHeaders:     cfg.AllowHeaders,
		ExposeHeaders:    cfg.ExposeHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           cfg.MaxAge,
	}

	if len(c.AllowOrigins) == 0 {
		zap.L().Warn("cors: allow-origins not configured, allowing all origins")
		c.AllowOrigins = []string{"*"}
	}
	if len(c.AllowMethods) == 0 {
		c.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	}
	if len(c.AllowHeaders) == 0 {
		c.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Requested-With"}
	}
	if c.MaxAge == 0 {
		c.MaxAge = 12 * time.Hour
	}
	return cors.New(c)
}
