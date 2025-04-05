package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestGitUtils(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "plannet-git-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize a git repository
	err = exec.Command("git", "init").SetDir(tempDir).Run()
	if err != nil {
		t.Fatalf("Failed to initialize git repository: %v", err)
	}

	// Test isGitRepo
	if !isGitRepo(tempDir) {
		t.Error("isGitRepo should return true for a git repository")
	}

	// Create a non-git directory
	nonGitDir := filepath.Join(tempDir, "non-git")
	err = os.Mkdir(nonGitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create non-git directory: %v", err)
	}

	if isGitRepo(nonGitDir) {
		t.Error("isGitRepo should return false for a non-git directory")
	}

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("Test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Add and commit the file
	err = exec.Command("git", "add", "test.txt").SetDir(tempDir).Run()
	if err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	err = exec.Command("git", "commit", "-m", "TEST-123: Initial commit").SetDir(tempDir).Run()
	if err != nil {
		t.Fatalf("Failed to commit file: %v", err)
	}

	// Test getCurrentBranch
	branch, err := getCurrentBranch(tempDir)
	if err != nil {
		t.Fatalf("Failed to get current branch: %v", err)
	}
	if branch != "main" && branch != "master" {
		t.Errorf("Expected branch to be 'main' or 'master', got '%s'", branch)
	}

	// Test extractTicketID
	ticketID := extractTicketID("TEST-123-feature-branch", []string{"TEST-"})
	if ticketID != "TEST-123" {
		t.Errorf("Expected ticket ID to be 'TEST-123', got '%s'", ticketID)
	}

	// Test with no ticket ID
	noTicketID := extractTicketID("feature-branch", []string{"TEST-"})
	if noTicketID != "" {
		t.Errorf("Expected empty ticket ID, got '%s'", noTicketID)
	}

	// Test getRecentCommits
	commits, err := getRecentCommits(tempDir, 10)
	if err != nil {
		t.Fatalf("Failed to get recent commits: %v", err)
	}
	if len(commits) != 1 {
		t.Errorf("Expected 1 commit, got %d", len(commits))
	}
	if commits[0].Message != "TEST-123: Initial commit" {
		t.Errorf("Expected commit message 'TEST-123: Initial commit', got '%s'", commits[0].Message)
	}

	// Test findSideQuests
	sideQuests := findSideQuests(commits, []string{"TEST-"})
	if len(sideQuests) != 0 {
		t.Errorf("Expected 0 side quests, got %d", len(sideQuests))
	}

	// Create a commit without a ticket ID
	err = os.WriteFile(testFile, []byte("Updated content"), 0644)
	if err != nil {
		t.Fatalf("Failed to update test file: %v", err)
	}

	err = exec.Command("git", "add", "test.txt").SetDir(tempDir).Run()
	if err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	err = exec.Command("git", "commit", "-m", "Update without ticket ID").SetDir(tempDir).Run()
	if err != nil {
		t.Fatalf("Failed to commit file: %v", err)
	}

	// Get commits again
	commits, err = getRecentCommits(tempDir, 10)
	if err != nil {
		t.Fatalf("Failed to get recent commits: %v", err)
	}
	if len(commits) != 2 {
		t.Errorf("Expected 2 commits, got %d", len(commits))
	}

	// Test findSideQuests with a commit without a ticket ID
	sideQuests = findSideQuests(commits, []string{"TEST-"})
	if len(sideQuests) != 1 {
		t.Errorf("Expected 1 side quest, got %d", len(sideQuests))
	}
	if sideQuests[0].Message != "Update without ticket ID" {
		t.Errorf("Expected side quest message 'Update without ticket ID', got '%s'", sideQuests[0].Message)
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
	err = exec.Command("git", "init").SetDir(tempDir).Run()
	if err != nil {
		t.Fatalf("Failed to initialize git repository: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("Test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Add and commit the file
	err = exec.Command("git", "add", "test.txt").SetDir(tempDir).Run()
	if err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	err = exec.Command("git", "commit", "-m", "TEST-123: Initial commit").SetDir(tempDir).Run()
	if err != nil {
		t.Fatalf("Failed to commit file: %v", err)
	}

	// Get the commit hash
	hashBytes, err := exec.Command("git", "rev-parse", "HEAD").SetDir(tempDir).Output()
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

	err = exec.Command("git", "add", "test.txt").SetDir(tempDir).Run()
	if err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	err = exec.Command("git", "commit", "-m", "TEST-456: Second commit").SetDir(tempDir).Run()
	if err != nil {
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
	err = exec.Command("git", "init").SetDir(tempDir).Run()
	if err != nil {
		t.Fatalf("Failed to initialize git repository: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("Test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Add and commit the file
	err = exec.Command("git", "add", "test.txt").SetDir(tempDir).Run()
	if err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	err = exec.Command("git", "commit", "-m", "TEST-123: Initial commit").SetDir(tempDir).Run()
	if err != nil {
		t.Fatalf("Failed to commit file: %v", err)
	}

	// Get the commit hash
	hashBytes, err := exec.Command("git", "rev-parse", "HEAD").SetDir(tempDir).Output()
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