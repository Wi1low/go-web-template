// Package auth implements JWT issuing and parsing. It has no lifecycle of its
// own: the demo module builds a TokenParser from configuration and publishes
// it through the kernel container.
package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/Wi1low/go-web-template/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

// Claims is the token payload. Extend it per project needs.
type Claims struct {
	UserID uint64 `json:"uid"`
	jwt.RegisteredClaims
}

// TokenParser issues and validates tokens.
type TokenParser interface {
	Issue(userID uint64) (string, error)
	Parse(token string) (*Claims, error)
}

// NewTokenParser builds a parser from the jwt config section, supporting
// HS256/384/512 (shared secret) and RS256/384/512 (PEM key files).
func NewTokenParser(cfg *config.JWT) (TokenParser, error) {
	method := cfg.SigningMethod
	if method == "" {
		method = "HS256"
	}
	expire := cfg.ExpireDuration
	if expire == 0 {
		expire = 24 * time.Hour
	}
	base := jwt.RegisteredClaims{
		Issuer:    cfg.Issuer,
		Subject:   "auth",
		Audience:  jwt.ClaimStrings{cfg.Issuer},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expire)),
		NotBefore: jwt.NewNumericDate(time.Now()),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	switch method {
	case "HS256", "HS384", "HS512":
		if cfg.SigningKey == "" {
			return nil, fmt.Errorf("auth: jwt.signing-key is empty for %s", method)
		}
		return &hsParser{method: method, secret: []byte(cfg.SigningKey), base: base, ttl: expire}, nil
	case "RS256", "RS384", "RS512":
		priv, err := loadRSAPrivate(cfg.PrivateKeyPath)
		if err != nil {
			return nil, err
		}
		pub, err := loadRSAPublic(cfg.PublicKeyPath)
		if err != nil {
			return nil, err
		}
		return &rsParser{method: method, priv: priv, pub: pub, base: base}, nil
	default:
		return nil, fmt.Errorf("auth: unsupported jwt.signing-method %q", method)
	}
}

type hsParser struct {
	method string
	secret []byte
	base   jwt.RegisteredClaims
	ttl    time.Duration
}

func (p *hsParser) Issue(userID uint64) (string, error) {
	claims := Claims{UserID: userID, RegisteredClaims: p.base}
	now := time.Now()
	claims.RegisteredClaims.IssuedAt = jwt.NewNumericDate(now)
	claims.RegisteredClaims.NotBefore = jwt.NewNumericDate(now)
	claims.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(now.Add(p.ttl))
	return jwt.NewWithClaims(jwt.GetSigningMethod(p.method), claims).SignedString(p.secret)
}

func (p *hsParser) Parse(token string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(token, claims,
		func(*jwt.Token) (any, error) { return p.secret, nil },
		jwt.WithValidMethods([]string{p.method}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

type rsParser struct {
	method string
	priv   *rsa.PrivateKey
	pub    *rsa.PublicKey
	base   jwt.RegisteredClaims
}

func (p *rsParser) Issue(userID uint64) (string, error) {
	claims := Claims{UserID: userID, RegisteredClaims: p.base}
	return jwt.NewWithClaims(jwt.GetSigningMethod(p.method), claims).SignedString(p.priv)
}

func (p *rsParser) Parse(token string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(token, claims,
		func(*jwt.Token) (any, error) { return p.pub, nil },
		jwt.WithValidMethods([]string{p.method}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

func loadRSAPrivate(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("auth: read private key: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("auth: no PEM block in %s", path)
	}
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("auth: %s is not an RSA key", path)
		}
		return rsaKey, nil
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func loadRSAPublic(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("auth: read public key: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("auth: no PEM block in %s", path)
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("auth: %s is not an RSA public key", path)
	}
	return rsaPub, nil
}
