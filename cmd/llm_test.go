package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/plannet-ai/plannet/config"
)

func TestLLMRequest(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check content type
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type to be application/json, got %s", contentType)
		}

		// Check authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-token" {
			t.Errorf("Expected Authorization to be Bearer test-token, got %s", authHeader)
		}

		// Parse request body
		var requestBody LLMRequest
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		// Check request body
		if requestBody.Model != "test-model" {
			t.Errorf("Expected model to be test-model, got %s", requestBody.Model)
		}
		if len(requestBody.Messages) != 2 {
			t.Errorf("Expected 2 messages, got %d", len(requestBody.Messages))
		}
		if requestBody.Messages[0].Role != "system" {
			t.Errorf("Expected first message role to be system, got %s", requestBody.Messages[0].Role)
		}
		if requestBody.Messages[0].Content != "test-system-prompt" {
			t.Errorf("Expected first message content to be test-system-prompt, got %s", requestBody.Messages[0].Content)
		}
		if requestBody.Messages[1].Role != "user" {
			t.Errorf("Expected second message role to be user, got %s", requestBody.Messages[1].Role)
		}
		if requestBody.Messages[1].Content != "test-prompt" {
			t.Errorf("Expected second message content to be test-prompt, got %s", requestBody.Messages[1].Content)
		}

		// Return a mock response
		response := LLMResponse{
			ID:      "test-id",
			Object:  "chat.completion",
			Created: 1234567890,
			Model:   "test-model",
			Choices: []struct {
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{
					Message: struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "assistant",
						Content: "test-response",
					},
					FinishReason: "stop",
				},
			},
			Usage: struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			}{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create a test config
	cfg := &config.Config{
		BaseURL: server.URL,
		Model:   "test-model",
		Headers: map[string]string{
			"Authorization": "Bearer test-token",
		},
		SystemPrompt: "test-system-prompt",
	}

	// Create test messages
	messages := []Message{
		{
			Role:    "system",
			Content: "test-system-prompt",
		},
		{
			Role:    "user",
			Content: "test-prompt",
		},
	}

	// Send request
	response, err := sendLLMRequest(cfg, messages)
	if err != nil {
		t.Fatalf("Failed to send LLM request: %v", err)
	}

	// Check response
	if response.ID != "test-id" {
		t.Errorf("Expected ID to be test-id, got %s", response.ID)
	}
	if response.Object != "chat.completion" {
		t.Errorf("Expected Object to be chat.completion, got %s", response.Object)
	}
	if response.Model != "test-model" {
		t.Errorf("Expected Model to be test-model, got %s", response.Model)
	}
	if len(response.Choices) != 1 {
		t.Errorf("Expected 1 choice, got %d", len(response.Choices))
	}
	if response.Choices[0].Message.Role != "assistant" {
		t.Errorf("Expected message role to be assistant, got %s", response.Choices[0].Message.Role)
	}
	if response.Choices[0].Message.Content != "test-response" {
		t.Errorf("Expected message content to be test-response, got %s", response.Choices[0].Message.Content)
	}
	if response.Choices[0].FinishReason != "stop" {
		t.Errorf("Expected finish reason to be stop, got %s", response.Choices[0].FinishReason)
	}
	if response.Usage.PromptTokens != 10 {
		t.Errorf("Expected prompt tokens to be 10, got %d", response.Usage.PromptTokens)
	}
	if response.Usage.CompletionTokens != 5 {
		t.Errorf("Expected completion tokens to be 5, got %d", response.Usage.CompletionTokens)
	}
	if response.Usage.TotalTokens != 15 {
		t.Errorf("Expected total tokens to be 15, got %d", response.Usage.TotalTokens)
	}
}

func TestLLMRequestError(t *testing.T) {
	// Create a mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	// Create a test config
	cfg := &config.Config{
		BaseURL: server.URL,
		Model:   "test-model",
		Headers: map[string]string{
			"Authorization": "Bearer test-token",
		},
	}

	// Create test messages
	messages := []Message{
		{
			Role:    "user",
			Content: "test-prompt",
		},
	}

	// Send request
	_, err := sendLLMRequest(cfg, messages)
	if err == nil {
		t.Error("Expected error, got nil")
	}
} 