package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// TestMixedConfigurationScenarios tests that local and remote servers work together
func TestMixedConfigurationScenarios(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := ioutil.TempDir("", "mcp_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test compose file with mixed local and remote servers
	composeContent := `services:
  local-time:
    command: uvx mcp-server-time --local-timezone=America/New_York
    labels:
      mcp.profile: default,programming

  container-server:
    image: my-mcp-server:latest
    environment:
      API_KEY: ${TEST_API_KEY}
    labels:
      mcp.profile: programming

  remote-server:
    command: https://example.com/mcp
    labels:
      mcp.profile: default,research
      mcp.grant-type: client_credentials
      mcp.token-endpoint: https://auth.example.com/oauth2/token
      mcp.client-id: ${REMOTE_CLIENT_ID}
      mcp.client-secret: ${REMOTE_CLIENT_SECRET}
`

	composeFile := filepath.Join(tempDir, "mcp-compose.yml")
	if err := ioutil.WriteFile(composeFile, []byte(composeContent), 0644); err != nil {
		t.Fatalf("Failed to write compose file: %v", err)
	}

	// Create .env file
	envContent := `TEST_API_KEY=test-key-123
REMOTE_CLIENT_ID=test-client-id
REMOTE_CLIENT_SECRET=test-client-secret
`
	envFile := filepath.Join(tempDir, ".env")
	if err := ioutil.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to write .env file: %v", err)
	}

	// Load configuration
	config, err := loadComposeFile(composeFile)
	if err != nil {
		t.Fatalf("Failed to load compose file: %v", err)
	}

	// Load environment variables
	envVars, err := loadEnvVars(composeFile)
	if err != nil {
		t.Fatalf("Failed to load env vars: %v", err)
	}

	// Test 1: Verify all servers are loaded
	if len(config.Services) != 3 {
		t.Errorf("Expected 3 services, got %d", len(config.Services))
	}

	// Test 2: Verify remote server detection
	remoteService := config.Services["remote-server"]
	if !IsRemoteServer(remoteService) {
		t.Error("remote-server should be detected as remote")
	}

	localService := config.Services["local-time"]
	if IsRemoteServer(localService) {
		t.Error("local-time should not be detected as remote")
	}

	containerService := config.Services["container-server"]
	if IsRemoteServer(containerService) {
		t.Error("container-server should not be detected as remote")
	}

	// Test 3: Test profile filtering with mixed server types
	defaultServers := filterServers(config, "", false)
	expectedDefault := []string{"local-time", "remote-server"}
	if len(defaultServers) != 2 {
		t.Errorf("Expected 2 default servers, got %d", len(defaultServers))
	}
	for _, name := range expectedDefault {
		if _, exists := defaultServers[name]; !exists {
			t.Errorf("Expected default server '%s' not found", name)
		}
	}

	programmingServers := filterServers(config, "programming", false)
	// When filtering by "programming", we get:
	// - local-time (has both default and programming profiles)
	// - container-server (has programming profile)
	// - remote-server (has default profile, so included when requesting specific profile)
	expectedProgramming := []string{"local-time", "container-server", "remote-server"}
	if len(programmingServers) != 3 {
		t.Errorf("Expected 3 programming servers, got %d", len(programmingServers))
	}
	for _, name := range expectedProgramming {
		if _, exists := programmingServers[name]; !exists {
			t.Errorf("Expected programming server '%s' not found", name)
		}
	}

	researchServers := filterServers(config, "research", false)
	expectedResearch := []string{"local-time", "remote-server"} // local-time is default, remote-server has research profile
	if len(researchServers) != 2 {
		t.Errorf("Expected 2 research servers, got %d", len(researchServers))
	}
	for _, name := range expectedResearch {
		if _, exists := researchServers[name]; !exists {
			t.Errorf("Expected research server '%s' not found", name)
		}
	}

	// Test 4: Verify environment variable expansion works for all server types
	expandedApiKey := expandEnvVars("${TEST_API_KEY}", envVars)
	if expandedApiKey != "test-key-123" {
		t.Errorf("Expected expanded API key 'test-key-123', got '%s'", expandedApiKey)
	}

	expandedClientId := expandEnvVars("${REMOTE_CLIENT_ID}", envVars)
	if expandedClientId != "test-client-id" {
		t.Errorf("Expected expanded client ID 'test-client-id', got '%s'", expandedClientId)
	}

	// Test 5: Verify OAuth config extraction for remote servers
	oauthConfig, err := ExtractOAuthConfig(remoteService, envVars)
	if err != nil {
		t.Fatalf("Failed to extract OAuth config: %v", err)
	}

	if oauthConfig.ClientID != "test-client-id" {
		t.Errorf("Expected client ID 'test-client-id', got '%s'", oauthConfig.ClientID)
	}

	if oauthConfig.ClientSecret != "test-client-secret" {
		t.Errorf("Expected client secret 'test-client-secret', got '%s'", oauthConfig.ClientSecret)
	}

	// Test 6: Verify tool compatibility validation
	mixedServers := map[string]Service{
		"local-time":    localService,
		"remote-server": remoteService,
	}

	// Should pass for supported tools
	if err := ValidateToolSupport("kiro", mixedServers); err != nil {
		t.Errorf("kiro should support mixed servers: %v", err)
	}

	if err := ValidateToolSupport("q-cli", mixedServers); err != nil {
		t.Errorf("q-cli should support mixed servers: %v", err)
	}

	// Should fail for unsupported tools
	if err := ValidateToolSupport("cursor", mixedServers); err == nil {
		t.Error("cursor should not support remote servers")
	}

	// Should pass when no tool is specified
	if err := ValidateToolSupport("", mixedServers); err != nil {
		t.Errorf("Empty tool should be allowed: %v", err)
	}
}

// TestExistingFunctionalityPreservation tests that existing functionality continues to work
func TestExistingFunctionalityPreservation(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := ioutil.TempDir("", "mcp_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test compose file with only local servers (existing functionality)
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

  no-profile-server:
    command: python -m my_mcp_server
`

	composeFile := filepath.Join(tempDir, "mcp-compose.yml")
	if err := ioutil.WriteFile(composeFile, []byte(composeContent), 0644); err != nil {
		t.Fatalf("Failed to write compose file: %v", err)
	}

	// Create .env file
	envContent := `TIMEZONE=America/New_York
API_KEY=secret-key-123
`
	envFile := filepath.Join(tempDir, ".env")
	if err := ioutil.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to write .env file: %v", err)
	}

	// Load configuration
	config, err := loadComposeFile(composeFile)
	if err != nil {
		t.Fatalf("Failed to load compose file: %v", err)
	}

	// Load environment variables
	envVars, err := loadEnvVars(composeFile)
	if err != nil {
		t.Fatalf("Failed to load env vars: %v", err)
	}

	// Test 1: Verify no servers are detected as remote
	for name, service := range config.Services {
		if IsRemoteServer(service) {
			t.Errorf("Server '%s' should not be detected as remote", name)
		}
	}

	// Test 2: Verify profile filtering works as before
	defaultServers := filterServers(config, "", false)
	expectedDefault := []string{"time-server", "no-profile-server"} // default profile + no profile
	if len(defaultServers) != 2 {
		t.Errorf("Expected 2 default servers, got %d", len(defaultServers))
	}
	for _, name := range expectedDefault {
		if _, exists := defaultServers[name]; !exists {
			t.Errorf("Expected default server '%s' not found", name)
		}
	}

	programmingServers := filterServers(config, "programming", false)
	expectedProgramming := []string{"time-server", "container-server", "no-profile-server"} // default + programming + no profile
	if len(programmingServers) != 3 {
		t.Errorf("Expected 3 programming servers, got %d", len(programmingServers))
	}
	for _, name := range expectedProgramming {
		if _, exists := programmingServers[name]; !exists {
			t.Errorf("Expected programming server '%s' not found", name)
		}
	}

	// Test 3: Verify environment variable expansion works
	expandedTimezone := expandEnvVars("${TIMEZONE}", envVars)
	if expandedTimezone != "America/New_York" {
		t.Errorf("Expected expanded timezone 'America/New_York', got '%s'", expandedTimezone)
	}

	expandedApiKey := expandEnvVars("${API_KEY}", envVars)
	if expandedApiKey != "secret-key-123" {
		t.Errorf("Expected expanded API key 'secret-key-123', got '%s'", expandedApiKey)
	}

	// Test 4: Verify MCP configuration generation for local servers
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

	expectedArgs := []string{"mcp-server-time", "--local-timezone=America/New_York"}
	if len(timeServer.Args) != len(expectedArgs) {
		t.Errorf("Expected %d args, got %d", len(expectedArgs), len(timeServer.Args))
	}
	for i, arg := range expectedArgs {
		if i >= len(timeServer.Args) || timeServer.Args[i] != arg {
			t.Errorf("Expected arg[%d] '%s', got '%s'", i, arg, timeServer.Args[i])
		}
	}

	if timeServer.Env["TIMEZONE"] != "America/New_York" {
		t.Errorf("Expected expanded TIMEZONE 'America/New_York', got '%s'", timeServer.Env["TIMEZONE"])
	}

	// Verify no remote server fields are set for local servers
	if timeServer.Type != "" {
		t.Errorf("Local server should not have Type field set, got '%s'", timeServer.Type)
	}
	if timeServer.URL != "" {
		t.Errorf("Local server should not have URL field set, got '%s'", timeServer.URL)
	}
	if timeServer.Headers != nil {
		t.Errorf("Local server should not have Headers field set")
	}

	// Test 5: Verify tool compatibility validation allows all tools for local servers
	localOnlyServers := map[string]Service{
		"time-server": config.Services["time-server"],
	}

	supportedTools := []string{"kiro", "q-cli", "cursor", "claude-desktop"}
	for _, tool := range supportedTools {
		if err := ValidateToolSupport(tool, localOnlyServers); err != nil {
			t.Errorf("Tool '%s' should support local servers: %v", tool, err)
		}
	}
}

// TestContainerServerBackwardCompatibility specifically tests container-based servers
func TestContainerServerBackwardCompatibility(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := ioutil.TempDir("", "mcp_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test compose file with container servers
	composeContent := `services:
  container-with-env:
    image: my-mcp-server:${VERSION}
    environment:
      API_KEY: ${API_KEY}
      DEBUG: "true"
    labels:
      mcp.profile: default
`

	composeFile := filepath.Join(tempDir, "mcp-compose.yml")
	if err := ioutil.WriteFile(composeFile, []byte(composeContent), 0644); err != nil {
		t.Fatalf("Failed to write compose file: %v", err)
	}

	// Create .env file
	envContent := `VERSION=v1.2.3
API_KEY=container-secret-123
`
	envFile := filepath.Join(tempDir, ".env")
	if err := ioutil.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to write .env file: %v", err)
	}

	// Load configuration
	config, err := loadComposeFile(composeFile)
	if err != nil {
		t.Fatalf("Failed to load compose file: %v", err)
	}

	// Load environment variables
	envVars, err := loadEnvVars(composeFile)
	if err != nil {
		t.Fatalf("Failed to load env vars: %v", err)
	}

	// Test 1: Verify container server is not detected as remote
	containerService := config.Services["container-with-env"]
	if IsRemoteServer(containerService) {
		t.Error("Container server should not be detected as remote")
	}

	// Test 2: Verify MCP configuration generation for container servers
	servers := filterServers(config, "", false)
	mcpConfig := convertToMCPConfig(servers, envVars)

	containerServer, exists := mcpConfig.MCPServers["container-with-env"]
	if !exists {
		t.Fatal("container-with-env not found in MCP config")
	}

	// Should use docker command by default
	if containerServer.Command != "docker" {
		t.Errorf("Expected command 'docker', got '%s'", containerServer.Command)
	}

	// Verify args include expanded image name and environment variables
	expectedArgsContain := []string{"run", "-i", "--rm", "-e", "API_KEY=container-secret-123", "-e", "DEBUG=true", "my-mcp-server:v1.2.3"}

	// Check that all expected args are present (order may vary for environment variables)
	for _, expectedArg := range expectedArgsContain {
		found := false
		for _, actualArg := range containerServer.Args {
			if actualArg == expectedArg {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected arg '%s' not found in %v", expectedArg, containerServer.Args)
		}
	}

	// Verify no remote server fields are set
	if containerServer.Type != "" {
		t.Errorf("Container server should not have Type field set, got '%s'", containerServer.Type)
	}
	if containerServer.URL != "" {
		t.Errorf("Container server should not have URL field set, got '%s'", containerServer.URL)
	}
	if containerServer.Headers != nil {
		t.Errorf("Container server should not have Headers field set")
	}
}
