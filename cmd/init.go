package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/plannet-ai/plannet/config"
	"github.com/plannet-ai/plannet/security"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Plannet configuration",
	Long: `Initialize Plannet by creating a configuration file.
This will set up your preferences for tracking work and integrating with git.`,
	Run: func(cmd *cobra.Command, args []string) {
		runInit()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit() {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error finding home directory:", err)
		return
	}

	// Define the path for the config file
	configPath := filepath.Join(homeDir, ".plannetrc")

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		prompt := promptui.Select{
			Label: "Configuration already exists. Do you want to overwrite it?",
			Items: []string{"Yes", "No"},
		}

		_, result, err := prompt.Run()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if result == "No" {
			fmt.Println("Initialization cancelled.")
			return
		}
	}

	// Create a new config
	cfg := &config.Config{
		GitIntegration: true, // Default to true
	}

	// Ask for ticket prefixes
	fmt.Println("Let's set up how Plannet identifies tickets in your work.")

	prefixPrompt := promptui.Prompt{
		Label:   "Enter ticket prefixes (comma-separated, e.g., JIRA-, DEV-, TICKET-)",
		Default: "JIRA-",
		Validate: func(input string) error {
			if input == "" {
				return fmt.Errorf("ticket prefixes cannot be empty")
			}
			return nil
		},
	}

	prefixesStr, err := prefixPrompt.Run()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Split the prefixes and clean them
	prefixes := strings.Split(prefixesStr, ",")
	for i, prefix := range prefixes {
		prefixes[i] = strings.TrimSpace(prefix)
	}
	cfg.TicketPrefixes = prefixes

	// Ask for preferred editor
	editorPrompt := promptui.Prompt{
		Label:   "What editor do you use for manual edits?",
		Default: "vim",
		Validate: func(input string) error {
			if input == "" {
				return fmt.Errorf("editor cannot be empty")
			}
			return nil
		},
	}

	editor, err := editorPrompt.Run()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	cfg.Editor = editor

	// Ask about git integration
	gitPrompt := promptui.Select{
		Label: "Enable git integration?",
		Items: []string{"Yes", "No"},
	}

	_, gitResult, err := gitPrompt.Run()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	cfg.GitIntegration = gitResult == "Yes"

	// Ask about copy preference
	copyPrompt := promptui.Select{
		Label: "How would you like to handle copying to clipboard?",
		Items: []string{
			"Ask every time",
			"Ask once per session",
			"Copy automatically",
			"Do not copy",
		},
	}

	_, copyResult, err := copyPrompt.Run()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Map the selection to the appropriate CopyPreference
	switch copyResult {
	case "Ask every time":
		cfg.CopyPreference = config.AskEveryTime
	case "Ask once per session":
		cfg.CopyPreference = config.AskOnce
	case "Copy automatically":
		cfg.CopyPreference = config.CopyAutomatically
	case "Do not copy":
		cfg.CopyPreference = config.DoNotCopy
	}

	// Ask about LLM integration
	llmPrompt := promptui.Select{
		Label: "Would you like to set up LLM integration?",
		Items: []string{"Yes", "No"},
	}

	_, llmResult, err := llmPrompt.Run()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if llmResult == "Yes" {
		// Ask for LLM provider
		providerPrompt := promptui.Select{
			Label: "Select LLM provider",
			Items: []string{"Plannet (brain.plannet.dev)", "Custom endpoint"},
		}

		_, providerResult, err := providerPrompt.Run()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if providerResult == "Plannet (brain.plannet.dev)" {
			// Set up Plannet LLM
			cfg.BaseURL = "https://brain.plannet.dev/v1/completions"
			cfg.Model = "plannet-default"

			fmt.Println("\nTo use Plannet's LLM, you need an API key.")
			fmt.Println("1. Visit https://plannet.dev/dashboard to set up your account")
			fmt.Println("2. Navigate to the API Keys section")
			fmt.Println("3. Create a new API key for brain.plannet.dev")
			fmt.Println("4. Copy the key and paste it below")

			apiKeyPrompt := promptui.Prompt{
				Label: "Plannet API Key",
				Mask:  '•',
				Validate: func(input string) error {
					return security.ValidateAPIKey(input)
				},
			}

			apiKey, err := apiKeyPrompt.Run()
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			// Store the API key in the config
			cfg.LLMToken = apiKey

			// Set up headers with API key
			cfg.Headers = map[string]string{
				"Authorization": "Bearer " + apiKey,
			}
		} else {
			// Ask for custom LLM API endpoint
			baseURLPrompt := promptui.Prompt{
				Label:   "Enter your LLM API endpoint",
				Default: "http://localhost:1234/v1/completions",
				Validate: func(input string) error {
					return security.ValidateURL(input)
				},
			}

			baseURL, err := baseURLPrompt.Run()
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			cfg.BaseURL = baseURL

			// Ask for model name
			modelPrompt := promptui.Prompt{
				Label:   "Enter model name",
				Default: "gpt-3.5-turbo",
				Validate: func(input string) error {
					if input == "" {
						return fmt.Errorf("model name cannot be empty")
					}
					return nil
				},
			}

			model, err := modelPrompt.Run()
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			cfg.Model = model

			// Ask for API key
			apiKeyPrompt := promptui.Prompt{
				Label: "Enter your API key",
				Mask:  '*',
				Validate: func(input string) error {
					return security.ValidateAPIKey(input)
				},
			}

			apiKey, err := apiKeyPrompt.Run()
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			// Store the API key in the config
			cfg.LLMToken = apiKey

			// Set up headers with API key
			cfg.Headers = map[string]string{
				"Authorization": "Bearer " + apiKey,
			}
		}

		// Optional system prompt
		systemPromptPrompt := promptui.Prompt{
			Label: "Enter system prompt (optional)",
		}

		systemPrompt, err := systemPromptPrompt.Run()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if systemPrompt != "" {
			cfg.SystemPrompt = systemPrompt
		}
	}

	// Ask about Jira integration
	jiraPrompt := promptui.Select{
		Label: "Would you like to set up Jira integration? (Optional)",
		Items: []string{"Yes", "No"},
	}

	_, jiraResult, err := jiraPrompt.Run()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if jiraResult == "Yes" {
		// Ask for Jira URL
		fmt.Println("\nPlease enter your Jira instance URL.")
		fmt.Println("Example: https://your-company.atlassian.net")

		jiraURLPrompt := promptui.Prompt{
			Label:   "Jira URL",
			Default: "https://your-instance.atlassian.net",
			Validate: func(input string) error {
				return security.ValidateURL(input)
			},
		}

		jiraURL, err := jiraURLPrompt.Run()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		cfg.JiraURL = jiraURL

		// Ask for Jira username/email
		fmt.Println("\nPlease enter your Jira username or email address.")

		jiraUserPrompt := promptui.Prompt{
			Label: "Jira Username/Email",
			Validate: func(input string) error {
				if input == "" {
					return fmt.Errorf("username cannot be empty")
				}
				return nil
			},
		}

		jiraUser, err := jiraUserPrompt.Run()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		cfg.JiraUser = jiraUser

		// Ask for Jira API token
		fmt.Println("\nTo use Jira, you need an API token.")
		fmt.Println("1. Visit https://id.atlassian.com/manage-profile/security/api-tokens")
		fmt.Println("2. Click 'Create API token'")
		fmt.Println("3. Give it a name (e.g., 'Plannet')")
		fmt.Println("4. Copy the token and paste it below")
		fmt.Println("\nNote: The token will be securely stored and masked when displayed.")

		jiraTokenPrompt := promptui.Prompt{
			Label: "Jira API Token",
			Mask:  '•',
			Validate: func(input string) error {
				return security.ValidateAPIKey(input)
			},
		}

		jiraToken, err := jiraTokenPrompt.Run()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		// Store the Jira token in the config
		cfg.JiraToken = jiraToken
	}

	// Save the configuration
	if err := config.Save(cfg); err != nil {
		fmt.Println("Error saving configuration:", err)
		return
	}

	fmt.Println("\nPlannet initialized successfully! No more un-tracked side quests.")
	fmt.Printf("Configuration saved to %s\n", configPath)

	// Display next steps
	fmt.Println("\nNext steps:")
	fmt.Println("1. Start tracking your work with 'plannet track'")
	fmt.Println("2. Generate content with 'plannet generate'")
	fmt.Println("3. View your current focus with 'plannet now'")
	fmt.Println("4. See your work timeline with 'plannet status'")
}
