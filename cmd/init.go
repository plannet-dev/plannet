package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"plannet/internal/systems/jira"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// Config represents the .plannetrc configuration
type Config struct {
	SystemType string `json:"system_type"`
	BaseURL    string `json:"base_url"`
	Username   string `json:"username"`
}

// System represents a supported ticket system
type System struct {
	ID          string
	Name        string
	Description string
	RequiresURL bool
	URLExample  string
}

// Available systems
var availableSystems = []System{
	{
		ID:          "jira",
		Name:        "Jira",
		Description: "Atlassian Jira - Project Management Tool",
		RequiresURL: true,
		URLExample:  "https://your-org.atlassian.net",
	},
	// Add more systems here as we implement them
}

func InitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize Plannet configuration",
		RunE:  runInit,
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	config, err := interactiveInit()
	if err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}

	if err := saveConfig(config); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println("\nConfiguration complete! You're all set to use plannet.")
	return nil
}

func interactiveInit() (*Config, error) {
	fmt.Println("Welcome to Plannet! Let's get you set up.")

	// System Selection
	prompt := promptui.Select{
		Label: "Select System",
		Items: availableSystems,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "\U0001F449 {{ .Name | cyan }} ({{ .Description }})",
			Inactive: "  {{ .Name | white }} ({{ .Description }})",
			Selected: "\U0001F44D {{ .Name | green }} selected",
		},
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return nil, fmt.Errorf("system selection failed: %w", err)
	}

	selectedSystem := availableSystems[idx]

	// Username Input
	usernamePrompt := promptui.Prompt{
		Label: "Email/Username",
		Validate: func(input string) error {
			if len(input) < 1 {
				return fmt.Errorf("username cannot be empty")
			}
			return nil
		},
	}

	username, err := usernamePrompt.Run()
	if err != nil {
		return nil, fmt.Errorf("username input failed: %w", err)
	}

	config := &Config{
		SystemType: selectedSystem.ID,
		Username:   username,
	}

	// URL Input if required
	if selectedSystem.RequiresURL {
		urlPrompt := promptui.Prompt{
			Label:    fmt.Sprintf("%s URL", selectedSystem.Name),
			Default:  selectedSystem.URLExample,
			Validate: validateURL,
		}

		url, err := urlPrompt.Run()
		if err != nil {
			return nil, fmt.Errorf("URL input failed: %w", err)
		}
		config.BaseURL = url
	}

	// Initialize the system and authenticate
	var system *jira.JiraTicketSystem
	switch selectedSystem.ID {
	case "jira":
		system = jira.NewTicketSystem(config.BaseURL)
	default:
		return nil, fmt.Errorf("unsupported system type: %s", selectedSystem.ID)
	}

	// Get API Token
	fmt.Println("\nTo authenticate with Jira, you'll need to create a Personal Access Token (PAT):")
	fmt.Printf("1. Go to %s/secure/ViewProfile.jspa\n", config.BaseURL)
	fmt.Println("2. Click on 'Security' in the left sidebar")
	fmt.Println("3. Click on 'Create and manage API tokens'")
	fmt.Println("4. Click 'Create API token'")
	fmt.Println("5. Give it a name (e.g., 'Plannet CLI')")
	fmt.Println("6. Copy the token and paste it here")

	tokenPrompt := promptui.Prompt{
		Label: "API Token",
		Mask:  '*',
	}

	token, err := tokenPrompt.Run()
	if err != nil {
		return nil, fmt.Errorf("token input failed: %w", err)
	}

	// Test authentication
	fmt.Println("\nTesting authentication...")
	if err := system.Authenticate(username, token); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	fmt.Println("Authentication successful!")
	return config, nil
}

func validateURL(input string) error {
	if len(input) < 1 {
		return fmt.Errorf("URL cannot be empty")
	}
	// Add more URL validation if needed
	return nil
}

func saveConfig(config *Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".plannetrc")

	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return fmt.Errorf("could not marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("could not write config file: %w", err)
	}

	fmt.Printf("\nConfiguration saved to %s\n", configPath)
	return nil
}

func loadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".plannetrc")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no configuration found. Please run 'plannet init' first")
		}
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("could not parse config file: %w", err)
	}

	return &config, nil
}
