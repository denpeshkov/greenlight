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

func (a *AuthService) CreateToken(ctx context.Context, userID int64) (token string, err error) {
	now := time.Now()
	claims := &jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(userID, 10),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
		Issuer:    "github.com./denpeshkov/greenlight",
		Audience:  []string{"github.com./denpeshkov/greenlight"},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(a.secret))
}

func (a *AuthService) ParseToken(tokenString string) (userID int64, err error) {
	token, err := jwt.Parse(
		tokenString,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(a.secret), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuer("github.com./denpeshkov/greenlight"),
		jwt.WithAudience("github.com./denpeshkov/greenlight"),
	)

	if err != nil {
		return 0, NewUnauthorizedError("Invalid or missing authentication token.")
	}
	return strconv.ParseInt(token.Claims.(*jwt.RegisteredClaims).Subject, 10, 64)
}
