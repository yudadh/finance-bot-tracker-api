package security

import (
	"strings"
	"testing"
	"time"

	"github.com/yudadh/finance-bot-tracker-api/internal/config"
)

func TestGenerateAndParseAdminToken(t *testing.T) {
	cfg := config.JWTConfig{
		Secret: "test-secret",
		Issuer: "finance-bot-test",
		TTL:    time.Hour,
	}

	token, err := GenerateAdminToken(cfg, 42, "admin@example.com")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	claims, err := ParseAdminToken(cfg, token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}

	if claims.AdminUserID != 42 {
		t.Fatalf("expected admin user ID 42, got %d", claims.AdminUserID)
	}

	if claims.Email != "admin@example.com" {
		t.Fatalf("expected email admin@example.com, got %q", claims.Email)
	}

	if claims.Issuer != cfg.Issuer {
		t.Fatalf("expected issuer %q, got %q", cfg.Issuer, claims.Issuer)
	}

	if claims.ExpiresAt == nil || !claims.ExpiresAt.After(claims.IssuedAt.Time) {
		t.Fatal("expected token expiration after issue time")
	}
}

func TestGenerateAndParseAdminToken_InvalidIssuer(t *testing.T) {
	cfg := config.JWTConfig{
		Secret: "test-secret",
		Issuer: "finance-bot-test",
		TTL:    time.Hour,
	}

	token, err := GenerateAdminToken(cfg, 42, "admin@example.com")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	cfg.Issuer = "invalid-issuer"
	_, err = ParseAdminToken(cfg, token)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

}

func TestGenerateAndParseAdminToken_MalformedToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "single segment",
			token: "abc",
		},
		{
			name:  "two segments",
			token: "abc.def",
		},
		{
			name:  "too many segments",
			token: "abc.def.ghi.jkl",
		},
		{
			name:  "empty payload segment",
			token: "abc..def",
		},
		{
			name:  "invalid base64 header",
			token: "@@@.eyJzdWIiOiIxIn0.signature",
		},
		{
			name:  "invalid base64 payload",
			token: "eyJhbGciOiJIUzI1NiJ9.@@@.signature",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := config.JWTConfig{
				Secret: "test-secret",
				Issuer: "finance-bot-test",
				TTL:    time.Hour,
			}
		
			_, err := ParseAdminToken(cfg, test.token)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestGenerateAndParseAdminToken_TamperedToken(t *testing.T) {
	cfg := config.JWTConfig{
		Secret: "test-secret",
		Issuer: "finance-bot-test",
		TTL:    time.Hour,
	}

	token, err := GenerateAdminToken(cfg, 42, "admin@example.com")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	stringLength := len(token)
	stringToReplace := token[stringLength-5:stringLength]
	
	tamperedToken := strings.Replace(token, stringToReplace, "abcde", 1)

	_, err = ParseAdminToken(cfg, tamperedToken)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGenerateAndParseAdminToken_ExpiredToken(t *testing.T) {
	cfg := config.JWTConfig{
		Secret: "test-secret",
		Issuer: "finance-bot-test",
		TTL:    time.Second,
	}

	token, err := GenerateAdminToken(cfg, 42, "admin@example.com")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	
	time.Sleep(time.Second * 2)
	_, err = ParseAdminToken(cfg, token)
	if err == nil {
		t.Fatal("expected error")
	}

}
func TestGenerateAndParseAdminToken_WrongSecretToken(t *testing.T) {
	cfg := config.JWTConfig{
		Secret: "test-secret",
		Issuer: "finance-bot-test",
		TTL:    time.Second,
	}

	token, err := GenerateAdminToken(cfg, 42, "admin@example.com")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	
	cfg.Secret = "wrong-secret"
	_, err = ParseAdminToken(cfg, token)
	if err == nil {
		t.Fatal("expected error")
	}
}
