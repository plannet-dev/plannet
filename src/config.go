package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// CopyPreference represents user preferences for copy behavior.
type CopyPreference struct {
	value string
}

// Predefined CopyPreferences (Strongly Typed)
var (
	AskEveryTime      = CopyPreference{"ask-every-time"}     // Always prompt before copying
	AskOnce           = CopyPreference{"ask-once"}           // Ask once per session
	CopyAutomatically = CopyPreference{"copy-automatically"} // Always copy without prompting
	DoNotCopy         = CopyPreference{"do-not-copy"}        // Never copy
)

// AllowedValues returns all valid enum options.
func AllowedValues() []CopyPreference {
	return []CopyPreference{AskEveryTime, AskOnce, CopyAutomatically, DoNotCopy}
}

// IsValid checks if the CopyPreference is valid.
func (c CopyPreference) IsValid() bool {
	for _, v := range AllowedValues() {
		if c == v {
			return true
		}
	}
	return false
}

// Default returns the recommended default value.
func DefaultCopyPreference() CopyPreference {
	return AskEveryTime
}

// String returns the string representation.
func (c CopyPreference) String() string {
	return c.value
}

// ParseCopyPreference safely converts a string to a CopyPreference.
func ParseCopyPreference(s string) (CopyPreference, error) {
	for _, v := range AllowedValues() {
		if v.value == s {
			return v, nil
		}
	}
	return CopyPreference{}, errors.New("invalid copy preference")
}

// JSON Serialization Support
func (c CopyPreference) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CopyPreference) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseCopyPreference(s)
	if err != nil {
		return err
	}
	*c = parsed
	return nil
}

// Config represents the application configuration
type Config struct {
	// LLM Configuration
	BaseURL      string            `json:"base-url"`
	Model        string            `json:"model"`
	SystemPrompt string            `json:"system-prompt"`
	Headers      map[string]string `json:"headers"`

	// Jira Configuration (all optional)
	JiraURL   string `json:"jira-url,omitempty"`
	JiraToken string `json:"jira-token,omitempty"`
	JiraUser  string `json:"jira-user,omitempty"`

	// User Preferences for copying
	CopyPreference CopyPreference `json:"copy-preference"`
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		BaseURL:        "http://localhost:1234/v1/completions",
		Headers:        make(map[string]string),
		CopyPreference: DefaultCopyPreference(), // Set default copy behavior
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

	// Validate CopyPreference
	if !c.CopyPreference.IsValid() {
		return fmt.Errorf("invalid copy_preference: %s", c.CopyPreference)
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
