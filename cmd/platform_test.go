package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetPlatformToolPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tests := []struct {
		name         string
		tool         string
		expectedPath string
	}{
		{
			name:         "q-cli",
			tool:         "q-cli",
			expectedPath: filepath.Join(homeDir, ".aws", "amazonq", "mcp.json"),
		},
		{
			name: "claude-desktop",
			tool: "claude-desktop",
			expectedPath: func() string {
				if runtime.GOOS == "windows" {
					return filepath.Join(homeDir, "AppData", "Roaming", "Claude", "claude_desktop_config.json")
				}
				return filepath.Join(homeDir, "Library", "Application Support", "Claude", "claude_desktop_config.json")
			}(),
		},
		{
			name:         "cursor",
			tool:         "cursor",
			expectedPath: filepath.Join(homeDir, ".cursor", "mcp.json"),
		},
		{
			name:         "kiro",
			tool:         "kiro",
			expectedPath: filepath.Join(homeDir, ".kiro", "settings", "mcp.json"),
		},
		{
			name:         "unknown tool",
			tool:         "unknown",
			expectedPath: "",
		},
		{
			name:         "empty tool",
			tool:         "",
			expectedPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPlatformToolPath(tt.tool)
			if result != tt.expectedPath {
				t.Errorf("getPlatformToolPath(%q) = %q, want %q", tt.tool, result, tt.expectedPath)
			}

			// For known tools, verify the path is absolute and starts with home directory
			if tt.expectedPath != "" {
				if !filepath.IsAbs(result) {
					t.Errorf("Expected absolute path for tool %q, got %q", tt.tool, result)
				}

				if !strings.HasPrefix(result, homeDir) {
					t.Errorf("Expected path for tool %q to start with home directory %q, got %q", tt.tool, homeDir, result)
				}
			}
		})
	}
}

func TestSupportedTools(t *testing.T) {
	expectedTools := []string{"q-cli", "claude-desktop", "cursor", "kiro"}

	if len(supportedTools) != len(expectedTools) {
		t.Errorf("Expected %d supported tools, got %d", len(expectedTools), len(supportedTools))
	}

	for _, expectedTool := range expectedTools {
		found := false
		for _, tool := range supportedTools {
			if tool == expectedTool {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected tool %q to be in supportedTools", expectedTool)
		}
	}

	// Verify all supported tools have valid paths
	for _, tool := range supportedTools {
		path := getPlatformToolPath(tool)
		if path == "" {
			t.Errorf("Tool %q should have a valid path", tool)
		}
	}
}
