package greenlight

import (
	"context"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AuthService is a service for managing authentication.
type AuthService struct {
	secret string
}

func NewAuthService(secret string) *AuthService {
	return &AuthService{
		secret: secret,
	}
}

func (a *AuthService) Create(ctx context.Context, id int64) (token string, err error) {
	now := time.Now()
	claims := &jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(id, 10),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
		Issuer:    "github.com./denpeshkov/greenlight",
		Audience:  []string{"github.com./denpeshkov/greenlight"},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(a.secret))
}
