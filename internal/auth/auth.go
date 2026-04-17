package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GenerateToken создает криптостойкий токен из 32 байт в base64url.
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b), nil
}
