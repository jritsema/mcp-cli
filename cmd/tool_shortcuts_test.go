package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestToolShortcutPaths tests that all existing tool shortcuts have correct paths
func TestToolShortcutPaths(t *testing.T) {
	// Get user home directory for path expansion
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	// Test all existing tool shortcuts
	tools := []string{"q-cli", "claude-desktop", "cursor", "kiro"}

	for _, tool := range tools {
		// Test path from getPlatformToolPath
		path := getPlatformToolPath(tool)
		if path == "" {
			t.Errorf("Tool shortcut '%s' returned empty path", tool)
			continue
		}

		// Verify path starts with home directory
		if !strings.HasPrefix(path, homeDir) {
			t.Errorf("Tool '%s': path should start with home directory, got '%s'", tool, path)
		}

		// Verify path is absolute
		if !filepath.IsAbs(path) {
			t.Errorf("Tool '%s': path should be absolute, got '%s'", tool, path)
		}
	}
}

// TestGetOutputPathWithToolShortcuts tests the getOutputPath function with tool shortcuts
func TestGetOutputPathWithToolShortcuts(t *testing.T) {
	// Test that getOutputPath works with tool shortcuts (using real paths)
	tools := []string{"q-cli", "claude-desktop", "cursor", "kiro"}
	
	for _, tool := range tools {
		// Simulate setting the tool shortcut
		originalToolShortcut := toolShortcut
		toolShortcut = tool

		// Test getOutputPath with empty env vars (will use real home dir)
		envVars := make(map[string]string)
		outputPath, err := getOutputPath(envVars)
		if err != nil {
			t.Errorf("getOutputPath failed for tool '%s': %v", tool, err)
			continue
		}

		// Verify path is not empty and is absolute
		if outputPath == "" {
			t.Errorf("Tool '%s': output path should not be empty", tool)
		}

		if !filepath.IsAbs(outputPath) {
			t.Errorf("Tool '%s': output path should be absolute, got '%s'", tool, outputPath)
		}

		// Verify path ends with expected filename
		expectedFilenames := map[string]string{
			"q-cli":          "mcp.json",
			"claude-desktop": "claude_desktop_config.json",
			"cursor":         "mcp.json",
			"kiro":           "mcp.json",
		}

		expectedFilename := expectedFilenames[tool]
		if !strings.HasSuffix(outputPath, expectedFilename) {
			t.Errorf("Tool '%s': path should end with '%s', got '%s'", tool, expectedFilename, outputPath)
		}

		toolShortcut = originalToolShortcut
	}
}

// TestToolCompatibilityWithLocalServers tests that all tools work with local servers
func TestToolCompatibilityWithLocalServers(t *testing.T) {
	// Create test servers (all local)
	localServers := map[string]Service{
		"command-server": {
			Command: "uvx mcp-server-time",
			Labels:  map[string]string{"mcp.profile": "default"},
		},
		"container-server": {
			Image:  "my-server:latest",
			Labels: map[string]string{"mcp.profile": "programming"},
		},
		"simple-server": {
			Command: "python -m server",
		},
	}

	// Test that all tool shortcuts work with local servers
	tools := []string{"q-cli", "claude-desktop", "cursor", "kiro"}
	for _, tool := range tools {
		err := ValidateToolSupport(tool, localServers)
		if err != nil {
			t.Errorf("Tool '%s' should support local servers: %v", tool, err)
		}
	}

	// Test empty tool shortcut (should always work)
	err := ValidateToolSupport("", localServers)
	if err != nil {
		t.Errorf("Empty tool shortcut should work with local servers: %v", err)
	}
}

// TestToolCompatibilityWithRemoteServers tests tool compatibility with remote servers
func TestToolCompatibilityWithRemoteServers(t *testing.T) {
	// Create test servers with remote server
	mixedServers := map[string]Service{
		"local-server": {
			Command: "uvx mcp-server-time",
			Labels:  map[string]string{"mcp.profile": "default"},
		},
		"remote-server": {
			Command: "https://example.com/mcp",
			Labels: map[string]string{
				"mcp.profile":        "default",
				"mcp.grant-type":     "client_credentials",
				"mcp.token-endpoint": "https://auth.example.com/oauth2/token",
				"mcp.client-id":      "test-id",
				"mcp.client-secret":  "test-secret",
			},
		},
	}

	// Test supported tools
	supportedTools := []string{"kiro", "q-cli"}
	for _, tool := range supportedTools {
		err := ValidateToolSupport(tool, mixedServers)
		if err != nil {
			t.Errorf("Tool '%s' should support remote servers: %v", tool, err)
		}
	}

	// Test unsupported tools
	unsupportedTools := []string{"cursor", "claude-desktop"}
	for _, tool := range unsupportedTools {
		err := ValidateToolSupport(tool, mixedServers)
		if err == nil {
			t.Errorf("Tool '%s' should NOT support remote servers", tool)
		}
	}

	// Test empty tool shortcut (should work - no validation needed)
	err := ValidateToolSupport("", mixedServers)
	if err != nil {
		t.Errorf("Empty tool shortcut should work with mixed servers: %v", err)
	}
}

// TestMCPConfigurationFormatConsistency tests that MCP config format is consistent across tools
func TestMCPConfigurationFormatConsistency(t *testing.T) {
	// Create test servers
	servers := map[string]Service{
		"command-server": {
			Command:     "uvx mcp-server-time --timezone=UTC",
			Environment: map[string]string{"DEBUG": "true"},
			Labels:      map[string]string{"mcp.profile": "default"},
		},
		"container-server": {
			Image:       "my-server:latest",
			Environment: map[string]string{"API_KEY": "test123"},
			Labels:      map[string]string{"mcp.profile": "programming"},
		},
	}

	envVars := map[string]string{
		"DEBUG":   "true",
		"API_KEY": "test123",
	}

	// Generate MCP configuration
	mcpConfig := convertToMCPConfig(servers, envVars)

	// Verify structure is consistent regardless of tool
	if len(mcpConfig.MCPServers) != 2 {
		t.Errorf("Expected 2 servers, got %d", len(mcpConfig.MCPServers))
	}

	// Verify command-based server
	commandServer, exists := mcpConfig.MCPServers["command-server"]
	if !exists {
		t.Fatal("command-server not found")
	}

	if commandServer.Command != "uvx" {
		t.Errorf("Expected command 'uvx', got '%s'", commandServer.Command)
	}

	expectedArgs := []string{"mcp-server-time", "--timezone=UTC"}
	if len(commandServer.Args) != len(expectedArgs) {
		t.Errorf("Expected %d args, got %d", len(expectedArgs), len(commandServer.Args))
	}

	if commandServer.Env["DEBUG"] != "true" {
		t.Errorf("Expected DEBUG=true, got '%s'", commandServer.Env["DEBUG"])
	}

	// Verify container-based server
	containerServer, exists := mcpConfig.MCPServers["container-server"]
	if !exists {
		t.Fatal("container-server not found")
	}

	if containerServer.Command != "docker" {
		t.Errorf("Expected command 'docker', got '%s'", containerServer.Command)
	}

	// Verify args contain expected elements
	argsStr := strings.Join(containerServer.Args, " ")
	expectedElements := []string{"run", "-i", "--rm", "my-server:latest", "API_KEY=test123"}
	for _, element := range expectedElements {
		if !strings.Contains(argsStr, element) {
			t.Errorf("Expected args to contain '%s', got: %v", element, containerServer.Args)
		}
	}

	// Verify no remote server fields are set for local servers
	for name, server := range mcpConfig.MCPServers {
		if server.Type != "" {
			t.Errorf("Local server '%s' should not have Type field", name)
		}
		if server.URL != "" {
			t.Errorf("Local server '%s' should not have URL field", name)
		}
		if server.Headers != nil {
			t.Errorf("Local server '%s' should not have Headers field", name)
		}
	}
}
