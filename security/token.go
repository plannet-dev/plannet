package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var (
	// ErrEncryptionKeyNotFound is returned when the encryption key is not found
	ErrEncryptionKeyNotFound = errors.New("encryption key not found")
)

// TokenStorage provides secure storage for API tokens
type TokenStorage struct {
	keyFile string
}

// NewTokenStorage creates a new TokenStorage instance
func NewTokenStorage() (*TokenStorage, error) {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error finding home directory: %w", err)
	}

	// Define the path for the key file
	keyFile := filepath.Join(homeDir, ".plannet_key")

	// Check if key file exists, if not create it
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		// Generate a random key
		key := make([]byte, 32) // 256 bits
		if _, err := io.ReadFull(rand.Reader, key); err != nil {
			return nil, fmt.Errorf("error generating encryption key: %w", err)
		}

		// Write the key to the file with restricted permissions
		if err := os.WriteFile(keyFile, key, 0600); err != nil {
			return nil, fmt.Errorf("error writing encryption key: %w", err)
		}
	}

	return &TokenStorage{
		keyFile: keyFile,
	}, nil
}

// StoreToken securely stores a token
func (ts *TokenStorage) StoreToken(key, token string) error {
	// Read the encryption key
	encryptionKey, err := os.ReadFile(ts.keyFile)
	if err != nil {
		return fmt.Errorf("error reading encryption key: %w", err)
	}

	// Create a new cipher block
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return fmt.Errorf("error creating cipher: %w", err)
	}

	// Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("error creating GCM: %w", err)
	}

	// Create a nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("error generating nonce: %w", err)
	}

	// Encrypt the token
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(token), nil)

	// Encode the ciphertext as base64
	encodedToken := base64.StdEncoding.EncodeToString(ciphertext)

	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error finding home directory: %w", err)
	}

	// Define the path for the token file
	tokenFile := filepath.Join(homeDir, ".plannet_tokens")

	// Read existing tokens
	tokens := make(map[string]string)
	if _, err := os.Stat(tokenFile); err == nil {
		data, err := os.ReadFile(tokenFile)
		if err != nil {
			return fmt.Errorf("error reading token file: %w", err)
		}

		if err := json.Unmarshal(data, &tokens); err != nil {
			return fmt.Errorf("error parsing token file: %w", err)
		}
	}

	// Add the new token
	tokens[key] = encodedToken

	// Write the tokens back to the file
	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling tokens: %w", err)
	}

	if err := os.WriteFile(tokenFile, data, 0600); err != nil {
		return fmt.Errorf("error writing token file: %w", err)
	}

	return nil
}

// GetToken retrieves a stored token
func (ts *TokenStorage) GetToken(key string) (string, error) {
	// Read the encryption key
	encryptionKey, err := os.ReadFile(ts.keyFile)
	if err != nil {
		return "", fmt.Errorf("error reading encryption key: %w", err)
	}

	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error finding home directory: %w", err)
	}

	// Define the path for the token file
	tokenFile := filepath.Join(homeDir, ".plannet_tokens")

	// Read the token file
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return "", fmt.Errorf("error reading token file: %w", err)
	}

	// Parse the tokens
	var tokens map[string]string
	if err := json.Unmarshal(data, &tokens); err != nil {
		return "", fmt.Errorf("error parsing token file: %w", err)
	}

	// Get the token
	encodedToken, ok := tokens[key]
	if !ok {
		return "", fmt.Errorf("token not found for key: %s", key)
	}

	// Decode the base64 token
	ciphertext, err := base64.StdEncoding.DecodeString(encodedToken)
	if err != nil {
		return "", fmt.Errorf("error decoding token: %w", err)
	}

	// Create a new cipher block
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("error creating cipher: %w", err)
	}

	// Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("error creating GCM: %w", err)
	}

	// Extract the nonce
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt the token
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("error decrypting token: %w", err)
	}

	return string(plaintext), nil
}
