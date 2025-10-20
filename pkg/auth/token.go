package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

const (
	tokenPrefix = "rlgl_"
	tokenLength = 32
)

func GenerateToken() (string, error) {
	randomBytes := make([]byte, tokenLength)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	encoded := base64.RawURLEncoding.EncodeToString(randomBytes)

	return tokenPrefix + encoded, nil
}

func ValidateTokenFormat(token string) bool {
	if len(token) < len(tokenPrefix)+1 {
		return false
	}

	return token[:len(tokenPrefix)] == tokenPrefix
}
