package cmd

import (
	"os"
	"os/exec"
	"strings"
	"time"
	"fmt"
)

// Commit represents a git commit
type Commit struct {
	Hash    string
	Message string
	Time    time.Time
}

// isGitRepo checks if the given directory is a git repository
func isGitRepo(dir string) bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = dir
	err := cmd.Run()
	return err == nil
}

// getCurrentBranch gets the current branch name
func getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// extractTicketID extracts a ticket ID from a branch name
func extractTicketID(branchName string, prefixes []string) string {
	for _, prefix := range prefixes {
		if strings.Contains(branchName, prefix) {
			parts := strings.Split(branchName, prefix)
			if len(parts) > 1 {
				// Extract the ticket ID (e.g., "123" from "feature/JIRA-123")
				ticketPart := parts[1]
				// Find the end of the ticket ID (usually a slash, hyphen, or end of string)
				endIndex := strings.IndexAny(ticketPart, "/-_")
				if endIndex == -1 {
					endIndex = len(ticketPart)
				}
				return prefix + ticketPart[:endIndex]
			}
		}
	}
	return ""
}

// getRecentCommits gets the most recent commits
func getRecentCommits(count int) ([]Commit, error) {
	cmd := exec.Command("git", "log", "-n", fmt.Sprintf("%d", count), "--format=%H|%s|%ct")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	commits := make([]Commit, 0, len(lines))

	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) >= 2 {
			hash := parts[0]
			message := parts[1]
			
			var time time.Time
			if len(parts) >= 3 {
				timestamp := parts[2]
				unixTime, err := time.Parse("1234567890", timestamp)
				if err == nil {
					time = time.Unix(unixTime.Unix(), 0)
				}
			}

			commits = append(commits, Commit{
				Hash:    hash,
				Message: message,
				Time:    time,
			})
		}
	}

	return commits, nil
}

// findSideQuests finds commits that don't contain ticket IDs
func findSideQuests(commits []Commit, prefixes []string) []Commit {
	sideQuests := []Commit{}
	
	for _, commit := range commits {
		hasTicketID := false
		for _, prefix := range prefixes {
			if strings.Contains(commit.Message, prefix) {
				hasTicketID = true
				break
			}
		}
		
		if !hasTicketID {
			sideQuests = append(sideQuests, commit)
		}
	}
	
	return sideQuests
} 