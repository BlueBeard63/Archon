package crypto

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeriveKey(t *testing.T) {
	tests := []struct {
		name   string
		apiKey string
		salt   []byte
	}{
		{
			name:   "consistent results with same inputs",
			apiKey: "test-api-key-12345",
			salt:   []byte("1234567890123456"), // 16 bytes
		},
		{
			name:   "longer api key",
			apiKey: "a-very-long-api-key-that-is-more-than-32-characters-long",
			salt:   []byte("abcdefghijklmnop"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key1, err := DeriveKey(tt.apiKey, tt.salt)
			require.NoError(t, err)
			assert.Len(t, key1, 32) // AES-256 requires 32 bytes

			// Same inputs should produce same key
			key2, err := DeriveKey(tt.apiKey, tt.salt)
			require.NoError(t, err)
			assert.Equal(t, key1, key2)
		})
	}
}

func TestDeriveKey_DifferentInputs(t *testing.T) {
	salt := []byte("1234567890123456")

	// Different API keys produce different keys
	key1, err := DeriveKey("api-key-1", salt)
	require.NoError(t, err)

	key2, err := DeriveKey("api-key-2", salt)
	require.NoError(t, err)

	assert.NotEqual(t, key1, key2)
}

func TestDeriveKey_DifferentSalts(t *testing.T) {
	apiKey := "same-api-key"

	key1, err := DeriveKey(apiKey, []byte("salt111111111111"))
	require.NoError(t, err)

	key2, err := DeriveKey(apiKey, []byte("salt222222222222"))
	require.NoError(t, err)

	assert.NotEqual(t, key1, key2)
}

func TestEncrypt(t *testing.T) {
	apiKey := "test-api-key-for-encryption"

	tests := []struct {
		name      string
		plaintext string
	}{
		{name: "short string", plaintext: "hello"},
		{name: "empty string", plaintext: ""},
		{name: "long string", plaintext: strings.Repeat("a", 1000)},
		{name: "unicode", plaintext: "こんにちは世界"},
		{name: "special chars", plaintext: "!@#$%^&*()_+-=[]{}|;':\",./<>?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := Encrypt(tt.plaintext, apiKey)
			require.NoError(t, err)

			// Should be valid base64
			_, err = base64.StdEncoding.DecodeString(ciphertext)
			require.NoError(t, err)

			// Different encryption of same plaintext should produce different ciphertext (random nonce)
			ciphertext2, err := Encrypt(tt.plaintext, apiKey)
			require.NoError(t, err)
			if tt.plaintext != "" { // Empty string might produce similar looking results due to padding
				assert.NotEqual(t, ciphertext, ciphertext2)
			}
		})
	}
}

func TestDecrypt(t *testing.T) {
	apiKey := "test-api-key-for-decryption"
	plaintext := "secret-password-123"

	// Encrypt first
	ciphertext, err := Encrypt(plaintext, apiKey)
	require.NoError(t, err)

	// Then decrypt
	decrypted, err := Decrypt(ciphertext, apiKey)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestDecrypt_WrongKey(t *testing.T) {
	plaintext := "secret-data"

	ciphertext, err := Encrypt(plaintext, "correct-api-key")
	require.NoError(t, err)

	// Try to decrypt with wrong key
	_, err = Decrypt(ciphertext, "wrong-api-key")
	assert.Error(t, err)
}

func TestDecrypt_CorruptedCiphertext(t *testing.T) {
	apiKey := "test-key"
	plaintext := "test-data"

	ciphertext, err := Encrypt(plaintext, apiKey)
	require.NoError(t, err)

	// Corrupt the ciphertext
	corrupted := ciphertext[:len(ciphertext)-5] + "XXXXX"

	_, err = Decrypt(corrupted, apiKey)
	assert.Error(t, err)
}

func TestDecrypt_TruncatedCiphertext(t *testing.T) {
	apiKey := "test-key"

	// Too short to contain salt + nonce + ciphertext
	_, err := Decrypt("dG9vLXNob3J0", apiKey) // "too-short" in base64
	assert.Error(t, err)
}

func TestDecrypt_InvalidBase64(t *testing.T) {
	apiKey := "test-key"

	_, err := Decrypt("not-valid-base64!!!", apiKey)
	assert.Error(t, err)
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
		apiKey    string
	}{
		{
			name:      "empty string",
			plaintext: "",
			apiKey:    "key1",
		},
		{
			name:      "short string",
			plaintext: "hi",
			apiKey:    "key2",
		},
		{
			name:      "medium string",
			plaintext: "This is a medium length password",
			apiKey:    "key3",
		},
		{
			name:      "long string",
			plaintext: strings.Repeat("password", 100),
			apiKey:    "key4",
		},
		{
			name:      "unicode characters",
			plaintext: "密码パスワード암호",
			apiKey:    "key5",
		},
		{
			name:      "special characters",
			plaintext: `!@#$%^&*()_+-={}[]|\:";'<>?,./~` + "`",
			apiKey:    "key6",
		},
		{
			name:      "newlines and tabs",
			plaintext: "line1\nline2\ttabbed",
			apiKey:    "key7",
		},
		{
			name:      "binary-like data",
			plaintext: string([]byte{0x00, 0x01, 0x02, 0xff, 0xfe, 0xfd}),
			apiKey:    "key8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := Encrypt(tt.plaintext, tt.apiKey)
			require.NoError(t, err)

			decrypted, err := Decrypt(ciphertext, tt.apiKey)
			require.NoError(t, err)

			assert.Equal(t, tt.plaintext, decrypted)
		})
	}
}
