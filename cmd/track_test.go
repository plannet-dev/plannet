package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/plannet-ai/plannet/config"
)

func TestTrackedWork(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "plannet-track-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test config
	cfg := &config.Config{
		TicketPrefixes: []string{"TEST-", "DEV-"},
		GitIntegration: true,
	}

	// Mock the config.Load function
	originalLoad := config.Load
	defer func() { config.Load = originalLoad }()
	config.Load = func() (*config.Config, error) {
		return cfg, nil
	}

	// Mock the getDBDir function
	originalGetDBDir := getDBDir
	defer func() { getDBDir = originalGetDBDir }()
	getDBDir = func() (string, error) {
		return tempDir, nil
	}

	// Mock the promptui.Prompt.Run function
	originalPromptRun := promptui.Prompt{}.Run
	defer func() { promptui.Prompt{}.Run = originalPromptRun }()
	promptui.Prompt{}.Run = func() (string, error) {
		return "Test description", nil
	}

	// Mock the promptui.Select.Run function
	originalSelectRun := promptui.Select{}.Run
	defer func() { promptui.Select{}.Run = originalSelectRun }()
	promptui.Select{}.Run = func() (int, string, error) {
		return 0, "Yes", nil
	}

	// Mock git functions
	originalGetCurrentBranch := getCurrentBranch
	defer func() { getCurrentBranch = originalGetCurrentBranch }()
	getCurrentBranch = func() (string, error) {
		return "feature/TEST-123", nil
	}

	originalGetRecentCommits := getRecentCommits
	defer func() { getRecentCommits = originalGetRecentCommits }()
	getRecentCommits = func(count int) ([]Commit, error) {
		return []Commit{{
			Hash:    "abc123",
			Message: "Test commit",
			Time:    time.Now(),
		}}, nil
	}

	originalGetFilesChanged := getFilesChanged
	defer func() { getFilesChanged = originalGetFilesChanged }()
	getFilesChanged = func(hash string) ([]string, error) {
		return []string{"test.go"}, nil
	}

	// Call the function
	runTrack()

	// Check if the file was created
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}

	// Read the file
	filePath := filepath.Join(tempDir, files[0].Name())
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Parse the file
	var trackedWork TrackedWork
	err = json.Unmarshal(fileContent, &trackedWork)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	// Check the tracked work
	if trackedWork.Description != "Test description" {
		t.Errorf("Expected description to be Test description, got %s", trackedWork.Description)
	}
	if trackedWork.TicketID != "TEST-123" {
		t.Errorf("Expected ticket ID to be TEST-123, got %s", trackedWork.TicketID)
	}
	if trackedWork.StartTime.IsZero() {
		t.Error("Expected start time to be set")
	}
	if !trackedWork.EndTime.IsZero() {
		t.Error("Expected end time to be zero")
	}
	if trackedWork.Status != "active" {
		t.Errorf("Expected status to be active, got %s", trackedWork.Status)
	}
	if trackedWork.Context.Branch != "feature/TEST-123" {
		t.Errorf("Expected branch to be feature/TEST-123, got %s", trackedWork.Context.Branch)
	}
	if trackedWork.Context.CommitHash != "abc123" {
		t.Errorf("Expected commit hash to be abc123, got %s", trackedWork.Context.CommitHash)
	}
	if len(trackedWork.Context.Files) != 1 || trackedWork.Context.Files[0] != "test.go" {
		t.Errorf("Expected files to be [test.go], got %v", trackedWork.Context.Files)
	}
}

func TestGetTrackedWork(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "plannet-track-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test config
	cfg := &config.Config{
		TicketPrefixes: []string{"TEST-", "DEV-"},
	}

	// Mock the config.Load function
	originalLoad := config.Load
	defer func() { config.Load = originalLoad }()
	config.Load = func() (*config.Config, error) {
		return cfg, nil
	}

	// Mock the getDBDir function
	originalGetDBDir := getDBDir
	defer func() { getDBDir = originalGetDBDir }()
	getDBDir = func() (string, error) {
		return tempDir, nil
	}

	// Create a test tracked work
	trackedWork := TrackedWork{
		ID:          "test-id",
		Description: "Test description",
		TicketID:    "TEST-123",
		StartTime:   time.Now(),
		EndTime:     time.Time{},
		Tags:        []string{"test", "work"},
	}

	// Convert to JSON
	trackedWorkJSON, err := json.Marshal(trackedWork)
	if err != nil {
		t.Fatalf("Failed to marshal tracked work: %v", err)
	}

	// Write to file
	filePath := filepath.Join(tempDir, "test-id.json")
	err = os.WriteFile(filePath, trackedWorkJSON, 0644)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Call the function
	trackedWorks, err := getTrackedWork()
	if err != nil {
		t.Fatalf("Failed to get tracked work: %v", err)
	}

	// Check the result
	if len(trackedWorks) != 1 {
		t.Errorf("Expected 1 tracked work, got %d", len(trackedWorks))
	}
	if trackedWorks[0].ID != "test-id" {
		t.Errorf("Expected ID to be test-id, got %s", trackedWorks[0].ID)
	}
	if trackedWorks[0].Description != "Test description" {
		t.Errorf("Expected description to be Test description, got %s", trackedWorks[0].Description)
	}
	if trackedWorks[0].TicketID != "TEST-123" {
		t.Errorf("Expected ticket ID to be TEST-123, got %s", trackedWorks[0].TicketID)
	}
	if len(trackedWorks[0].Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(trackedWorks[0].Tags))
	}
}

func TestSaveTrackedWork(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "plannet-track-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Mock the getDBDir function
	originalGetDBDir := getDBDir
	defer func() { getDBDir = originalGetDBDir }()
	getDBDir = func() (string, error) {
		return tempDir, nil
	}

	// Create a test tracked work
	trackedWork := TrackedWork{
		ID:          "test-id",
		Description: "Test description",
		TicketID:    "TEST-123",
		StartTime:   time.Now(),
		EndTime:     time.Time{},
		Tags:        []string{"test", "work"},
	}

	// Call the function
	err = saveTrackedWork(&trackedWork)
	if err != nil {
		t.Fatalf("Failed to save tracked work: %v", err)
	}

	// Check if the file was created
	filePath := filepath.Join(tempDir, "test-id.json")
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Parse the file
	var savedTrackedWork TrackedWork
	err = json.Unmarshal(fileContent, &savedTrackedWork)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	// Check the tracked work
	if savedTrackedWork.ID != "test-id" {
		t.Errorf("Expected ID to be test-id, got %s", savedTrackedWork.ID)
	}
	if savedTrackedWork.Description != "Test description" {
		t.Errorf("Expected description to be Test description, got %s", savedTrackedWork.Description)
	}
	if savedTrackedWork.TicketID != "TEST-123" {
		t.Errorf("Expected ticket ID to be TEST-123, got %s", savedTrackedWork.TicketID)
	}
	if len(savedTrackedWork.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(savedTrackedWork.Tags))
	}
}

func TestActiveWork(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "plannet-track-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Mock the getDBDir function
	originalGetDBDir := getDBDir
	defer func() { getDBDir = originalGetDBDir }()
	getDBDir = func() (string, error) {
		return tempDir, nil
	}

	// Create a test tracked work
	trackedWork := TrackedWork{
		ID:          "test-id",
		Description: "Test description",
		TicketID:    "TEST-123",
		StartTime:   time.Now(),
		Status:      "active",
	}

	// Convert to JSON
	trackedWorkJSON, err := json.Marshal(trackedWork)
	if err != nil {
		t.Fatalf("Failed to marshal tracked work: %v", err)
	}

	// Write to file
	filePath := filepath.Join(tempDir, "test-id.json")
	err = os.WriteFile(filePath, trackedWorkJSON, 0644)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Call the function
	activeWork, err := getActiveWork()
	if err != nil {
		t.Fatalf("Failed to get active work: %v", err)
	}

	// Check the result
	if activeWork == nil {
		t.Error("Expected active work to be found")
	}
	if activeWork.ID != "test-id" {
		t.Errorf("Expected ID to be test-id, got %s", activeWork.ID)
	}
	if activeWork.Description != "Test description" {
		t.Errorf("Expected description to be Test description, got %s", activeWork.Description)
	}
	if activeWork.TicketID != "TEST-123" {
		t.Errorf("Expected ticket ID to be TEST-123, got %s", activeWork.TicketID)
	}
	if activeWork.Status != "active" {
		t.Errorf("Expected status to be active, got %s", activeWork.Status)
	}
} 