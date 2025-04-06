package cmd

import (
	"fmt"
	"strings"

	"github.com/plannet-ai/plannet/config"
	"github.com/plannet-ai/plannet/llm"
	"github.com/plannet-ai/plannet/output"
)

// prompt is the prompt for content generation
var prompt string

// runGenerate executes the generate command
func runGenerate(args []string) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Error loading configuration:", err)
		fmt.Println("Run 'plannet init' to set up your configuration.")
		return
	}

	// Get prompt from args or flag
	userPrompt := prompt
	if len(args) > 0 {
		userPrompt = strings.Join(args, " ")
	}

	if userPrompt == "" {
		fmt.Println("Error: No prompt provided.")
		fmt.Println("Usage: plannet generate [prompt] or plannet generate --prompt \"your prompt\"")
		return
	}

	// Create generator
	generator := llm.NewGenerator(cfg)

	// Generate content
	content, err := generator.Generate(userPrompt)
	if err != nil {
		fmt.Println("Error generating content:", err)
		return
	}

	// Handle output
	if err := output.HandleOutput(content, cfg); err != nil {
		fmt.Println("Error handling output:", err)
		return
	}
}
