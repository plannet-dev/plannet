package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/manifoldco/promptui"
)

// JiraClient handles all Jira-related operations
type JiraClient struct {
	config *Config
	client *http.Client
}

// JiraIssue represents a Jira ticket
type JiraIssue struct {
	Key    string `json:"key"`
	Fields struct {
		Summary     string `json:"summary"`
		Description string `json:"description"`
		Status      struct {
			Name string `json:"name"`
		} `json:"status"`
		Priority struct {
			Name string `json:"name"`
		} `json:"priority"`
		Assignee struct {
			DisplayName  string `json:"displayName"`
			EmailAddress string `json:"emailAddress"`
		} `json:"assignee"`
	} `json:"fields"`
}

// JiraSearchResponse represents the response from Jira search API
type JiraSearchResponse struct {
	StartAt    int         `json:"startAt"`
	MaxResults int         `json:"maxResults"`
	Total      int         `json:"total"`
	Issues     []JiraIssue `json:"issues"`
}

// NewJiraClient creates a new Jira client instance
func NewJiraClient(config *Config) *JiraClient {
	return &JiraClient{
		config: config,
		client: &http.Client{},
	}
}

// IsConfigured checks if Jira is properly configured
func (j *JiraClient) IsConfigured() bool {
	return j.config.JiraURL != "" &&
		j.config.JiraToken != "" &&
		j.config.JiraUser != ""
}

// SelectTicket presents a list of tickets and returns the selected one
func (j *JiraClient) SelectTicket() (*JiraIssue, error) {
	issues, err := j.FetchTickets()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tickets: %w", err)
	}

	if len(issues) == 0 {
		return nil, fmt.Errorf("no tickets found")
	}

	return j.promptForTicket(issues)
}

// FetchTickets retrieves all tickets assigned to the user
func (j *JiraClient) FetchTickets() ([]JiraIssue, error) {
	// Build JQL query
	jql := fmt.Sprintf("assignee = %s AND status != Done ORDER BY updated DESC",
		url.QueryEscape(j.config.JiraUser))

	endpoint := fmt.Sprintf("%s/rest/api/2/search", j.config.JiraURL)

	// Create request
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	q.Add("jql", jql)
	q.Add("fields", "summary,description,status,priority,assignee")
	req.URL.RawQuery = q.Encode()

	// Add authentication
	req.Header.Set("Authorization", "Bearer "+j.config.JiraToken)
	req.Header.Set("Content-Type", "application/json")

	// Make request
	resp, err := j.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Jira API returned status: %d", resp.StatusCode)
	}

	// Parse response
	var searchResp JiraSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return searchResp.Issues, nil
}

// promptForTicket displays an interactive prompt for ticket selection
func (j *JiraClient) promptForTicket(issues []JiraIssue) (*JiraIssue, error) {
	templates := &promptui.SelectTemplates{
		Label: "{{ . }}",
		Active: fmt.Sprintf(`%s {{ .Key | cyan }} {{ .Fields.Status.Name | yellow }} {{ .Fields.Priority.Name | red }} 
    {{ .Fields.Summary | cyan }}`, promptui.IconSelect),
		Inactive: "  {{ .Key | white }} {{ .Fields.Status.Name }} {{ .Fields.Priority.Name }} {{ .Fields.Summary }}",
		Selected: fmt.Sprintf(`{{ "%s" | green }} {{ .Key | faint }} {{ .Fields.Summary }}`, promptui.IconGood),
		Details: `
--------- Ticket Details ----------
{{ "Key:" | faint }}	{{ .Key }}
{{ "Summary:" | faint }}	{{ .Fields.Summary }}
{{ "Status:" | faint }}	{{ .Fields.Status.Name }}
{{ "Priority:" | faint }}	{{ .Fields.Priority.Name }}
{{ "Assignee:" | faint }}	{{ .Fields.Assignee.DisplayName }}
{{ "Description:" | faint }}	{{ .Fields.Description }}`,
	}

	searcher := func(input string, index int) bool {
		issue := issues[index]
		name := fmt.Sprintf("%s %s", issue.Key, issue.Fields.Summary)
		input = strings.ToLower(input)
		name = strings.ToLower(name)
		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:     "Select a ticket",
		Items:     issues,
		Templates: templates,
		Size:      10,
		Searcher:  searcher,
	}

	index, _, err := prompt.Run()
	if err != nil {
		return nil, fmt.Errorf("prompt failed: %w", err)
	}

	return &issues[index], nil
}

// GetTicketDetails retrieves full details for a specific ticket
func (j *JiraClient) GetTicketDetails(key string) (*JiraIssue, error) {
	endpoint := fmt.Sprintf("%s/rest/api/2/issue/%s", j.config.JiraURL, key)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+j.config.JiraToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := j.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Jira API returned status: %d", resp.StatusCode)
	}

	var issue JiraIssue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &issue, nil
}
