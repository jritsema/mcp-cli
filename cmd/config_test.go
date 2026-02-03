package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetConfigDir(t *testing.T) {
	result := getConfigDir()

	// Should return a path under the user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	expected := filepath.Join(homeDir, ".config", "mcp")
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// Should be an absolute path
	if !filepath.IsAbs(result) {
		t.Errorf("Expected absolute path, got %s", result)
	}
}
