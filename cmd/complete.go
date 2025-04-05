package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// completeCmd represents the complete command
var completeCmd = &cobra.Command{
	Use:   "complete [id]",
	Short: "Mark tracked work as complete",
	Long: `Mark tracked work as complete.
This command allows you to mark a piece of tracked work as finished,
recording the end time.`,
	Run: func(cmd *cobra.Command, args []string) {
		runComplete(args)
	},
}

func init() {
	rootCmd.AddCommand(completeCmd)
}

func runComplete(args []string) {
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

	// Filter for incomplete work
	var incompleteWork []TrackedWork
	for _, work := range trackedWork {
		if work.EndTime.IsZero() {
			incompleteWork = append(incompleteWork, work)
		}
	}

	if len(incompleteWork) == 0 {
		fmt.Println("No incomplete work found.")
		return
	}

	// Get work ID from args or prompt
	var workID string
	if len(args) > 0 {
		workID = args[0]
	} else {
		// Create items for the prompt
		var items []string
		for _, work := range incompleteWork {
			items = append(items, fmt.Sprintf("%s: %s", work.ID, work.Description))
		}

		// Create the prompt
		prompt := promptui.Select{
			Label: "Select work to complete",
			Items: items,
		}

		// Run the prompt
		index, _, err := prompt.Run()
		if err != nil {
			fmt.Println("Error selecting work:", err)
			return
		}

		workID = incompleteWork[index].ID
	}

	// Find the work
	var work *TrackedWork
	for i := range incompleteWork {
		if incompleteWork[i].ID == workID {
			work = &incompleteWork[i]
			break
		}
	}

	if work == nil {
		fmt.Printf("Work with ID %s not found.\n", workID)
		return
	}

	// Mark work as complete
	work.EndTime = time.Now()

	// Save the work
	err = saveTrackedWork(*work)
	if err != nil {
		fmt.Println("Error saving work:", err)
		return
	}

	fmt.Println("Work marked as complete!")
	fmt.Printf("ID: %s\n", work.ID)
	fmt.Printf("Description: %s\n", work.Description)
	if work.TicketID != "" {
		fmt.Printf("Ticket ID: %s\n", work.TicketID)
	}
	fmt.Printf("Start time: %s\n", work.StartTime.Format("2006-01-02 15:04"))
	fmt.Printf("End time: %s\n", work.EndTime.Format("2006-01-02 15:04"))
	if len(work.Tags) > 0 {
		fmt.Printf("Tags: %s\n", strings.Join(work.Tags, ", "))
	}
} 