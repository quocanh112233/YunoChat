package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const Issuer = "yunochat"

var (
	ErrInvalidToken = errors.New("invalid or expired token")
)

// GenerateAccessToken generates a new JWT token for the user.
// Claims contained: sub (userID), exp (expiration time), iat (issued at).
func GenerateAccessToken(userID uuid.UUID, secret string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		Issuer:    Issuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken parses the token and extracts the subject (userID).
func ParseToken(tokenString string, secret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the alg is what we expect: HS256
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return uuid.Nil, ErrInvalidToken
	}

	// Double-check Issuer just in case
	if claims.Issuer != Issuer {
		return uuid.Nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	return userID, nil
}
