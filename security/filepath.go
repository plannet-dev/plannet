package security

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SanitizeFilePath sanitizes a file path to prevent path traversal attacks
func SanitizeFilePath(baseDir, filePath string) (string, error) {
	// Check if the file path is empty
	if filePath == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}

	// Check for path traversal attempts
	if strings.Contains(filePath, "..") {
		return "", fmt.Errorf("file path contains forbidden path traversal pattern")
	}

	// Resolve the absolute path
	absPath, err := filepath.Abs(filepath.Join(baseDir, filePath))
	if err != nil {
		return "", fmt.Errorf("error resolving absolute path: %w", err)
	}

	// Check if the resolved path is within the base directory
	if !strings.HasPrefix(absPath, baseDir) {
		return "", fmt.Errorf("file path is outside the base directory")
	}

	return absPath, nil
}

// SafeReadFile safely reads a file with path sanitization
func SafeReadFile(baseDir, filePath string) ([]byte, error) {
	// Sanitize the file path
	safePath, err := SanitizeFilePath(baseDir, filePath)
	if err != nil {
		return nil, err
	}

	// Read the file
	return os.ReadFile(safePath)
}

// SafeWriteFile safely writes to a file with path sanitization
func SafeWriteFile(baseDir, filePath string, data []byte, perm os.FileMode) error {
	// Sanitize the file path
	safePath, err := SanitizeFilePath(baseDir, filePath)
	if err != nil {
		return err
	}

	// Create the directory if it doesn't exist
	dir := filepath.Dir(safePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	// Write the file
	return os.WriteFile(safePath, data, perm)
}

// SafeRemoveFile safely removes a file with path sanitization
func SafeRemoveFile(baseDir, filePath string) error {
	// Sanitize the file path
	safePath, err := SanitizeFilePath(baseDir, filePath)
	if err != nil {
		return err
	}

	// Remove the file
	return os.Remove(safePath)
}

// SafeCreateTempFile safely creates a temporary file with path sanitization
func SafeCreateTempFile(baseDir, pattern string) (*os.File, error) {
	// Sanitize the pattern
	safePattern, err := SanitizeFilePath(baseDir, pattern)
	if err != nil {
		return nil, err
	}

	// Create the temporary file
	return os.CreateTemp(baseDir, filepath.Base(safePattern))
}
