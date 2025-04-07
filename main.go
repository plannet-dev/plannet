package main

import (
	"os"

	"github.com/plannet-ai/plannet/cmd"
	"github.com/plannet-ai/plannet/logger"
)

func main() {
	// Set up logging
	if os.Getenv("DEBUG") == "1" {
		logger.SetLevel(logger.DebugLevel)
	}

	// Execute the root command
	if err := cmd.Execute(); err != nil {
		logger.Fatal("Failed to execute root command: %v", err)
	}
}
