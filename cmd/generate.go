package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/plannet-ai/plannet/config"
	"github.com/plannet-ai/plannet/llm"
	"github.com/plannet-ai/plannet/output"
)

var (
	prompt string
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate [prompt]",
	Short: "Generate content using the LLM",
	Long: `Generate content using the LLM.
This command allows you to generate content based on a prompt.
If no prompt is provided, it will use the --prompt flag.`,
	Run: func(cmd *cobra.Command, args []string) {
		runGenerate(args)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// Add flags
	generateCmd.Flags().StringVarP(&prompt, "prompt", "p", "", "Prompt for content generation")
}

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