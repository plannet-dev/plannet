package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/plannet-ai/plannet/config"
)

// nowCmd represents the now command
var nowCmd = &cobra.Command{
	Use:   "now",
	Short: "Show what you're currently working on",
	Long: `Show what you're currently working on based on your git activity.
This command looks at your current branch and recent commits to determine
what you're focused on, including any "side quests" that aren't tracked
in your ticketing system.`,
	Run: func(cmd *cobra.Command, args []string) {
		runNow()
	},
}

func init() {
	rootCmd.AddCommand(nowCmd)
}

func runNow() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Error loading configuration:", err)
		fmt.Println("Run 'plannet init' to set up your configuration.")
		return
	}

	// Check if git integration is enabled
	if !cfg.GitIntegration {
		fmt.Println("Git integration is disabled. Enable it in your configuration.")
		return
	}

	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	// Check if we're in a git repository
	if !isGitRepo(currentDir) {
		fmt.Println("Not in a git repository. Plannet works best in git repositories.")
		return
	}

	// Get current branch
	branchName, err := getCurrentBranch()
	if err != nil {
		fmt.Println("Error getting current branch:", err)
		return
	}

	// Extract ticket ID from branch name
	ticketID := extractTicketID(branchName, cfg.TicketPrefixes)

	// Get recent commits
	commits, err := getRecentCommits(5)
	if err != nil {
		fmt.Println("Error getting recent commits:", err)
		return
	}

	// Display current focus
	fmt.Println("Current focus:")
	if ticketID != "" {
		fmt.Printf("  Branch: %s (%s)\n", branchName, ticketID)
	} else {
		fmt.Printf("  Branch: %s (untracked work)\n", branchName)
	}

	// Display recent activity
	fmt.Println("\nRecent activity:")
	for _, commit := range commits {
		// Check if commit has a ticket ID
		commitTicketID := extractTicketIDFromMessage(commit.Message, cfg.TicketPrefixes)
		
		if commitTicketID != "" {
			fmt.Printf("  %s: %s\n", commitTicketID, commit.Message)
		} else {
			fmt.Printf("  [untracked]: %s\n", commit.Message)
		}
	}

	// Find and display side quests
	sideQuests := findSideQuests(commits, cfg.TicketPrefixes)
	if len(sideQuests) > 0 {
		fmt.Println("\nSide quests:")
		for _, quest := range sideQuests {
			fmt.Printf("  %s\n", quest.Message)
		}
	}
} 