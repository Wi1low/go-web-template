package config

import "time"

// CORSConfig 跨域配置
type CORS struct {
	AllowOrigins     []string      `mapstructure:"allow-origins"`     // 允许的源，如 ["https://example.com"]
	AllowMethods     []string      `mapstructure:"allow-methods"`     // 允许的 HTTP 方法，如 ["GET", "POST"]
	AllowHeaders     []string      `mapstructure:"allow-headers"`     // 允许的请求头，如 ["Content-Type", "Authorization"]
	ExposeHeaders    []string      `mapstructure:"expose-headers"`    // 暴露给客户端的响应头
	AllowCredentials bool          `mapstructure:"allow-credentials"` // 是否允许携带凭证（如 Cookies）
	MaxAge           time.Duration `mapstructure:"max-age"`           // 预检请求的缓存时间
}
