package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setupGitRepo(t *testing.T) (string, func()) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "plannet-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Initialize git repo
	if err := exec.Command("git", "init").Run(); err != nil {
		os.Chdir(originalDir)
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	// Configure git user for commits
	if err := exec.Command("git", "config", "user.name", "Test User").Run(); err != nil {
		os.Chdir(originalDir)
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to configure git user name: %v", err)
	}
	if err := exec.Command("git", "config", "user.email", "test@example.com").Run(); err != nil {
		os.Chdir(originalDir)
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to configure git user email: %v", err)
	}

	cleanup := func() {
		os.Chdir(originalDir)
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func TestGitIntegration(t *testing.T) {
	tempDir, cleanup := setupGitRepo(t)
	defer cleanup()

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Add and commit the file
	if err := exec.Command("git", "add", "test.txt").Run(); err != nil {
		t.Fatalf("Failed to add file to git: %v", err)
	}

	if err := exec.Command("git", "commit", "-m", "TEST-123: Initial commit").Run(); err != nil {
		t.Fatalf("Failed to commit file: %v", err)
	}

	// Test getCurrentBranch
	branch, err := getCurrentBranch()
	if err != nil {
		t.Fatalf("Failed to get current branch: %v", err)
	}
	if branch != "main" && branch != "master" {
		t.Errorf("Expected branch to be 'main' or 'master', got '%s'", branch)
	}

	// Test getRecentCommits
	commits, err := getRecentCommits(5)
	if err != nil {
		t.Fatalf("Failed to get recent commits: %v", err)
	}
	if len(commits) == 0 {
		t.Error("Expected at least one commit")
	}

	// Test extractTicketID
	ticketID := extractTicketID("feature/TEST-123-add-feature", []string{"TEST-"})
	if ticketID != "TEST-123" {
		t.Errorf("Expected ticket ID to be 'TEST-123', got '%s'", ticketID)
	}

	// Test findSideQuests
	// Create another commit without a ticket ID
	if err := exec.Command("git", "add", "test.txt").Run(); err != nil {
		t.Fatalf("Failed to add file to git: %v", err)
	}

	if err := exec.Command("git", "commit", "-m", "Update without ticket ID").Run(); err != nil {
		t.Fatalf("Failed to commit file: %v", err)
	}

	commits, err = getRecentCommits(5)
	if err != nil {
		t.Fatalf("Failed to get recent commits: %v", err)
	}

	sideQuests := findSideQuests(commits, []string{"TEST-"})
	if len(sideQuests) == 0 {
		t.Error("Expected at least one side quest")
	}
}

func TestIsGitRepo(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test with non-git directory
	if isGitRepo(tempDir) {
		t.Error("Expected non-git directory to return false")
	}

	// Initialize git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	// Test with git directory
	if !isGitRepo(tempDir) {
		t.Error("Expected git directory to return true")
	}
}

func TestGetCurrentBranch(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	// Create a file and commit it
	filePath := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Test getting current branch
	branch, err := getCurrentBranch()
	if err != nil {
		t.Fatalf("Failed to get current branch: %v", err)
	}

	if branch != "main" && branch != "master" {
		t.Errorf("Expected branch to be 'main' or 'master', got '%s'", branch)
	}
}

func TestExtractTicketID(t *testing.T) {
	tests := []struct {
		name     string
		branch   string
		prefixes []string
		want     string
	}{
		{
			name:     "JIRA ticket in branch name",
			branch:   "feature/JIRA-123",
			prefixes: []string{"JIRA-"},
			want:     "JIRA-123",
		},
		{
			name:     "Multiple prefixes",
			branch:   "bugfix/DEV-456",
			prefixes: []string{"JIRA-", "DEV-"},
			want:     "DEV-456",
		},
		{
			name:     "No ticket ID",
			branch:   "feature/new-feature",
			prefixes: []string{"JIRA-"},
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTicketID(tt.branch, tt.prefixes)
			if got != tt.want {
				t.Errorf("extractTicketID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRecentCommits(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	// Create and commit some files
	for i := 0; i < 3; i++ {
		filePath := filepath.Join(tempDir, fmt.Sprintf("test%d.txt", i))
		if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		cmd = exec.Command("git", "add", fmt.Sprintf("test%d.txt", i))
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add file: %v", err)
		}

		cmd = exec.Command("git", "commit", "-m", fmt.Sprintf("Commit %d", i))
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to commit: %v", err)
		}

		// Sleep to ensure different timestamps
		time.Sleep(time.Second)
	}

	// Test getting recent commits
	commits, err := getRecentCommits(2)
	if err != nil {
		t.Fatalf("Failed to get recent commits: %v", err)
	}

	if len(commits) != 2 {
		t.Errorf("Expected 2 commits, got %d", len(commits))
	}

	// Check commit messages
	expectedMessages := []string{"Commit 2", "Commit 1"}
	for i, commit := range commits {
		if commit.Message != expectedMessages[i] {
			t.Errorf("Expected message '%s', got '%s'", expectedMessages[i], commit.Message)
		}
	}
}

func TestGetCommitsSince(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	// Create and commit a file
	filePath := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Test getting commits since midnight
	commits, err := getCommitsSince(tempDir, "midnight")
	if err != nil {
		t.Fatalf("Failed to get commits since midnight: %v", err)
	}

	if len(commits) != 1 {
		t.Errorf("Expected 1 commit, got %d", len(commits))
	}

	if commits[0].Message != "Initial commit" {
		t.Errorf("Expected message 'Initial commit', got '%s'", commits[0].Message)
	}
}

func TestGetFilesChanged(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	// Create and commit a file
	filePath := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Get the commit hash
	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = tempDir
	hashBytes, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get commit hash: %v", err)
	}
	hash := strings.TrimSpace(string(hashBytes))

	// Test getting files changed
	files, err := getFilesChanged(tempDir, hash)
	if err != nil {
		t.Fatalf("Failed to get files changed: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}

	if files[0] != "test.txt" {
		t.Errorf("Expected file 'test.txt', got '%s'", files[0])
	}
}
