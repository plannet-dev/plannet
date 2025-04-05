package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tracked work",
	Long: `List tracked work, showing both git-based and manually tracked work.
This command gives you a comprehensive view of your work history.`,
	Run: func(cmd *cobra.Command, args []string) {
		runList()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Error loading configuration:", err)
		fmt.Println("Run 'plannet init' to set up your configuration.")
		return
	}

	// Get tracked work
	trackedWork, err := getTrackedWork()
	if err != nil {
		fmt.Println("Error getting tracked work:", err)
		return
	}

	// Sort tracked work by start time (newest first)
	sort.Slice(trackedWork, func(i, j int) bool {
		return trackedWork[i].StartTime.After(trackedWork[j].StartTime)
	})

	// Display tracked work
	if len(trackedWork) == 0 {
		fmt.Println("No tracked work found.")
		return
	}

	fmt.Println("Tracked work:")
	for _, work := range trackedWork {
		// Format time
		startTime := work.StartTime.Format("2006-01-02 15:04")
		var timeStr string
		if work.EndTime.IsZero() {
			timeStr = fmt.Sprintf("%s (ongoing)", startTime)
		} else {
			endTime := work.EndTime.Format("2006-01-02 15:04")
			timeStr = fmt.Sprintf("%s – %s", startTime, endTime)
		}

		// Display work
		fmt.Printf("\n%s\n", timeStr)
		fmt.Printf("  %s\n", work.Description)
		if work.TicketID != "" {
			fmt.Printf("  Ticket: %s\n", work.TicketID)
		}
		if len(work.Tags) > 0 {
			fmt.Printf("  Tags: %s\n", strings.Join(work.Tags, ", "))
		}
	}
}

// getTrackedWork gets all tracked work from the database
func getTrackedWork() ([]TrackedWork, error) {
	// Get the database directory
	dbDir, err := getDBDir()
	if err != nil {
		return nil, err
	}

	// Check if the database directory exists
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		return []TrackedWork{}, nil
	}

	// Read all files in the database directory
	files, err := os.ReadDir(dbDir)
	if err != nil {
		return nil, err
	}

	// Read each file
	var trackedWork []TrackedWork
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		// Read the file
		filePath := filepath.Join(dbDir, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", filePath, err)
			continue
		}

		// Parse the JSON
		var work TrackedWork
		err = json.Unmarshal(data, &work)
		if err != nil {
			fmt.Printf("Error parsing file %s: %v\n", filePath, err)
			continue
		}

		trackedWork = append(trackedWork, work)
	}

	return trackedWork, nil
} 