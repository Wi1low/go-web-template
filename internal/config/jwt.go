package config

import "time"

// JWTConfig JWT 配置
type JWT struct {
	SigningKey     string        `mapstructure:"signing-key"`     // 签名密钥（必填）
	Issuer         string        `mapstructure:"issuer"`          // 签发者
	ExpireDuration time.Duration `mapstructure:"expire-duration"` // 过期时长（如 24h）
	BufferTime     time.Duration `mapstructure:"buffer-time"`     // 缓冲时间（用于自动刷新）
	SigningMethod  string        `mapstructure:"signing-method"`  // 签名方法，如 HS256、RS256
	// 如果使用 RSA，可以增加私钥/公钥路径
	PrivateKeyPath string `mapstructure:"private-key-path"` // RSA 私钥路径
	PublicKeyPath  string `mapstructure:"public-key-path"`  // RSA 公钥路径
}
