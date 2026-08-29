package security

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yudadh/finance-bot-tracker-api/internal/config"
)

type AdminClaims struct {
	AdminUserID uint64 `json:"admin_user_id"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func GenerateAdminToken(cfg config.JWTConfig, adminUserID uint64, email string) (string, error) {
	now := time.Now()

	claims := AdminClaims{
		AdminUserID: adminUserID,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: cfg.Issuer,
			Subject: email,
			IssuedAt: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.TTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(cfg.Secret))
	if err != nil {
		return "", err
	}

	return t, err
}

func ParseAdminToken(cfg config.JWTConfig, jwtToken string) (*AdminClaims, error) {
	token, err := jwt.ParseWithClaims(jwtToken, &AdminClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidSigningMethod
		}
		return []byte(cfg.Secret), nil
	})

	if err != nil {
		return nil, err
	}
	
	claims, ok := token.Claims.(*AdminClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}