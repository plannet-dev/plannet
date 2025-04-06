package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/plannet-ai/plannet/config"
)

// setupTest creates a temporary test environment and returns a cleanup function
func setupTest(t *testing.T) (string, func()) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "plannet-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create the database directory
	dbDir := filepath.Join(tempDir, ".plannet", "db")
	err = os.MkdirAll(dbDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create db dir: %v", err)
	}

	// Create a test config
	testConfig := &config.Config{
		GitIntegration: true,
		TicketPrefixes: []string{"JIRA-", "DEV-"},
	}

	// Override the config path for testing
	originalConfigPath := config.GetConfigPath()
	configPath := filepath.Join(tempDir, ".plannetrc")
	config.SetConfigPath(configPath)

	// Save the test config
	err = config.Save(testConfig)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Set up environment for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)

	// Return the temp directory and cleanup function
	return tempDir, func() {
		os.Setenv("HOME", originalHome)
		config.SetConfigPath(originalConfigPath)
		os.RemoveAll(tempDir)
	}
}

func TestSaveAndGetTrackedWork(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()

	// Create a test work item
	work := TrackedWork{
		ID:          "test-1",
		Description: "Test work",
		StartTime:   time.Now(),
		Status:      "active",
	}

	// Save the work
	err := saveTrackedWork(work)
	if err != nil {
		t.Fatalf("Failed to save tracked work: %v", err)
	}

	// Verify work was saved
	savedWork, err := getTrackedWork()
	if err != nil {
		t.Fatalf("Failed to get tracked work: %v", err)
	}

	if len(savedWork) != 1 {
		t.Errorf("Expected 1 tracked work item, got %d", len(savedWork))
	}

	if savedWork[0].ID != work.ID {
		t.Errorf("Expected ID %s, got %s", work.ID, savedWork[0].ID)
	}

	if savedWork[0].Description != work.Description {
		t.Errorf("Expected description %s, got %s", work.Description, savedWork[0].Description)
	}
}

func TestGetActiveWork(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()

	// Create test work items
	work1 := TrackedWork{
		ID:          "test-1",
		Description: "Active work",
		StartTime:   time.Now(),
		Status:      "active",
	}

	work2 := TrackedWork{
		ID:          "test-2",
		Description: "Completed work",
		StartTime:   time.Now().Add(-1 * time.Hour),
		EndTime:     time.Now(),
		Status:      "completed",
	}

	// Save the work items
	err := saveTrackedWork(work1)
	if err != nil {
		t.Fatalf("Failed to save work1: %v", err)
	}

	err = saveTrackedWork(work2)
	if err != nil {
		t.Fatalf("Failed to save work2: %v", err)
	}

	// Test getting active work
	activeWork, err := getActiveWork()
	if err != nil {
		t.Fatalf("Failed to get active work: %v", err)
	}

	if activeWork == nil {
		t.Fatal("Expected active work, got nil")
	}

	if activeWork.ID != work1.ID {
		t.Errorf("Expected ID %s, got %s", work1.ID, activeWork.ID)
	}

	if activeWork.Description != work1.Description {
		t.Errorf("Expected description %s, got %s", work1.Description, activeWork.Description)
	}
}

func TestValidateTicketID(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "Valid JIRA ticket",
			input:   "JIRA-123",
			wantErr: false,
		},
		{
			name:    "Valid DEV ticket",
			input:   "DEV-456",
			wantErr: false,
		},
		{
			name:    "Invalid ticket",
			input:   "INVALID-789",
			wantErr: true,
		},
		{
			name:    "Empty ticket",
			input:   "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTicketID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTicketID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
