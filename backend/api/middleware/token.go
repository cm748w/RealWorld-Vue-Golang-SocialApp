package middleware

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenTTL is how long an issued access token stays valid.
const TokenTTL = 24 * time.Hour

// ErrNoSecret is returned when JWT_SECRET is not configured.
var ErrNoSecret = errors.New("JWT_SECRET is not configured")

// IssueToken signs an HS256 token whose issuer is the user id.
//
// The issuer is the only identity claim the API trusts (AuthMiddleware reads it
// back into Locals("userId")), so nothing else is embedded. Signing with an
// empty key is refused outright: see JWTSecret for why that matters.
func IssueToken(userID string) (string, error) {
	secret, ok := JWTSecret()
	if !ok {
		return "", ErrNoSecret
	}
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    userID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(TokenTTL)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}
