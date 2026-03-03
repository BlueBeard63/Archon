use aes_gcm::{
    aead::{Aead, KeyInit},
    Aes256Gcm, Nonce,
};
use base64::{engine::general_purpose::STANDARD as BASE64, Engine};
use pbkdf2::pbkdf2_hmac;
use rand::RngCore;
use sha2::Sha256;
use thiserror::Error;

const SALT_SIZE: usize = 16;
const NONCE_SIZE: usize = 12;
const KEY_SIZE: usize = 32;
const PBKDF2_ITERATIONS: u32 = 100_000;
const GCM_TAG_SIZE: usize = 16;

#[derive(Debug, Error)]
pub enum KdfError {
    #[error("salt must be at least {SALT_SIZE} bytes")]
    SaltTooShort,
    #[error("failed to generate random bytes")]
    RandomGeneration,
    #[error("encryption failed: {0}")]
    Encryption(String),
    #[error("decryption failed: {0}")]
    Decryption(String),
    #[error("ciphertext too short")]
    CiphertextTooShort,
    #[error("failed to decode base64: {0}")]
    Base64Decode(#[from] base64::DecodeError),
}

/// Derives a 32-byte key from an API key and salt using PBKDF2-HMAC-SHA256.
pub fn derive_key(api_key: &str, salt: &[u8]) -> Result<[u8; KEY_SIZE], KdfError> {
    if salt.len() < SALT_SIZE {
        return Err(KdfError::SaltTooShort);
    }
    let mut key = [0u8; KEY_SIZE];
    pbkdf2_hmac::<Sha256>(api_key.as_bytes(), salt, PBKDF2_ITERATIONS, &mut key);
    Ok(key)
}

/// Encrypts plaintext using AES-256-GCM with a key derived from the API key.
/// Returns base64-encoded: salt[16] || nonce[12] || ciphertext+tag
///
/// This produces output compatible with the Go implementation in
/// `archon/internal/crypto/kdf.go` and `node/internal/crypto/kdf.go`.
pub fn encrypt(plaintext: &str, api_key: &str) -> Result<String, KdfError> {
    let mut rng = rand::thread_rng();

    // Generate random salt
    let mut salt = [0u8; SALT_SIZE];
    rng.try_fill_bytes(&mut salt)
        .map_err(|_| KdfError::RandomGeneration)?;

    // Derive key
    let key = derive_key(api_key, &salt)?;

    // Create AES-256-GCM cipher
    let cipher = Aes256Gcm::new_from_slice(&key)
        .map_err(|e| KdfError::Encryption(e.to_string()))?;

    // Generate random nonce
    let mut nonce_bytes = [0u8; NONCE_SIZE];
    rng.try_fill_bytes(&mut nonce_bytes)
        .map_err(|_| KdfError::RandomGeneration)?;
    let nonce = Nonce::from_slice(&nonce_bytes);

    // Encrypt
    let ciphertext = cipher
        .encrypt(nonce, plaintext.as_bytes())
        .map_err(|e| KdfError::Encryption(e.to_string()))?;

    // Concatenate: salt || nonce || ciphertext+tag
    let mut result = Vec::with_capacity(SALT_SIZE + NONCE_SIZE + ciphertext.len());
    result.extend_from_slice(&salt);
    result.extend_from_slice(&nonce_bytes);
    result.extend_from_slice(&ciphertext);

    Ok(BASE64.encode(&result))
}

/// Decrypts base64-encoded ciphertext using AES-256-GCM with a key derived from the API key.
/// Expects format: salt[16] || nonce[12] || ciphertext+tag
///
/// Compatible with the Go implementation's Encrypt output.
pub fn decrypt(ciphertext_b64: &str, api_key: &str) -> Result<String, KdfError> {
    let data = BASE64.decode(ciphertext_b64)?;

    // Minimum length: salt + nonce + GCM auth tag
    let min_len = SALT_SIZE + NONCE_SIZE + GCM_TAG_SIZE;
    if data.len() < min_len {
        return Err(KdfError::CiphertextTooShort);
    }

    // Split components
    let salt = &data[..SALT_SIZE];
    let nonce_bytes = &data[SALT_SIZE..SALT_SIZE + NONCE_SIZE];
    let ciphertext = &data[SALT_SIZE + NONCE_SIZE..];

    // Derive key
    let key = derive_key(api_key, salt)?;

    // Create cipher and decrypt
    let cipher = Aes256Gcm::new_from_slice(&key)
        .map_err(|e| KdfError::Decryption(e.to_string()))?;
    let nonce = Nonce::from_slice(nonce_bytes);

    let plaintext = cipher
        .decrypt(nonce, ciphertext)
        .map_err(|e| KdfError::Decryption(e.to_string()))?;

    String::from_utf8(plaintext).map_err(|e| KdfError::Decryption(e.to_string()))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_encrypt_decrypt_roundtrip() {
        let api_key = "test-api-key-12345";
        let plaintext = "hello world";

        let encrypted = encrypt(plaintext, api_key).unwrap();
        let decrypted = decrypt(&encrypted, api_key).unwrap();

        assert_eq!(decrypted, plaintext);
    }

    #[test]
    fn test_encrypt_produces_different_output() {
        let api_key = "test-key";
        let plaintext = "same text";

        let enc1 = encrypt(plaintext, api_key).unwrap();
        let enc2 = encrypt(plaintext, api_key).unwrap();

        // Different random salt/nonce each time
        assert_ne!(enc1, enc2);

        // Both decrypt to the same value
        assert_eq!(decrypt(&enc1, api_key).unwrap(), plaintext);
        assert_eq!(decrypt(&enc2, api_key).unwrap(), plaintext);
    }

    #[test]
    fn test_wrong_key_fails() {
        let encrypted = encrypt("secret", "correct-key").unwrap();
        let result = decrypt(&encrypted, "wrong-key");
        assert!(result.is_err());
    }

    #[test]
    fn test_ciphertext_too_short() {
        let result = decrypt(&BASE64.encode(b"tooshort"), "key");
        assert!(matches!(result, Err(KdfError::CiphertextTooShort)));
    }

    #[test]
    fn test_empty_plaintext() {
        let api_key = "key";
        let encrypted = encrypt("", api_key).unwrap();
        let decrypted = decrypt(&encrypted, api_key).unwrap();
        assert_eq!(decrypted, "");
    }

    #[test]
    fn test_derive_key_salt_too_short() {
        let result = derive_key("key", &[0u8; 8]);
        assert!(matches!(result, Err(KdfError::SaltTooShort)));
    }
}
