package cmd

import (
	"os"

	"github.com/plannet-ai/plannet/logger"
	"github.com/spf13/cobra"
)

var philosophyCmd = &cobra.Command{
	Use:   "philosophy",
	Short: "Display Plannet's philosophy",
	Long:  `Display the core philosophy and principles behind Plannet`,
	Run: func(cmd *cobra.Command, args []string) {
		content, err := os.ReadFile("philosophy.md")
		if err != nil {
			logger.WithContext(cmd.Context()).Error("Failed to read philosophy: %v", err)
			return
		}
		cmd.Print(string(content))
	},
}

func init() {
	rootCmd.AddCommand(philosophyCmd)
}
