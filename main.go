package main

import (
	"log"
	"os"

	"github.com/plannet-ai/plannet/cmd"
)

func main() {
	// Set up logging
	log.SetOutput(os.Stderr)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Execute the root command
	cmd.Execute()
} 