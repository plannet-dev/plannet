package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/plannet-ai/plannet/config"
)

// mockConfig is used to mock the config for testing
var mockConfig *config.Config

func init() {
	// Initialize mock config
	mockConfig = &config.Config{
		JiraURL:   "http://localhost:8080",
		JiraToken: "test-token",
		JiraUser:  "test-user",
	}
}

// getConfig returns the mock config for testing
func getConfig() (*config.Config, error) {
	return mockConfig, nil
}

func TestJiraList(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Basic dGVzdC11c2VyOnRlc3QtdG9rZW4=" { // Base64 encoded "test-user:test-token"
			t.Errorf("Expected Authorization to be Basic dGVzdC11c2VyOnRlc3QtdG9rZW4=, got %s", authHeader)
		}

		// Check URL
		expectedURL := "/rest/api/2/search"
		if r.URL.Path != expectedURL {
			t.Errorf("Expected URL to be %s, got %s", expectedURL, r.URL.Path)
		}

		// Return a mock response
		response := struct {
			Issues []struct {
				Key    string `json:"key"`
				Fields struct {
					Summary string `json:"summary"`
					Status  struct {
						Name string `json:"name"`
					} `json:"status"`
				} `json:"fields"`
			} `json:"issues"`
		}{
			Issues: []struct {
				Key    string `json:"key"`
				Fields struct {
					Summary string `json:"summary"`
					Status  struct {
						Name string `json:"name"`
					} `json:"status"`
				} `json:"fields"`
			}{
				{
					Key: "TEST-123",
					Fields: struct {
						Summary string `json:"summary"`
						Status  struct {
							Name string `json:"name"`
						} `json:"status"`
					}{
						Summary: "Test ticket",
						Status: struct {
							Name string `json:"name"`
						}{
							Name: "In Progress",
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Update mock config with server URL
	mockConfig.JiraURL = server.URL

	// Call the function
	runJiraList()
}

func TestJiraView(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Basic dGVzdC11c2VyOnRlc3QtdG9rZW4=" { // Base64 encoded "test-user:test-token"
			t.Errorf("Expected Authorization to be Basic dGVzdC11c2VyOnRlc3QtdG9rZW4=, got %s", authHeader)
		}

		// Check URL
		expectedURL := "/rest/api/2/issue/TEST-123"
		if r.URL.Path != expectedURL {
			t.Errorf("Expected URL to be %s, got %s", expectedURL, r.URL.Path)
		}

		// Return a mock response
		response := struct {
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
		}{
			Key: "TEST-123",
			Fields: struct {
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
			}{
				Summary:     "Test ticket",
				Description: "Test description",
				Status: struct {
					Name string `json:"name"`
				}{
					Name: "In Progress",
				},
				Type: struct {
					Name string `json:"name"`
				}{
					Name: "Task",
				},
				Priority: struct {
					Name string `json:"name"`
				}{
					Name: "High",
				},
				Assignee: struct {
					DisplayName string `json:"displayName"`
				}{
					DisplayName: "Test User",
				},
				Created: "2023-01-01T00:00:00.000Z",
				Updated: "2023-01-02T00:00:00.000Z",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Update mock config with server URL
	mockConfig.JiraURL = server.URL

	// Call the function
	runJiraView("TEST-123")
}

func TestJiraCreate(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Basic dGVzdC11c2VyOnRlc3QtdG9rZW4=" { // Base64 encoded "test-user:test-token"
			t.Errorf("Expected Authorization to be Basic dGVzdC11c2VyOnRlc3QtdG9rZW4=, got %s", authHeader)
		}

		// Check URL
		expectedURL := "/rest/api/2/issue"
		if r.URL.Path != expectedURL {
			t.Errorf("Expected URL to be %s, got %s", expectedURL, r.URL.Path)
		}

		// Parse request body
		var requestBody map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		// Check request body
		fields, ok := requestBody["fields"].(map[string]interface{})
		if !ok {
			t.Error("Expected fields to be a map")
		}

		project, ok := fields["project"].(map[string]interface{})
		if !ok {
			t.Error("Expected project to be a map")
		}

		projectKey, ok := project["key"].(string)
		if !ok {
			t.Error("Expected project key to be a string")
		}
		if projectKey != "TEST" {
			t.Errorf("Expected project key to be TEST, got %s", projectKey)
		}

		summary, ok := fields["summary"].(string)
		if !ok {
			t.Error("Expected summary to be a string")
		}
		if summary != "Test summary" {
			t.Errorf("Expected summary to be Test summary, got %s", summary)
		}

		description, ok := fields["description"].(string)
		if !ok {
			t.Error("Expected description to be a string")
		}
		if description != "Test description" {
			t.Errorf("Expected description to be Test description, got %s", description)
		}

		issuetype, ok := fields["issuetype"].(map[string]interface{})
		if !ok {
			t.Error("Expected issuetype to be a map")
		}

		typeName, ok := issuetype["name"].(string)
		if !ok {
			t.Error("Expected type name to be a string")
		}
		if typeName != "Task" {
			t.Errorf("Expected type name to be Task, got %s", typeName)
		}

		// Return a mock response
		response := struct {
			Key string `json:"key"`
		}{
			Key: "TEST-123",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Update mock config with server URL
	mockConfig.JiraURL = server.URL

	// Call the function
	runJiraCreate("Test summary", "Test description", "Task", "TEST")
} 