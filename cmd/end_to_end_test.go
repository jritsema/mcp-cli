package cmd

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// TestEndToEndBackwardCompatibility tests a complete end-to-end scenario with local servers
func TestEndToEndBackwardCompatibility(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := ioutil.TempDir("", "mcp_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a realistic compose file with various local server types
	composeContent := `services:
  time-server:
    command: uvx mcp-server-time --local-timezone=${TIMEZONE}
    environment:
      DEBUG: ${DEBUG_MODE}
      LOG_LEVEL: info
    labels:
      mcp.profile: default,productivity

  weather-server:
    image: weather-mcp:${VERSION}
    environment:
      API_KEY: ${WEATHER_API_KEY}
      UNITS: metric
    labels:
      mcp.profile: productivity

  file-server:
    command: python -m file_mcp_server --root-dir=${HOME}/Documents
    environment:
      MAX_FILE_SIZE: "10MB"
    labels:
      mcp.profile: development

  simple-calculator:
    command: node calculator-server.js
    labels:
      mcp.profile: default

  no-profile-server:
    command: basic-server --port=8080
`

	composeFile := filepath.Join(tempDir, "mcp-compose.yml")
	if err := ioutil.WriteFile(composeFile, []byte(composeContent), 0644); err != nil {
		t.Fatalf("Failed to write compose file: %v", err)
	}

	// Create .env file with realistic environment variables
	envContent := `TIMEZONE=America/New_York
DEBUG_MODE=false
VERSION=v2.1.0
WEATHER_API_KEY=sk-weather-123456789
HOME=/Users/testuser
`
	envFile := filepath.Join(tempDir, ".env")
	if err := ioutil.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to write .env file: %v", err)
	}

	// Test the complete workflow that existing users would experience

	// Step 1: Load configuration (like all commands do)
	config, err := loadComposeFile(composeFile)
	if err != nil {
		t.Fatalf("Failed to load compose file: %v", err)
	}

	if len(config.Services) != 5 {
		t.Errorf("Expected 5 services, got %d", len(config.Services))
	}

	// Step 2: Load environment variables (like all commands do)
	envVars, err := loadEnvVars(composeFile)
	if err != nil {
		t.Fatalf("Failed to load env vars: %v", err)
	}

	// Verify environment variables are loaded correctly
	// Note: System env vars take precedence over .env file
	expectedEnvVars := map[string]string{
		"TIMEZONE":        "America/New_York",
		"DEBUG_MODE":      "false",
		"VERSION":         "v2.1.0",
		"WEATHER_API_KEY": "sk-weather-123456789",
	}

	for key, expected := range expectedEnvVars {
		if envVars[key] != expected {
			t.Errorf("Expected env var %s='%s', got '%s'", key, expected, envVars[key])
		}
	}

	// HOME should be set (either from system or .env file)
	if envVars["HOME"] == "" {
		t.Error("HOME environment variable should be set")
	}

	// Step 3: Test profile filtering (like list and set commands do)
	testCases := []struct {
		profile  string
		expected []string
	}{
		// Default profile: servers with default profile + servers with no profile
		{"", []string{"time-server", "simple-calculator", "no-profile-server"}},
		// Productivity profile: servers with productivity profile + default servers + no-profile servers
		{"productivity", []string{"time-server", "weather-server", "simple-calculator", "no-profile-server"}},
		// Development profile: servers with development profile + default servers + no-profile servers
		{"development", []string{"time-server", "file-server", "simple-calculator", "no-profile-server"}},
	}

	for _, tc := range testCases {
		servers := filterServers(config, tc.profile, false)
		if len(servers) != len(tc.expected) {
			t.Errorf("Profile '%s': expected %d servers, got %d", tc.profile, len(tc.expected), len(servers))
		}
		for _, expectedName := range tc.expected {
			if _, exists := servers[expectedName]; !exists {
				t.Errorf("Profile '%s': expected server '%s' not found", tc.profile, expectedName)
			}
		}

		// Verify all servers are local (not remote)
		for name, service := range servers {
			if IsRemoteServer(service) {
				t.Errorf("Server '%s' should not be detected as remote", name)
			}
		}
	}

	// Step 4: Test MCP configuration generation (like set command does)
	defaultServers := filterServers(config, "", false)
	mcpConfig := convertToMCPConfig(defaultServers, envVars)

	if len(mcpConfig.MCPServers) != 3 {
		t.Errorf("Expected 3 MCP servers, got %d", len(mcpConfig.MCPServers))
	}

	// Verify command-based server with environment variable expansion
	timeServer, exists := mcpConfig.MCPServers["time-server"]
	if !exists {
		t.Fatal("time-server not found in MCP config")
	}

	if timeServer.Command != "uvx" {
		t.Errorf("Expected command 'uvx', got '%s'", timeServer.Command)
	}

	expectedArgs := []string{"mcp-server-time", "--local-timezone=America/New_York"}
	if len(timeServer.Args) != len(expectedArgs) {
		t.Errorf("Expected %d args, got %d", len(expectedArgs), len(timeServer.Args))
	}
	for i, arg := range expectedArgs {
		if i >= len(timeServer.Args) || timeServer.Args[i] != arg {
			t.Errorf("Expected arg[%d] '%s', got '%s'", i, arg, timeServer.Args[i])
		}
	}

	if timeServer.Env["DEBUG"] != "false" {
		t.Errorf("Expected DEBUG='false', got '%s'", timeServer.Env["DEBUG"])
	}

	if timeServer.Env["LOG_LEVEL"] != "info" {
		t.Errorf("Expected LOG_LEVEL='info', got '%s'", timeServer.Env["LOG_LEVEL"])
	}

	// Verify simple command server
	calculatorServer, exists := mcpConfig.MCPServers["simple-calculator"]
	if !exists {
		t.Fatal("simple-calculator not found in MCP config")
	}

	if calculatorServer.Command != "node" {
		t.Errorf("Expected command 'node', got '%s'", calculatorServer.Command)
	}

	if len(calculatorServer.Args) != 1 || calculatorServer.Args[0] != "calculator-server.js" {
		t.Errorf("Expected args ['calculator-server.js'], got %v", calculatorServer.Args)
	}

	// Verify no remote server fields are set
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

	// Step 5: Test JSON serialization (like writeMCPConfig does)
	jsonData, err := json.MarshalIndent(mcpConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal MCP config: %v", err)
	}

	// Step 6: Test writing and reading config file
	configFile := filepath.Join(tempDir, "test-mcp.json")
	if err := ioutil.WriteFile(configFile, jsonData, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Verify file was written correctly
	readData, err := ioutil.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var readConfig MCPConfig
	if err := json.Unmarshal(readData, &readConfig); err != nil {
		t.Fatalf("Failed to unmarshal read config: %v", err)
	}

	if len(readConfig.MCPServers) != len(mcpConfig.MCPServers) {
		t.Errorf("Read config should have same number of servers as original")
	}

	// Step 7: Test tool compatibility (all tools should work with local servers)
	allTools := []string{"q-cli", "claude-desktop", "cursor", "kiro"}
	for _, tool := range allTools {
		if err := ValidateToolSupport(tool, defaultServers); err != nil {
			t.Errorf("Tool '%s' should support local servers: %v", tool, err)
		}
	}

	// Step 8: Test that all existing functionality works with container servers
	productivityServers := filterServers(config, "productivity", false)
	productivityConfig := convertToMCPConfig(productivityServers, envVars)

	weatherServer, exists := productivityConfig.MCPServers["weather-server"]
	if !exists {
		t.Fatal("weather-server not found in productivity config")
	}

	if weatherServer.Command != "docker" {
		t.Errorf("Expected container command 'docker', got '%s'", weatherServer.Command)
	}

	// Verify container args include expanded image name and environment variables
	argsStr := ""
	for _, arg := range weatherServer.Args {
		argsStr += arg + " "
	}

	expectedElements := []string{"run", "-i", "--rm", "weather-mcp:v2.1.0", "API_KEY=sk-weather-123456789", "UNITS=metric"}
	for _, element := range expectedElements {
		if !contains(weatherServer.Args, element) {
			t.Errorf("Expected container args to contain '%s', got: %v", element, weatherServer.Args)
		}
	}
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}