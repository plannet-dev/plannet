package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/plannet-ai/plannet/config"
)

// Generator handles all LLM interaction and prompt generation
type Generator struct {
	config *config.Config
	client *http.Client
}

// Request represents the request body for the LLM API
type Request struct {
	Model     string `json:"model"`
	Prompt    string `json:"prompt"`
	MaxTokens int    `json:"max_tokens,omitempty"`
}

// Response represents the response from the LLM API
type Response struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int    `json:"index"`
		Text         string `json:"text"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// NewGenerator creates a new Generator instance
func NewGenerator(cfg *config.Config) *Generator {
	return &Generator{
		config: cfg,
		client: &http.Client{},
	}
}

// Generate takes a prompt and returns the generated text
func (g *Generator) Generate(prompt string) (string, error) {
	formattedPrompt := g.formatPrompt(prompt)

	reqBody := Request{
		Model:  g.config.Model,
		Prompt: formattedPrompt,
	}

	response, err := g.makeRequest(reqBody)
	if err != nil {
		return "", fmt.Errorf("generation failed: %w", err)
	}

	return g.extractResponse(response)
}

// formatPrompt formats the prompt according to the model's expected format
func (g *Generator) formatPrompt(prompt string) string {
	if g.config.SystemPrompt != "" {
		return fmt.Sprintf("%s\n\nUser: %s\n\nAssistant:", g.config.SystemPrompt, prompt)
	}
	return fmt.Sprintf("User: %s\n\nAssistant:", prompt)
}

// makeRequest sends a request to the LLM API
func (g *Generator) makeRequest(reqBody Request) (*Response, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", g.config.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for key, value := range g.config.Headers {
		req.Header.Set(key, value)
	}

	// Send request
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// Parse response
	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return &response, nil
}

// extractResponse extracts the generated text from the response
func (g *Generator) extractResponse(response *Response) (string, error) {
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return response.Choices[0].Text, nil
}
