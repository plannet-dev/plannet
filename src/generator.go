package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Generator handles all LLM interaction and prompt generation
type Generator struct {
	config *Config
	client *http.Client
}

// LLMRequest represents the request body for the LLM API
type LLMRequest struct {
	Model     string `json:"model"`
	Prompt    string `json:"prompt"`
	MaxTokens int    `json:"max_tokens,omitempty"`
}

// LLMResponse represents the response from the LLM API
type LLMResponse struct {
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
func NewGenerator(config *Config) *Generator {
	return &Generator{
		config: config,
		client: &http.Client{},
	}
}

// Generate takes a prompt and returns the generated text
func (generator *Generator) Generate(prompt string) (string, error) {
	formattedPrompt := generator.formatPrompt(prompt)

	reqBody := LLMRequest{
		Model:  generator.config.Model,
		Prompt: formattedPrompt,
	}

	response, err := generator.makeRequest(reqBody)
	if err != nil {
		return "", fmt.Errorf("generation failed: %w", err)
	}

	return generator.extractResponse(response)
}

// formatPrompt formats the prompt according to the model's expected format
func (generator *Generator) formatPrompt(userPrompt string) string {
	const (
		beginText       = "<|begin_of_text|>"
		headerSystem    = "<|start_header_id|>system<|end_header_id|>\n"
		headerUser      = "<|start_header_id|>user<|end_header_id|>\n"
		headerAssistant = "<|start_header_id|>assistant<|end_header_id|>\n"
		eotId           = "<|eot_id|>"
	)

	var prompt string
	prompt = beginText

	// Add system message if present
	if generator.config.SystemPrompt != "" {
		prompt += headerSystem + generator.config.SystemPrompt + eotId
	}

	// Add user message
	prompt += headerUser + "Respond only with an example ticket. Don't include an ID or number. Create a ticket for" + userPrompt + eotId + headerAssistant + "\n"

	return prompt
}

// makeRequest sends the request to the LLM API
func (generator *Generator) makeRequest(reqBody LLMRequest) (*LLMResponse, error) {
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", generator.config.BaseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for key, value := range generator.config.Headers {
		req.Header.Set(key, value)
	}

	// Make the request
	resp, err := generator.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-200 status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var llmResp LLMResponse
	if err := json.Unmarshal(body, &llmResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	fmt.Printf("%+v\n", llmResp)

	return &llmResp, nil
}

// extractResponse extracts the generated text from the API response
func (generator *Generator) extractResponse(resp *LLMResponse) (string, error) {
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	if resp.Choices[0].Text == "" {
		return "", fmt.Errorf("empty response text")
	}

	return resp.Choices[0].Text, nil
}
