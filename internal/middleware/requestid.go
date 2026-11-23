package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// HeaderRequestID carries the trace id in and out of the service.
	HeaderRequestID = "X-Request-Id"
	// CtxTraceID is the gin context key read by pkg/response.
	CtxTraceID = "trace_id"
)

// RequestID must be the first middleware in the chain: every later stage
// (logger, recovery, response) relies on the trace id it installs. An inbound
// X-Request-Id header is honored so calls can be joined across services.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(HeaderRequestID)
		if id == "" {
			id = newTraceID()
		}
		c.Set(CtxTraceID, id)
		c.Header(HeaderRequestID, id)
		c.Next()
	}
}

func newTraceID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

// TraceID returns the request trace id.
func TraceID(c *gin.Context) string { return c.GetString(CtxTraceID) }
