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
	keyFile := filepath.Join(homeDir, ".plannetrc")

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

// StoreToken securely stores a token in the config file
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

	// Define the path for the config file
	configPath := filepath.Join(homeDir, ".plannetrc")

	// Read the config file
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	// Parse the config
	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("error parsing config file: %w", err)
	}

	// Store the token in the config based on the key
	switch key {
	case "llm":
		config["llm_token"] = encodedToken
	case "jira":
		config["jira_token"] = encodedToken
	default:
		return fmt.Errorf("unknown token key: %s", key)
	}

	// Write the config back to the file
	updatedConfig, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	if err := os.WriteFile(configPath, updatedConfig, 0644); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

// GetToken retrieves a stored token from the config file
func (ts *TokenStorage) GetToken(key string) (string, error) {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error finding home directory: %w", err)
	}

	// Define the path for the config file
	configPath := filepath.Join(homeDir, ".plannetrc")

	// Read the config file
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("error reading config file: %w", err)
	}

	// Parse the config
	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		return "", fmt.Errorf("error parsing config file: %w", err)
	}

	// Get the token from the config based on the key
	var encodedToken string
	switch key {
	case "llm":
		if token, ok := config["llm_token"].(string); ok {
			encodedToken = token
		}
	case "jira":
		if token, ok := config["jira_token"].(string); ok {
			encodedToken = token
		}
	default:
		return "", fmt.Errorf("unknown token key: %s", key)
	}

	if encodedToken == "" {
		return "", fmt.Errorf("token not found for key: %s", key)
	}

	// Read the encryption key
	encryptionKey, err := os.ReadFile(ts.keyFile)
	if err != nil {
		return "", fmt.Errorf("error reading encryption key: %w", err)
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
