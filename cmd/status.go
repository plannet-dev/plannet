package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
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
		// Format time range
		startTime := block.StartTime.Format("15:04")
		endTime := block.EndTime.Format("15:04")
		if endTime == "00:00" {
			endTime = "now"
		}

		// Display time block
		fmt.Printf("%s – %s: ", startTime, endTime)
		
		// Display ticket ID if available
		if block.TicketID != "" {
			fmt.Printf("%s", block.TicketID)
		} else {
			fmt.Printf("untracked work")
		}

		// Display commit count
		fmt.Printf(" (%d commits)\n", len(block.Commits))
		
		// Display file changes if available
		if len(block.Files) > 0 {
			fmt.Printf("  Files: %s\n", strings.Join(block.Files, ", "))
		}
	}
}

// TimeBlock represents a block of time with associated commits
type TimeBlock struct {
	StartTime time.Time
	EndTime   time.Time
	TicketID  string
	Commits   []Commit
	Files     []string
}

// getCommitsSince gets commits since a given time reference
func getCommitsSince(dir string, since string) ([]Commit, error) {
	// Convert "midnight" to the appropriate git time reference
	timeRef := "midnight"
	if since == "midnight" {
		timeRef = "00:00:00"
	}

	cmd := exec.Command("git", "log", "--since", timeRef, "--format=%H|%s|%ct")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	commits := make([]Commit, 0, len(lines))

	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) >= 2 {
			hash := parts[0]
			message := parts[1]
			
			var time time.Time
			if len(parts) >= 3 {
				timestamp := parts[2]
				unixTime, err := time.Parse("1234567890", timestamp)
				if err == nil {
					time = time.Unix(unixTime.Unix(), 0)
				}
			}

			commits = append(commits, Commit{
				Hash:    hash,
				Message: message,
				Time:    time,
			})
		}
	}

	return commits, nil
}

// groupCommitsByTimeBlock groups commits into time blocks
func groupCommitsByTimeBlock(commits []Commit) []TimeBlock {
	if len(commits) == 0 {
		return []TimeBlock{}
	}

	// Sort commits by time (newest first)
	sortedCommits := make([]Commit, len(commits))
	copy(sortedCommits, commits)
	
	// Group commits into time blocks (30-minute intervals)
	timeBlocks := []TimeBlock{}
	
	// Get the current time
	now := time.Now()
	
	// Start with the most recent commit
	currentBlock := TimeBlock{
		StartTime: commits[0].Time,
		EndTime:   now,
		Commits:   []Commit{commits[0]},
	}
	
	// Extract ticket ID from the first commit
	cfg, _ := config.Load()
	if cfg != nil {
		currentBlock.TicketID = extractTicketIDFromMessage(commits[0].Message, cfg.TicketPrefixes)
	}
	
	// Get files changed in the first commit
	files, _ := getFilesChanged(commits[0].Hash)
	currentBlock.Files = files
	
	// Process the rest of the commits
	for i := 1; i < len(commits); i++ {
		commit := commits[i]
		
		// If this commit is within 30 minutes of the previous block's start time,
		// add it to the current block
		if currentBlock.StartTime.Sub(commit.Time) < 30*time.Minute {
			currentBlock.StartTime = commit.Time
			currentBlock.Commits = append(currentBlock.Commits, commit)
			
			// Update ticket ID if this commit has one and the current block doesn't
			if currentBlock.TicketID == "" {
				if cfg != nil {
					currentBlock.TicketID = extractTicketIDFromMessage(commit.Message, cfg.TicketPrefixes)
				}
			}
			
			// Add files changed in this commit
			files, _ := getFilesChanged(commit.Hash)
			currentBlock.Files = append(currentBlock.Files, files...)
		} else {
			// Start a new block
			timeBlocks = append(timeBlocks, currentBlock)
			
			currentBlock = TimeBlock{
				StartTime: commit.Time,
				EndTime:   commits[i-1].Time,
				Commits:   []Commit{commit},
			}
			
			// Extract ticket ID from the commit
			if cfg != nil {
				currentBlock.TicketID = extractTicketIDFromMessage(commit.Message, cfg.TicketPrefixes)
			}
			
			// Get files changed in this commit
			files, _ := getFilesChanged(commit.Hash)
			currentBlock.Files = files
		}
	}
	
	// Add the last block
	timeBlocks = append(timeBlocks, currentBlock)
	
	return timeBlocks
}

// extractTicketIDFromMessage extracts a ticket ID from a commit message
func extractTicketIDFromMessage(message string, prefixes []string) string {
	for _, prefix := range prefixes {
		if strings.Contains(message, prefix) {
			parts := strings.Split(message, prefix)
			if len(parts) > 1 {
				// Extract the ticket ID (e.g., "123" from "JIRA-123: Fix bug")
				ticketPart := parts[1]
				// Find the end of the ticket ID (usually a colon, space, or end of string)
				endIndex := strings.IndexAny(ticketPart, ": ")
				if endIndex == -1 {
					endIndex = len(ticketPart)
				}
				return prefix + ticketPart[:endIndex]
			}
		}
	}
	return ""
}

// getFilesChanged gets the files changed in a commit
func getFilesChanged(hash string) ([]string, error) {
	cmd := exec.Command("git", "show", "--name-only", "--format=", hash)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	return lines, nil
} 