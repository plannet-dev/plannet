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

// TrackedWork represents a piece of work tracked by the user
type TrackedWork struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	TicketID    string    `json:"ticket_id,omitempty"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	Status      string    `json:"status"` // "active", "paused", "completed"
	Context     struct {
		Branch     string   `json:"branch,omitempty"`
		Files      []string `json:"files,omitempty"`
		CommitHash string   `json:"commit_hash,omitempty"`
	} `json:"context,omitempty"`
}

// Context represents the git context of tracked work
type Context struct {
	Branch     string   `json:"branch,omitempty"`
	Files      []string `json:"files,omitempty"`
	CommitHash string   `json:"commit_hash,omitempty"`
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
		fmt.Println("Error loading configuration:", err)
		fmt.Println("Run 'plannet init' to set up your configuration.")
		return
	}

	// Check for active work
	activeWork, err := getActiveWork()
	if err != nil {
		fmt.Println("Error checking active work:", err)
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
			fmt.Println("Error:", err)
			return
		}

		switch index {
		case 0: // Complete current work
			activeWork.EndTime = time.Now()
			activeWork.Status = "completed"
			err = saveTrackedWork(*activeWork)
			if err != nil {
				fmt.Println("Error completing work:", err)
				return
			}
		case 1: // Pause current work
			activeWork.Status = "paused"
			err = saveTrackedWork(*activeWork)
			if err != nil {
				fmt.Println("Error pausing work:", err)
				return
			}
		case 2: // Cancel new work
			return
		}
	}

	// Get description from args or prompt
	var description string
	if len(args) > 0 {
		description = args[0]
	} else {
		prompt := promptui.Prompt{
			Label: "What are you working on?",
		}
		result, err := prompt.Run()
		if err != nil {
			fmt.Println("Error getting description:", err)
			return
		}
		description = result
	}

	// Try to infer ticket ID from current branch
	var ticketID string
	if cfg.GitIntegration {
		branch, err := getCurrentBranch()
		if err == nil {
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
		if err != nil && err != promptui.ErrInterrupt {
			fmt.Println("Error getting ticket ID:", err)
			return
		}
		ticketID = result
	}

	// Ask for tags
	var tags []string
	for {
		prompt := promptui.Prompt{
			Label: "Add a tag (leave empty to finish)",
		}
		result, err := prompt.Run()
		if err != nil {
			fmt.Println("Error getting tag:", err)
			return
		}
		if result == "" {
			break
		}
		tags = append(tags, result)
	}

	// Get context if in git repo
	var context Context
	if cfg.GitIntegration {
		if branch, err := getCurrentBranch(); err == nil {
			context.Branch = branch
		}
		if commits, err := getRecentCommits(1); err == nil && len(commits) > 0 {
			context.CommitHash = commits[0].Hash
			if files, err := getFilesChanged(commits[0].Hash); err == nil {
				context.Files = files
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

	fmt.Println("Work tracked successfully!")
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
	trackedWork, err := getTrackedWork()
	if err != nil {
		return nil, err
	}

	for _, work := range trackedWork {
		if work.Status == "active" {
			return &work, nil
		}
	}

	return nil, nil
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
	// Get the database directory
	dbDir, err := getDBDir()
	if err != nil {
		return err
	}

	// Create the database directory if it doesn't exist
	err = os.MkdirAll(dbDir, 0755)
	if err != nil {
		return err
	}

	// Create the file path
	filePath := filepath.Join(dbDir, fmt.Sprintf("%s.json", work.ID))

	// Convert to JSON
	data, err := json.MarshalIndent(work, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// getDBDir gets the directory for the tracked work database
func getDBDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".plannet", "db"), nil
} 