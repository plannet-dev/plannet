package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type JiraTicketSystem struct {
	baseURL    string
	httpClient *http.Client
	token      string
	username   string
}

type Ticket struct {
	ID          string
	Title       string
	Description string
	Status      string
}

// NewTicketSystem creates a new Jira ticket system instance
func NewTicketSystem(baseURL string) *JiraTicketSystem {
	return &JiraTicketSystem{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Authenticate stores credentials and validates them
func (j *JiraTicketSystem) Authenticate(username, token string) error {
	j.username = username
	j.token = token

	// Test authentication
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/rest/api/3/myself", j.baseURL), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	j.addAuthHeader(req)

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("authentication test failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed with status: %d", resp.StatusCode)
	}

	return nil
}

// GetAssignedTickets retrieves tickets assigned to the authenticated user
func (j *JiraTicketSystem) GetAssignedTickets() ([]Ticket, error) {
	// JQL query to find assigned issues
	jqlQuery := fmt.Sprintf("assignee = %s AND resolution = Unresolved", j.username)

	url := fmt.Sprintf("%s/rest/api/3/search", j.baseURL)

	reqBody := map[string]interface{}{
		"jql":        jqlQuery,
		"maxResults": 50,
		"fields":     []string{"summary", "description", "status"},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	j.addAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tickets: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch tickets, status: %d", resp.StatusCode)
	}

	var result struct {
		Issues []struct {
			ID     string `json:"id"`
			Key    string `json:"key"`
			Fields struct {
				Summary     string `json:"summary"`
				Description string `json:"description"`
				Status      struct {
					Name string `json:"name"`
				} `json:"status"`
			} `json:"fields"`
		} `json:"issues"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	tickets := make([]Ticket, len(result.Issues))
	for i, issue := range result.Issues {
		tickets[i] = Ticket{
			ID:          issue.Key,
			Title:       issue.Fields.Summary,
			Description: issue.Fields.Description,
			Status:      issue.Fields.Status.Name,
		}
	}

	return tickets, nil
}

// Logout clears stored credentials
func (j *JiraTicketSystem) Logout() error {
	j.token = ""
	j.username = ""
	return nil
}

// addAuthHeader adds basic auth header to request
func (j *JiraTicketSystem) addAuthHeader(req *http.Request) {
	req.SetBasicAuth(j.username, j.token)
}
