package output

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/plannet-ai/plannet/config"
)

// Manager handles displaying and copying output
type Manager struct {
	useColors   bool
	config      *config.Config
	sessionCopy *bool      // Stores session preference for AskOnce
	sessionLock sync.Mutex // Ensures safe concurrent access
}

// NewManager creates a new OutputManager instance
func NewManager(useColors bool, cfg *config.Config) *Manager {
	return &Manager{
		useColors: useColors,
		config:    cfg,
	}
}

// HandleOutput manages the display and optional copying of generated output
func (m *Manager) HandleOutput(output string) error {
	// Display the output
	if err := m.displayOutput(output); err != nil {
		return fmt.Errorf("failed to display output: %w", err)
	}

	if shouldCopy := m.shouldCopyBasedOnPreference(); shouldCopy {
		if err := m.copyToClipboard(output); err != nil {
			return fmt.Errorf("failed to copy to clipboard: %w", err)
		}
		m.showCopyConfirmation()
	}
	return nil
}

// shouldCopyBasedOnPreference determines whether to copy based on CopyPreference
func (m *Manager) shouldCopyBasedOnPreference() bool {
	// Get the string value of the preference
	prefValue := m.config.CopyPreference.String()

	// Parse the string preference into a CopyPreference type
	pref, err := config.ParseCopyPreference(prefValue)
	if err != nil {
		// If parsing fails, default to asking every time
		return m.promptForCopy()
	}

	switch pref {
	case config.AskOnce:
		return m.shouldCopyForSession()
	case config.CopyAutomatically:
		return true
	case config.DoNotCopy:
		return false
	case config.AskEveryTime:
		fallthrough
	default:
		return m.promptForCopy() // Fallback to asking every time
	}
}

// shouldCopyForSession ensures "AskOnce" prompts only once per execution
func (m *Manager) shouldCopyForSession() bool {
	m.sessionLock.Lock()
	defer m.sessionLock.Unlock()

	// If already set, return stored session choice
	if m.sessionCopy != nil {
		return *m.sessionCopy
	}

	// Otherwise, prompt the user and store the choice for the session
	shouldCopy := m.promptForCopy()
	m.sessionCopy = &shouldCopy
	return shouldCopy
}

// displayOutput shows the generated output with optional formatting
func (m *Manager) displayOutput(output string) error {
	if m.useColors {
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
func (m *Manager) promptForCopy() bool {
	var response string
	prompt := "Copy to clipboard? [y/n]: "

	if m.useColors {
		prompt = color.YellowString(prompt)
	}

	fmt.Print(prompt)
	fmt.Scanln(&response)

	return strings.ToLower(response) == "y"
}

// copyToClipboard attempts to copy text to clipboard using available system commands
func (m *Manager) copyToClipboard(text string) error {
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
func (m *Manager) showCopyConfirmation() {
	message := "\n✓ Copied to clipboard!"

	if m.useColors {
		// Show an animated confirmation
		color.Green(message)
	} else {
		fmt.Printf("\n%s\n", message)
	}
}

// HandleOutput is a convenience function for simple output handling
func HandleOutput(output string, cfg *config.Config) error {
	manager := NewManager(true, cfg) // Enable colors by default
	return manager.HandleOutput(output)
}
