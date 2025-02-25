package main

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/fatih/color"
)

// OutputManager handles displaying and copying output
type OutputManager struct {
	useColors   bool
	config      *Config
	sessionCopy *bool      // Stores session preference for AskOnce
	sessionLock sync.Mutex // Ensures safe concurrent access
}

// NewOutputManager creates a new OutputManager instance
func NewOutputManager(useColors bool, config *Config) *OutputManager {
	return &OutputManager{
		useColors: useColors,
		config:    config,
	}
}

// HandleOutput manages the display and optional copying of generated output
func (outputManager *OutputManager) HandleOutput(output string) error {
	// Display the output
	if err := outputManager.displayOutput(output); err != nil {
		return fmt.Errorf("failed to display output: %w", err)
	}

	if shouldCopy := outputManager.shouldCopyBasedOnPreference(); shouldCopy {
		if err := outputManager.copyToClipboard(output); err != nil {
			return fmt.Errorf("failed to copy to clipboard: %w", err)
		}
		outputManager.showCopyConfirmation()
	}
	return nil
}

// shouldCopyBasedOnPreference determines whether to copy based on CopyPreference
func (outputManager *OutputManager) shouldCopyBasedOnPreference() bool {
	switch outputManager.config.CopyPreference {
	case AskOnce:
		return outputManager.shouldCopyForSession()
	case CopyAutomatically:
		return true
	case DoNotCopy:
		return false
	case AskEveryTime:
		fallthrough
	default:
		return outputManager.promptForCopy() // Fallback to asking every time
	}
}

// shouldCopyForSession ensures "AskOnce" prompts only once per execution
func (outputManager *OutputManager) shouldCopyForSession() bool {
	outputManager.sessionLock.Lock()
	defer outputManager.sessionLock.Unlock()

	// If already set, return stored session choice
	if outputManager.sessionCopy != nil {
		return *outputManager.sessionCopy
	}

	// Otherwise, prompt the user and store the choice for the session
	shouldCopy := outputManager.shouldCopyBasedOnPreference()
	outputManager.sessionCopy = &shouldCopy
	return shouldCopy
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
	message := "\n✓ Copied to clipboard!"

	if outputManager.useColors {
		// Show an animated confirmation
		color.Green(message)
	} else {
		fmt.Printf("\n%s\n", message)
	}
}

// Convenience function for simple output handling
func handleOutput(output string, config *Config) error {
	manager := NewOutputManager(true, config) // Enable colors by default
	return manager.HandleOutput(output)
}
