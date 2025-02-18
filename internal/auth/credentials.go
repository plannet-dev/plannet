package auth

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/zalando/go-keyring"
)

const (
	servicePrefix = "plannet"
)

// Credentials stores authentication information
type Credentials struct {
	Username  string    `json:"username"`
	Token     string    `json:"token"`
	System    string    `json:"system"` // e.g., "jira", "asana"
	ExpiresAt time.Time `json:"expires_at"`
}

// CredentialManager handles secure storage and retrieval of credentials
type CredentialManager struct {
	system string
}

// NewCredentialManager creates a new credential manager for a specific system
func NewCredentialManager(system string) *CredentialManager {
	return &CredentialManager{
		system: system,
	}
}

// StoreCredentials saves credentials securely in the system keyring
func (cm *CredentialManager) StoreCredentials(creds *Credentials) error {
	// Ensure system matches
	if creds.System != cm.system {
		return fmt.Errorf("credential system mismatch: expected %s, got %s", cm.system, creds.System)
	}

	// Convert credentials to JSON
	credsJSON, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	// Store in system keyring
	service := fmt.Sprintf("%s-%s", servicePrefix, cm.system)
	err = keyring.Set(service, creds.Username, string(credsJSON))
	if err != nil {
		return fmt.Errorf("failed to store credentials: %w", err)
	}

	return nil
}

// GetCredentials retrieves credentials from the system keyring
func (cm *CredentialManager) GetCredentials(username string) (*Credentials, error) {
	service := fmt.Sprintf("%s-%s", servicePrefix, cm.system)
	credsJSON, err := keyring.Get(service, username)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credentials: %w", err)
	}

	var creds Credentials
	if err := json.Unmarshal([]byte(credsJSON), &creds); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	// Check if credentials have expired
	if !creds.ExpiresAt.IsZero() && time.Now().After(creds.ExpiresAt) {
		return nil, fmt.Errorf("credentials have expired")
	}

	return &creds, nil
}

// DeleteCredentials removes credentials from the system keyring
func (cm *CredentialManager) DeleteCredentials(username string) error {
	service := fmt.Sprintf("%s-%s", servicePrefix, cm.system)
	err := keyring.Delete(service, username)
	if err != nil {
		return fmt.Errorf("failed to delete credentials: %w", err)
	}
	return nil
}

// IsAuthenticated checks if valid credentials exist for the given username
func (cm *CredentialManager) IsAuthenticated(username string) bool {
	creds, err := cm.GetCredentials(username)
	if err != nil {
		return false
	}

	// Check expiration
	if !creds.ExpiresAt.IsZero() && time.Now().After(creds.ExpiresAt) {
		return false
	}

	return true
}

// Example usage in a ticket system:
/*
func (j *JiraTicketSystem) Authenticate(username, password string) error {
    // Perform Jira authentication and get token
    token, expiresAt, err := j.authenticateWithJira(username, password)
    if err != nil {
        return err
    }

    // Store credentials
    cm := auth.NewCredentialManager("jira")
    creds := &auth.Credentials{
        Username:  username,
        Token:     token,
        System:    "jira",
        ExpiresAt: expiresAt,
    }

    return cm.StoreCredentials(creds)
}
*/
