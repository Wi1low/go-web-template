package middleware

import (
	"runtime/debug"

	"github.com/Wi1low/go-web-template/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery turns panics inside handlers into a 500 response instead of a
// crashed process, logging the stack with the trace id for correlation.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				zap.L().Error("panic recovered",
					zap.String("trace_id", TraceID(c)),
					zap.Any("error", err),
					zap.ByteString("stack", debug.Stack()),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)
				response.InternalServerError(c, "")
				c.Abort()
			}
		}()
		c.Next()
	}
}
