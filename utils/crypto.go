package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"io"
	"os"
)

func GenerateRandomHash(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func getAESKey() []byte {
	keyStr := os.Getenv("AES_KEY")
	if keyStr == "" {
		return []byte("12345678901234567890123456789012")
	}

	if len(keyStr) == 64 {
		if decoded, err := hex.DecodeString(keyStr); err == nil {
			return decoded
		}
	}

	return []byte(keyStr)
}

func EncryptAES(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	block, err := aes.NewCipher(getAESKey())
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := aesgcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

func DecryptAES(ciphertextHex string) (string, error) {
	if ciphertextHex == "" {
		return "", nil
	}

	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil || len(ciphertext) < 12 {
		return ciphertextHex, nil
	}

	block, err := aes.NewCipher(getAESKey())
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := aesgcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return ciphertextHex, nil
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return ciphertextHex, nil
	}
	return string(plaintext), nil
}

func MaskPESEL(pesel string) string {
	if len(pesel) != 11 {
		return pesel
	}
	return "XXXXXXX" + pesel[7:]
}
