package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "plannet-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override the config path for testing
	originalConfigPath := configPath
	configPath = filepath.Join(tempDir, ".plannetrc")
	defer func() { configPath = originalConfigPath }()

	// Reset global config
	globalConfig = nil

	// Test IsInitialized when no config exists
	if IsInitialized() {
		t.Error("IsInitialized should return false when no config exists")
	}

	// Create a test config
	testConfig := &Config{
		TicketPrefixes: []string{"TEST-", "DEV-"},
		Editor:         "vim",
		GitIntegration: true,
		BaseURL:        "https://test.example.com/v1/completions",
		Model:          "test-model",
		Headers: map[string]string{
			"Authorization": "Bearer test-token",
		},
	}

	// Test Save
	err = Save(testConfig)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Test IsInitialized after config is created
	if !IsInitialized() {
		t.Error("IsInitialized should return true after config is created")
	}

	// Test Load
	loadedConfig, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify loaded config matches test config
	if loadedConfig.Editor != testConfig.Editor {
		t.Errorf("Editor mismatch: got %s, want %s", loadedConfig.Editor, testConfig.Editor)
	}
	if loadedConfig.GitIntegration != testConfig.GitIntegration {
		t.Errorf("GitIntegration mismatch: got %v, want %v", loadedConfig.GitIntegration, testConfig.GitIntegration)
	}
	if loadedConfig.BaseURL != testConfig.BaseURL {
		t.Errorf("BaseURL mismatch: got %s, want %s", loadedConfig.BaseURL, testConfig.BaseURL)
	}
	if loadedConfig.Model != testConfig.Model {
		t.Errorf("Model mismatch: got %s, want %s", loadedConfig.Model, testConfig.Model)
	}
	if loadedConfig.Headers["Authorization"] != testConfig.Headers["Authorization"] {
		t.Errorf("Headers mismatch: got %s, want %s", loadedConfig.Headers["Authorization"], testConfig.Headers["Authorization"])
	}

	// Test Get
	getConfig, err := Get()
	if err != nil {
		t.Fatalf("Failed to get config: %v", err)
	}
	if getConfig != loadedConfig {
		t.Error("Get should return the same config as Load")
	}

	// Test GetConfigPath
	if GetConfigPath() != configPath {
		t.Errorf("GetConfigPath returned %s, want %s", GetConfigPath(), configPath)
	}
}

func TestConfigValidation(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "plannet-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override the config path for testing
	originalConfigPath := configPath
	configPath = filepath.Join(tempDir, ".plannetrc")
	defer func() { configPath = originalConfigPath }()

	// Reset global config
	globalConfig = nil

	// Test with invalid JSON
	err = os.WriteFile(configPath, []byte("invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// Test Load with invalid JSON
	_, err = Load()
	if err == nil {
		t.Error("Load should return an error for invalid JSON")
	}
} 