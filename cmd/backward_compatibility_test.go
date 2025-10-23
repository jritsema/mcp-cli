package cmd

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// TestAllCommandsWithLocalServers tests that all existing commands work unchanged with local servers
func TestAllCommandsWithLocalServers(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := ioutil.TempDir("", "mcp_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test compose file with various local server types
	composeContent := `services:
  command-server:
    command: uvx mcp-server-time --local-timezone=America/New_York
    environment:
      TIMEZONE: ${TIMEZONE}
    labels:
      mcp.profile: default,programming

  container-server:
    image: my-mcp-server:latest
    environment:
      API_KEY: ${API_KEY}
    labels:
      mcp.profile: programming

  simple-command:
    command: python -m my_server
    labels:
      mcp.profile: research

  no-profile-server:
    command: node server.js
`

	composeFile := filepath.Join(tempDir, "mcp-compose.yml")
	if err := ioutil.WriteFile(composeFile, []byte(composeContent), 0644); err != nil {
		t.Fatalf("Failed to write compose file: %v", err)
	}

	// Create .env file
	envContent := `TIMEZONE=America/New_York
API_KEY=test-api-key-123
`
	envFile := filepath.Join(tempDir, ".env")
	if err := ioutil.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to write .env file: %v", err)
	}

	// Test loadComposeFile function
	config, err := loadComposeFile(composeFile)
	if err != nil {
		t.Fatalf("loadComposeFile failed: %v", err)
	}

	if len(config.Services) != 4 {
		t.Errorf("Expected 4 services, got %d", len(config.Services))
	}

	// Test loadEnvVars function
	envVars, err := loadEnvVars(composeFile)
	if err != nil {
		t.Fatalf("loadEnvVars failed: %v", err)
	}

	if envVars["TIMEZONE"] != "America/New_York" {
		t.Errorf("Expected TIMEZONE 'America/New_York', got '%s'", envVars["TIMEZONE"])
	}

	if envVars["API_KEY"] != "test-api-key-123" {
		t.Errorf("Expected API_KEY 'test-api-key-123', got '%s'", envVars["API_KEY"])
	}

	// Test expandEnvVars function
	testString := "Server running on ${TIMEZONE} with key ${API_KEY}"
	expanded := expandEnvVars(testString, envVars)
	expected := "Server running on America/New_York with key test-api-key-123"
	if expanded != expected {
		t.Errorf("Expected '%s', got '%s'", expected, expanded)
	}

	// Test filterServers function with different profiles
	testCases := []struct {
		profile  string
		expected []string
	}{
		{"", []string{"command-server", "no-profile-server"}},                                // default servers
		{"programming", []string{"command-server", "container-server", "no-profile-server"}}, // programming + default + no-profile
		{"research", []string{"command-server", "simple-command", "no-profile-server"}},      // research + default + no-profile
		{"nonexistent", []string{"command-server", "no-profile-server"}},                     // only default + no-profile
	}

	for _, tc := range testCases {
		filtered := filterServers(config, tc.profile, false)
		if len(filtered) != len(tc.expected) {
			t.Errorf("Profile '%s': expected %d servers, got %d", tc.profile, len(tc.expected), len(filtered))
		}
		for _, expectedName := range tc.expected {
			if _, exists := filtered[expectedName]; !exists {
				t.Errorf("Profile '%s': expected server '%s' not found", tc.profile, expectedName)
			}
		}
	}

	// Test filterServers with all=true
	allServers := filterServers(config, "", true)
	if len(allServers) != 4 {
		t.Errorf("Expected all 4 servers with all=true, got %d", len(allServers))
	}

	// Test convertToMCPConfig function
	defaultServers := filterServers(config, "", false)
	mcpConfig := convertToMCPConfig(defaultServers, envVars)

	if len(mcpConfig.MCPServers) != 2 {
		t.Errorf("Expected 2 MCP servers, got %d", len(mcpConfig.MCPServers))
	}

	// Verify command-based server configuration
	commandServer, exists := mcpConfig.MCPServers["command-server"]
	if !exists {
		t.Fatal("command-server not found in MCP config")
	}

	if commandServer.Command != "uvx" {
		t.Errorf("Expected command 'uvx', got '%s'", commandServer.Command)
	}

	expectedArgs := []string{"mcp-server-time", "--local-timezone=America/New_York"}
	if len(commandServer.Args) != len(expectedArgs) {
		t.Errorf("Expected %d args, got %d", len(expectedArgs), len(commandServer.Args))
	}

	if commandServer.Env["TIMEZONE"] != "America/New_York" {
		t.Errorf("Expected expanded TIMEZONE, got '%s'", commandServer.Env["TIMEZONE"])
	}

	// Verify simple command server
	noProfileServer, exists := mcpConfig.MCPServers["no-profile-server"]
	if !exists {
		t.Fatal("no-profile-server not found in MCP config")
	}

	if noProfileServer.Command != "node" {
		t.Errorf("Expected command 'node', got '%s'", noProfileServer.Command)
	}

	if len(noProfileServer.Args) != 1 || noProfileServer.Args[0] != "server.js" {
		t.Errorf("Expected args ['server.js'], got %v", noProfileServer.Args)
	}
}

// TestToolShortcutsBackwardCompatibility tests that existing tool shortcuts work with local servers
func TestToolShortcutsBackwardCompatibility(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := ioutil.TempDir("", "mcp_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test compose file with local servers only
	composeContent := `services:
  local-server:
    command: uvx mcp-server-time
    labels:
      mcp.profile: default
`

	composeFile := filepath.Join(tempDir, "mcp-compose.yml")
	if err := ioutil.WriteFile(composeFile, []byte(composeContent), 0644); err != nil {
		t.Fatalf("Failed to write compose file: %v", err)
	}

	// Load configuration
	config, err := loadComposeFile(composeFile)
	if err != nil {
		t.Fatalf("Failed to load compose file: %v", err)
	}

	_, err = loadEnvVars(composeFile)
	if err != nil {
		t.Fatalf("Failed to load env vars: %v", err)
	}

	servers := filterServers(config, "", false)

	// Test that all tool shortcuts work with local servers
	toolShortcuts := []string{"q-cli", "claude-desktop", "cursor", "kiro"}

	for _, tool := range toolShortcuts {
		err := ValidateToolSupport(tool, servers)
		if err != nil {
			t.Errorf("Tool '%s' should support local servers: %v", tool, err)
		}
	}

	// Test that empty tool shortcut works (no validation needed)
	err = ValidateToolSupport("", servers)
	if err != nil {
		t.Errorf("Empty tool shortcut should work: %v", err)
	}
}

// TestEnvironmentVariableExpansionBackwardCompatibility tests that env var expansion continues to work
func TestEnvironmentVariableExpansionBackwardCompatibility(t *testing.T) {
	// Test various environment variable formats
	envVars := map[string]string{
		"SIMPLE_VAR":    "simple_value",
		"PATH_VAR":      "/usr/local/bin",
		"COMPLEX_VAR":   "value-with-dashes_and_underscores.123",
		"EMPTY_VAR":     "",
		"SPECIAL_CHARS": "value!@#$%^&*()",
	}

	testCases := []struct {
		input    string
		expected string
	}{
		{"${SIMPLE_VAR}", "simple_value"},
		{"$SIMPLE_VAR", "simple_value"},
		{"prefix-${SIMPLE_VAR}-suffix", "prefix-simple_value-suffix"},
		{"${PATH_VAR}/bin", "/usr/local/bin/bin"},
		{"${COMPLEX_VAR}", "value-with-dashes_and_underscores.123"},
		{"${EMPTY_VAR}", ""},
		{"${SPECIAL_CHARS}", "value!@#$%^&*()"},
		{"${NONEXISTENT_VAR}", "${NONEXISTENT_VAR}"}, // Should remain unchanged
		{"$NONEXISTENT_VAR", "$NONEXISTENT_VAR"},     // Should remain unchanged
		{"no variables here", "no variables here"},
		{"${SIMPLE_VAR} and ${PATH_VAR}", "simple_value and /usr/local/bin"},
		{"$SIMPLE_VAR and $PATH_VAR", "simple_value and /usr/local/bin"},
		{"mixed ${SIMPLE_VAR} and $PATH_VAR", "mixed simple_value and /usr/local/bin"},
	}

	for _, tc := range testCases {
		result := expandEnvVars(tc.input, envVars)
		if result != tc.expected {
			t.Errorf("expandEnvVars('%s'): expected '%s', got '%s'", tc.input, tc.expected, result)
		}
	}
}

// TestConfigurationFileFormatsBackwardCompatibility tests that existing config file formats work
func TestConfigurationFileFormatsBackwardCompatibility(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := ioutil.TempDir("", "mcp_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test various YAML formats that should be supported
	yamlFormats := []struct {
		name    string
		content string
	}{
		{
			name: "minimal",
			content: `services:
  server1:
    command: uvx mcp-server-time`,
		},
		{
			name: "with-environment",
			content: `services:
  server1:
    command: uvx mcp-server-time
    environment:
      KEY1: value1
      KEY2: value2`,
		},
		{
			name: "with-labels",
			content: `services:
  server1:
    command: uvx mcp-server-time
    labels:
      mcp.profile: default`,
		},
		{
			name: "with-image",
			content: `services:
  server1:
    image: my-server:latest`,
		},
		{
			name: "complex",
			content: `services:
  server1:
    command: uvx mcp-server-time --timezone=UTC
    environment:
      API_KEY: ${API_KEY}
      DEBUG: "true"
    labels:
      mcp.profile: default,programming
  server2:
    image: my-server:${VERSION}
    environment:
      CONFIG_PATH: /app/config
    labels:
      mcp.profile: research`,
		},
	}

	for _, format := range yamlFormats {
		t.Run(format.name, func(t *testing.T) {
			composeFile := filepath.Join(tempDir, format.name+"-compose.yml")
			if err := ioutil.WriteFile(composeFile, []byte(format.content), 0644); err != nil {
				t.Fatalf("Failed to write compose file: %v", err)
			}

			// Should be able to load without errors
			config, err := loadComposeFile(composeFile)
			if err != nil {
				t.Fatalf("Failed to load %s format: %v", format.name, err)
			}

			// Should have at least one service
			if len(config.Services) == 0 {
				t.Errorf("No services loaded from %s format", format.name)
			}

			// Should be able to process with existing functions
			envVars := map[string]string{
				"API_KEY": "test-key",
				"VERSION": "v1.0.0",
			}

			servers := filterServers(config, "", false)
			mcpConfig := convertToMCPConfig(servers, envVars)

			// Should generate valid MCP configuration
			if len(mcpConfig.MCPServers) == 0 {
				t.Errorf("No MCP servers generated from %s format", format.name)
			}

			// Should be able to marshal to JSON
			_, err = json.Marshal(mcpConfig)
			if err != nil {
				t.Errorf("Failed to marshal MCP config from %s format: %v", format.name, err)
			}
		})
	}
}
