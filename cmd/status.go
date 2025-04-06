package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/plannet-ai/plannet/config"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show a timeline overview of your work",
	Long: `Show a timeline overview of your work based on your git activity.
This command looks at your recent commits and organizes them by time blocks
to give you a clear picture of what you've been working on.`,
	Run: func(cmd *cobra.Command, args []string) {
		runStatus()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus() {
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

	// Get commits from today
	commits, err := getCommitsSince(currentDir, "midnight")
	if err != nil {
		fmt.Println("Error getting commits:", err)
		return
	}

	if len(commits) == 0 {
		fmt.Println("No commits found today.")
		return
	}

	// Group commits by time blocks
	timeBlocks := groupCommitsByTimeBlock(commits)

	// Display timeline
	fmt.Println("Today's map:")
	for _, block := range timeBlocks {
		fmt.Printf("\n%s - %s\n", block.StartTime.Format("15:04"), block.EndTime.Format("15:04"))
		fmt.Printf("Focus: %s\n", block.Focus)
		if len(block.Files) > 0 {
			fmt.Println("Files changed:")
			for _, file := range block.Files {
				fmt.Printf("  - %s\n", file)
			}
		}
	}
}

// TimeBlock represents a period of focused work
type TimeBlock struct {
	StartTime time.Time
	EndTime   time.Time
	Focus     string
	Files     []string
}

// groupCommitsByTimeBlock groups commits into time blocks of focused work
func groupCommitsByTimeBlock(commits []Commit) []TimeBlock {
	if len(commits) == 0 {
		return []TimeBlock{}
	}

	var blocks []TimeBlock
	currentBlock := TimeBlock{
		StartTime: commits[0].Time,
		EndTime:   commits[0].Time,
		Focus:     commits[0].Message,
	}

	// Get files changed in the first commit
	if files, err := getFilesChanged(".", commits[0].Hash); err == nil {
		currentBlock.Files = files
	}

	for i := 1; i < len(commits); i++ {
		commit := commits[i]
		timeDiff := currentBlock.StartTime.Sub(commit.Time)

		// If commits are within 30 minutes of each other, consider them part of the same block
		if timeDiff < 30*time.Minute {
			currentBlock.StartTime = commit.Time
			currentBlock.Focus = commit.Message

			// Add files changed in this commit
			if files, err := getFilesChanged(".", commit.Hash); err == nil {
				currentBlock.Files = append(currentBlock.Files, files...)
			}
		} else {
			// Start a new block
			blocks = append(blocks, currentBlock)
			currentBlock = TimeBlock{
				StartTime: commit.Time,
				EndTime:   commit.Time,
				Focus:     commit.Message,
			}

			// Get files changed in this commit
			if files, err := getFilesChanged(".", commit.Hash); err == nil {
				currentBlock.Files = files
			}
		}
	}

	// Add the last block
	blocks = append(blocks, currentBlock)

	return blocks
}
