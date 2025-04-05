package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
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
}

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
	config := Config{
		GitIntegration: true, // Default to true
	}

	// Ask for ticket prefixes
	fmt.Println("Let's set up how Plannet identifies tickets in your work.")
	
	prefixPrompt := promptui.Prompt{
		Label:   "Enter ticket prefixes (comma-separated, e.g., JIRA-, DEV-, TICKET-)",
		Default: "JIRA-",
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
	config.TicketPrefixes = prefixes

	// Ask for preferred editor
	editorPrompt := promptui.Prompt{
		Label:   "What editor do you use for manual edits?",
		Default: "vim",
	}

	editor, err := editorPrompt.Run()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	config.Editor = editor

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

	config.GitIntegration = gitResult == "Yes"

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
			config.BaseURL = "https://brain.plannet.dev/v1/completions"
			config.Model = "plannet-default"
			
			fmt.Println("\nTo use Plannet's LLM, you need an API key.")
			fmt.Println("Visit https://plannet.dev/console to get your API key.")
			
			apiKeyPrompt := promptui.Prompt{
				Label: "Enter your Plannet API key",
				Mask:  '*',
			}

			apiKey, err := apiKeyPrompt.Run()
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			// Set up headers with API key
			config.Headers = map[string]string{
				"Authorization": "Bearer " + apiKey,
			}
		} else {
			// Ask for custom LLM API endpoint
			baseURLPrompt := promptui.Prompt{
				Label:   "Enter your LLM API endpoint",
				Default: "http://localhost:1234/v1/completions",
			}

			baseURL, err := baseURLPrompt.Run()
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			config.BaseURL = baseURL

			// Ask for model name
			modelPrompt := promptui.Prompt{
				Label:   "Enter model name",
				Default: "gpt-3.5-turbo",
			}

			model, err := modelPrompt.Run()
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			config.Model = model

			// Ask for API key
			apiKeyPrompt := promptui.Prompt{
				Label: "Enter your API key",
				Mask:  '*',
			}

			apiKey, err := apiKeyPrompt.Run()
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			// Set up headers with API key
			config.Headers = map[string]string{
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
			config.SystemPrompt = systemPrompt
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
		jiraURLPrompt := promptui.Prompt{
			Label:   "Enter your Jira instance URL",
			Default: "https://your-instance.atlassian.net",
		}

		jiraURL, err := jiraURLPrompt.Run()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		config.JiraURL = jiraURL

		// Ask for Jira username/email
		jiraUserPrompt := promptui.Prompt{
			Label: "Enter your Jira username/email",
		}

		jiraUser, err := jiraUserPrompt.Run()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		config.JiraUser = jiraUser

		// Ask for Jira API token
		fmt.Println("\nTo use Jira, you need an API token.")
		fmt.Println("Visit https://id.atlassian.com/manage-profile/security/api-tokens to create one.")
		
		jiraTokenPrompt := promptui.Prompt{
			Label: "Enter your Jira API token",
			Mask:  '*',
		}

		jiraToken, err := jiraTokenPrompt.Run()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		config.JiraToken = jiraToken
	}

	// Convert config to JSON
	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Println("Error creating configuration:", err)
		return
	}

	// Write config to file
	err = os.WriteFile(configPath, configJSON, 0644)
	if err != nil {
		fmt.Println("Error writing configuration file:", err)
		return
	}

	fmt.Println("\nPlannet initialized successfully! No more un-tracked side quests.")
	fmt.Printf("Configuration saved to %s\n", configPath)
	
	// Display next steps
	fmt.Println("\nNext steps:")
	fmt.Println("1. Run 'plannet now' to see what you're currently working on")
	fmt.Println("2. Run 'plannet track' to manually track work")
	fmt.Println("3. Run 'plannet status' to see a timeline of your work")
	fmt.Println("4. Run 'plannet list' to see all tracked work")
	fmt.Println("5. Run 'plannet complete' to mark work as complete")
	fmt.Println("6. Run 'plannet export' to export your work data")
} 