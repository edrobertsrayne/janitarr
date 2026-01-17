package crypto

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateKey(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	// AES-256 requires 32-byte key
	if len(key) != 32 {
		t.Errorf("GenerateKey() returned %d bytes, want 32", len(key))
	}

	// Keys should be random - generate another and verify they're different
	key2, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() second call error = %v", err)
	}

	if string(key) == string(key2) {
		t.Error("GenerateKey() returned same key twice - randomness issue")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	testCases := []struct {
		name      string
		plaintext string
	}{
		{"empty string", ""},
		{"simple text", "hello world"},
		{"api key format", "abc123def456ghi789"},
		{"unicode", "日本語テスト"},
		{"long text", strings.Repeat("x", 1000)},
		{"special chars", "!@#$%^&*()_+-=[]{}|;':\",./<>?"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encrypted, err := Encrypt(tc.plaintext, key)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			decrypted, err := Decrypt(encrypted, key)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if decrypted != tc.plaintext {
				t.Errorf("round-trip failed: got %q, want %q", decrypted, tc.plaintext)
			}
		})
	}
}

func TestEncryptFormat(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	plaintext := "test-api-key"
	encrypted, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Format should be IV_BASE64:CIPHERTEXT_BASE64
	parts := strings.Split(encrypted, ":")
	if len(parts) != 2 {
		t.Errorf("Encrypt() format = %q, want IV_BASE64:CIPHERTEXT_BASE64", encrypted)
	}

	if parts[0] == "" || parts[1] == "" {
		t.Errorf("Encrypt() has empty parts: IV=%q, ciphertext=%q", parts[0], parts[1])
	}

	// Same plaintext should produce different ciphertext (due to random IV)
	encrypted2, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt() second call error = %v", err)
	}

	if encrypted == encrypted2 {
		t.Error("Encrypt() produced identical output for same input - IV not random")
	}
}

func TestDecryptWrongKey(t *testing.T) {
	key1, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	key2, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() second call error = %v", err)
	}

	plaintext := "secret-api-key"
	encrypted, err := Encrypt(plaintext, key1)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Decrypting with wrong key should fail
	_, err = Decrypt(encrypted, key2)
	if err == nil {
		t.Error("Decrypt() with wrong key should return error")
	}
}

func TestDecryptInvalidFormat(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	testCases := []struct {
		name       string
		ciphertext string
	}{
		{"empty string", ""},
		{"no delimiter", "abc123"},
		{"empty IV", ":ciphertext"},
		{"empty ciphertext", "iv:"},
		{"invalid base64 IV", "not-base64!:Y2lwaGVydGV4dA=="},
		{"invalid base64 ciphertext", "YWJjMTIz:not-base64!"},
		{"too many colons", "a:b:c"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Decrypt(tc.ciphertext, key)
			if err == nil {
				t.Errorf("Decrypt(%q) should return error", tc.ciphertext)
			}
		})
	}
}

func TestLoadOrCreateKey_NewKey(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test.key")

	key, err := LoadOrCreateKey(keyPath)
	if err != nil {
		t.Fatalf("LoadOrCreateKey() error = %v", err)
	}

	if len(key) != 32 {
		t.Errorf("LoadOrCreateKey() returned %d bytes, want 32", len(key))
	}

	// File should be created
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Error("LoadOrCreateKey() did not create key file")
	}
}

func TestLoadOrCreateKey_ExistingKey(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test.key")

	// Create key first time
	key1, err := LoadOrCreateKey(keyPath)
	if err != nil {
		t.Fatalf("LoadOrCreateKey() first call error = %v", err)
	}

	// Load existing key
	key2, err := LoadOrCreateKey(keyPath)
	if err != nil {
		t.Fatalf("LoadOrCreateKey() second call error = %v", err)
	}

	// Keys should be identical
	if string(key1) != string(key2) {
		t.Error("LoadOrCreateKey() returned different key for same file")
	}
}

func TestLoadOrCreateKey_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "subdir", "nested", "test.key")

	key, err := LoadOrCreateKey(keyPath)
	if err != nil {
		t.Fatalf("LoadOrCreateKey() error = %v", err)
	}

	if len(key) != 32 {
		t.Errorf("LoadOrCreateKey() returned %d bytes, want 32", len(key))
	}

	// File should exist
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Error("LoadOrCreateKey() did not create key file in nested directory")
	}
}

func TestLoadOrCreateKey_CorruptedFile(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test.key")

	// Write invalid key data (wrong length)
	if err := os.WriteFile(keyPath, []byte("short"), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := LoadOrCreateKey(keyPath)
	if err == nil {
		t.Error("LoadOrCreateKey() should return error for corrupted key file")
	}
}
