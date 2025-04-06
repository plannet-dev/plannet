package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/plannet-ai/plannet/config"
	"github.com/spf13/cobra"
)

// WorkContext represents the git context of tracked work
type WorkContext struct {
	Branch     string   `json:"branch,omitempty"`
	Files      []string `json:"files,omitempty"`
	CommitHash string   `json:"commit_hash,omitempty"`
}

// TrackedWork represents a piece of work tracked by the user
type TrackedWork struct {
	ID          string      `json:"id"`
	Description string      `json:"description"`
	TicketID    string      `json:"ticket_id,omitempty"`
	StartTime   time.Time   `json:"start_time"`
	EndTime     time.Time   `json:"end_time,omitempty"`
	Tags        []string    `json:"tags,omitempty"`
	Status      string      `json:"status"` // "active", "paused", "completed"
	Context     WorkContext `json:"context,omitempty"`
}

// trackCmd represents the track command
var trackCmd = &cobra.Command{
	Use:   "track [description]",
	Short: "Track a piece of work manually",
	Long: `Track a piece of work manually that isn't captured by git.
This command allows you to record work that doesn't involve code changes,
such as meetings, documentation, or research.`,
	Run: func(cmd *cobra.Command, args []string) {
		runTrack(args)
	},
}

func init() {
	rootCmd.AddCommand(trackCmd)
}

func runTrack(args []string) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\nRun 'plannet init' to set up your configuration.\n", err)
		return
	}

	// Check for active work
	activeWork, err := getActiveWork()
	if err != nil {
		fmt.Printf("Failed to check active work: %v\n", err)
		return
	}

	if activeWork != nil {
		fmt.Println("You have active work:")
		fmt.Printf("Description: %s\n", activeWork.Description)
		if activeWork.TicketID != "" {
			fmt.Printf("Ticket: %s\n", activeWork.TicketID)
		}
		fmt.Printf("Started: %s\n", activeWork.StartTime.Format("15:04"))

		// Ask what to do with active work
		prompt := promptui.Select{
			Label: "What would you like to do?",
			Items: []string{
				"Complete current work and start new",
				"Pause current work and start new",
				"Cancel new work",
			},
		}

		index, _, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				fmt.Println("\nOperation cancelled by user.")
				return
			}
			fmt.Printf("Failed to get user selection: %v\n", err)
			return
		}

		switch index {
		case 0: // Complete current work
			activeWork.EndTime = time.Now()
			activeWork.Status = "completed"
			if err := saveTrackedWork(*activeWork); err != nil {
				fmt.Printf("Failed to complete work: %v\n", err)
				return
			}
		case 1: // Pause current work
			activeWork.Status = "paused"
			if err := saveTrackedWork(*activeWork); err != nil {
				fmt.Printf("Failed to pause work: %v\n", err)
				return
			}
		case 2: // Cancel new work
			return
		}
	}

	// Get description from args or prompt
	var description string
	if len(args) > 0 {
		description = strings.Join(args, " ")
	} else {
		prompt := promptui.Prompt{
			Label: "What are you working on?",
			Validate: func(input string) error {
				if strings.TrimSpace(input) == "" {
					return fmt.Errorf("description cannot be empty")
				}
				return nil
			},
		}
		result, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				fmt.Println("\nOperation cancelled by user.")
				return
			}
			fmt.Printf("Failed to get work description: %v\n", err)
			return
		}
		description = result
	}

	// Try to infer ticket ID from current branch if git integration is enabled
	var ticketID string
	if cfg.GitIntegration {
		currentDir, err := os.Getwd()
		if err != nil {
			fmt.Printf("Failed to get current directory: %v\n", err)
			return
		}
		if isGitRepo(currentDir) {
			branch, err := getCurrentBranch()
			if err != nil {
				fmt.Printf("Failed to get current branch: %v\n", err)
				return
			}
			ticketID = extractTicketID(branch, cfg.TicketPrefixes)
		}
	}

	// If no ticket ID found, ask for one
	if ticketID == "" && len(cfg.TicketPrefixes) > 0 {
		prompt := promptui.Prompt{
			Label:    "Ticket ID (optional)",
			Validate: validateTicketID,
		}
		result, err := prompt.Run()
		if err != nil {
			if err != promptui.ErrInterrupt && err != promptui.ErrAbort {
				fmt.Printf("Failed to get ticket ID: %v\n", err)
				return
			}
			// If user cancelled, just continue without a ticket ID
		} else {
			ticketID = result
		}
	}

	// Ask for tags
	var tags []string
	for {
		prompt := promptui.Prompt{
			Label: "Add a tag (leave empty to finish)",
		}
		result, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				break
			}
			fmt.Println("Error getting tag:", err)
			return
		}
		if result == "" {
			break
		}
		tags = append(tags, strings.TrimSpace(result))
	}

	// Get context if in git repo and git integration is enabled
	var context WorkContext
	if cfg.GitIntegration {
		currentDir, err := os.Getwd()
		if err == nil && isGitRepo(currentDir) {
			// Get current branch
			if branch, err := getCurrentBranch(); err == nil {
				context.Branch = branch
			}

			// Get most recent commit and changed files
			if commits, err := getRecentCommits(1); err == nil && len(commits) > 0 {
				context.CommitHash = commits[0].Hash
				if files, err := getFilesChanged(currentDir, commits[0].Hash); err == nil {
					context.Files = files
				}
			}
		}
	}

	// Create tracked work
	work := TrackedWork{
		ID:          generateID(),
		Description: description,
		TicketID:    ticketID,
		StartTime:   time.Now(),
		Tags:        tags,
		Status:      "active",
		Context:     context,
	}

	// Save tracked work
	err = saveTrackedWork(work)
	if err != nil {
		fmt.Println("Error saving tracked work:", err)
		return
	}

	fmt.Println("\nWork tracked successfully!")
	fmt.Printf("ID: %s\n", work.ID)
	fmt.Printf("Description: %s\n", work.Description)
	if work.TicketID != "" {
		fmt.Printf("Ticket ID: %s\n", work.TicketID)
	}
	if len(work.Tags) > 0 {
		fmt.Printf("Tags: %s\n", strings.Join(work.Tags, ", "))
	}
	if work.Context.Branch != "" {
		fmt.Printf("Branch: %s\n", work.Context.Branch)
	}
}

// getActiveWork returns the currently active work, if any
func getActiveWork() (*TrackedWork, error) {
	dbDir, err := getDBDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get database directory: %w", err)
	}

	activeFile := filepath.Join(dbDir, "active.json")
	if _, err := os.Stat(activeFile); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(activeFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read active work file: %w", err)
	}

	var work TrackedWork
	if err := json.Unmarshal(data, &work); err != nil {
		return nil, fmt.Errorf("failed to parse active work data: %w", err)
	}

	return &work, nil
}

// validateTicketID validates a ticket ID against the configured prefixes
func validateTicketID(input string) error {
	if input == "" {
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	for _, prefix := range cfg.TicketPrefixes {
		if strings.HasPrefix(input, prefix) {
			return nil
		}
	}

	return fmt.Errorf("ticket ID must start with one of: %s", strings.Join(cfg.TicketPrefixes, ", "))
}

// generateID generates a unique ID for tracked work
func generateID() string {
	return fmt.Sprintf("tw-%d", time.Now().UnixNano())
}

// saveTrackedWork saves a piece of tracked work to the database
func saveTrackedWork(work TrackedWork) error {
	dbDir, err := getDBDir()
	if err != nil {
		return fmt.Errorf("failed to get database directory: %w", err)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Save to active.json if work is active or paused
	if work.Status == "active" || work.Status == "paused" {
		activeFile := filepath.Join(dbDir, "active.json")
		data, err := json.MarshalIndent(work, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal work data: %w", err)
		}

		if err := os.WriteFile(activeFile, data, 0644); err != nil {
			return fmt.Errorf("failed to write active work file: %w", err)
		}
	}

	// Save to completed.json if work is completed
	if work.Status == "completed" {
		completedFile := filepath.Join(dbDir, "completed.json")
		var completed []TrackedWork

		// Read existing completed work
		if data, err := os.ReadFile(completedFile); err == nil {
			if err := json.Unmarshal(data, &completed); err != nil {
				return fmt.Errorf("failed to parse completed work data: %w", err)
			}
		}

		// Append new work
		completed = append(completed, work)

		// Write back to file
		data, err := json.MarshalIndent(completed, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal completed work data: %w", err)
		}

		if err := os.WriteFile(completedFile, data, 0644); err != nil {
			return fmt.Errorf("failed to write completed work file: %w", err)
		}

		// Remove from active.json if it exists
		activeFile := filepath.Join(dbDir, "active.json")
		if err := os.Remove(activeFile); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove active work file: %w", err)
		}
	}

	return nil
}

// getDBDir gets the directory for the tracked work database
func getDBDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	dbDir := filepath.Join(homeDir, ".plannet", "db")
	return dbDir, nil
}
