package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGetDescription tests the GetDescription function
func TestGetDescription(t *testing.T) {
	tests := []struct {
		name     string
		service  Service
		expected string
	}{
		{
			name: "service with description label",
			service: Service{
				Command: "uvx mcp-server-time",
				Labels: map[string]string{
					"mcp.profile":     "default",
					"mcp.description": "A time server for MCP",
				},
			},
			expected: "A time server for MCP",
		},
		{
			name: "service without description label",
			service: Service{
				Command: "uvx mcp-server-time",
				Labels: map[string]string{
					"mcp.profile": "default",
				},
			},
			expected: "",
		},
		{
			name: "service with nil labels",
			service: Service{
				Command: "uvx mcp-server-time",
				Labels:  nil,
			},
			expected: "",
		},
		{
			name: "service with empty labels map",
			service: Service{
				Command: "uvx mcp-server-time",
				Labels:  map[string]string{},
			},
			expected: "",
		},
		{
			name: "service with empty description",
			service: Service{
				Command: "uvx mcp-server-time",
				Labels: map[string]string{
					"mcp.description": "",
				},
			},
			expected: "",
		},
		{
			name: "service with description containing special characters",
			service: Service{
				Command: "uvx mcp-server-time",
				Labels: map[string]string{
					"mcp.description": "Server with special chars: @#$%^&*()!",
				},
			},
			expected: "Server with special chars: @#$%^&*()!",
		},
		{
			name: "service with description containing spaces and punctuation",
			service: Service{
				Command: "uvx mcp-server-time",
				Labels: map[string]string{
					"mcp.description": "This is a server, with punctuation. And spaces!",
				},
			},
			expected: "This is a server, with punctuation. And spaces!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDescription(tt.service)
			if result != tt.expected {
				t.Errorf("GetDescription() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestTruncateDescription tests the TruncateDescription function
func TestTruncateDescription(t *testing.T) {
	tests := []struct {
		name     string
		desc     string
		maxLen   int
		expected string
	}{
		{
			name:     "empty string",
			desc:     "",
			maxLen:   60,
			expected: "",
		},
		{
			name:     "string shorter than max length",
			desc:     "A short description",
			maxLen:   60,
			expected: "A short description",
		},
		{
			name:     "string exactly at max length",
			desc:     "This is exactly sixty characters long, no more and no less!!", // 60 chars
			maxLen:   60,
			expected: "This is exactly sixty characters long, no more and no less!!",
		},
		{
			name:     "string one character over max length",
			desc:     "This is exactly sixty-one characters long, no more no less!!!", // 61 chars
			maxLen:   60,
			expected: "This is exactly sixty-one characters long, no more no les...",
		},
		{
			name:     "long string requiring truncation",
			desc:     "This is a very long description that exceeds the maximum allowed length and should be truncated with ellipsis",
			maxLen:   60,
			expected: "This is a very long description that exceeds the maximum ...",
		},
		{
			name:     "string with special characters truncated",
			desc:     "Server with special chars: @#$%^&*()! and more text that makes it too long to display",
			maxLen:   60,
			expected: "Server with special chars: @#$%^&*()! and more text that ...",
		},
		{
			name:     "using MaxDescriptionLength constant",
			desc:     "This description is longer than sixty characters and will be truncated at the default max",
			maxLen:   MaxDescriptionLength,
			expected: "This description is longer than sixty characters and will...",
		},
		{
			name:     "custom max length shorter than default",
			desc:     "A medium length description",
			maxLen:   20,
			expected: "A medium length d...",
		},
		{
			name:     "string exactly at custom max length",
			desc:     "Exactly twenty chars",
			maxLen:   20,
			expected: "Exactly twenty chars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateDescription(tt.desc, tt.maxLen)
			if result != tt.expected {
				t.Errorf("TruncateDescription(%q, %d) = %q, want %q", tt.desc, tt.maxLen, result, tt.expected)
			}
			// Verify truncated strings don't exceed maxLen
			if len(result) > tt.maxLen {
				t.Errorf("TruncateDescription(%q, %d) returned string of length %d, exceeds max %d", tt.desc, tt.maxLen, len(result), tt.maxLen)
			}
			// Verify truncated strings end with "..." when truncation occurred
			if len(tt.desc) > tt.maxLen && len(result) >= 3 {
				if result[len(result)-3:] != "..." {
					t.Errorf("TruncateDescription(%q, %d) = %q, expected to end with '...'", tt.desc, tt.maxLen, result)
				}
			}
		})
	}
}

// TestMaxDescriptionLengthConstant verifies the constant value
func TestMaxDescriptionLengthConstant(t *testing.T) {
	if MaxDescriptionLength != 60 {
		t.Errorf("MaxDescriptionLength = %d, want 60", MaxDescriptionLength)
	}
}

// TestLoadComposeFile tests the loadComposeFile function
func TestLoadComposeFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mcp-compose-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("valid compose file", func(t *testing.T) {
		composeContent := `
services:
  time-server:
    command: uvx mcp-server-time
    environment:
      DEBUG: "true"
    labels:
      mcp.profile: "default"
      mcp.description: "A time server for MCP"
  
  container-server:
    image: my-server:latest
    environment:
      API_KEY: "secret"
    labels:
      mcp.profile: "programming"
`
		composePath := filepath.Join(tempDir, "valid-compose.yml")
		if err := os.WriteFile(composePath, []byte(composeContent), 0644); err != nil {
			t.Fatalf("Failed to create compose file: %v", err)
		}

		config, err := loadComposeFile(composePath)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if len(config.Services) != 2 {
			t.Errorf("Expected 2 services, got %d", len(config.Services))
		}

		timeServer, exists := config.Services["time-server"]
		if !exists {
			t.Error("Expected time-server to exist")
		}

		if timeServer.Command != "uvx mcp-server-time" {
			t.Errorf("Expected command 'uvx mcp-server-time', got %s", timeServer.Command)
		}

		if timeServer.Environment["DEBUG"] != "true" {
			t.Errorf("Expected DEBUG=true, got %s", timeServer.Environment["DEBUG"])
		}

		if timeServer.Labels["mcp.profile"] != "default" {
			t.Errorf("Expected profile 'default', got %s", timeServer.Labels["mcp.profile"])
		}

		containerServer, exists := config.Services["container-server"]
		if !exists {
			t.Error("Expected container-server to exist")
		}

		if containerServer.Image != "my-server:latest" {
			t.Errorf("Expected image 'my-server:latest', got %s", containerServer.Image)
		}
	})

	t.Run("file does not exist", func(t *testing.T) {
		nonExistentPath := filepath.Join(tempDir, "does-not-exist.yml")

		_, err := loadComposeFile(nonExistentPath)
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("invalid YAML", func(t *testing.T) {
		invalidContent := `
services:
  invalid-yaml:
    command: uvx server
    invalid: yaml: content: [
`
		invalidPath := filepath.Join(tempDir, "invalid.yml")
		if err := os.WriteFile(invalidPath, []byte(invalidContent), 0644); err != nil {
			t.Fatalf("Failed to create invalid file: %v", err)
		}

		_, err := loadComposeFile(invalidPath)
		if err == nil {
			t.Error("Expected error for invalid YAML")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		emptyPath := filepath.Join(tempDir, "empty.yml")
		if err := os.WriteFile(emptyPath, []byte(""), 0644); err != nil {
			t.Fatalf("Failed to create empty file: %v", err)
		}

		config, err := loadComposeFile(emptyPath)
		if err != nil {
			t.Errorf("Expected no error for empty file, got %v", err)
		}

		// Empty YAML file should result in zero-value struct, Services will be nil
		if config.Services != nil && len(config.Services) != 0 {
			t.Error("Expected Services to be nil or empty for empty file")
		}
	})
}

// TestFilterServers tests the filterServers function
func TestFilterServers(t *testing.T) {
	config := &ComposeConfig{
		Services: map[string]Service{
			"default-server": {
				Command: "uvx default-server",
				Labels: map[string]string{
					"mcp.profile": "default",
				},
			},
			"no-profile-server": {
				Command: "uvx no-profile-server",
				Labels:  map[string]string{},
			},
			"programming-server": {
				Command: "uvx programming-server",
				Labels: map[string]string{
					"mcp.profile": "programming",
				},
			},
			"multi-profile-server": {
				Command: "uvx multi-profile-server",
				Labels: map[string]string{
					"mcp.profile": "default,programming,research",
				},
			},
			"research-server": {
				Command: "uvx research-server",
				Labels: map[string]string{
					"mcp.profile": "research",
				},
			},
		},
	}

	t.Run("filter by default profile", func(t *testing.T) {
		result := filterServers(config, "", false)

		// Should include: default-server, no-profile-server, multi-profile-server
		expectedServers := []string{"default-server", "no-profile-server", "multi-profile-server"}

		if len(result) != len(expectedServers) {
			t.Errorf("Expected %d servers, got %d", len(expectedServers), len(result))
		}

		for _, serverName := range expectedServers {
			if _, exists := result[serverName]; !exists {
				t.Errorf("Expected server %s to be included", serverName)
			}
		}
	})

	t.Run("filter by programming profile", func(t *testing.T) {
		result := filterServers(config, "programming", false)

		// Should include: default-server, no-profile-server, multi-profile-server, programming-server
		expectedServers := []string{"default-server", "no-profile-server", "multi-profile-server", "programming-server"}

		if len(result) != len(expectedServers) {
			t.Errorf("Expected %d servers, got %d", len(expectedServers), len(result))
		}

		for _, serverName := range expectedServers {
			if _, exists := result[serverName]; !exists {
				t.Errorf("Expected server %s to be included", serverName)
			}
		}
	})

	t.Run("filter by research profile", func(t *testing.T) {
		result := filterServers(config, "research", false)

		// Should include: default-server, no-profile-server, multi-profile-server, research-server
		expectedServers := []string{"default-server", "no-profile-server", "multi-profile-server", "research-server"}

		if len(result) != len(expectedServers) {
			t.Errorf("Expected %d servers, got %d", len(expectedServers), len(result))
		}

		for _, serverName := range expectedServers {
			if _, exists := result[serverName]; !exists {
				t.Errorf("Expected server %s to be included", serverName)
			}
		}
	})

	t.Run("filter by non-existent profile", func(t *testing.T) {
		result := filterServers(config, "non-existent", false)

		// Should include only default servers: default-server, no-profile-server, multi-profile-server
		expectedServers := []string{"default-server", "no-profile-server", "multi-profile-server"}

		if len(result) != len(expectedServers) {
			t.Errorf("Expected %d servers, got %d", len(expectedServers), len(result))
		}

		for _, serverName := range expectedServers {
			if _, exists := result[serverName]; !exists {
				t.Errorf("Expected server %s to be included", serverName)
			}
		}
	})

	t.Run("show all servers", func(t *testing.T) {
		result := filterServers(config, "", true)

		// Should include all servers
		if len(result) != len(config.Services) {
			t.Errorf("Expected %d servers, got %d", len(config.Services), len(result))
		}

		for serverName := range config.Services {
			if _, exists := result[serverName]; !exists {
				t.Errorf("Expected server %s to be included", serverName)
			}
		}
	})

	t.Run("show all servers with specific profile", func(t *testing.T) {
		result := filterServers(config, "programming", true)

		// Should include all servers regardless of profile when all=true
		if len(result) != len(config.Services) {
			t.Errorf("Expected %d servers, got %d", len(config.Services), len(result))
		}
	})

	t.Run("empty config", func(t *testing.T) {
		emptyConfig := &ComposeConfig{
			Services: map[string]Service{},
		}

		result := filterServers(emptyConfig, "", false)

		if len(result) != 0 {
			t.Errorf("Expected 0 servers, got %d", len(result))
		}
	})
}
