package security

import (
	"errors"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

var ErrPasswordPolicy = errors.New("password must be 12 to 72 bytes and must not contain leading or trailing whitespace")

func ValidatePassword(password string) error {
	bytes := len([]byte(password))
	if !utf8.ValidString(password) || bytes < 12 || bytes > 72 || strings.TrimSpace(password) != password {
		return ErrPasswordPolicy
	}
	return nil
}

func HashPassword(password string, cost int) (string, error) {
	if err := ValidatePassword(password); err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(hash), err
}

func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
