package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// supportedTools lists all supported tool shortcuts
var supportedTools = []string{"q-cli", "claude-desktop", "cursor", "kiro"}

// getPlatformToolPath returns the platform-appropriate path for a tool
// Hard fails on error, consistent with getConfigDir() in config.go
func getPlatformToolPath(tool string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting user home directory: %v\n", err)
		os.Exit(1)
	}

	switch tool {
	case "q-cli":
		return filepath.Join(homeDir, ".aws", "amazonq", "mcp.json")
	case "claude-desktop":
		if runtime.GOOS == "windows" {
			return filepath.Join(homeDir, "AppData", "Roaming", "Claude", "claude_desktop_config.json")
		}
		return filepath.Join(homeDir, "Library", "Application Support", "Claude", "claude_desktop_config.json")
	case "cursor":
		return filepath.Join(homeDir, ".cursor", "mcp.json")
	case "kiro":
		return filepath.Join(homeDir, ".kiro", "settings", "mcp.json")
	default:
		return ""
	}
}
