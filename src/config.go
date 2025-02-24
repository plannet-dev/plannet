package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	// LLM Configuration
	BaseURL      string            `json:"base_url"`
	Model        string            `json:"model"`
	SystemPrompt string            `json:"system_prompt"`
	Headers      map[string]string `json:"headers"`

	// Jira Configuration (all optional)
	JiraURL   string `json:"jira_url,omitempty"`
	JiraToken string `json:"jira_token,omitempty"`
	JiraUser  string `json:"jira_user,omitempty"`

	// User Preferences for copying go here
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		BaseURL: "http://localhost:1234/v1/completions",
		Headers: make(map[string]string),
	}
}

// LoadConfig reads and parses the configuration file
func LoadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	// Start with default configuration
	config := DefaultConfig()

	// Read configuration file if it exists
	if fileExists(configPath) {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}

		if err := json.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("error parsing config file: %w", err)
		}
	}

	// Validate the configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// SaveConfig writes the configuration to the config file
func (c *Config) SaveConfig() error {
	configPath, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.BaseURL == "" {
		return fmt.Errorf("base_url is required")
	}

	if c.Model == "" {
		return fmt.Errorf("model is required")
	}

	// Validate Jira configuration if any Jira field is set
	if c.JiraURL != "" || c.JiraToken != "" || c.JiraUser != "" {
		if c.JiraURL == "" {
			return fmt.Errorf("jira_url is required when using Jira integration")
		}
		if c.JiraToken == "" {
			return fmt.Errorf("jira_token is required when using Jira integration")
		}
		if c.JiraUser == "" {
			return fmt.Errorf("jira_user is required when using Jira integration")
		}
	}

	return nil
}

// getConfigPath returns the path to the configuration file
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %w", err)
	}

	return filepath.Join(homeDir, ".plannetrc"), nil
}

// fileExists checks if a file exists and is not a directory
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
