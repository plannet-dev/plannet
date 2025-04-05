package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
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
	tempDir, cleanup := setupGitRepo(t)
	defer cleanup()

	// Test with git directory
	if !isGitRepo(tempDir) {
		t.Error("Expected git directory to return true")
	}

	// Create and test non-git directory
	nonGitDir := filepath.Join(tempDir, "non-git")
	if err := os.Mkdir(nonGitDir, 0755); err != nil {
		t.Fatalf("Failed to create non-git directory: %v", err)
	}

	if isGitRepo(nonGitDir) {
		t.Error("Expected non-git directory to return false")
	}
}

func TestGetCommitsSince(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "plannet-git-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize a git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repository: %v", err)
	}

	// Configure git user for commits
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to configure git user name: %v", err)
	}

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to configure git user email: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("Test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Add and commit the file
	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "TEST-123: Initial commit")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit file: %v", err)
	}

	// Get the commit hash
	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = tempDir
	hashBytes, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get commit hash: %v", err)
	}
	hash := string(hashBytes)
	hash = hash[:len(hash)-1] // Remove newline

	// Wait a bit
	time.Sleep(1 * time.Second)

	// Create another commit
	err = os.WriteFile(testFile, []byte("Updated content"), 0644)
	if err != nil {
		t.Fatalf("Failed to update test file: %v", err)
	}

	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "TEST-456: Second commit")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit file: %v", err)
	}

	// Test getCommitsSince
	commits, err := getCommitsSince(tempDir, hash)
	if err != nil {
		t.Fatalf("Failed to get commits since: %v", err)
	}
	if len(commits) != 1 {
		t.Errorf("Expected 1 commit, got %d", len(commits))
	}
	if commits[0].Message != "TEST-456: Second commit" {
		t.Errorf("Expected commit message 'TEST-456: Second commit', got '%s'", commits[0].Message)
	}
}

func TestGetFilesChanged(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "plannet-git-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize a git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repository: %v", err)
	}

	// Configure git user for commits
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to configure git user name: %v", err)
	}

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to configure git user email: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("Test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Add and commit the file
	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "TEST-123: Initial commit")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit file: %v", err)
	}

	// Get the commit hash
	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = tempDir
	hashBytes, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get commit hash: %v", err)
	}
	hash := string(hashBytes)
	hash = hash[:len(hash)-1] // Remove newline

	// Test getFilesChanged
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