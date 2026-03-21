package auth

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashToken applies a SHA-256 hash to the raw refresh token string.
// This is used before saving the token to the database and looking it up
// to prevent raw token exposure if the database is compromised.
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
