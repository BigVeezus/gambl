package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

var (
	// ErrEncryptionFailed indicates encryption failed
	ErrEncryptionFailed = errors.New("encryption failed")
	
	// ErrDecryptionFailedInternal indicates decryption failed
	ErrDecryptionFailedInternal = errors.New("decryption failed")
)

// CryptoEncrypter provides encryption and decryption functions
type CryptoEncrypter interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// AESEncrypter implements CryptoEncrypter using AES encryption
type AESEncrypter struct {
	secretKey []byte
}

// NewAESEncrypter creates a new AES encrypter with the given secret key
func NewAESEncrypter(secretKey string) *AESEncrypter {
	// Ensure the key is exactly 32 bytes (256 bits) for AES-256
	key := make([]byte, 32)
	
	// If the key is shorter than 32 bytes, it will be padded with zeros
	// If longer, it will be truncated
	copy(key, []byte(secretKey))
	
	return &AESEncrypter{
		secretKey: key,
	}
}

// Encrypt encrypts plaintext using AES-GCM
func (e *AESEncrypter) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(e.secretKey)
	if err != nil {
		return "", err
	}
	
	// GCM is an authenticated encryption mode
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	
	// Create a random nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	
	// Encrypt and authenticate data
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	
	// Return base64-encoded result for storage
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext using AES-GCM
func (e *AESEncrypter) Decrypt(encryptedData string) (string, error) {
	// Decode base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", err
	}
	
	block, err := aes.NewCipher(e.secretKey)
	if err != nil {
		return "", err
	}
	
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	
	// Extract nonce from the front of the ciphertext
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", ErrDecryptionFailedInternal
	}
	
	nonce, encryptedText := ciphertext[:nonceSize], ciphertext[nonceSize:]
	
	// Decrypt the data
	plaintext, err := aesGCM.Open(nil, nonce, encryptedText, nil)
	if err != nil {
		return "", ErrDecryptionFailedInternal
	}
	
	return string(plaintext), nil
}