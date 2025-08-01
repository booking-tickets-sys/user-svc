package token

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashToken creates a SHA-256 hash of the token for secure storage
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// ValidateTokenHash validates if a token matches its hash
func ValidateTokenHash(token, hash string) bool {
	expectedHash := HashToken(token)
	return expectedHash == hash
}

// GenerateTokenHash generates a hash for a token and returns both token and hash
func GenerateTokenHash(token string) (string, string) {
	hash := HashToken(token)
	return token, hash
}
