package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// Password hashing and verification utilities.

// HashPassword creates a secure hash from a plaintext password using Argon2id.
func HashPassword(password string) (string, error) {
	if len(password) < 6 {
		return "", errors.New("password must be at least 6 characters")
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	return fmt.Sprintf("$argon2id$v=19$m=65536,t=1,p=4$%s$%s", b64Salt, b64Hash), nil
}

// VerifyPassword checks a password against an Argon2id hash.
func VerifyPassword(password, encodedHash string) bool {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return bcryptVerify(password, encodedHash)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(hash) != 32 {
		return false
	}

	computed := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return subtle.ConstantTimeCompare(hash, computed) == 1
}

// bcryptVerify falls back to bcrypt verification for legacy hashes.
func bcryptVerify(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// PasswordStrength checks if a password meets minimum requirements.
func PasswordStrength(password string) error {
	if len(password) < 6 {
		return errors.New("password must be at least 6 characters")
	}
	if len(password) > 128 {
		return errors.New("password must be at most 128 characters")
	}
	return nil
}

// GenerateAPIKey creates a cryptographically random API key.
func GenerateAPIKey() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "vort-" + base64.RawURLEncoding.EncodeToString(buf), nil
}

// GenerateSecret creates a random hex string for TOTP and other secrets.
func GenerateSecret(length int) (string, error) {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", buf), nil
}
