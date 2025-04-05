package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/plannet-ai/plannet/config"
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
		fmt.Println("Error loading configuration:", err)
		fmt.Println("Run 'plannet init' to set up your configuration.")
		return
	}

	// Check if Jira integration is configured
	if cfg.JiraURL == "" || cfg.JiraToken == "" || cfg.JiraUser == "" {
		fmt.Println("Jira integration is not configured.")
		fmt.Println("Run 'plannet init' to set up Jira integration.")
		return
	}

	// Create HTTP client
	client := &http.Client{}

	// Create request
	url := fmt.Sprintf("%s/rest/api/2/search?jql=assignee=%s+ORDER+BY+updated+DESC", cfg.JiraURL, cfg.JiraUser)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Set headers
	req.Header.Set("Authorization", "Basic "+cfg.JiraToken)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: Jira API returned status %d: %s\n", resp.StatusCode, string(body))
		return
	}

	// Parse response
	var result struct {
		Issues []struct {
			Key    string `json:"key"`
			Fields struct {
				Summary     string `json:"summary"`
				Status     struct {
					Name string `json:"name"`
				} `json:"status"`
				Type struct {
					Name string `json:"name"`
				} `json:"type"`
				Priority struct {
					Name string `json:"name"`
				} `json:"priority"`
				Assignee struct {
					DisplayName string `json:"displayName"`
				} `json:"assignee"`
			} `json:"fields"`
		} `json:"issues"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("Error parsing response:", err)
		return
	}

	// Display tickets
	if len(result.Issues) == 0 {
		fmt.Println("No Jira tickets found.")
		return
	}

	fmt.Println("Your Jira tickets:")
	fmt.Println("-----------------")
	for _, issue := range result.Issues {
		fmt.Printf("%s: %s\n", issue.Key, issue.Fields.Summary)
		fmt.Printf("  Status: %s, Type: %s, Priority: %s\n", 
			issue.Fields.Status.Name, 
			issue.Fields.Type.Name, 
			issue.Fields.Priority.Name)
		fmt.Printf("  URL: %s/browse/%s\n\n", cfg.JiraURL, issue.Key)
	}
}

// runJiraView displays details of a specific Jira ticket
func runJiraView(ticketKey string) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Error loading configuration:", err)
		fmt.Println("Run 'plannet init' to set up your configuration.")
		return
	}

	// Check if Jira integration is configured
	if cfg.JiraURL == "" || cfg.JiraToken == "" || cfg.JiraUser == "" {
		fmt.Println("Jira integration is not configured.")
		fmt.Println("Run 'plannet init' to set up Jira integration.")
		return
	}

	// Create HTTP client
	client := &http.Client{}

	// Create request
	url := fmt.Sprintf("%s/rest/api/2/issue/%s", cfg.JiraURL, ticketKey)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Set headers
	req.Header.Set("Authorization", "Basic "+cfg.JiraToken)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: Jira API returned status %d: %s\n", resp.StatusCode, string(body))
		return
	}

	// Parse response
	var issue struct {
		Key    string `json:"key"`
		Fields struct {
			Summary     string `json:"summary"`
			Description string `json:"description"`
			Status     struct {
				Name string `json:"name"`
			} `json:"status"`
			Type struct {
				Name string `json:"name"`
			} `json:"type"`
			Priority struct {
				Name string `json:"name"`
			} `json:"priority"`
			Assignee struct {
				DisplayName string `json:"displayName"`
			} `json:"assignee"`
			Created string `json:"created"`
			Updated string `json:"updated"`
		} `json:"fields"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		fmt.Println("Error parsing response:", err)
		return
	}

	// Display ticket details
	fmt.Printf("Ticket: %s\n", issue.Key)
	fmt.Printf("Summary: %s\n", issue.Fields.Summary)
	fmt.Printf("Status: %s\n", issue.Fields.Status.Name)
	fmt.Printf("Type: %s\n", issue.Fields.Type.Name)
	fmt.Printf("Priority: %s\n", issue.Fields.Priority.Name)
	fmt.Printf("Assignee: %s\n", issue.Fields.Assignee.DisplayName)
	fmt.Printf("Created: %s\n", issue.Fields.Created)
	fmt.Printf("Updated: %s\n", issue.Fields.Updated)
	fmt.Printf("URL: %s/browse/%s\n\n", cfg.JiraURL, issue.Key)

	// Display description
	if issue.Fields.Description != "" {
		fmt.Println("Description:")
		fmt.Println("------------")
		fmt.Println(issue.Fields.Description)
	}
}

// runJiraCreate creates a new Jira ticket
func runJiraCreate() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Error loading configuration:", err)
		fmt.Println("Run 'plannet init' to set up your configuration.")
		return
	}

	// Check if Jira integration is configured
	if cfg.JiraURL == "" || cfg.JiraToken == "" || cfg.JiraUser == "" {
		fmt.Println("Jira integration is not configured.")
		fmt.Println("Run 'plannet init' to set up Jira integration.")
		return
	}

	// Prompt for ticket details
	prompt := promptui.Prompt{
		Label: "Project key (e.g., PROJ)",
	}

	projectKey, err := prompt.Run()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	prompt = promptui.Prompt{
		Label: "Summary",
	}

	summary, err := prompt.Run()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	prompt = promptui.Prompt{
		Label: "Description",
	}

	description, err := prompt.Run()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Prompt for issue type
	typePrompt := promptui.Select{
		Label: "Issue type",
		Items: []string{"Task", "Bug", "Story", "Epic"},
	}

	_, issueType, err := typePrompt.Run()
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
		fmt.Println("Error creating request body:", err)
		return
	}

	// Create HTTP client
	client := &http.Client{}

	// Create request
	url := fmt.Sprintf("%s/rest/api/2/issue", cfg.JiraURL)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Set headers
	req.Header.Set("Authorization", "Basic "+cfg.JiraToken)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: Jira API returned status %d: %s\n", resp.StatusCode, string(body))
		return
	}

	// Parse response
	var result struct {
		Key string `json:"key"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("Error parsing response:", err)
		return
	}

	// Display success message
	fmt.Printf("Ticket created successfully: %s\n", result.Key)
	fmt.Printf("URL: %s/browse/%s\n", cfg.JiraURL, result.Key)
} 