// internal/auth/jwt.go
package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrNoAuthHeader      = errors.New("authorization header required")
	ErrInvalidAuthHeader = errors.New("invalid authorization header format")
)

// ExtractBearerToken extracts a bearer token from the Authorization header.
func ExtractBearerToken(r *http.Request) (string, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", ErrNoAuthHeader
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", ErrInvalidAuthHeader
	}

	return parts[1], nil
}

// NewClaims creates a standard set of JWT claims for a user.
func NewClaims(userID uuid.UUID, email, issuer string, duration time.Duration) *Claims {
	exp := time.Now().Add(duration)
	return &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
		},
	}
}

// SignAccessToken signs provided claims with the given secret and returns the token string.
func SignAccessToken(secret string, claims *Claims) (string, error) {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(secret))
}

// ParseAccessToken parses and validates a token string using the secret and returns the claims.
func ParseAccessToken(tokenStr, secret string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return claims, nil
}
