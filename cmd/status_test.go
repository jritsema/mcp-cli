package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists(t *testing.T) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "test-file")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "existing file",
			path:     tempFile.Name(),
			expected: true,
		},
		{
			name:     "non-existing file",
			path:     "/path/that/does/not/exist",
			expected: false,
		},
		{
			name:     "empty path",
			path:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fileExists(tt.path)
			if result != tt.expected {
				t.Errorf("fileExists(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestLoadToolConfig(t *testing.T) {
	// Create a temporary directory for test configs
	tempDir, err := os.MkdirTemp("", "mcp-tool-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a valid MCP config file
	validConfig := MCPConfig{
		MCPServers: map[string]MCPServer{
			"test-server": {
				Command: "uvx",
				Args:    []string{"mcp-server-time"},
				Env:     map[string]string{"DEBUG": "true"},
			},
		},
	}

	validConfigPath := filepath.Join(tempDir, "valid-config.json")
	validConfigData, _ := json.MarshalIndent(validConfig, "", "  ")
	if err := os.WriteFile(validConfigPath, validConfigData, 0644); err != nil {
		t.Fatalf("Failed to create valid config file: %v", err)
	}

	// Create an invalid JSON file
	invalidConfigPath := filepath.Join(tempDir, "invalid-config.json")
	if err := os.WriteFile(invalidConfigPath, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to create invalid config file: %v", err)
	}

	t.Run("unknown tool", func(t *testing.T) {
		config, path, err := loadToolConfig("unknown-tool")
		if err == nil {
			t.Error("Expected error for unknown tool")
		}
		if path != "" {
			t.Errorf("Expected empty path for unknown tool, got %s", path)
		}
		if len(config.MCPServers) != 0 {
			t.Error("Expected empty config for unknown tool")
		}
	})

	// We can't easily test the other cases without mocking getPlatformToolPath
	// which would require refactoring the code to make it more testable
}

func TestGetToolConfigs(t *testing.T) {
	tools := []string{"unknown-tool-1", "unknown-tool-2"}

	configs := getToolConfigs(tools)

	if len(configs) != 2 {
		t.Errorf("Expected 2 configs, got %d", len(configs))
	}

	for _, tool := range tools {
		config, exists := configs[tool]
		if !exists {
			t.Errorf("Expected config for tool %s", tool)
			continue
		}

		if config.Exists {
			t.Errorf("Expected tool %s to not exist", tool)
		}

		if config.Error == "" {
			t.Errorf("Expected error for unknown tool %s", tool)
		}
	}
}

func TestCompareHeaders(t *testing.T) {
	tests := []struct {
		name     string
		expected map[string]string
		actual   map[string]string
		want     bool
	}{
		{
			name:     "identical headers",
			expected: map[string]string{"Authorization": "Bearer token", "X-API-Key": "key123"},
			actual:   map[string]string{"Authorization": "Bearer token", "X-API-Key": "key123"},
			want:     true,
		},
		{
			name:     "different values",
			expected: map[string]string{"Authorization": "Bearer token1"},
			actual:   map[string]string{"Authorization": "Bearer token2"},
			want:     false,
		},
		{
			name:     "missing header in actual",
			expected: map[string]string{"Authorization": "Bearer token", "X-API-Key": "key123"},
			actual:   map[string]string{"Authorization": "Bearer token"},
			want:     false,
		},
		{
			name:     "extra header in actual",
			expected: map[string]string{"Authorization": "Bearer token"},
			actual:   map[string]string{"Authorization": "Bearer token", "X-API-Key": "key123"},
			want:     false,
		},
		{
			name:     "both empty",
			expected: map[string]string{},
			actual:   map[string]string{},
			want:     true,
		},
		{
			name:     "both nil",
			expected: nil,
			actual:   nil,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareHeaders(tt.expected, tt.actual)
			if result != tt.want {
				t.Errorf("compareHeaders() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestCompareEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		expected map[string]string
		actual   map[string]string
		want     bool
	}{
		{
			name:     "identical env vars",
			expected: map[string]string{"DEBUG": "true", "API_KEY": "secret"},
			actual:   map[string]string{"DEBUG": "true", "API_KEY": "secret"},
			want:     true,
		},
		{
			name:     "different values",
			expected: map[string]string{"DEBUG": "true"},
			actual:   map[string]string{"DEBUG": "false"},
			want:     false,
		},
		{
			name:     "missing var in actual",
			expected: map[string]string{"DEBUG": "true", "API_KEY": "secret"},
			actual:   map[string]string{"DEBUG": "true"},
			want:     false,
		},
		{
			name:     "extra var in actual",
			expected: map[string]string{"DEBUG": "true"},
			actual:   map[string]string{"DEBUG": "true", "API_KEY": "secret"},
			want:     false,
		},
		{
			name:     "both empty",
			expected: map[string]string{},
			actual:   map[string]string{},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareEnvVars(tt.expected, tt.actual)
			if result != tt.want {
				t.Errorf("compareEnvVars() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestCompareStringSlices(t *testing.T) {
	tests := []struct {
		name string
		a    []string
		b    []string
		want bool
	}{
		{
			name: "identical slices",
			a:    []string{"arg1", "arg2", "arg3"},
			b:    []string{"arg1", "arg2", "arg3"},
			want: true,
		},
		{
			name: "different order",
			a:    []string{"arg1", "arg2"},
			b:    []string{"arg2", "arg1"},
			want: false,
		},
		{
			name: "different lengths",
			a:    []string{"arg1", "arg2"},
			b:    []string{"arg1"},
			want: false,
		},
		{
			name: "different values",
			a:    []string{"arg1", "arg2"},
			b:    []string{"arg1", "arg3"},
			want: false,
		},
		{
			name: "both empty",
			a:    []string{},
			b:    []string{},
			want: true,
		},
		{
			name: "both nil",
			a:    nil,
			b:    nil,
			want: true,
		},
		{
			name: "one nil one empty",
			a:    nil,
			b:    []string{},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareStringSlices(tt.a, tt.b)
			if result != tt.want {
				t.Errorf("compareStringSlices() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestNormalizeToolName(t *testing.T) {
	tests := []struct {
		name     string
		tool     string
		expected string
	}{
		{
			name:     "q-cli",
			tool:     "q-cli",
			expected: "Q-CLI",
		},
		{
			name:     "claude-desktop",
			tool:     "claude-desktop",
			expected: "CLAUDE",
		},
		{
			name:     "cursor",
			tool:     "cursor",
			expected: "CURSOR",
		},
		{
			name:     "kiro",
			tool:     "kiro",
			expected: "KIRO",
		},
		{
			name:     "unknown tool",
			tool:     "unknown",
			expected: "UNKNOWN",
		},
		{
			name:     "empty string",
			tool:     "",
			expected: "",
		},
		{
			name:     "mixed case",
			tool:     "MyTool",
			expected: "MYTOOL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeToolName(tt.tool)
			if result != tt.expected {
				t.Errorf("normalizeToolName(%q) = %q, want %q", tt.tool, result, tt.expected)
			}
		})
	}
}

func TestCompareServerConfig(t *testing.T) {
	envVars := map[string]string{
		"DEBUG": "true",
	}

	tests := []struct {
		name           string
		serverName     string
		composeService Service
		deployedServer MCPServer
		expectedStatus string
	}{
		{
			name:       "not configured - empty deployed server",
			serverName: "test-server",
			composeService: Service{
				Command: "uvx mcp-server-time",
			},
			deployedServer: MCPServer{},
			expectedStatus: "not-configured",
		},
		{
			name:       "remote server configured",
			serverName: "remote-server",
			composeService: Service{
				Command: "https://api.example.com/mcp",
				Labels: map[string]string{
					"mcp.header.Authorization": "Bearer token123",
				},
			},
			deployedServer: MCPServer{
				Type: "http",
				URL:  "https://api.example.com/mcp",
				Headers: map[string]string{
					"Authorization": "Bearer token123",
				},
			},
			expectedStatus: "configured",
		},
		{
			name:       "local server configured",
			serverName: "local-server",
			composeService: Service{
				Command:     "uvx mcp-server-time",
				Environment: map[string]string{"DEBUG": "true"},
			},
			deployedServer: MCPServer{
				Command: "uvx",
				Args:    []string{"mcp-server-time"},
				Env:     map[string]string{"DEBUG": "true"},
			},
			expectedStatus: "configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, _ := compareServerConfig(tt.serverName, tt.composeService, tt.deployedServer, envVars)
			if status != tt.expectedStatus {
				t.Errorf("compareServerConfig() status = %q, want %q", status, tt.expectedStatus)
			}
		})
	}
}
func TestLoadToolConfigDetailed(t *testing.T) {
	// Create a temporary directory for test configs
	tempDir, err := os.MkdirTemp("", "mcp-tool-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a valid MCP config file
	validConfig := MCPConfig{
		MCPServers: map[string]MCPServer{
			"test-server": {
				Command: "uvx",
				Args:    []string{"mcp-server-time"},
				Env:     map[string]string{"DEBUG": "true"},
			},
		},
	}

	validConfigData, _ := json.MarshalIndent(validConfig, "", "  ")

	// Mock getPlatformToolPath by creating files at expected locations
	homeDir, _ := os.UserHomeDir()
	testConfigPath := filepath.Join(homeDir, ".test-tool", "mcp.json")

	// Create directory and file
	if err := os.MkdirAll(filepath.Dir(testConfigPath), 0755); err != nil {
		t.Fatalf("Failed to create test config dir: %v", err)
	}
	defer os.RemoveAll(filepath.Join(homeDir, ".test-tool"))

	if err := os.WriteFile(testConfigPath, validConfigData, 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	t.Run("unknown tool", func(t *testing.T) {
		config, path, err := loadToolConfig("unknown-tool")
		if err == nil {
			t.Error("Expected error for unknown tool")
		}
		if path != "" {
			t.Errorf("Expected empty path for unknown tool, got %s", path)
		}
		if len(config.MCPServers) != 0 {
			t.Error("Expected empty config for unknown tool")
		}
	})

	// Test with a known tool that has a config file
	t.Run("known tool with config", func(t *testing.T) {
		// We can test with kiro since we know its path structure
		kiroPath := filepath.Join(homeDir, ".kiro", "settings", "mcp.json")

		// Create the config file
		if err := os.MkdirAll(filepath.Dir(kiroPath), 0755); err != nil {
			t.Fatalf("Failed to create kiro config dir: %v", err)
		}
		defer os.RemoveAll(filepath.Join(homeDir, ".kiro"))

		if err := os.WriteFile(kiroPath, validConfigData, 0644); err != nil {
			t.Fatalf("Failed to create kiro config file: %v", err)
		}

		config, path, err := loadToolConfig("kiro")
		if err != nil {
			t.Errorf("Expected no error for kiro tool, got %v", err)
		}

		if path != kiroPath {
			t.Errorf("Expected path %s, got %s", kiroPath, path)
		}

		if len(config.MCPServers) != 1 {
			t.Errorf("Expected 1 server, got %d", len(config.MCPServers))
		}
	})

	t.Run("known tool without config file", func(t *testing.T) {
		// Test cursor which won't have a config file
		config, path, err := loadToolConfig("cursor")
		if err != nil {
			t.Errorf("Expected no error for missing config, got %v", err)
		}

		expectedPath := filepath.Join(homeDir, ".cursor", "mcp.json")
		if path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, path)
		}

		if len(config.MCPServers) != 0 {
			t.Errorf("Expected empty config for missing file, got %d servers", len(config.MCPServers))
		}
	})
}

func TestCompareRemoteServers(t *testing.T) {
	envVars := map[string]string{
		"API_TOKEN": "secret123",
	}

	tests := []struct {
		name           string
		composeService Service
		deployedServer MCPServer
		expectedStatus string
		expectDiffs    bool
	}{
		{
			name: "matching remote server with headers",
			composeService: Service{
				Command: "https://api.example.com/mcp",
				Labels: map[string]string{
					"mcp.header.Authorization": "Bearer ${API_TOKEN}",
				},
				Environment: map[string]string{
					"API_TOKEN": "${API_TOKEN}",
				},
			},
			deployedServer: MCPServer{
				Type: "http",
				URL:  "https://api.example.com/mcp",
				Headers: map[string]string{
					"Authorization": "Bearer secret123",
				},
			},
			expectedStatus: "configured",
			expectDiffs:    false,
		},
		{
			name: "URL mismatch",
			composeService: Service{
				Command: "https://api.example.com/mcp",
				Labels: map[string]string{
					"mcp.header.Authorization": "Bearer token123",
				},
			},
			deployedServer: MCPServer{
				Type: "http",
				URL:  "https://different.example.com/mcp",
				Headers: map[string]string{
					"Authorization": "Bearer token123",
				},
			},
			expectedStatus: "different",
			expectDiffs:    true,
		},
		{
			name: "type mismatch",
			composeService: Service{
				Command: "https://api.example.com/mcp",
				Labels: map[string]string{
					"mcp.header.Authorization": "Bearer token123",
				},
			},
			deployedServer: MCPServer{
				Type: "websocket",
				URL:  "https://api.example.com/mcp",
				Headers: map[string]string{
					"Authorization": "Bearer token123",
				},
			},
			expectedStatus: "different",
			expectDiffs:    true,
		},
		{
			name: "OAuth server with valid headers",
			composeService: Service{
				Command: "https://api.example.com/mcp",
				Labels: map[string]string{
					"mcp.grant-type":     "client_credentials",
					"mcp.token-endpoint": "https://auth.example.com/token",
					"mcp.client-id":      "client123",
					"mcp.client-secret":  "secret456",
				},
			},
			deployedServer: MCPServer{
				Type: "http",
				URL:  "https://api.example.com/mcp",
				Headers: map[string]string{
					"Authorization": "Bearer oauth-token-123",
				},
			},
			expectedStatus: "configured",
			expectDiffs:    false,
		},
		{
			name: "OAuth server with missing headers",
			composeService: Service{
				Command: "https://api.example.com/mcp",
				Labels: map[string]string{
					"mcp.grant-type":     "client_credentials",
					"mcp.token-endpoint": "https://auth.example.com/token",
					"mcp.client-id":      "client123",
					"mcp.client-secret":  "secret456",
				},
			},
			deployedServer: MCPServer{
				Type: "http",
				URL:  "https://api.example.com/mcp",
			},
			expectedStatus: "different",
			expectDiffs:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, differences := compareRemoteServers(tt.composeService, tt.deployedServer, envVars)

			if status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, status)
			}

			if tt.expectDiffs && len(differences) == 0 {
				t.Error("Expected differences but got none")
			}

			if !tt.expectDiffs && len(differences) > 0 {
				t.Errorf("Expected no differences but got: %v", differences)
			}
		})
	}
}

func TestCompareLocalServers(t *testing.T) {
	envVars := map[string]string{
		"DEBUG":   "true",
		"API_KEY": "secret123",
	}

	tests := []struct {
		name           string
		serverName     string
		composeService Service
		deployedServer MCPServer
		expectedStatus string
		expectDiffs    bool
	}{
		{
			name:       "matching command server",
			serverName: "test-server",
			composeService: Service{
				Command:     "uvx mcp-server-time --debug",
				Environment: map[string]string{"DEBUG": "${DEBUG}"},
			},
			deployedServer: MCPServer{
				Command: "uvx",
				Args:    []string{"mcp-server-time", "--debug"},
				Env:     map[string]string{"DEBUG": "true"},
			},
			expectedStatus: "configured",
			expectDiffs:    false,
		},
		{
			name:       "command mismatch",
			serverName: "test-server",
			composeService: Service{
				Command: "uvx mcp-server-time",
			},
			deployedServer: MCPServer{
				Command: "python",
				Args:    []string{"-m", "mcp-server-time"},
			},
			expectedStatus: "different",
			expectDiffs:    true,
		},
		{
			name:       "args mismatch",
			serverName: "test-server",
			composeService: Service{
				Command: "uvx mcp-server-time --debug",
			},
			deployedServer: MCPServer{
				Command: "uvx",
				Args:    []string{"mcp-server-time"},
			},
			expectedStatus: "different",
			expectDiffs:    true,
		},
		{
			name:       "env vars mismatch",
			serverName: "test-server",
			composeService: Service{
				Command:     "uvx mcp-server-time",
				Environment: map[string]string{"DEBUG": "true"},
			},
			deployedServer: MCPServer{
				Command: "uvx",
				Args:    []string{"mcp-server-time"},
				Env:     map[string]string{"DEBUG": "false"},
			},
			expectedStatus: "different",
			expectDiffs:    true,
		},
		{
			name:       "matching container server",
			serverName: "container-server",
			composeService: Service{
				Image:       "my-server:latest",
				Environment: map[string]string{"API_KEY": "${API_KEY}"},
			},
			deployedServer: MCPServer{
				Command: "docker",
				Args:    []string{"run", "-i", "--rm", "-e", "API_KEY=secret123", "my-server:latest"},
				Env:     map[string]string{"API_KEY": "secret123"},
			},
			expectedStatus: "configured",
			expectDiffs:    false,
		},
		{
			name:       "container missing run args",
			serverName: "container-server",
			composeService: Service{
				Image: "my-server:latest",
			},
			deployedServer: MCPServer{
				Command: "docker",
				Args:    []string{"my-server:latest"},
			},
			expectedStatus: "different",
			expectDiffs:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, differences := compareLocalServers(tt.serverName, tt.composeService, tt.deployedServer, envVars)

			if status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, status)
			}

			if tt.expectDiffs && len(differences) == 0 {
				t.Error("Expected differences but got none")
			}

			if !tt.expectDiffs && len(differences) > 0 {
				t.Errorf("Expected no differences but got: %v", differences)
			}
		})
	}
}

func TestGetServerStatus(t *testing.T) {
	envVars := map[string]string{
		"DEBUG": "true",
	}

	composeService := Service{
		Command:     "uvx mcp-server-time",
		Environment: map[string]string{"DEBUG": "${DEBUG}"},
	}

	toolConfigs := map[string]ToolConfig{
		"kiro": {
			Config: MCPConfig{
				MCPServers: map[string]MCPServer{
					"test-server": {
						Command: "uvx",
						Args:    []string{"mcp-server-time"},
						Env:     map[string]string{"DEBUG": "true"},
					},
				},
			},
			Path:   "/home/user/.kiro/settings/mcp.json",
			Exists: true,
			Error:  "",
		},
		"cursor": {
			Config: MCPConfig{
				MCPServers: map[string]MCPServer{},
			},
			Path:   "/home/user/.cursor/mcp.json",
			Exists: true,
			Error:  "",
		},
		"broken-tool": {
			Config: MCPConfig{},
			Path:   "/invalid/path",
			Exists: false,
			Error:  "file not found",
		},
	}

	result := getServerStatus("test-server", composeService, toolConfigs, envVars)

	// Check kiro status (should be configured)
	kiroStatus, exists := result["kiro"]
	if !exists {
		t.Error("Expected kiro status to exist")
	} else {
		if kiroStatus.Status != "configured" {
			t.Errorf("Expected kiro status 'configured', got %s", kiroStatus.Status)
		}
		if kiroStatus.Tool != "kiro" {
			t.Errorf("Expected tool 'kiro', got %s", kiroStatus.Tool)
		}
	}

	// Check cursor status (should be not-configured)
	cursorStatus, exists := result["cursor"]
	if !exists {
		t.Error("Expected cursor status to exist")
	} else {
		if cursorStatus.Status != "not-configured" {
			t.Errorf("Expected cursor status 'not-configured', got %s", cursorStatus.Status)
		}
	}

	// Check broken tool status (should be not-configured since file doesn't exist)
	brokenStatus, exists := result["broken-tool"]
	if !exists {
		t.Error("Expected broken-tool status to exist")
	} else {
		if brokenStatus.Status != "not-configured" {
			t.Errorf("Expected broken-tool status 'not-configured', got %s", brokenStatus.Status)
		}
	}
}
