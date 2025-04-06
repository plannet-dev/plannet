package security

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ValidateURL validates a URL string
func ValidateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use HTTP or HTTPS scheme")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL must have a host")
	}

	return nil
}

// ValidateTicketKey validates a Jira ticket key
func ValidateTicketKey(ticketKey string) error {
	if ticketKey == "" {
		return fmt.Errorf("ticket key cannot be empty")
	}

	// Jira ticket keys typically follow the pattern PROJECT-123
	// where PROJECT is uppercase letters and numbers, and 123 is a number
	pattern := regexp.MustCompile(`^[A-Z0-9]+-\d+$`)
	if !pattern.MatchString(ticketKey) {
		return fmt.Errorf("invalid ticket key format: %s. Expected format: PROJECT-123", ticketKey)
	}

	return nil
}

// ValidateAPIKey validates an API key
func ValidateAPIKey(apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	// Basic validation - API keys should not contain whitespace
	if strings.ContainsAny(apiKey, " \t\n\r") {
		return fmt.Errorf("API key cannot contain whitespace")
	}

	// API keys should be reasonably long
	if len(apiKey) < 10 {
		return fmt.Errorf("API key is too short")
	}

	return nil
}

// ValidateFilePath validates a file path to prevent path traversal attacks
func ValidateFilePath(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	// Check for path traversal attempts
	if strings.Contains(filePath, "..") {
		return fmt.Errorf("file path contains forbidden path traversal pattern")
	}

	// Check for absolute paths
	if strings.HasPrefix(filePath, "/") || (len(filePath) > 1 && filePath[1] == ':') {
		return fmt.Errorf("absolute file paths are not allowed")
	}

	return nil
}

// SanitizeInput sanitizes user input to prevent injection attacks
func SanitizeInput(input string) string {
	// Replace potentially dangerous characters
	// This is a simple example and should be customized based on your specific needs
	return strings.Map(func(r rune) rune {
		switch r {
		case '<', '>', '"', '\'', '&':
			return ' '
		default:
			return r
		}
	}, input)
}
