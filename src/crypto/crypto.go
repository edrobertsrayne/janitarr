package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// KeySize is the size of AES-256 keys in bytes
	KeySize = 32
	// IVLength is the recommended IV length for AES-GCM (12 bytes)
	IVLength = 12
)

// GenerateKey generates a new random 32-byte key for AES-256 encryption.
func GenerateKey() ([]byte, error) {
	key := make([]byte, KeySize)
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("generating random key: %w", err)
	}
	return key, nil
}

// LoadOrCreateKey loads an existing key from the given path, or creates a new one
// if the file doesn't exist. It also creates any necessary parent directories.
func LoadOrCreateKey(path string) ([]byte, error) {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("creating key directory: %w", err)
	}

	// Try to read existing key
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Generate new key
			key, err := GenerateKey()
			if err != nil {
				return nil, err
			}

			// Save to file with secure permissions
			if err := os.WriteFile(path, key, 0600); err != nil {
				return nil, fmt.Errorf("saving key file: %w", err)
			}

			return key, nil
		}
		return nil, fmt.Errorf("reading key file: %w", err)
	}

	// Validate key length
	if len(data) != KeySize {
		return nil, fmt.Errorf("invalid key file: expected %d bytes, got %d", KeySize, len(data))
	}

	return data, nil
}

// Encrypt encrypts plaintext using AES-256-GCM and returns the result as
// "IV_BASE64:CIPHERTEXT_BASE64".
func Encrypt(plaintext string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("creating cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("creating GCM: %w", err)
	}

	// Generate random IV
	iv := make([]byte, IVLength)
	if _, err := rand.Read(iv); err != nil {
		return "", fmt.Errorf("generating IV: %w", err)
	}

	// Encrypt
	ciphertext := aesGCM.Seal(nil, iv, []byte(plaintext), nil)

	// Encode as IV_BASE64:CIPHERTEXT_BASE64
	ivBase64 := base64.StdEncoding.EncodeToString(iv)
	ciphertextBase64 := base64.StdEncoding.EncodeToString(ciphertext)

	return ivBase64 + ":" + ciphertextBase64, nil
}

// Decrypt decrypts ciphertext in the format "IV_BASE64:CIPHERTEXT_BASE64"
// using AES-256-GCM.
func Decrypt(ciphertextWithIV string, key []byte) (string, error) {
	if ciphertextWithIV == "" {
		return "", errors.New("empty ciphertext")
	}

	parts := strings.Split(ciphertextWithIV, ":")
	if len(parts) != 2 {
		return "", errors.New("invalid ciphertext format: expected IV_BASE64:CIPHERTEXT_BASE64")
	}

	ivBase64, ciphertextBase64 := parts[0], parts[1]
	if ivBase64 == "" || ciphertextBase64 == "" {
		return "", errors.New("invalid ciphertext format: empty IV or ciphertext")
	}

	iv, err := base64.StdEncoding.DecodeString(ivBase64)
	if err != nil {
		return "", fmt.Errorf("decoding IV: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", fmt.Errorf("decoding ciphertext: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("creating cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("creating GCM: %w", err)
	}

	plaintext, err := aesGCM.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypting: %w", err)
	}

	return string(plaintext), nil
}
