package config

import "time"

// JWT configures the token parser in internal/auth. An empty signing key
// disables the authentication demo routes.
type JWT struct {
	SigningKey     string        `mapstructure:"signing-key"`     // HS* secret; inject via APP_JWT_SIGNING_KEY in production
	Issuer         string        `mapstructure:"issuer"`          // iss claim
	ExpireDuration time.Duration `mapstructure:"expire-duration"` // e.g. 24h
	BufferTime     time.Duration `mapstructure:"buffer-time"`     // refresh window (application-level)
	SigningMethod  string        `mapstructure:"signing-method"`  // HS256, HS384, HS512, RS256

	PrivateKeyPath string `mapstructure:"private-key-path"` // RS*: PEM private key
	PublicKeyPath  string `mapstructure:"public-key-path"`  // RS*: PEM public key
}
