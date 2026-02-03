package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExecute(t *testing.T) {
	// Test that Execute function works without panicking
	// We can't easily test the actual execution without complex setup
	// but we can test that the function exists and is callable
	// Note: We don't actually call Execute() here as it would try to run the CLI
}

func TestGetDefaultComposeFile(t *testing.T) {
	// Save original working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mcp-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	t.Run("local file exists", func(t *testing.T) {
		// Create local mcp-compose.yml
		localFile := "mcp-compose.yml"
		if err := os.WriteFile(localFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create local compose file: %v", err)
		}
		defer os.Remove(localFile)

		result := getDefaultComposeFile()
		if result != localFile {
			t.Errorf("Expected %s, got %s", localFile, result)
		}
	})

	t.Run("local file does not exist", func(t *testing.T) {
		// Ensure no local file exists
		os.Remove("mcp-compose.yml")

		result := getDefaultComposeFile()

		// Should return global config path
		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("Failed to get home dir: %v", err)
		}
		expected := filepath.Join(homeDir, ".config", "mcp", "mcp-compose.yml")

		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})
}
