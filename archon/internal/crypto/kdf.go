package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// SaltSize is the size of the salt in bytes
	SaltSize = 16
	// NonceSize is the size of the GCM nonce in bytes
	NonceSize = 12
	// KeySize is the size of the derived key (AES-256)
	KeySize = 32
	// PBKDF2Iterations is the number of iterations for key derivation
	PBKDF2Iterations = 100000
)

// DeriveKey derives a 32-byte key from an API key and salt using PBKDF2
func DeriveKey(apiKey string, salt []byte) ([]byte, error) {
	if len(salt) < SaltSize {
		return nil, fmt.Errorf("salt must be at least %d bytes", SaltSize)
	}
	return pbkdf2.Key([]byte(apiKey), salt, PBKDF2Iterations, KeySize, sha256.New), nil
}

// Encrypt encrypts plaintext using AES-256-GCM with a key derived from the API key
// Returns base64-encoded: salt[16] + nonce[12] + ciphertext
func Encrypt(plaintext, apiKey string) (string, error) {
	// Generate random salt
	salt := make([]byte, SaltSize)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive key from API key and salt
	key, err := DeriveKey(apiKey, salt)
	if err != nil {
		return "", fmt.Errorf("failed to derive key: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the plaintext
	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

	// Combine salt + nonce + ciphertext
	result := make([]byte, 0, SaltSize+NonceSize+len(ciphertext))
	result = append(result, salt...)
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	return base64.StdEncoding.EncodeToString(result), nil
}

// Decrypt decrypts base64-encoded ciphertext using AES-256-GCM with a key derived from the API key
// Expects format: salt[16] + nonce[12] + ciphertext
func Decrypt(ciphertextB64, apiKey string) (string, error) {
	// Decode base64
	data, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Check minimum length (salt + nonce + at least empty ciphertext with auth tag)
	minLen := SaltSize + NonceSize + 16 // GCM auth tag is 16 bytes
	if len(data) < minLen {
		return "", fmt.Errorf("ciphertext too short")
	}

	// Extract salt, nonce, and ciphertext
	salt := data[:SaltSize]
	nonce := data[SaltSize : SaltSize+NonceSize]
	ciphertext := data[SaltSize+NonceSize:]

	// Derive key from API key and salt
	key, err := DeriveKey(apiKey, salt)
	if err != nil {
		return "", fmt.Errorf("failed to derive key: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	return string(plaintext), nil
}
