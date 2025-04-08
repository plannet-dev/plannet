package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/plannet-ai/plannet/config"
	"github.com/plannet-ai/plannet/logger"
	"github.com/plannet-ai/plannet/security"
	"github.com/spf13/cobra"
)

// Message represents a message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Response represents the response from the LLM API
type Response struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// llmCmd represents the llm command
var llmCmd = &cobra.Command{
	Use:   "llm",
	Short: "Interact with the LLM",
	Long:  `Interact with the LLM to get help with your tasks`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		prompt, _ := cmd.Flags().GetString("prompt")
		if prompt != "" {
			return runLLMWithPrompt(ctx, prompt)
		}
		return runLLMInteractive(ctx)
	},
}

func init() {
	rootCmd.AddCommand(llmCmd)
	llmCmd.Flags().String("prompt", "", "Single prompt to send to the LLM")
}

// runLLMInteractive starts an interactive session with the LLM
func runLLMInteractive(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration: %v", err)
		return err
	}

	if cfg.BaseURL == "" || cfg.Model == "" {
		logger.Error("LLM integration is not configured. Please run 'plannet init' first")
		return fmt.Errorf("LLM integration not configured")
	}

	// Get LLM token from config
	token := cfg.LLMToken
	if token == "" {
		fmt.Println("Error: LLM token not found. Please run 'plannet init' to set up LLM integration.")
		return fmt.Errorf("LLM token not found")
	}

	logger.Info("Starting interactive session with LLM. Type 'exit' to quit.")
	logger.Info("Type your message and press Enter:")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			var input string
			fmt.Print("> ")
			fmt.Scanln(&input)

			if strings.ToLower(input) == "exit" {
				return nil
			}

			response, err := sendLLMRequest(ctx, cfg, input)
			if err != nil {
				logger.Error("Failed to get response: %v", err)
				continue
			}

			logger.Info("LLM: %s", response)
		}
	}
}

// runLLMWithPrompt sends a single prompt to the LLM
func runLLMWithPrompt(ctx context.Context, prompt string) error {
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration: %v", err)
		return err
	}

	if cfg.BaseURL == "" || cfg.Model == "" {
		logger.Error("LLM integration is not configured. Please run 'plannet init' first")
		return fmt.Errorf("LLM integration not configured")
	}

	// Get LLM token from config
	token := cfg.LLMToken
	if token == "" {
		fmt.Println("Error: LLM token not found. Please run 'plannet init' to set up LLM integration.")
		return fmt.Errorf("LLM token not found")
	}

	response, err := sendLLMRequest(ctx, cfg, prompt)
	if err != nil {
		logger.Error("Failed to get response: %v", err)
		return err
	}

	logger.Info("LLM: %s", response)
	return nil
}

// sendLLMRequest sends a request to the LLM API
func sendLLMRequest(ctx context.Context, cfg *config.Config, prompt string) (string, error) {
	// Get LLM token from config
	token := cfg.LLMToken
	if token == "" {
		fmt.Println("Error: LLM token not found. Please run 'plannet init' to set up LLM integration.")
		return "", fmt.Errorf("LLM token not found")
	}

	// Create rate limiter: 5 requests per minute
	rateLimiter := security.NewHTTPRateLimiter(5, time.Minute)
	baseClient := &http.Client{}
	client := rateLimiter.WrapHTTPClient(baseClient, "llm")

	messages := []Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	requestBody, err := json.Marshal(map[string]interface{}{
		"model":    cfg.Model,
		"messages": messages,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", cfg.BaseURL, bytes.NewReader(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	return response.Choices[0].Message.Content, nil
}
