package utils

import (
	"os"
	"testing"
)

func TestAESEncryptionDecryption(t *testing.T) {
	// Set a known environment variable for the test to ensure consistency
	os.Setenv("AES_KEY", "61122b0510e39e95735f9d4d3f34b8f4afae15e5015c5576b6c772fe30d0a265")

	originalText := "06301268369"

	// 1. Test Encryption
	encryptedHex, err := EncryptAES(originalText)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}
	if encryptedHex == "" {
		t.Fatal("Encrypted string is empty")
	}
	if encryptedHex == originalText {
		t.Fatal("Encrypted string is identical to plaintext")
	}

	// 2. Test Decryption
	decryptedText, err := DecryptAES(encryptedHex)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}
	if decryptedText != originalText {
		t.Fatalf("Decrypted text does not match original. Expected %s, got %s", originalText, decryptedText)
	}
}

func TestDecryptAESUnencryptedFallback(t *testing.T) {
	// Test that our migration safety net works (unencrypted 11-digit string)
	unencryptedPESEL := "12345678901"

	result, err := DecryptAES(unencryptedPESEL)
	if err != nil {
		t.Fatalf("DecryptAES returned error for unencrypted string: %v", err)
	}
	if result != unencryptedPESEL {
		t.Fatalf("Expected fallback to return %s, got %s", unencryptedPESEL, result)
	}
}

func TestMaskPESEL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Valid 11 digit PESEL",
			input:    "06301268369",
			expected: "XXXXXXX8369",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Too short string",
			input:    "123",
			expected: "123", // Should return as-is
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := MaskPESEL(tc.input)
			if result != tc.expected {
				t.Fatalf("MaskPESEL(%s) = %s, expected %s", tc.input, result, tc.expected)
			}
		})
	}
}
