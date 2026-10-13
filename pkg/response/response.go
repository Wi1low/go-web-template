// Package response provides the uniform JSON envelope used by all handlers.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response is the standard envelope. TraceID is filled from the request
// context set by the RequestID middleware.
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	TraceID string `json:"trace_id,omitempty"`
}

// Success writes 200 with data.
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
		TraceID: c.GetString("trace_id"),
	})
}

// Fail writes an error envelope with the given HTTP status.
func Fail(c *gin.Context, code int, message string) {
	if message == "" {
		message = http.StatusText(code)
	}
	c.JSON(code, Response{
		Code:    code,
		Message: message,
		TraceID: c.GetString("trace_id"),
	})
}

// BadRequest writes a 400.
func BadRequest(c *gin.Context, message string) { Fail(c, http.StatusBadRequest, message) }

// Unauthorized writes a 401.
func Unauthorized(c *gin.Context, message string) { Fail(c, http.StatusUnauthorized, message) }

// Forbidden writes a 403.
func Forbidden(c *gin.Context, message string) { Fail(c, http.StatusForbidden, message) }

// NotFound writes a 404.
func NotFound(c *gin.Context, message string) { Fail(c, http.StatusNotFound, message) }

// InternalServerError writes a 500.
func InternalServerError(c *gin.Context, message string) {
	Fail(c, http.StatusInternalServerError, message)
}
