package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"vortexuipro/internal/database"
)

// ─── Constants ───────────────────────────────────────────────────────

const (
	AESKeySize = 32 // AES-256
)

// ─── BackupCryptoService ─────────────────────────────────────────────

// BackupCryptoService handles encryption and decryption of backup archives.
type BackupCryptoService struct{}

// NewBackupCryptoService creates a new backup crypto service.
func NewBackupCryptoService() *BackupCryptoService {
	return &BackupCryptoService{}
}

// GenerateKey generates a new random AES-256 key and returns it as base64.
func (s *BackupCryptoService) GenerateKey() (string, error) {
	key := make([]byte, AESKeySize)
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("generate key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// SaveKey saves a new encryption key to the database.
func (s *BackupCryptoService) SaveKey(name, keyData string) (*database.BackupEncryptionKey, error) {
	// Deactivate all existing keys
	database.DB.Model(&database.BackupEncryptionKey{}).Where("active = ?", true).Update("active", false)

	key := &database.BackupEncryptionKey{
		Name:    name,
		KeyData: keyData,
		Active:  true,
	}
	if err := database.DB.Create(key).Error; err != nil {
		return nil, fmt.Errorf("save key: %w", err)
	}
	return key, nil
}

// GetActiveKey returns the active encryption key.
func (s *BackupCryptoService) GetActiveKey() (*database.BackupEncryptionKey, error) {
	var key database.BackupEncryptionKey
	if err := database.DB.Where("active = ?", true).First(&key).Error; err != nil {
		return nil, fmt.Errorf("no active key found: %w", err)
	}
	return &key, nil
}

// ListKeys returns all encryption keys.
func (s *BackupCryptoService) ListKeys() ([]database.BackupEncryptionKey, error) {
	var keys []database.BackupEncryptionKey
	if err := database.DB.Order("created_at desc").Find(&keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}

// DeleteKey deletes an encryption key by ID.
func (s *BackupCryptoService) DeleteKey(id int64) error {
	return database.DB.Delete(&database.BackupEncryptionKey{}, id).Error
}

// EncryptFile encrypts a file using AES-256-GCM and writes to outPath.
// Returns the encryption metadata (key ID, nonce) needed for decryption.
// The encrypted file format: [nonce (12 bytes)][ciphertext]
func (s *BackupCryptoService) EncryptFile(inPath, outPath string, keyID int64) error {
	var keyRecord database.BackupEncryptionKey
	if err := database.DB.First(&keyRecord, keyID).Error; err != nil {
		return fmt.Errorf("encryption key not found: %w", err)
	}

	key, err := base64.StdEncoding.DecodeString(keyRecord.KeyData)
	if err != nil {
		return fmt.Errorf("decode key: %w", err)
	}

	if len(key) != AESKeySize {
		return fmt.Errorf("invalid key size: %d", len(key))
	}

	// Read plaintext
	plaintext, err := os.ReadFile(inPath)
	if err != nil {
		return fmt.Errorf("read input file: %w", err)
	}

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := aesGCM.Seal(nil, nonce, plaintext, nil)

	// Write: nonce + ciphertext
	outFile, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer outFile.Close()

	if _, err := outFile.Write(nonce); err != nil {
		return fmt.Errorf("write nonce: %w", err)
	}
	if _, err := outFile.Write(ciphertext); err != nil {
		return fmt.Errorf("write ciphertext: %w", err)
	}

	return nil
}

// DecryptFile decrypts a file encrypted with EncryptFile.
func (s *BackupCryptoService) DecryptFile(inPath, outPath string, keyID int64) error {
	var keyRecord database.BackupEncryptionKey
	if err := database.DB.First(&keyRecord, keyID).Error; err != nil {
		return fmt.Errorf("encryption key not found: %w", err)
	}

	key, err := base64.StdEncoding.DecodeString(keyRecord.KeyData)
	if err != nil {
		return fmt.Errorf("decode key: %w", err)
	}

	// Read encrypted data
	ciphertext, err := os.ReadFile(inPath)
	if err != nil {
		return fmt.Errorf("read encrypted file: %w", err)
	}

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("create GCM: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return fmt.Errorf("ciphertext too short")
	}

	nonce := ciphertext[:nonceSize]
	encrypted := ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := aesGCM.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return fmt.Errorf("decrypt: %w", err)
	}

	if err := os.WriteFile(outPath, plaintext, 0644); err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	return nil
}
