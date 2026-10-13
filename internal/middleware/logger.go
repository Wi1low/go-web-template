package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ZapLogger logs one structured line per request, keyed by trace id. Level is
// chosen by status so error logs stand out: >=500 error, >=400 warn, else info.
func ZapLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		fields := []zap.Field{
			zap.String("trace_id", TraceID(c)),
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.String("ip", c.ClientIP()),
			zap.Duration("latency", time.Since(start)),
			zap.String("user_agent", c.Request.UserAgent()),
		}
		if errs := c.Errors.ByType(gin.ErrorTypePrivate); len(errs) > 0 {
			fields = append(fields, zap.String("errors", errs.String()))
		}

		status := c.Writer.Status()
		switch {
		case status >= 500:
			logger.Error("http request", fields...)
		case status >= 400:
			logger.Warn("http request", fields...)
		default:
			logger.Info("http request", fields...)
		}
	}
}
