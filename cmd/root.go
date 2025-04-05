package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/plannet-ai/plannet/config"
	"github.com/plannet-ai/plannet/llm"
	"github.com/plannet-ai/plannet/output"
)

var (
	// Version is the version of Plannet
	Version = "0.1.0"
	// Debug mode flag
	debug bool
	// Prompt for content generation
	prompt string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "plannet",
	Short: "A command-line tool that tracks the work Jira misses",
	Long: `Plannet is a command-line tool that helps you stay on top of your workload 
and backlog from where you work, the command line.

It tracks what you're working on, even when it doesn't make it into Jira or other
ticketing systems. No more un-tracked side quests.`,
	Version: Version,
	Run: func(cmd *cobra.Command, args []string) {
		// If a prompt is provided, generate content
		if prompt != "" {
			runGenerate(args)
			return
		}

		// Otherwise, show help
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Add global flags
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")
	
	// Add version flag
	rootCmd.Flags().BoolP("version", "v", false, "Show version information")
	
	// Add prompt flag
	rootCmd.Flags().StringVarP(&prompt, "prompt", "p", "", "Prompt for content generation")
	
	// Add custom version template
	rootCmd.SetVersionTemplate(`Plannet version {{.Version}}
`)
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
		fmt.Println("Usage: plannet --prompt \"your prompt\" or plannet generate [prompt]")
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