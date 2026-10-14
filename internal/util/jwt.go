package util

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateRandomToken creates a cryptographically secure hex-encoded string
func GenerateJWT(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
