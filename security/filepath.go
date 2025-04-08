package security

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SanitizeFilePath sanitizes a file path to prevent path traversal attacks
func SanitizeFilePath(baseDir, filePath string) (string, error) {
	fmt.Printf("Debug: Input baseDir=%s, filePath=%s\n", baseDir, filePath)

	// Check if the file path is empty
	if filePath == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}

	// Check for path traversal attempts
	if strings.Contains(filePath, "..") {
		return "", fmt.Errorf("file path contains forbidden path traversal pattern")
	}

	// Resolve the absolute paths and any symbolic links
	resolvedBase, err := filepath.EvalSymlinks(baseDir)
	if err != nil {
		fmt.Printf("Debug: Error resolving base directory: %v\n", err)
		resolvedBase = baseDir // Fallback to unresolved path
	}
	fmt.Printf("Debug: Resolved base directory: %s\n", resolvedBase)

	absPath, err := filepath.Abs(filepath.Join(baseDir, filePath))
	if err != nil {
		return "", fmt.Errorf("error resolving absolute path: %w", err)
	}
	fmt.Printf("Debug: Absolute path: %s\n", absPath)

	resolvedPath, err := filepath.EvalSymlinks(filepath.Dir(absPath))
	if err != nil && !os.IsNotExist(err) {
		fmt.Printf("Debug: Error resolving path: %v\n", err)
		resolvedPath = filepath.Dir(absPath) // Fallback to unresolved path
	}
	fmt.Printf("Debug: Resolved path: %s\n", resolvedPath)

	// If the file doesn't exist yet, check its parent directory
	if os.IsNotExist(err) {
		resolvedPath = filepath.Dir(resolvedPath)
		fmt.Printf("Debug: Using parent directory: %s\n", resolvedPath)
	}

	// Clean both paths for comparison
	resolvedBase = filepath.Clean(resolvedBase)
	resolvedPath = filepath.Clean(resolvedPath)
	fmt.Printf("Debug: Cleaned paths - base=%s, path=%s\n", resolvedBase, resolvedPath)

	// Check if the resolved path is within the base directory
	if !strings.HasPrefix(resolvedPath, resolvedBase) {
		return "", fmt.Errorf("file path is outside the base directory (base=%s, path=%s)", resolvedBase, resolvedPath)
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
