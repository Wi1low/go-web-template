package middleware

import (
	"strings"

	"github.com/Wi1low/go-web-template/internal/auth"
	"github.com/Wi1low/go-web-template/pkg/response"
	"github.com/gin-gonic/gin"
)

// CtxUserID stores the authenticated user id in the gin context.
const CtxUserID = "user_id"

// Auth guards routes with a bearer JWT.
func Auth(parser auth.TokenParser) gin.HandlerFunc {
	const prefix = "Bearer "
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, prefix) {
			response.Unauthorized(c, "missing Authorization: Bearer <token> header")
			c.Abort()
			return
		}
		claims, err := parser.Parse(strings.TrimPrefix(header, prefix))
		if err != nil {
			response.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}
		c.Set(CtxUserID, claims.UserID)
		c.Next()
	}
}

// CurrentUserID reads the authenticated user id set by Auth.
func CurrentUserID(c *gin.Context) (uint64, bool) {
	v, ok := c.Get(CtxUserID)
	if !ok {
		return 0, false
	}
	id, ok := v.(uint64)
	return id, ok
}
