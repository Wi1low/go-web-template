package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/Wi1low/go-web-template/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RecoveryMiddleware 自定义恢复中间件
// 捕获程序运行时的 panic，防止服务崩溃，并返回适当的错误响应
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录错误信息和堆栈信息
				stack := string(debug.Stack())
				errorMsg := fmt.Sprintf("%v", err)
				zap.L().Error("程序发生 panic",
					zap.String("error", errorMsg),
					zap.String("stack", stack),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)

				// 返回 500 错误响应
				response.InternalServerError(c, "服务器内部错误")
				c.Abort()
			}
		}()

		c.Next()
	}
}
