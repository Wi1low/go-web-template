package auth

import (
	"testing"
	"time"

	"github.com/Wi1low/go-web-template/internal/config"
)

func TestHSParserRoundTrip(t *testing.T) {
	cfg := &config.JWT{
		SigningKey:     "unit-test-secret",
		Issuer:         "template",
		ExpireDuration: time.Hour,
		SigningMethod:  "HS256",
	}
	p, err := NewTokenParser(cfg)
	if err != nil {
		t.Fatal(err)
	}

	token, err := p.Issue(42)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := p.Parse(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != 42 {
		t.Fatalf("claims.UserID = %d, want 42", claims.UserID)
	}
}

func TestExpiredTokenRejected(t *testing.T) {
	cfg := &config.JWT{
		SigningKey:     "unit-test-secret",
		ExpireDuration: -time.Hour, // deterministic: already expired at issue
		SigningMethod:  "HS256",
	}
	p, err := NewTokenParser(cfg)
	if err != nil {
		t.Fatal(err)
	}
	token, err := p.Issue(1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.Parse(token); err == nil {
		t.Fatal("expired token accepted")
	}
}

func TestEmptyKeyRejected(t *testing.T) {
	if _, err := NewTokenParser(&config.JWT{SigningMethod: "HS256"}); err == nil {
		t.Fatal("empty signing key should fail")
	}
}
