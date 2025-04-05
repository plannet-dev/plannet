package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	JiraToken      string            `json:"jira_token,omitempty"`
	JiraUser       string            `json:"jira_user,omitempty"`
	CopyPreference CopyPreference    `json:"copy_preference,omitempty"`
}

var (
	// Global config instance
	globalConfig *Config
	// Config file path
	configPath string
)

func init() {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// If we can't get the home directory, use the current directory
		homeDir = "."
	}
	configPath = filepath.Join(homeDir, ".plannetrc")
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

	// Read the config file
	configData, err := os.ReadFile(configPath)
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

	// Write config to file
	if err := os.WriteFile(configPath, configJSON, 0644); err != nil {
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

// IsInitialized checks if Plannet is initialized
func IsInitialized() bool {
	_, err := os.Stat(configPath)
	return !os.IsNotExist(err)
} 