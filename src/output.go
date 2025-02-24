package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/fatih/color"
)

// OutputManager handles displaying and copying output
type OutputManager struct {
	useColors bool
	config    *Config
}

// NewOutputManager creates a new OutputManager instance
func NewOutputManager(useColors bool) *OutputManager {
	return &OutputManager{
		useColors: useColors,
	}
}

// HandleOutput manages the display and optional copying of generated output
func (outputManager *OutputManager) HandleOutput(output string) error {
	// Display the output
	if err := outputManager.displayOutput(output); err != nil {
		return fmt.Errorf("failed to display output: %w", err)
	}

	// get alwaysCopy from config and return early
	// if outputManager.config.alwaysCopy {
	// 	if err := outputManager.copyToClipboard(output); err != nil {
	// 		return fmt.Errorf("failed to copy to clipboard: %w", err)
	// 	}
	// 	outputManager.showCopyConfirmation()
	// 	return nil
	// }

	// Ask about clipboard copy
	if shouldCopy := outputManager.promptForCopy(); shouldCopy {
		if err := outputManager.copyToClipboard(output); err != nil {
			return fmt.Errorf("failed to copy to clipboard: %w", err)
		}
		outputManager.showCopyConfirmation()
	}

	return nil
}

// displayOutput shows the generated output with optional formatting
func (outputManager *OutputManager) displayOutput(output string) error {
	if outputManager.useColors {
		// Add a separator line before output
		color.Blue("\n" + strings.Repeat("-", 80) + "\n")

		// Display output with subtle highlighting
		fmt.Println(color.CyanString(output))

		// Add a separator line after output
		color.Blue("\n" + strings.Repeat("-", 80) + "\n")
	} else {
		fmt.Printf("\n%s\n\n", output)
	}

	return nil
}

// promptForCopy asks the user if they want to copy the output
func (outputManager *OutputManager) promptForCopy() bool {
	var response string
	prompt := "Copy to clipboard? [y/n]: "

	if outputManager.useColors {
		prompt = color.YellowString(prompt)
	}

	fmt.Print(prompt)
	fmt.Scanln(&response)

	return strings.ToLower(response) == "y"
}

// copyToClipboard attempts to copy text to clipboard using available system commands
func (outputManager *OutputManager) copyToClipboard(text string) error {
	// Try different clipboard commands based on OS
	commands := []struct {
		name string
		args []string
	}{
		{"pbcopy", nil}, // macOS
		{"xclip", []string{"-selection", "clipboard"}}, // Linux
		{"xsel", []string{"--clipboard", "--input"}},   // Linux alternative
		{"clip", nil}, // Windows
	}

	for _, cmd := range commands {
		command := exec.Command(cmd.name, cmd.args...)
		command.Stdin = strings.NewReader(text)

		if err := command.Run(); err == nil {
			return nil
		}
	}

	return fmt.Errorf("no clipboard command available")
}

// showCopyConfirmation displays a confirmation message about successful copying
func (outputManager *OutputManager) showCopyConfirmation() {
	message := "Copied to clipboard!"

	if outputManager.useColors {
		// Show an animated confirmation
		color.Green("\n✓ ")
		time.Sleep(100 * time.Millisecond)
		color.Green(message)
	} else {
		fmt.Printf("\n%s\n", message)
	}
}

// Convenience function for simple output handling
func handleOutput(output string) error {
	manager := NewOutputManager(true) // Enable colors by default
	return manager.HandleOutput(output)
}
