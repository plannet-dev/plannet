package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export [format] [output]",
	Short: "Export tracked work",
	Long: `Export tracked work to various formats.
This command allows you to export your tracked work to CSV, JSON, or other formats
for use in other tools or for reporting.`,
	Run: func(cmd *cobra.Command, args []string) {
		runExport(args)
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
}

func runExport(args []string) {
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

	if len(trackedWork) == 0 {
		fmt.Println("No tracked work found.")
		return
	}

	// Get format from args or default to CSV
	format := "csv"
	if len(args) > 0 {
		format = args[0]
	}

	// Get output path from args or default to stdout
	outputPath := ""
	if len(args) > 1 {
		outputPath = args[1]
	}

	// Export based on format
	switch format {
	case "csv":
		err = exportCSV(trackedWork, outputPath)
	case "json":
		err = exportJSON(trackedWork, outputPath)
	default:
		fmt.Printf("Unsupported format: %s\n", format)
		fmt.Println("Supported formats: csv, json")
		return
	}

	if err != nil {
		fmt.Println("Error exporting work:", err)
		return
	}

	fmt.Println("Export completed successfully!")
}

// exportCSV exports tracked work to CSV
func exportCSV(work []TrackedWork, outputPath string) error {
	// Create a new CSV writer
	var writer *csv.Writer
	var file *os.File
	var err error

	if outputPath == "" {
		writer = csv.NewWriter(os.Stdout)
	} else {
		file, err = os.Create(outputPath)
		if err != nil {
			return err
		}
		defer file.Close()
		writer = csv.NewWriter(file)
	}
	defer writer.Flush()

	// Write header
	err = writer.Write([]string{
		"ID",
		"Description",
		"Ticket ID",
		"Start Time",
		"End Time",
		"Tags",
	})
	if err != nil {
		return err
	}

	// Write data
	for _, w := range work {
		// Format times
		startTime := w.StartTime.Format("2006-01-02 15:04:05")
		var endTime string
		if w.EndTime.IsZero() {
			endTime = ""
		} else {
			endTime = w.EndTime.Format("2006-01-02 15:04:05")
		}

		// Write row
		err = writer.Write([]string{
			w.ID,
			w.Description,
			w.TicketID,
			startTime,
			endTime,
			strings.Join(w.Tags, ";"),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// exportJSON exports tracked work to JSON
func exportJSON(work []TrackedWork, outputPath string) error {
	// Convert to JSON
	data, err := json.MarshalIndent(work, "", "  ")
	if err != nil {
		return err
	}

	// Write to file or stdout
	if outputPath == "" {
		fmt.Println(string(data))
	} else {
		err = os.WriteFile(outputPath, data, 0644)
		if err != nil {
			return err
		}
	}

	return nil
} 