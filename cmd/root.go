package cmd

import (
	"context"
	"os"

	"github.com/google/uuid"
	"github.com/plannet-ai/plannet/logger"
	"github.com/spf13/cobra"
)

var (
	// Version is the version of Plannet
	Version = "0.1.0"
	// Debug mode flag
	debug bool
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
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Create a context with trace ID
		ctx := context.WithValue(cmd.Context(), "trace_id", uuid.New().String())
		cmd.SetContext(ctx)

		// Set debug level if flag is set
		if debug {
			logger.SetLevel(logger.DebugLevel)
			logger.Debug("Debug mode enabled")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		log := logger.WithContext(cmd.Context())
		log.Info("Plannet version %s", Version)
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	// Create initial context
	ctx := context.Background()
	rootCmd.SetContext(ctx)

	if err := rootCmd.Execute(); err != nil {
		logger.WithContext(ctx).Error("Failed to execute command: %v", err)
		os.Exit(1)
	}
	return nil
}

func init() {
	// Add global flags
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")

	// Add version flag
	rootCmd.Flags().BoolP("version", "v", false, "Show version information")

	// Add custom version template
	rootCmd.SetVersionTemplate(`Plannet version {{.Version}}
`)
}
