package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

const CookieName = "starter_session"

func RandomToken() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}

func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func CSRFToken(secret, sessionToken string) (string, error) {
	if len(secret) < 32 || sessionToken == "" {
		return "", errors.New("session security is not configured")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(sessionToken))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func ValidCSRFToken(secret, sessionToken, candidate string) bool {
	expected, err := CSRFToken(secret, sessionToken)
	if err != nil {
		return false
	}
	expectedBytes, err := hex.DecodeString(expected)
	if err != nil {
		return false
	}
	candidateBytes, err := hex.DecodeString(strings.TrimSpace(candidate))
	return err == nil && hmac.Equal(expectedBytes, candidateBytes)
}
