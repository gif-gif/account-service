package security

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

func GenerateAPIKey() (string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate api key: %w", err)
	}
	return "acct_" + base64.RawURLEncoding.EncodeToString(randomBytes), nil
}

func HashAPIKey(apiKey string) string {
	sum := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(sum[:])
}

func VerifyAPIKey(apiKey string, hash string) bool {
	got := HashAPIKey(apiKey)
	return subtle.ConstantTimeCompare([]byte(got), []byte(hash)) == 1
}
