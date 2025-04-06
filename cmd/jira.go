package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"regexp"

	"github.com/manifoldco/promptui"
	"github.com/plannet-ai/plannet/config"
	"github.com/plannet-ai/plannet/security"
	"github.com/spf13/cobra"
)

// JiraTicket represents a Jira ticket
type JiraTicket struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Type        string `json:"type"`
	Priority    string `json:"priority"`
	Assignee    string `json:"assignee"`
	URL         string `json:"url"`
}

// jiraCmd represents the jira command
var jiraCmd = &cobra.Command{
	Use:   "jira [subcommand]",
	Short: "Interact with Jira tickets",
	Long: `Interact with Jira tickets.
This command allows you to view, create, and update Jira tickets directly from the command line.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			runJiraList()
		} else {
			switch args[0] {
			case "list":
				runJiraList()
			case "view":
				if len(args) < 2 {
					fmt.Println("Error: Please provide a ticket key (e.g., JIRA-123)")
					return
				}
				runJiraView(args[1])
			case "create":
				runJiraCreate()
			default:
				fmt.Printf("Unknown subcommand: %s\n", args[0])
				fmt.Println("Available subcommands: list, view, create")
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(jiraCmd)
}

// runJiraList lists the user's assigned Jira tickets
func runJiraList() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\nRun 'plannet init' to set up your configuration.\n", err)
		return
	}

	// Check if Jira integration is configured
	if cfg.JiraURL == "" || cfg.JiraUser == "" {
		fmt.Println("Jira integration is not configured.")
		fmt.Println("Run 'plannet init' to set up Jira integration.")
		return
	}

	// Get Jira token from secure storage
	jiraToken, err := config.GetJiraToken()
	if err != nil {
		fmt.Printf("Failed to get Jira token: %v\n", err)
		return
	}

	// Create HTTP client with rate limiting
	rateLimiter := security.NewHTTPRateLimiter(10, time.Minute) // 10 requests per minute
	client := rateLimiter.WrapHTTPClient(&http.Client{}, "jira")

	// Create request
	url := fmt.Sprintf("%s/rest/api/2/search?jql=assignee=%s+ORDER+BY+updated+DESC", cfg.JiraURL, cfg.JiraUser)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Failed to create Jira API request: %v\n", err)
		return
	}

	// Set headers
	req.Header.Set("Authorization", "Basic "+jiraToken)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to send Jira API request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Jira API returned status %d: %s\n", resp.StatusCode, string(body))
		return
	}

	// Parse response
	var result struct {
		Issues []JiraTicket `json:"issues"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Failed to parse Jira API response: %v\n", err)
		return
	}

	// Display tickets
	if len(result.Issues) == 0 {
		fmt.Println("No tickets found.")
		return
	}

	fmt.Println("Your Jira tickets:")
	fmt.Println("-----------------")
	for _, ticket := range result.Issues {
		fmt.Printf("%s: %s (%s)\n", ticket.Key, ticket.Summary, ticket.Status)
	}
}

// runJiraView views a specific Jira ticket
func runJiraView(ticketKey string) {
	// Validate ticket key
	if err := security.ValidateTicketKey(ticketKey); err != nil {
		fmt.Printf("Invalid ticket key: %v\n", err)
		return
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\nRun 'plannet init' to set up your configuration.\n", err)
		return
	}

	// Check if Jira integration is configured
	if cfg.JiraURL == "" || cfg.JiraUser == "" {
		fmt.Println("Jira integration is not configured.")
		fmt.Println("Run 'plannet init' to set up Jira integration.")
		return
	}

	// Get Jira token from secure storage
	jiraToken, err := config.GetJiraToken()
	if err != nil {
		fmt.Printf("Failed to get Jira token: %v\n", err)
		return
	}

	// Create HTTP client with rate limiting
	rateLimiter := security.NewHTTPRateLimiter(10, time.Minute) // 10 requests per minute
	client := rateLimiter.WrapHTTPClient(&http.Client{}, "jira")

	// Create request
	url := fmt.Sprintf("%s/rest/api/2/issue/%s", cfg.JiraURL, ticketKey)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Failed to create Jira API request: %v\n", err)
		return
	}

	// Set headers
	req.Header.Set("Authorization", "Basic "+jiraToken)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to send Jira API request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Jira API returned status %d: %s\n", resp.StatusCode, string(body))
		return
	}

	// Parse response
	var ticket JiraTicket
	if err := json.NewDecoder(resp.Body).Decode(&ticket); err != nil {
		fmt.Printf("Failed to parse Jira API response: %v\n", err)
		return
	}

	// Display ticket details
	fmt.Printf("Ticket: %s\n", ticket.Key)
	fmt.Printf("Summary: %s\n", ticket.Summary)
	fmt.Printf("Status: %s\n", ticket.Status)
	fmt.Printf("Type: %s\n", ticket.Type)
	fmt.Printf("Priority: %s\n", ticket.Priority)
	fmt.Printf("Assignee: %s\n", ticket.Assignee)
	fmt.Printf("URL: %s\n", ticket.URL)
	fmt.Println("\nDescription:")
	fmt.Println(ticket.Description)
}

// runJiraCreate creates a new Jira ticket
func runJiraCreate() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\nRun 'plannet init' to set up your configuration.\n", err)
		return
	}

	// Check if Jira integration is configured
	if cfg.JiraURL == "" || cfg.JiraUser == "" {
		fmt.Println("Jira integration is not configured.")
		fmt.Println("Run 'plannet init' to set up Jira integration.")
		return
	}

	// Get Jira token from secure storage
	jiraToken, err := config.GetJiraToken()
	if err != nil {
		fmt.Printf("Failed to get Jira token: %v\n", err)
		return
	}

	// Ask for project key
	projectPrompt := promptui.Prompt{
		Label: "Enter project key (e.g., PROJ)",
		Validate: func(input string) error {
			if input == "" {
				return fmt.Errorf("project key cannot be empty")
			}
			// Project keys are typically uppercase letters and numbers
			pattern := regexp.MustCompile(`^[A-Z0-9]+$`)
			if !pattern.MatchString(input) {
				return fmt.Errorf("project key must contain only uppercase letters and numbers")
			}
			return nil
		},
	}

	projectKey, err := projectPrompt.Run()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Ask for issue type
	issueTypePrompt := promptui.Select{
		Label: "Select issue type",
		Items: []string{"Task", "Bug", "Story", "Epic"},
	}

	_, issueType, err := issueTypePrompt.Run()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Ask for summary
	summaryPrompt := promptui.Prompt{
		Label: "Enter summary",
		Validate: func(input string) error {
			if input == "" {
				return fmt.Errorf("summary cannot be empty")
			}
			return nil
		},
	}

	summary, err := summaryPrompt.Run()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Ask for description
	descriptionPrompt := promptui.Prompt{
		Label: "Enter description",
		Validate: func(input string) error {
			if input == "" {
				return fmt.Errorf("description cannot be empty")
			}
			return nil
		},
	}

	description, err := descriptionPrompt.Run()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Create request body
	requestBody := map[string]interface{}{
		"fields": map[string]interface{}{
			"project": map[string]string{
				"key": projectKey,
			},
			"summary":     summary,
			"description": description,
			"issuetype": map[string]string{
				"name": issueType,
			},
		},
	}

	// Convert to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Printf("Failed to create request body: %v\n", err)
		return
	}

	// Create HTTP client with rate limiting
	rateLimiter := security.NewHTTPRateLimiter(10, time.Minute) // 10 requests per minute
	client := rateLimiter.WrapHTTPClient(&http.Client{}, "jira")

	// Create request
	url := fmt.Sprintf("%s/rest/api/2/issue", cfg.JiraURL)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		fmt.Printf("Failed to create Jira API request: %v\n", err)
		return
	}

	// Set headers
	req.Header.Set("Authorization", "Basic "+jiraToken)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to send Jira API request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Jira API returned status %d: %s\n", resp.StatusCode, string(body))
		return
	}

	// Parse response
	var result struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Failed to parse Jira API response: %v\n", err)
		return
	}

	fmt.Printf("Ticket created successfully: %s\n", result.Key)
}
