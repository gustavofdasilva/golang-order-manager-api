package security

import (
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {

	bytes, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		14,
	)

	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func ComparePassword(hash string, password string) error {

	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// CompareToken checks if a raw token matches a hashed token.
func CompareToken(rawToken, hashedToken string) bool {
	return HashToken(rawToken) == hashedToken
}
