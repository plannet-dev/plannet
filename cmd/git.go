package cmd

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
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
		return "", fmt.Errorf("failed to get current branch: %w", err)
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
		return nil, fmt.Errorf("failed to get recent commits: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	commits := make([]Commit, 0, len(lines))

	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) >= 2 {
			hash := parts[0]
			message := parts[1]

			var commitTime time.Time
			if len(parts) >= 3 {
				timestamp := parts[2]
				if unixSeconds, err := strconv.ParseInt(timestamp, 10, 64); err == nil {
					commitTime = time.Unix(unixSeconds, 0)
				}
			}

			commits = append(commits, Commit{
				Hash:    hash,
				Message: message,
				Time:    commitTime,
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

// getFilesChanged gets the list of files changed since a specific commit
func getFilesChanged(dir string, commitHash string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", commitHash)
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get changed files: %w", err)
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(files) == 1 && files[0] == "" {
		return []string{}, nil
	}
	return files, nil
}

// getCommitsSince gets all commits since a specific time
func getCommitsSince(dir string, since string) ([]Commit, error) {
	cmd := exec.Command("git", "log", "--since", since, "--format=%H|%s|%ct")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commits since %s: %w", since, err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	commits := make([]Commit, 0, len(lines))

	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) >= 2 {
			hash := parts[0]
			message := parts[1]

			var commitTime time.Time
			if len(parts) >= 3 {
				timestamp := parts[2]
				if unixSeconds, err := strconv.ParseInt(timestamp, 10, 64); err == nil {
					commitTime = time.Unix(unixSeconds, 0)
				}
			}

			commits = append(commits, Commit{
				Hash:    hash,
				Message: message,
				Time:    commitTime,
			})
		}
	}

	return commits, nil
}

// extractTicketIDFromMessage extracts a ticket ID from a commit message
func extractTicketIDFromMessage(message string, prefixes []string) string {
	for _, prefix := range prefixes {
		if strings.Contains(message, prefix) {
			// Find the start of the ticket ID
			startIndex := strings.Index(message, prefix)
			if startIndex == -1 {
				continue
			}

			// Extract the part after the prefix
			ticketPart := message[startIndex:]

			// Find the end of the ticket ID (usually a space, colon, or end of string)
			endIndex := strings.IndexAny(ticketPart, " :")
			if endIndex == -1 {
				endIndex = len(ticketPart)
			}

			return ticketPart[:endIndex]
		}
	}
	return ""
}
