package cmd

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// TestSetCommandWithLocalServers tests that the set command works unchanged with local servers
func TestSetCommandWithLocalServers(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := ioutil.TempDir("", "mcp_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test compose file with local servers only
	composeContent := `services:
  time-server:
    command: uvx mcp-server-time --local-timezone=America/New_York
    environment:
      TIMEZONE: ${TIMEZONE}
    labels:
      mcp.profile: default

  container-server:
    image: my-mcp-server:latest
    environment:
      API_KEY: ${API_KEY}
    labels:
      mcp.profile: programming

  simple-server:
    command: python -m my_server
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

	// Test the core functionality that the set command uses
	config, err := loadComposeFile(composeFile)
	if err != nil {
		t.Fatalf("Failed to load compose file: %v", err)
	}

	envVars, err := loadEnvVars(composeFile)
	if err != nil {
		t.Fatalf("Failed to load env vars: %v", err)
	}

	// Test default profile filtering (what set command does with no args)
	defaultServers := filterServers(config, "", false)
	expectedDefault := []string{"time-server", "simple-server"} // default + no profile
	if len(defaultServers) != 2 {
		t.Errorf("Expected 2 default servers, got %d", len(defaultServers))
	}
	for _, name := range expectedDefault {
		if _, exists := defaultServers[name]; !exists {
			t.Errorf("Expected default server '%s' not found", name)
		}
	}

	// Test specific profile filtering (what set command does with profile arg)
	programmingServers := filterServers(config, "programming", false)
	if len(programmingServers) != 3 {
		t.Errorf("Expected 3 programming servers, got %d", len(programmingServers))
	}

	// Test OAuth validation (should pass for local servers)
	for name, service := range defaultServers {
		if IsRemoteServer(service) {
			t.Errorf("Local server '%s' incorrectly detected as remote", name)
		}
		// OAuth validation should not be called for local servers, but if it is, it should not error
	}

	// Test tool compatibility validation (should pass for all tools with local servers)
	supportedTools := []string{"q-cli", "claude-desktop", "cursor", "kiro", ""}
	for _, tool := range supportedTools {
		if err := ValidateToolSupport(tool, defaultServers); err != nil {
			t.Errorf("Tool '%s' should support local servers: %v", tool, err)
		}
	}

	// Test MCP configuration generation
	mcpConfig := convertToMCPConfig(defaultServers, envVars)
	if len(mcpConfig.MCPServers) != 2 {
		t.Errorf("Expected 2 MCP servers, got %d", len(mcpConfig.MCPServers))
	}

	// Verify command-based server
	timeServer, exists := mcpConfig.MCPServers["time-server"]
	if !exists {
		t.Fatal("time-server not found in MCP config")
	}

	if timeServer.Command != "uvx" {
		t.Errorf("Expected command 'uvx', got '%s'", timeServer.Command)
	}

	if timeServer.Type != "" {
		t.Errorf("Local server should not have Type field, got '%s'", timeServer.Type)
	}

	if timeServer.URL != "" {
		t.Errorf("Local server should not have URL field, got '%s'", timeServer.URL)
	}

	if timeServer.Headers != nil {
		t.Errorf("Local server should not have Headers field")
	}

	// Test JSON marshaling (what writeMCPConfig does)
	jsonData, err := json.MarshalIndent(mcpConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal MCP config: %v", err)
	}

	// Verify JSON structure
	var parsedConfig MCPConfig
	if err := json.Unmarshal(jsonData, &parsedConfig); err != nil {
		t.Fatalf("Failed to unmarshal generated JSON: %v", err)
	}

	if len(parsedConfig.MCPServers) != 2 {
		t.Errorf("Parsed config should have 2 servers, got %d", len(parsedConfig.MCPServers))
	}
}

// TestListCommandCompatibility tests that list command functionality works with local servers
func TestListCommandCompatibility(t *testing.T) {
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
      TIMEZONE: America/New_York
    labels:
      mcp.profile: default,programming

  container-server:
    image: my-mcp-server:latest
    environment:
      API_KEY: secret123
    labels:
      mcp.profile: programming

  simple-server:
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

	// Test the core functionality that the list command uses
	config, err := loadComposeFile(composeFile)
	if err != nil {
		t.Fatalf("Failed to load compose file: %v", err)
	}

	// Test filtering by different profiles (what list command does)
	testCases := []struct {
		profile  string
		all      bool
		expected int
	}{
		{"", false, 2},            // default servers: command-server (default), no-profile-server (no profile)
		{"programming", false, 3}, // programming + default + no-profile: command-server, container-server, no-profile-server
		{"research", false, 3},    // research + default + no-profile: command-server, simple-server, no-profile-server
		{"nonexistent", false, 2}, // only default + no-profile: command-server, no-profile-server
		{"", true, 4},             // all servers
	}

	for _, tc := range testCases {
		servers := filterServers(config, tc.profile, tc.all)
		if len(servers) != tc.expected {
			t.Errorf("Profile '%s' (all=%v): expected %d servers, got %d", tc.profile, tc.all, tc.expected, len(servers))
		}

		// Verify all returned servers are local (not remote)
		for name, service := range servers {
			if IsRemoteServer(service) {
				t.Errorf("Server '%s' should not be detected as remote", name)
			}
		}
	}
}

// TestClearCommandCompatibility tests that clear command functionality works
func TestClearCommandCompatibility(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := ioutil.TempDir("", "mcp_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test MCP config file
	mcpConfig := MCPConfig{
		MCPServers: map[string]MCPServer{
			"test-server": {
				Command: "uvx",
				Args:    []string{"mcp-server-time"},
				Env:     map[string]string{"TIMEZONE": "UTC"},
			},
		},
	}

	configFile := filepath.Join(tempDir, "mcp.json")
	jsonData, err := json.MarshalIndent(mcpConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}

	if err := ioutil.WriteFile(configFile, jsonData, 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Fatal("Test config file should exist")
	}

	// Test file removal (what clear command does)
	if err := os.Remove(configFile); err != nil {
		t.Fatalf("Failed to remove config file: %v", err)
	}

	// Verify file is gone
	if _, err := os.Stat(configFile); !os.IsNotExist(err) {
		t.Error("Config file should be removed")
	}
}

// TestConfigCommandCompatibility tests that config command functionality works
func TestConfigCommandCompatibility(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := ioutil.TempDir("", "mcp_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test CLI config structure (what config command uses)
	cliConfig := CLIConfig{
		Tool:          filepath.Join(tempDir, "test-mcp.json"),
		ContainerTool: "podman",
	}

	configFile := filepath.Join(tempDir, "config.json")
	jsonData, err := json.MarshalIndent(cliConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal CLI config: %v", err)
	}

	if err := ioutil.WriteFile(configFile, jsonData, 0644); err != nil {
		t.Fatalf("Failed to write CLI config: %v", err)
	}

	// Test reading config (what config command does)
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read CLI config: %v", err)
	}

	var parsedConfig CLIConfig
	if err := json.Unmarshal(data, &parsedConfig); err != nil {
		t.Fatalf("Failed to unmarshal CLI config: %v", err)
	}

	if parsedConfig.Tool != cliConfig.Tool {
		t.Errorf("Expected tool '%s', got '%s'", cliConfig.Tool, parsedConfig.Tool)
	}

	if parsedConfig.ContainerTool != cliConfig.ContainerTool {
		t.Errorf("Expected container tool '%s', got '%s'", cliConfig.ContainerTool, parsedConfig.ContainerTool)
	}
}
