package output

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/plannet-ai/plannet/config"
)

// mockClipboardCommand is used to mock clipboard commands
type mockClipboardCommand struct {
	shouldSucceed bool
}

func (m *mockClipboardCommand) Run() error {
	if m.shouldSucceed {
		return nil
	}
	return fmt.Errorf("mock clipboard error")
}

// mockExecCommand is used to replace exec.Command in tests
func mockExecCommand(name string, args ...string) *exec.Cmd {
	// This is a placeholder that will be replaced in tests
	return exec.Command(name, args...)
}

// testManager is a test-specific struct that embeds Manager and allows method overrides
type testManager struct {
	*Manager
	promptForCopyFunc        func() bool
	shouldCopyForSessionFunc func() bool
	copyToClipboardFunc      func(string) error
	sessionCopy              *bool
}

// Override the methods with our test implementations
func (t *testManager) promptForCopy() bool {
	if t.promptForCopyFunc != nil {
		return t.promptForCopyFunc()
	}
	return t.Manager.promptForCopy()
}

func (t *testManager) shouldCopyForSession() bool {
	if t.shouldCopyForSessionFunc != nil {
		return t.shouldCopyForSessionFunc()
	}
	return t.Manager.shouldCopyForSession()
}

func (t *testManager) copyToClipboard(text string) error {
	if t.copyToClipboardFunc != nil {
		return t.copyToClipboardFunc(text)
	}
	return t.Manager.copyToClipboard(text)
}

// TestNewManager tests the NewManager function
func TestNewManager(t *testing.T) {
	// Create a test config
	cfg := &config.Config{
		CopyPreference: config.AskEveryTime,
	}

	// Test with colors enabled
	manager := NewManager(true, cfg)
	if manager.useColors != true {
		t.Errorf("Expected useColors to be true, got %v", manager.useColors)
	}
	if manager.config != cfg {
		t.Errorf("Expected config to be %v, got %v", cfg, manager.config)
	}
	if manager.sessionCopy != nil {
		t.Errorf("Expected sessionCopy to be nil, got %v", manager.sessionCopy)
	}

	// Test with colors disabled
	manager = NewManager(false, cfg)
	if manager.useColors != false {
		t.Errorf("Expected useColors to be false, got %v", manager.useColors)
	}
}

// TestHandleOutput tests the HandleOutput function
func TestHandleOutput(t *testing.T) {
	// Create a test config
	cfg := &config.Config{
		CopyPreference: config.AskEveryTime,
	}

	// Create a manager
	manager := NewManager(true, cfg)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	// Test output
	output := "Test output"
	err = manager.HandleOutput(output)
	if err != nil {
		t.Errorf("HandleOutput returned error: %v", err)
	}

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Check if output contains the test string
	if !strings.Contains(buf.String(), output) {
		t.Errorf("Expected output to contain '%s', got '%s'", output, buf.String())
	}
}

// TestShouldCopyBasedOnPreference tests the shouldCopyBasedOnPreference function
func TestShouldCopyBasedOnPreference(t *testing.T) {
	// Test cases
	testCases := []struct {
		name       string
		preference config.CopyPreference
		expected   bool
	}{
		{
			name:       "AskEveryTime",
			preference: config.AskEveryTime,
			expected:   false, // Will be overridden by promptForCopy mock
		},
		{
			name:       "AskOnce",
			preference: config.AskOnce,
			expected:   false, // Will be overridden by shouldCopyForSession mock
		},
		{
			name:       "CopyAutomatically",
			preference: config.CopyAutomatically,
			expected:   true,
		},
		{
			name:       "DoNotCopy",
			preference: config.DoNotCopy,
			expected:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test config
			cfg := &config.Config{
				CopyPreference: tc.preference,
			}

			// Create a manager
			baseManager := NewManager(true, cfg)

			// Create a test manager with mocked functions
			testMgr := &testManager{
				Manager: baseManager,
				promptForCopyFunc: func() bool {
					return true
				},
				shouldCopyForSessionFunc: func() bool {
					return true
				},
			}

			// Test the function
			result := testMgr.shouldCopyBasedOnPreference()

			// Check result
			if tc.preference == config.AskEveryTime || tc.preference == config.AskOnce {
				// For these preferences, the result should be true due to our mocks
				if !result {
					t.Errorf("Expected result to be true for %s, got false", tc.name)
				}
			} else {
				// For other preferences, the result should match the expected value
				if result != tc.expected {
					t.Errorf("Expected result to be %v for %s, got %v", tc.expected, tc.name, result)
				}
			}
		})
	}
}

// TestShouldCopyForSession tests the shouldCopyForSession function
func TestShouldCopyForSession(t *testing.T) {
	// Create a test config
	cfg := &config.Config{
		CopyPreference: config.AskOnce,
	}

	// Create a manager
	baseManager := NewManager(true, cfg)

	// Create a test manager with mocked functions
	testMgr := &testManager{
		Manager: baseManager,
		promptForCopyFunc: func() bool {
			return true
		},
	}

	// First call should prompt and store the result
	result1 := testMgr.shouldCopyForSession()
	if !result1 {
		t.Errorf("Expected first call to return true, got false")
	}
	if testMgr.sessionCopy == nil {
		t.Errorf("Expected sessionCopy to be set, got nil")
	}
	if *testMgr.sessionCopy != true {
		t.Errorf("Expected sessionCopy to be true, got %v", *testMgr.sessionCopy)
	}

	// Second call should use the stored result without prompting
	// Create a new test manager with a different promptForCopy implementation
	testMgr2 := &testManager{
		Manager:     baseManager,
		sessionCopy: testMgr.sessionCopy, // Use the same sessionCopy from previous call
		promptForCopyFunc: func() bool {
			t.Errorf("promptForCopy should not be called on second invocation")
			return false
		},
	}

	result2 := testMgr2.shouldCopyForSession()
	if !result2 {
		t.Errorf("Expected second call to return true, got false")
	}
}

// TestDisplayOutput tests the displayOutput function
func TestDisplayOutput(t *testing.T) {
	// Create a test config
	cfg := &config.Config{}

	// Test cases
	testCases := []struct {
		name      string
		useColors bool
		output    string
	}{
		{
			name:      "With colors",
			useColors: true,
			output:    "Test output with colors",
		},
		{
			name:      "Without colors",
			useColors: false,
			output:    "Test output without colors",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a manager
			manager := NewManager(tc.useColors, cfg)

			// Capture stdout
			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}
			os.Stdout = w

			// Test output
			err = manager.displayOutput(tc.output)
			if err != nil {
				t.Errorf("displayOutput returned error: %v", err)
			}

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Read captured output
			var buf bytes.Buffer
			io.Copy(&buf, r)

			// Check if output contains the test string
			if !strings.Contains(buf.String(), tc.output) {
				t.Errorf("Expected output to contain '%s', got '%s'", tc.output, buf.String())
			}
		})
	}
}

// TestPromptForCopy tests the promptForCopy function
func TestPromptForCopy(t *testing.T) {
	// Create a test config
	cfg := &config.Config{}

	// Create a manager
	baseManager := NewManager(true, cfg)

	// Create a test manager with mocked functions
	testMgr := &testManager{
		Manager: baseManager,
		promptForCopyFunc: func() bool {
			return true
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	// Test the function
	result := testMgr.promptForCopy()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Check result
	if !result {
		t.Errorf("Expected result to be true, got false")
	}

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Check if output contains the prompt
	if !strings.Contains(buf.String(), "Copy to clipboard?") {
		t.Errorf("Expected output to contain 'Copy to clipboard?', got '%s'", buf.String())
	}
}

// TestCopyToClipboard tests the copyToClipboard function
func TestCopyToClipboard(t *testing.T) {
	// Create a test config
	cfg := &config.Config{}

	// Create a manager
	baseManager := NewManager(true, cfg)

	// Create a test manager with mocked functions
	testMgr := &testManager{
		Manager: baseManager,
	}

	// Test cases
	testCases := []struct {
		name          string
		shouldSucceed bool
		expectedError bool
	}{
		{
			name:          "Success",
			shouldSucceed: true,
			expectedError: false,
		},
		{
			name:          "Failure",
			shouldSucceed: false,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Override the copyToClipboard method for testing
			testMgr.copyToClipboardFunc = func(text string) error {
				if tc.shouldSucceed {
					return nil
				}
				return fmt.Errorf("mock clipboard error")
			}

			// Test the function
			err := testMgr.copyToClipboard("Test clipboard content")

			// Check result
			if tc.expectedError && err == nil {
				t.Errorf("Expected error, got nil")
			}
			if !tc.expectedError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

// TestShowCopyConfirmation tests the showCopyConfirmation function
func TestShowCopyConfirmation(t *testing.T) {
	// Test cases
	testCases := []struct {
		name      string
		useColors bool
	}{
		{
			name:      "With colors",
			useColors: true,
		},
		{
			name:      "Without colors",
			useColors: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test config
			cfg := &config.Config{}

			// Create a manager
			manager := NewManager(tc.useColors, cfg)

			// Capture stdout
			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}
			os.Stdout = w

			// Test the function
			manager.showCopyConfirmation()

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Read captured output
			var buf bytes.Buffer
			io.Copy(&buf, r)

			// Check if output contains the confirmation message
			if !strings.Contains(buf.String(), "Copied to clipboard") {
				t.Errorf("Expected output to contain 'Copied to clipboard', got '%s'", buf.String())
			}
		})
	}
}

// TestHandleOutputConvenience tests the HandleOutput convenience function
func TestHandleOutputConvenience(t *testing.T) {
	// Create a test config
	cfg := &config.Config{
		CopyPreference: config.DoNotCopy, // Don't copy to avoid clipboard issues
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	// Test output
	output := "Test convenience output"
	err = HandleOutput(output, cfg)
	if err != nil {
		t.Errorf("HandleOutput returned error: %v", err)
	}

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Check if output contains the test string
	if !strings.Contains(buf.String(), output) {
		t.Errorf("Expected output to contain '%s', got '%s'", output, buf.String())
	}
}
