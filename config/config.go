package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/plannet-ai/plannet/security"
)

// Config represents the Plannet configuration
type Config struct {
	TicketPrefixes []string          `json:"ticket_prefixes"`
	Editor         string            `json:"editor"`
	GitIntegration bool              `json:"git_integration"`
	Headers        map[string]string `json:"headers,omitempty"`
	BaseURL        string            `json:"base_url,omitempty"`
	Model          string            `json:"model,omitempty"`
	SystemPrompt   string            `json:"system_prompt,omitempty"`
	JiraURL        string            `json:"jira_url,omitempty"`
	JiraUser       string            `json:"jira_user,omitempty"`
	CopyPreference CopyPreference    `json:"copy_preference,omitempty"`
	// JiraToken is no longer stored directly in the config
	// It will be stored securely using the TokenStorage
}

var (
	// Global config instance
	globalConfig *Config
	// Config file path
	configPath string
	// Token storage
	tokenStorage *security.TokenStorage
	// Base directory for file operations
	baseDir string
)

func init() {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// If we can't get the home directory, use the current directory
		homeDir = "."
	}
	configPath = filepath.Join(homeDir, ".plannetrc")
	baseDir = homeDir

	// Initialize token storage
	var initErr error
	tokenStorage, initErr = security.NewTokenStorage()
	if initErr != nil {
		fmt.Printf("Warning: Failed to initialize secure token storage: %v\n", initErr)
	}
}

// Load loads the configuration from the .plannetrc file
func Load() (*Config, error) {
	// If config is already loaded, return it
	if globalConfig != nil {
		return globalConfig, nil
	}

	// Check if config exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found. Run 'plannet init' to create one")
	}

	// Read the config file safely
	configData, err := security.SafeReadFile(baseDir, configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file: %w", err)
	}

	// Parse the config
	config := &Config{}
	if err := json.Unmarshal(configData, config); err != nil {
		return nil, fmt.Errorf("error parsing configuration: %w", err)
	}

	// Store the config globally
	globalConfig = config
	return config, nil
}

// Save saves the configuration to the .plannetrc file
func Save(config *Config) error {
	// Convert config to JSON
	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error creating configuration: %w", err)
	}

	// Write config to file safely
	if err := security.SafeWriteFile(baseDir, configPath, configJSON, 0644); err != nil {
		return fmt.Errorf("error writing configuration file: %w", err)
	}

	// Update global config
	globalConfig = config
	return nil
}

// Get returns the current configuration
func Get() (*Config, error) {
	if globalConfig == nil {
		return Load()
	}
	return globalConfig, nil
}

// GetConfigPath returns the path to the configuration file
func GetConfigPath() string {
	return configPath
}

// SetConfigPath sets the path to the configuration file
// This is primarily used for testing
func SetConfigPath(path string) {
	configPath = path
	globalConfig = nil // Reset the global config to force a reload
}

// IsInitialized checks if Plannet is initialized
func IsInitialized() bool {
	_, err := os.Stat(configPath)
	return !os.IsNotExist(err)
}

// GetJiraToken retrieves the Jira API token from secure storage
func GetJiraToken() (string, error) {
	if tokenStorage == nil {
		return "", fmt.Errorf("token storage not initialized")
	}
	return tokenStorage.GetToken("jira")
}

// SetJiraToken stores the Jira API token in secure storage
func SetJiraToken(token string) error {
	if tokenStorage == nil {
		return fmt.Errorf("token storage not initialized")
	}
	return tokenStorage.StoreToken("jira", token)
}

// GetLLMToken retrieves the LLM API token from secure storage
func GetLLMToken() (string, error) {
	if tokenStorage == nil {
		return "", fmt.Errorf("token storage not initialized")
	}
	return tokenStorage.GetToken("llm")
}

// SetLLMToken stores the LLM API token in secure storage
func SetLLMToken(token string) error {
	if tokenStorage == nil {
		return fmt.Errorf("token storage not initialized")
	}
	return tokenStorage.StoreToken("llm", token)
}
