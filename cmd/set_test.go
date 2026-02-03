package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteMCPConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mcp-write-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := MCPConfig{
		MCPServers: map[string]MCPServer{
			"test-server": {
				Command: "uvx",
				Args:    []string{"mcp-server-time"},
				Env:     map[string]string{"DEBUG": "true"},
			},
		},
	}

	testPath := filepath.Join(tempDir, "test-config.json")

	t.Run("successful write", func(t *testing.T) {
		err := writeMCPConfig(config, testPath)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify file was created and has correct content
		if !fileExists(testPath) {
			t.Error("Config file was not created")
		}

		// Read and verify content
		data, err := os.ReadFile(testPath)
		if err != nil {
			t.Fatalf("Failed to read config file: %v", err)
		}

		var readConfig MCPConfig
		if err := json.Unmarshal(data, &readConfig); err != nil {
			t.Fatalf("Failed to parse config file: %v", err)
		}

		if len(readConfig.MCPServers) != 1 {
			t.Errorf("Expected 1 server, got %d", len(readConfig.MCPServers))
		}

		server, exists := readConfig.MCPServers["test-server"]
		if !exists {
			t.Error("Expected test-server to exist")
		}

		if server.Command != "uvx" {
			t.Errorf("Expected command 'uvx', got %s", server.Command)
		}
	})

	t.Run("write to invalid path", func(t *testing.T) {
		invalidPath := filepath.Join("/invalid/path/that/does/not/exist", "config.json")
		err := writeMCPConfig(config, invalidPath)
		if err == nil {
			t.Error("Expected error for invalid path")
		}
	})
}

func TestGetOutputPath(t *testing.T) {
	// Save original flag values
	originalConfigFile := configFile
	originalToolShortcut := toolShortcut

	// Restore original values after test
	defer func() {
		configFile = originalConfigFile
		toolShortcut = originalToolShortcut
	}()

	envVars := map[string]string{
		"HOME": "/home/testuser",
	}

	t.Run("explicit config file", func(t *testing.T) {
		configFile = "/path/to/config.json"
		toolShortcut = ""

		result, err := getOutputPath(envVars)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result != "/path/to/config.json" {
			t.Errorf("Expected /path/to/config.json, got %s", result)
		}
	})

	t.Run("config file with env vars", func(t *testing.T) {
		configFile = "${HOME}/my-config.json"
		toolShortcut = ""

		result, err := getOutputPath(envVars)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result != "/home/testuser/my-config.json" {
			t.Errorf("Expected /home/testuser/my-config.json, got %s", result)
		}
	})

	t.Run("valid tool shortcut", func(t *testing.T) {
		configFile = ""
		toolShortcut = "kiro"

		result, err := getOutputPath(envVars)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Should return the platform-specific path for kiro
		homeDir, _ := os.UserHomeDir()
		expected := filepath.Join(homeDir, ".kiro", "settings", "mcp.json")
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("invalid tool shortcut", func(t *testing.T) {
		configFile = ""
		toolShortcut = "invalid-tool"

		_, err := getOutputPath(envVars)
		if err == nil {
			t.Error("Expected error for invalid tool shortcut")
		}
	})

	t.Run("no config file or tool shortcut", func(t *testing.T) {
		configFile = ""
		toolShortcut = ""

		// Use a temporary home directory to avoid interfering with real config
		originalHome := os.Getenv("HOME")
		tempHome, err := os.MkdirTemp("", "test-home")
		if err != nil {
			t.Fatalf("Failed to create temp home: %v", err)
		}
		defer func() {
			os.Setenv("HOME", originalHome)
			os.RemoveAll(tempHome)
		}()
		os.Setenv("HOME", tempHome)

		_, err = getOutputPath(envVars)
		if err == nil {
			t.Error("Expected error when neither config file nor tool shortcut is specified")
		}
	})
}

func TestConvertToMCPConfig(t *testing.T) {
	envVars := map[string]string{
		"DEBUG":   "true",
		"API_KEY": "secret123",
		"IMAGE":   "my-server:latest",
	}

	t.Run("command-based server", func(t *testing.T) {
		servers := map[string]Service{
			"command-server": {
				Command:     "uvx mcp-server-time --debug",
				Environment: map[string]string{"DEBUG": "${DEBUG}"},
			},
		}

		result := convertToMCPConfig(servers, envVars)

		if len(result.MCPServers) != 1 {
			t.Errorf("Expected 1 server, got %d", len(result.MCPServers))
		}

		server, exists := result.MCPServers["command-server"]
		if !exists {
			t.Fatal("Expected command-server to exist")
		}

		if server.Command != "uvx" {
			t.Errorf("Expected command 'uvx', got %s", server.Command)
		}

		expectedArgs := []string{"mcp-server-time", "--debug"}
		if len(server.Args) != len(expectedArgs) {
			t.Errorf("Expected %d args, got %d", len(expectedArgs), len(server.Args))
		}

		for i, expectedArg := range expectedArgs {
			if i < len(server.Args) && server.Args[i] != expectedArg {
				t.Errorf("Expected arg[%d] = %s, got %s", i, expectedArg, server.Args[i])
			}
		}

		if server.Env["DEBUG"] != "true" {
			t.Errorf("Expected DEBUG=true, got %s", server.Env["DEBUG"])
		}
	})

	t.Run("container-based server", func(t *testing.T) {
		servers := map[string]Service{
			"container-server": {
				Image:       "${IMAGE}",
				Environment: map[string]string{"API_KEY": "${API_KEY}"},
				Volumes:     []string{"/host:/container"},
			},
		}

		result := convertToMCPConfig(servers, envVars)

		server, exists := result.MCPServers["container-server"]
		if !exists {
			t.Fatal("Expected container-server to exist")
		}

		if server.Command != "docker" {
			t.Errorf("Expected command 'docker', got %s", server.Command)
		}

		// Check that args contain expected elements
		argsStr := ""
		for _, arg := range server.Args {
			argsStr += arg + " "
		}

		expectedElements := []string{"run", "-i", "--rm", "my-server:latest", "API_KEY=secret123", "/host:/container"}
		for _, element := range expectedElements {
			if !strings.Contains(argsStr, element) {
				t.Errorf("Expected args to contain '%s', got: %v", element, server.Args)
			}
		}
	})

	t.Run("remote server with headers", func(t *testing.T) {
		servers := map[string]Service{
			"remote-server": {
				Command: "https://api.example.com/mcp",
				Labels: map[string]string{
					"mcp.header.Authorization": "Bearer ${API_KEY}",
				},
				Environment: map[string]string{"API_KEY": "${API_KEY}"},
			},
		}

		result := convertToMCPConfig(servers, envVars)

		server, exists := result.MCPServers["remote-server"]
		if !exists {
			t.Fatal("Expected remote-server to exist")
		}

		if server.Type != "http" {
			t.Errorf("Expected type 'http', got %s", server.Type)
		}

		if server.URL != "https://api.example.com/mcp" {
			t.Errorf("Expected URL 'https://api.example.com/mcp', got %s", server.URL)
		}

		if server.Headers["Authorization"] != "Bearer secret123" {
			t.Errorf("Expected Authorization header 'Bearer secret123', got %s", server.Headers["Authorization"])
		}
	})

	t.Run("empty servers", func(t *testing.T) {
		servers := map[string]Service{}

		result := convertToMCPConfig(servers, envVars)

		if len(result.MCPServers) != 0 {
			t.Errorf("Expected 0 servers, got %d", len(result.MCPServers))
		}
	})
}

// TestGetOutputPathWithDefaultTool tests getOutputPath with a default tool configured
func TestGetOutputPathWithDefaultTool(t *testing.T) {
	// Save original flag values
	originalConfigFile := configFile
	originalToolShortcut := toolShortcut

	// Restore original values after test
	defer func() {
		configFile = originalConfigFile
		toolShortcut = originalToolShortcut
	}()

	configFile = ""
	toolShortcut = ""

	// Create a temporary home directory
	tempHome, err := os.MkdirTemp("", "test-home")
	if err != nil {
		t.Fatalf("Failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tempHome)

	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)

	// Create config directory and file with default tool
	configDir := filepath.Join(tempHome, ".config", "mcp")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.json")
	defaultToolPath := filepath.Join(tempHome, "default-tool.json")

	config := CLIConfig{
		Tool: defaultToolPath,
	}

	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	envVars := map[string]string{}
	result, err := getOutputPath(envVars)
	if err != nil {
		t.Errorf("Expected no error with default tool configured, got %v", err)
	}

	if result != defaultToolPath {
		t.Errorf("Expected %s, got %s", defaultToolPath, result)
	}
}

// TestConvertToMCPConfigAdvanced tests more complex scenarios
func TestConvertToMCPConfigAdvanced(t *testing.T) {
	envVars := map[string]string{
		"DEBUG":       "true",
		"API_KEY":     "secret123",
		"IMAGE":       "my-server:latest",
		"CUSTOM_TOOL": "podman",
	}

	t.Run("remote server with OAuth", func(t *testing.T) {
		servers := map[string]Service{
			"oauth-server": {
				Command: "https://api.example.com/mcp",
				Labels: map[string]string{
					"mcp.grant-type":     "client_credentials",
					"mcp.token-endpoint": "https://auth.example.com/token",
					"mcp.client-id":      "${API_KEY}",
					"mcp.client-secret":  "secret456",
				},
				Environment: map[string]string{
					"API_KEY": "${API_KEY}",
				},
			},
		}

		// Note: This test would normally fail due to OAuth network calls
		// In a real implementation, we'd mock the HTTP client
		// For now, we'll skip the actual conversion and just test the structure

		// Verify the service is recognized as remote
		if !IsRemoteServer(servers["oauth-server"]) {
			t.Error("Expected oauth-server to be recognized as remote")
		}

		// Verify OAuth config extraction works
		envVarsWithExpansion := make(map[string]string)
		for k, v := range envVars {
			envVarsWithExpansion[k] = v
		}
		for key, value := range servers["oauth-server"].Environment {
			expandedValue := expandEnvVars(value, envVars)
			envVarsWithExpansion[key] = expandedValue
		}

		oauthConfig, err := ExtractOAuthConfig(servers["oauth-server"], envVarsWithExpansion)
		if err != nil {
			t.Errorf("Failed to extract OAuth config: %v", err)
		}

		if oauthConfig.ClientID != "secret123" {
			t.Errorf("Expected ClientID=secret123, got %s", oauthConfig.ClientID)
		}
	})

	t.Run("container server with custom tool", func(t *testing.T) {
		// Create a temporary config to test custom container tool
		tempHome, err := os.MkdirTemp("", "test-home")
		if err != nil {
			t.Fatalf("Failed to create temp home: %v", err)
		}
		defer os.RemoveAll(tempHome)

		originalHome := os.Getenv("HOME")
		defer os.Setenv("HOME", originalHome)
		os.Setenv("HOME", tempHome)

		// Create config with custom container tool
		configDir := filepath.Join(tempHome, ".config", "mcp")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config dir: %v", err)
		}

		configPath := filepath.Join(configDir, "config.json")
		config := CLIConfig{
			ContainerTool: "podman",
		}

		configData, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal config: %v", err)
		}

		if err := os.WriteFile(configPath, configData, 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		servers := map[string]Service{
			"podman-server": {
				Image:       "${IMAGE}",
				Environment: map[string]string{"DEBUG": "${DEBUG}"},
				Volumes:     []string{"/host:/container"},
			},
		}

		result := convertToMCPConfig(servers, envVars)

		server, exists := result.MCPServers["podman-server"]
		if !exists {
			t.Fatal("Expected podman-server to exist")
		}

		if server.Command != "podman" {
			t.Errorf("Expected command 'podman', got %s", server.Command)
		}

		// Check that args contain expected elements
		argsStr := strings.Join(server.Args, " ")
		expectedElements := []string{"run", "-i", "--rm", "my-server:latest", "DEBUG=true", "/host:/container"}
		for _, element := range expectedElements {
			if !strings.Contains(argsStr, element) {
				t.Errorf("Expected args to contain '%s', got: %v", element, server.Args)
			}
		}
	})

	t.Run("mixed server types", func(t *testing.T) {
		servers := map[string]Service{
			"command-server": {
				Command:     "uvx mcp-server-time",
				Environment: map[string]string{"DEBUG": "${DEBUG}"},
			},
			"container-server": {
				Image:       "my-server:latest",
				Environment: map[string]string{"API_KEY": "${API_KEY}"},
			},
			"remote-server": {
				Command: "https://api.example.com/mcp",
				Labels: map[string]string{
					"mcp.header.Authorization": "Bearer ${API_KEY}",
				},
				Environment: map[string]string{
					"API_KEY": "${API_KEY}",
				},
			},
		}

		result := convertToMCPConfig(servers, envVars)

		if len(result.MCPServers) != 3 {
			t.Errorf("Expected 3 servers, got %d", len(result.MCPServers))
		}

		// Check command server
		commandServer := result.MCPServers["command-server"]
		if commandServer.Command != "uvx" {
			t.Errorf("Expected command server command 'uvx', got %s", commandServer.Command)
		}

		// Check container server
		containerServer := result.MCPServers["container-server"]
		if containerServer.Command != "docker" {
			t.Errorf("Expected container server command 'docker', got %s", containerServer.Command)
		}

		// Check remote server
		remoteServer := result.MCPServers["remote-server"]
		if remoteServer.Type != "http" {
			t.Errorf("Expected remote server type 'http', got %s", remoteServer.Type)
		}
		if remoteServer.URL != "https://api.example.com/mcp" {
			t.Errorf("Expected remote server URL 'https://api.example.com/mcp', got %s", remoteServer.URL)
		}
	})

	t.Run("server with complex command", func(t *testing.T) {
		servers := map[string]Service{
			"complex-server": {
				Command:     "python -m server --host localhost --port 8080 --debug",
				Environment: map[string]string{"PYTHONPATH": "/custom/path"},
			},
		}

		result := convertToMCPConfig(servers, envVars)

		server := result.MCPServers["complex-server"]
		if server.Command != "python" {
			t.Errorf("Expected command 'python', got %s", server.Command)
		}

		expectedArgs := []string{"-m", "server", "--host", "localhost", "--port", "8080", "--debug"}
		if len(server.Args) != len(expectedArgs) {
			t.Errorf("Expected %d args, got %d", len(expectedArgs), len(server.Args))
		}

		for i, expectedArg := range expectedArgs {
			if i < len(server.Args) && server.Args[i] != expectedArg {
				t.Errorf("Expected arg[%d] = %s, got %s", i, expectedArg, server.Args[i])
			}
		}

		if server.Env["PYTHONPATH"] != "/custom/path" {
			t.Errorf("Expected PYTHONPATH=/custom/path, got %s", server.Env["PYTHONPATH"])
		}
	})
}
