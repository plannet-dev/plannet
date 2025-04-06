package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/plannet-ai/plannet/config"
)

// llmTestConfig is used to create a test configuration
type llmTestConfig struct {
	*config.Config
}

func (m *llmTestConfig) Load() error {
	return nil
}

// makeLLMRequest sends a request to the LLM API
func makeLLMRequest(cfg *llmTestConfig, prompt string) (string, error) {
	// Check if LLM integration is configured
	if cfg.Config.BaseURL == "" || cfg.Config.Model == "" || cfg.Config.SystemPrompt == "" {
		return "", fmt.Errorf("LLM integration is not configured")
	}

	// Create request body
	requestBody := map[string]interface{}{
		"model": cfg.Config.Model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": cfg.Config.SystemPrompt,
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	// Marshal request body
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("POST", cfg.Config.BaseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create LLM API request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if cfg.Config.Headers != nil {
		if apiKey, ok := cfg.Config.Headers["Authorization"]; ok {
			req.Header.Set("Authorization", apiKey)
		}
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send LLM API request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LLM API returned status %d", resp.StatusCode)
	}

	// Parse response
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse LLM API response: %w", err)
	}

	// Check if we have a response
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM API")
	}

	return result.Choices[0].Message.Content, nil
}

// TestLLMRequest tests the makeLLMRequest function
func TestLLMRequest(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		prompt         string
		serverResponse string
		statusCode     int
		expectError    bool
	}{
		{
			name:   "Success",
			prompt: "Test prompt",
			serverResponse: `{
				"choices": [
					{
						"message": {
							"content": "Test response"
						}
					}
				]
			}`,
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name:           "Server Error",
			prompt:         "Test prompt",
			serverResponse: `{"error": "Internal Server Error"}`,
			statusCode:     http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:           "Invalid JSON",
			prompt:         "Test prompt",
			serverResponse: `{invalid json}`,
			statusCode:     http.StatusOK,
			expectError:    true,
		},
		{
			name:           "Empty Response",
			prompt:         "Test prompt",
			serverResponse: `{"choices": []}`,
			statusCode:     http.StatusOK,
			expectError:    true,
		},
		{
			name:   "Missing Content",
			prompt: "Test prompt",
			serverResponse: `{
				"choices": [
					{
						"message": {}
					}
				]
			}`,
			statusCode:  http.StatusOK,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST request, got %s", r.Method)
				}

				// Verify content type
				contentType := r.Header.Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", contentType)
				}

				// Set response
				w.WriteHeader(tc.statusCode)
				w.Write([]byte(tc.serverResponse))
			}))
			defer server.Close()

			// Create mock config
			cfg := &llmTestConfig{
				Config: &config.Config{
					BaseURL:      server.URL,
					Model:        "test-model",
					SystemPrompt: "test-system-prompt",
					Headers: map[string]string{
						"Authorization": "Bearer test-api-key",
					},
				},
			}

			// Test makeLLMRequest
			response, err := makeLLMRequest(cfg, tc.prompt)

			// Check error
			if tc.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			// Check response
			if !tc.expectError && response != "Test response" {
				t.Errorf("Expected response 'Test response', got '%s'", response)
			}
		})
	}
}

// TestLLMIntegration tests the LLM integration configuration
func TestLLMIntegration(t *testing.T) {
	// Test cases
	testCases := []struct {
		name        string
		config      *config.Config
		expectError bool
	}{
		{
			name: "Valid Configuration",
			config: &config.Config{
				BaseURL:      "https://api.test.com",
				Model:        "test-model",
				SystemPrompt: "test-system-prompt",
				Headers: map[string]string{
					"Authorization": "Bearer test-api-key",
				},
			},
			expectError: false,
		},
		{
			name: "Missing Base URL",
			config: &config.Config{
				Model:        "test-model",
				SystemPrompt: "test-system-prompt",
				Headers: map[string]string{
					"Authorization": "Bearer test-api-key",
				},
			},
			expectError: true,
		},
		{
			name: "Missing Model",
			config: &config.Config{
				BaseURL:      "https://api.test.com",
				SystemPrompt: "test-system-prompt",
				Headers: map[string]string{
					"Authorization": "Bearer test-api-key",
				},
			},
			expectError: true,
		},
		{
			name: "Missing System Prompt",
			config: &config.Config{
				BaseURL: "https://api.test.com",
				Model:   "test-model",
				Headers: map[string]string{
					"Authorization": "Bearer test-api-key",
				},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock config
			cfg := &llmTestConfig{
				Config: tc.config,
			}

			// Test makeLLMRequest (which checks integration)
			_, err := makeLLMRequest(cfg, "test prompt")

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

// TestLLMResponseParsing tests the parsing of LLM API responses
func TestLLMResponseParsing(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		response       string
		expectedFields []string
		expectError    bool
	}{
		{
			name: "Valid Response",
			response: `{
				"choices": [
					{
						"message": {
							"content": "Test response"
						}
					}
				]
			}`,
			expectedFields: []string{"choices", "message", "content"},
			expectError:    false,
		},
		{
			name: "Missing Choices",
			response: `{
				"message": {
					"content": "Test response"
				}
			}`,
			expectedFields: []string{"choices"},
			expectError:    true,
		},
		{
			name: "Empty Choices",
			response: `{
				"choices": []
			}`,
			expectedFields: []string{"choices"},
			expectError:    true,
		},
		{
			name: "Missing Message",
			response: `{
				"choices": [
					{}
				]
			}`,
			expectedFields: []string{"message"},
			expectError:    true,
		},
		{
			name: "Missing Content",
			response: `{
				"choices": [
					{
						"message": {}
					}
				]
			}`,
			expectedFields: []string{"content"},
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
				if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
					if choice, ok := choices[0].(map[string]interface{}); ok {
						if message, ok := choice["message"].(map[string]interface{}); ok {
							checkFields(message, tc.expectedFields)
						}
					}
				}
			}
		})
	}
}
