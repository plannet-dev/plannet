package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Config represents the .plannetrc configuration
type Config struct {
	BaseURL      string            `json:"base_url"`
	Model        string            `json:"model"`
	SystemPrompt string            `json:"system_prompt"`
	Headers      map[string]string `json:"headers"`
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

func formatPrompt(systemPrompt, userPrompt string) string {
	const beginText = "<|begin_of_text|>"
	const headerSystem = "<|start_header_id|>system<|end_header_id|>\n"
	const headerUser = "<|start_header_id|>user<|end_header_id|>\n"
	const headerAssistant = "<|start_header_id|>assistant<|end_header_id|>\n"
	const eotId = "<|eot_id|>"

	var prompt string
	prompt = beginText

	// Add system message if present
	if systemPrompt != "" {
		prompt += headerSystem + systemPrompt + eotId
	}

	// Add user message
	prompt += headerUser + userPrompt + eotId + headerAssistant + "\n"

	return prompt
}

func loadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error getting home directory: %v", err)
	}

	configPath := filepath.Join(homeDir, ".plannetrc")
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	// Set defaults if not specified
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:1234/v1/completions"
	}

	return &config, nil
}

func main() {
	prompt := flag.String("prompt", "", "The prompt to send to the LLM")
	flag.Parse()

	if *prompt == "" {
		fmt.Println("Error: prompt is required")
		flag.Usage()
		os.Exit(1)
	}

	config, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Format the prompt according to the template
	formattedPrompt := formatPrompt(config.SystemPrompt, *prompt)

	reqBody := LLMRequest{
		Model:  config.Model,
		Prompt: formattedPrompt,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}

	req, err := http.NewRequest("POST", config.BaseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}

	// Set headers from config
	req.Header.Set("Content-Type", "application/json")
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	var llmResp LLMResponse
	if err := json.Unmarshal(body, &llmResp); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	var output string
	if len(llmResp.Choices) > 0 {
		output = llmResp.Choices[0].Text
		fmt.Println(output)

		fmt.Print("\nCopy to clipboard? [y/n]: ")
		var response string
		fmt.Scanln(&response)

		if response == "y" || response == "Y" {
			cmd := exec.Command("pbcopy") // for macOS
			cmd.Stdin = strings.NewReader(output)
			err := cmd.Run()
			if err != nil {
				// Try xclip for Linux
				cmd = exec.Command("xclip", "-selection", "clipboard")
				cmd.Stdin = strings.NewReader(output)
				err = cmd.Run()
				if err != nil {
					// Try clip for Windows
					cmd = exec.Command("clip")
					cmd.Stdin = strings.NewReader(output)
					err = cmd.Run()
					if err != nil {
						fmt.Printf("Error copying to clipboard: %v\n", err)
						return
					}
				}
			}
			fmt.Println("Copied to clipboard!")
		}
	}
}
