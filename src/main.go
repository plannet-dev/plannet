package main

import (
	"flag"
	"fmt"
	"os"

	"runtime/debug"
)

var (
	// Command-line flags
	customPrompt = flag.String("prompt", "", "Custom prompt (if not using Jira)")
	showVersion  = flag.Bool("version", false, "Show version information")
	isDebugging  = flag.Bool("debug", false, "Enable debug mode")
)

// Version information
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Defer panic handler
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Fatal error: %v\n", r)
			if *isDebugging {
				fmt.Fprintf(os.Stderr, "Stack trace:\n%s\n", debug.Stack())
			}
			os.Exit(1)
		}
	}()

	// Parse command line flags
	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("plannet version %s\n", version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("built at: %s\n", date)
		os.Exit(0)
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create and run application
	app := NewApp(config)
	if err := app.Run(*customPrompt); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// init sets up any necessary initialization
func init() {
	// Customize flag usage
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of plannet:\n")
		fmt.Fprintf(os.Stderr, "\nFlags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample usage:\n")
		fmt.Fprintf(os.Stderr, "  plannet --prompt \"Generate a test plan\"\n")
		fmt.Fprintf(os.Stderr, "  plannet                              # Use Jira integration if configured\n")
	}
}
