package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/plannet-ai/plannet/config"
	"github.com/plannet-ai/plannet/security"
)

// jiraTestConfig is used to create a test configuration
type jiraTestConfig struct {
	*config.Config
	// Mock token for testing
	mockToken string
}

func (m *jiraTestConfig) Load() error {
	return nil
}

// Mock GetJiraToken for testing
func (m *jiraTestConfig) GetJiraToken() (string, error) {
	if m.mockToken == "" {
		return "", fmt.Errorf("token not set")
	}
	return m.mockToken, nil
}

// JiraList lists the user's assigned Jira tickets
func JiraList(cfg *jiraTestConfig) error {
	// Check if Jira integration is configured
	if cfg.JiraURL == "" || cfg.JiraUser == "" {
		return fmt.Errorf("Jira integration is not configured")
	}

	// Get Jira token from secure storage
	jiraToken, err := cfg.GetJiraToken()
	if err != nil {
		return fmt.Errorf("failed to get Jira token: %w", err)
	}

	// Create HTTP client with rate limiting
	rateLimiter := security.NewHTTPRateLimiter(10, time.Minute) // 10 requests per minute
	client := rateLimiter.WrapHTTPClient(&http.Client{}, "jira")

	// Create request
	url := fmt.Sprintf("%s/rest/api/2/search?jql=assignee=%s+ORDER+BY+updated+DESC", cfg.JiraURL, cfg.JiraUser)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create Jira API request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Basic "+jiraToken)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Jira API request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Jira API returned status %d", resp.StatusCode)
	}

	// Parse response
	var result struct {
		Issues []struct {
			Key    string `json:"key"`
			Fields struct {
				Summary string `json:"summary"`
				Status  struct {
					Name string `json:"name"`
				} `json:"status"`
			} `json:"fields"`
		} `json:"issues"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse Jira API response: %w", err)
	}

	return nil
}

// JiraView displays details of a specific Jira ticket
func JiraView(cfg *jiraTestConfig, ticketKey string) error {
	// Validate ticket key
	if err := security.ValidateTicketKey(ticketKey); err != nil {
		return fmt.Errorf("invalid ticket key: %w", err)
	}

	// Check if Jira integration is configured
	if cfg.JiraURL == "" || cfg.JiraUser == "" {
		return fmt.Errorf("Jira integration is not configured")
	}

	// Get Jira token from secure storage
	jiraToken, err := cfg.GetJiraToken()
	if err != nil {
		return fmt.Errorf("failed to get Jira token: %w", err)
	}

	// Create HTTP client with rate limiting
	rateLimiter := security.NewHTTPRateLimiter(10, time.Minute) // 10 requests per minute
	client := rateLimiter.WrapHTTPClient(&http.Client{}, "jira")

	// Create request
	url := fmt.Sprintf("%s/rest/api/2/issue/%s", cfg.JiraURL, ticketKey)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create Jira API request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Basic "+jiraToken)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Jira API request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Jira API returned status %d", resp.StatusCode)
	}

	// Parse response
	var issue struct {
		Key    string `json:"key"`
		Fields struct {
			Summary     string `json:"summary"`
			Description string `json:"description"`
			Status      struct {
				Name string `json:"name"`
			} `json:"status"`
		} `json:"fields"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return fmt.Errorf("failed to parse Jira API response: %w", err)
	}

	return nil
}

// TestJiraList tests the JiraList function
func TestJiraList(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		serverResponse string
		statusCode     int
		expectError    bool
	}{
		{
			name: "Success",
			serverResponse: `{
				"issues": [
					{
						"key": "PROJ-123",
						"fields": {
							"summary": "Test Issue",
							"status": {"name": "In Progress"}
						}
					}
				]
			}`,
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name:           "Server Error",
			serverResponse: `{"error": "Internal Server Error"}`,
			statusCode:     http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:           "Invalid JSON",
			serverResponse: `{invalid json}`,
			statusCode:     http.StatusOK,
			expectError:    true,
		},
		{
			name:           "Empty Response",
			serverResponse: `{"issues": []}`,
			statusCode:     http.StatusOK,
			expectError:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET request, got %s", r.Method)
				}

				// Verify authorization header
				auth := r.Header.Get("Authorization")
				if auth == "" {
					t.Error("Expected Authorization header")
				}

				// Set response
				w.WriteHeader(tc.statusCode)
				w.Write([]byte(tc.serverResponse))
			}))
			defer server.Close()

			// Create mock config
			cfg := &jiraTestConfig{
				Config: &config.Config{
					JiraURL:  server.URL,
					JiraUser: "test-user",
				},
				mockToken: "test-token",
			}

			// Test JiraList
			err := JiraList(cfg)

			// Check error
			if tc.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

// TestJiraView tests the JiraView function
func TestJiraView(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		issueKey       string
		serverResponse string
		statusCode     int
		expectError    bool
	}{
		{
			name:     "Success",
			issueKey: "PROJ-123",
			serverResponse: `{
				"key": "PROJ-123",
				"fields": {
					"summary": "Test Issue",
					"description": "Test Description",
					"status": {"name": "In Progress"},
					"assignee": {"displayName": "Test User"}
				}
			}`,
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name:           "Server Error",
			issueKey:       "PROJ-123",
			serverResponse: `{"error": "Internal Server Error"}`,
			statusCode:     http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:           "Invalid JSON",
			issueKey:       "PROJ-123",
			serverResponse: `{invalid json}`,
			statusCode:     http.StatusOK,
			expectError:    true,
		},
		{
			name:           "Issue Not Found",
			issueKey:       "PROJ-999",
			serverResponse: `{"error": "Issue not found"}`,
			statusCode:     http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "Invalid Ticket Key",
			issueKey:       "INVALID",
			serverResponse: `{}`,
			statusCode:     http.StatusOK,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET request, got %s", r.Method)
				}

				// Verify authorization header
				auth := r.Header.Get("Authorization")
				if auth == "" {
					t.Error("Expected Authorization header")
				}

				// Verify URL path
				expectedPath := fmt.Sprintf("/rest/api/2/issue/%s", tc.issueKey)
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				// Set response
				w.WriteHeader(tc.statusCode)
				w.Write([]byte(tc.serverResponse))
			}))
			defer server.Close()

			// Create mock config
			cfg := &jiraTestConfig{
				Config: &config.Config{
					JiraURL:  server.URL,
					JiraUser: "test-user",
				},
				mockToken: "test-token",
			}

			// Test JiraView
			err := JiraView(cfg, tc.issueKey)

			// Check error
			if tc.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

// TestJiraIntegration tests the Jira integration configuration
func TestJiraIntegration(t *testing.T) {
	// Test cases
	testCases := []struct {
		name        string
		config      *config.Config
		expectError bool
	}{
		{
			name: "Valid Configuration",
			config: &config.Config{
				JiraURL:  "https://test.atlassian.net",
				JiraUser: "test-user",
			},
			expectError: false,
		},
		{
			name: "Missing URL",
			config: &config.Config{
				JiraUser: "test-user",
			},
			expectError: true,
		},
		{
			name: "Missing User",
			config: &config.Config{
				JiraURL: "https://test.atlassian.net",
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock config
			cfg := &jiraTestConfig{
				Config:    tc.config,
				mockToken: "test-token",
			}

			// Test JiraList (which checks integration)
			err := JiraList(cfg)

			// Check error
			if tc.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

// TestJiraResponseParsing tests the parsing of Jira API responses
func TestJiraResponseParsing(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		response       string
		expectedFields []string
		expectError    bool
	}{
		{
			name: "Valid Issue List",
			response: `{
				"issues": [
					{
						"key": "PROJ-123",
						"fields": {
							"summary": "Test Issue 1",
							"status": {"name": "In Progress"}
						}
					},
					{
						"key": "PROJ-124",
						"fields": {
							"summary": "Test Issue 2",
							"status": {"name": "Done"}
						}
					}
				]
			}`,
			expectedFields: []string{"key", "summary", "status"},
			expectError:    false,
		},
		{
			name: "Valid Single Issue",
			response: `{
				"key": "PROJ-123",
				"fields": {
					"summary": "Test Issue",
					"description": "Test Description",
					"status": {"name": "In Progress"},
					"assignee": {"displayName": "Test User"}
				}
			}`,
			expectedFields: []string{"key", "summary", "description", "status", "assignee"},
			expectError:    false,
		},
		{
			name:           "Missing Required Fields",
			response:       `{"issues": [{"key": "PROJ-123"}]}`,
			expectedFields: []string{"fields"},
			expectError:    true,
		},
		{
			name:           "Invalid Field Types",
			response:       `{"issues": [{"key": 123, "fields": {"summary": 456}}]}`,
			expectedFields: []string{"key", "summary"},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse response
			var result map[string]interface{}
			err := json.Unmarshal([]byte(tc.response), &result)

			// Check error
			if tc.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			// If no error expected, verify fields
			if !tc.expectError {
				// Helper function to check field existence
				checkFields := func(obj map[string]interface{}, fields []string) {
					for _, field := range fields {
						if _, exists := obj[field]; !exists {
							t.Errorf("Expected field %s to exist", field)
						}
					}
				}

				// Check fields in response
				if issues, ok := result["issues"].([]interface{}); ok {
					// Check first issue fields
					if len(issues) > 0 {
						if issue, ok := issues[0].(map[string]interface{}); ok {
							checkFields(issue, tc.expectedFields)
						}
					}
				} else if fields, ok := result["fields"].(map[string]interface{}); ok {
					// Check single issue fields
					checkFields(fields, tc.expectedFields)
				}
			}
		})
	}
}
