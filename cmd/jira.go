package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/plannet-ai/plannet/config"
	"github.com/plannet-ai/plannet/logger"
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
	Use:   "jira",
	Short: "Interact with Jira",
	Long: `Interact with Jira to view and manage tickets.
This command allows you to list your assigned tickets, view ticket details,
and create new tickets.`,
	Run: func(cmd *cobra.Command, args []string) {
		log := logger.WithContext(cmd.Context())
		log.Info("Use one of the subcommands: list, view, create")
		cmd.Help()
	},
}

// jiraListCmd represents the jira list command
var jiraListCmd = &cobra.Command{
	Use:   "list",
	Short: "List your Jira tickets",
	Long:  `List all Jira tickets assigned to you.`,
	Run: func(cmd *cobra.Command, args []string) {
		runJiraList(cmd.Context())
	},
}

// jiraViewCmd represents the jira view command
var jiraViewCmd = &cobra.Command{
	Use:   "view [ticket]",
	Short: "View a Jira ticket",
	Long:  `View details of a specific Jira ticket.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runJiraView(cmd.Context(), args[0])
	},
}

// jiraCreateCmd represents the jira create command
var jiraCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new Jira ticket",
	Long:  `Create a new Jira ticket with the specified details.`,
	Run: func(cmd *cobra.Command, args []string) {
		runJiraCreate(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(jiraCmd)
	jiraCmd.AddCommand(jiraListCmd)
	jiraCmd.AddCommand(jiraViewCmd)
	jiraCmd.AddCommand(jiraCreateCmd)
}

// runJiraList lists all Jira tickets assigned to you
func runJiraList(ctx context.Context) {
	log := logger.WithContext(ctx)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Error("Failed to load configuration: %v", err)
		log.Info("Run 'plannet init' to set up your configuration.")
		return
	}

	// Check if Jira integration is configured
	if cfg.JiraURL == "" || cfg.JiraUser == "" {
		log.Error("Jira integration is not configured")
		log.Info("Run 'plannet init' to set up Jira integration.")
		return
	}

	// Get Jira token from config
	token := cfg.JiraToken
	if token == "" {
		fmt.Println("Error: Jira token not found. Please run 'plannet init' to set up Jira integration.")
		return
	}

	// Create HTTP client with rate limiting
	rateLimiter := security.NewHTTPRateLimiter(10, time.Minute) // 10 requests per minute
	client := rateLimiter.WrapHTTPClient(&http.Client{}, "jira")

	// Create request
	url := cfg.JiraURL + "/rest/api/2/search?jql=assignee=" + cfg.JiraUser + "+ORDER+BY+updated+DESC"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error("Failed to create Jira API request: %v", err)
		return
	}

	// Set headers
	req.Header.Set("Authorization", "Basic "+token)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		log.Error("Failed to send Jira API request: %v", err)
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Error("Jira API returned status %d: %s", resp.StatusCode, string(body))
		return
	}

	// Parse response
	var result struct {
		Issues []JiraTicket `json:"issues"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Error("Failed to parse Jira API response: %v", err)
		return
	}

	// Display tickets
	if len(result.Issues) == 0 {
		log.Info("No tickets found.")
		return
	}

	log.Info("Your Jira tickets:")
	log.Info("-----------------")
	for _, ticket := range result.Issues {
		log.Info("%s: %s (%s)", ticket.Key, ticket.Summary, ticket.Status)
	}
}

// runJiraView views a specific Jira ticket
func runJiraView(ctx context.Context, ticketKey string) {
	log := logger.WithContext(ctx)

	// Validate ticket key
	if err := security.ValidateTicketKey(ticketKey); err != nil {
		log.Error("Invalid ticket key: %v", err)
		return
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Error("Failed to load configuration: %v", err)
		log.Info("Run 'plannet init' to set up your configuration.")
		return
	}

	// Check if Jira integration is configured
	if cfg.JiraURL == "" || cfg.JiraUser == "" {
		log.Error("Jira integration is not configured")
		log.Info("Run 'plannet init' to set up Jira integration.")
		return
	}

	// Get Jira token from config
	token := cfg.JiraToken
	if token == "" {
		fmt.Println("Error: Jira token not found. Please run 'plannet init' to set up Jira integration.")
		return
	}

	// Create HTTP client with rate limiting
	rateLimiter := security.NewHTTPRateLimiter(10, time.Minute) // 10 requests per minute
	client := rateLimiter.WrapHTTPClient(&http.Client{}, "jira")

	// Create request
	url := cfg.JiraURL + "/rest/api/2/issue/" + ticketKey
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error("Failed to create Jira API request: %v", err)
		return
	}

	// Set headers
	req.Header.Set("Authorization", "Basic "+token)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		log.Error("Failed to send Jira API request: %v", err)
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Error("Jira API returned status %d: %s", resp.StatusCode, string(body))
		return
	}

	// Parse response
	var ticket JiraTicket
	if err := json.NewDecoder(resp.Body).Decode(&ticket); err != nil {
		log.Error("Failed to parse Jira API response: %v", err)
		return
	}

	// Display ticket details
	log.Info("Ticket: %s", ticket.Key)
	log.Info("Summary: %s", ticket.Summary)
	log.Info("Status: %s", ticket.Status)
	log.Info("Type: %s", ticket.Type)
	log.Info("Priority: %s", ticket.Priority)
	log.Info("Assignee: %s", ticket.Assignee)
	log.Info("URL: %s", ticket.URL)
	log.Info("\nDescription:")
	log.Info(ticket.Description)
}

// runJiraCreate creates a new Jira ticket
func runJiraCreate(ctx context.Context) {
	log := logger.WithContext(ctx)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Error("Failed to load configuration: %v", err)
		log.Info("Run 'plannet init' to set up your configuration.")
		return
	}

	// Check if Jira integration is configured
	if cfg.JiraURL == "" || cfg.JiraUser == "" {
		log.Error("Jira integration is not configured")
		log.Info("Run 'plannet init' to set up Jira integration.")
		return
	}

	// Get Jira token from config
	token := cfg.JiraToken
	if token == "" {
		fmt.Println("Error: Jira token not found. Please run 'plannet init' to set up Jira integration.")
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
		log.Error("Error: %v", err)
		return
	}

	// Ask for issue type
	issueTypePrompt := promptui.Select{
		Label: "Select issue type",
		Items: []string{"Task", "Bug", "Story", "Epic"},
	}

	_, issueType, err := issueTypePrompt.Run()
	if err != nil {
		log.Error("Error: %v", err)
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
		log.Error("Error: %v", err)
		return
	}

	// Ask for description
	descriptionPrompt := promptui.Prompt{
		Label: "Enter description",
	}

	description, err := descriptionPrompt.Run()
	if err != nil {
		log.Error("Error: %v", err)
		return
	}

	// Create ticket
	ticket := map[string]interface{}{
		"fields": map[string]interface{}{
			"project": map[string]string{
				"key": projectKey,
			},
			"issuetype": map[string]string{
				"name": issueType,
			},
			"summary":     summary,
			"description": description,
		},
	}

	// Marshal ticket data
	ticketData, err := json.Marshal(ticket)
	if err != nil {
		log.Error("Failed to marshal ticket data: %v", err)
		return
	}

	// Create HTTP client with rate limiting
	rateLimiter := security.NewHTTPRateLimiter(10, time.Minute) // 10 requests per minute
	client := rateLimiter.WrapHTTPClient(&http.Client{}, "jira")

	// Create request
	url := cfg.JiraURL + "/rest/api/2/issue"
	req, err := http.NewRequest("POST", url, bytes.NewReader(ticketData))
	if err != nil {
		log.Error("Failed to create Jira API request: %v", err)
		return
	}

	// Set headers
	req.Header.Set("Authorization", "Basic "+token)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		log.Error("Failed to send Jira API request: %v", err)
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Error("Jira API returned status %d: %s", resp.StatusCode, string(body))
		return
	}

	// Parse response
	var result struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Error("Failed to parse Jira API response: %v", err)
		return
	}

	log.Info("Successfully created ticket %s", result.Key)
	log.Info("URL: %s/browse/%s", cfg.JiraURL, result.Key)
}
