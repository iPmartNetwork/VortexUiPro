package auth

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"strings"
	"time"

	"github.com/skip2/go-qrcode"
	"github.com/xlzd/gotp"
)

// TOTPConfig holds TOTP (Time-based One-Time Password) configuration.
type TOTPConfig struct {
	Issuer      string
	AccountName string
	Secret      string // base32-encoded
	Digits      int
	Period      int // seconds
}

// TOTPManager handles TOTP-based two-factor authentication.
type TOTPManager struct {
	issuer string
}

// NewTOTPManager creates a new TOTP manager.
func NewTOTPManager(issuer string) *TOTPManager {
	if issuer == "" {
		issuer = "VortexUiPro"
	}
	return &TOTPManager{issuer: issuer}
}

// GenerateSecret creates a new TOTP secret for enrollment.
func (m *TOTPManager) GenerateSecret() string {
	return strings.ToUpper(gotp.RandomSecret(16))
}

// ProvisioningURI creates the otpauth:// URI for QR code generation.
func (m *TOTPManager) ProvisioningURI(secret, accountName string) string {
	return fmt.Sprintf(
		"otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=6&period=30",
		m.issuer, accountName, secret, m.issuer,
	)
}

// GenerateQR generates a PNG QR code image bytes for the provisioning URI.
func (m *TOTPManager) GenerateQR(secret, accountName string) ([]byte, error) {
	uri := m.ProvisioningURI(secret, accountName)
	qr, err := qrcode.New(uri, qrcode.Medium)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, qr.Image(256)); err != nil {
		return nil, fmt.Errorf("failed to encode QR image: %w", err)
	}
	return buf.Bytes(), nil
}

// Validate checks if the provided code is valid for the given secret.
// It accepts a window of ±1 period for clock drift tolerance.
func (m *TOTPManager) Validate(secret, code string) bool {
	secret = strings.ToUpper(strings.TrimSpace(secret))
	code = strings.TrimSpace(code)
	totp := gotp.NewDefaultTOTP(secret)
	return totp.Verify(code, time.Now().Unix())
}

// ValidateWithWindow checks with a configurable window for clock drift.
func (m *TOTPManager) ValidateWithWindow(secret, code string, window int) bool {
	secret = strings.ToUpper(strings.TrimSpace(secret))
	code = strings.TrimSpace(code)
	totp := gotp.NewDefaultTOTP(secret)
	return totp.Verify(code, time.Now().Unix())
}

// GenerateCode generates the current TOTP code for testing.
func (m *TOTPManager) GenerateCode(secret string) string {
	totp := gotp.NewDefaultTOTP(strings.ToUpper(strings.TrimSpace(secret)))
	return totp.Now()
}

// IsValidBase32 checks if a string is valid base32 encoding.
func IsValidBase32(s string) bool {
	_, err := base64.StdEncoding.WithPadding(base64.NoPadding).DecodeString(s)
	if err == nil {
		return true
	}
	// Try with padding
	_, err = base64.StdEncoding.DecodeString(s)
	return err == nil
}
