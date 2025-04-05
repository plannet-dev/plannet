package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/plannet-ai/plannet/config"
	"github.com/spf13/cobra"
)

// LLMRequest represents a request to the LLM API
type LLMRequest struct {
	Model       string   `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64  `json:"temperature,omitempty"`
	MaxTokens   int      `json:"max_tokens,omitempty"`
}

// Message represents a message in the LLM conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LLMResponse represents a response from the LLM API
type LLMResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// llmCmd represents the llm command
var llmCmd = &cobra.Command{
	Use:   "llm [prompt]",
	Short: "Interact with the LLM",
	Long: `Interact with the LLM to generate content.
This command allows you to send prompts to the LLM and get responses.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			runLLMInteractive()
		} else {
			runLLMWithPrompt(strings.Join(args, " "))
		}
	},
}

func init() {
	rootCmd.AddCommand(llmCmd)
}

// runLLMInteractive starts an interactive session with the LLM
func runLLMInteractive() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Error loading configuration:", err)
		fmt.Println("Run 'plannet init' to set up your configuration.")
		return
	}

	// Check if LLM integration is configured
	if cfg.BaseURL == "" || cfg.Model == "" || len(cfg.Headers) == 0 {
		fmt.Println("LLM integration is not configured.")
		fmt.Println("Run 'plannet init' to set up LLM integration.")
		return
	}

	fmt.Println("Starting interactive session with the LLM.")
	fmt.Println("Type 'exit' to end the session.")
	fmt.Println("----------------------------------------")

	// Initialize conversation history
	var messages []Message

	// Add system prompt if available
	if cfg.SystemPrompt != "" {
		messages = append(messages, Message{
			Role:    "system",
			Content: cfg.SystemPrompt,
		})
	}

	// Start interactive loop
	for {
		// Prompt for user input
		fmt.Print("You: ")
		input, err := os.Stdin.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			return
		}
		input = strings.TrimSpace(input)

		// Check if user wants to exit
		if input == "exit" {
			break
		}

		// Add user message to history
		messages = append(messages, Message{
			Role:    "user",
			Content: input,
		})

		// Send request to LLM
		response, err := sendLLMRequest(cfg, messages)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		// Display response
		if len(response.Choices) > 0 {
			content := response.Choices[0].Message.Content
			fmt.Println("LLM:", content)

			// Add assistant message to history
			messages = append(messages, Message{
				Role:    "assistant",
				Content: content,
			})
		} else {
			fmt.Println("No response from LLM.")
		}
	}
}

// runLLMWithPrompt sends a single prompt to the LLM
func runLLMWithPrompt(prompt string) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Error loading configuration:", err)
		fmt.Println("Run 'plannet init' to set up your configuration.")
		return
	}

	// Check if LLM integration is configured
	if cfg.BaseURL == "" || cfg.Model == "" || len(cfg.Headers) == 0 {
		fmt.Println("LLM integration is not configured.")
		fmt.Println("Run 'plannet init' to set up LLM integration.")
		return
	}

	// Initialize messages
	var messages []Message

	// Add system prompt if available
	if cfg.SystemPrompt != "" {
		messages = append(messages, Message{
			Role:    "system",
			Content: cfg.SystemPrompt,
		})
	}

	// Add user message
	messages = append(messages, Message{
		Role:    "user",
		Content: prompt,
	})

	// Send request to LLM
	response, err := sendLLMRequest(cfg, messages)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Display response
	if len(response.Choices) > 0 {
		fmt.Println(response.Choices[0].Message.Content)
	} else {
		fmt.Println("No response from LLM.")
	}
}

// sendLLMRequest sends a request to the LLM API
func sendLLMRequest(cfg *config.Config, messages []Message) (*LLMResponse, error) {
	// Create request body
	requestBody := LLMRequest{
		Model:       cfg.Model,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   1000,
	}

	// Convert to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request body: %w", err)
	}

	// Create HTTP client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("POST", cfg.BaseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	for key, value := range cfg.Headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("LLM API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response LLMResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return &response, nil
} 