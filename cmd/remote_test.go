package cmd

import (
	"strings"
	"testing"
)

func TestIsRemoteServer(t *testing.T) {
	tests := []struct {
		name     string
		service  Service
		expected bool
	}{
		{
			name: "https URL",
			service: Service{
				Command: "https://api.example.com/mcp",
			},
			expected: true,
		},
		{
			name: "http URL",
			service: Service{
				Command: "http://localhost:8080/mcp",
			},
			expected: true,
		},
		{
			name: "local command",
			service: Service{
				Command: "uvx mcp-server-time",
			},
			expected: false,
		},
		{
			name: "empty command",
			service: Service{
				Command: "",
			},
			expected: false,
		},
		{
			name: "command with https in middle",
			service: Service{
				Command: "python -m server --url https://example.com",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRemoteServer(tt.service)
			if result != tt.expected {
				t.Errorf("IsRemoteServer() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestUsesHeadersAuth(t *testing.T) {
	tests := []struct {
		name     string
		service  Service
		expected bool
	}{
		{
			name: "has header labels",
			service: Service{
				Labels: map[string]string{
					"mcp.header.Authorization": "Bearer token123",
					"mcp.profile":              "default",
				},
			},
			expected: true,
		},
		{
			name: "multiple header labels",
			service: Service{
				Labels: map[string]string{
					"mcp.header.Authorization": "Bearer token123",
					"mcp.header.X-API-Key":     "key123",
				},
			},
			expected: true,
		},
		{
			name: "no header labels",
			service: Service{
				Labels: map[string]string{
					"mcp.profile": "default",
				},
			},
			expected: false,
		},
		{
			name: "empty labels",
			service: Service{
				Labels: map[string]string{},
			},
			expected: false,
		},
		{
			name: "nil labels",
			service: Service{
				Labels: nil,
			},
			expected: false,
		},
		{
			name: "has OAuth labels but no headers",
			service: Service{
				Labels: map[string]string{
					"mcp.grant-type":     "client_credentials",
					"mcp.token-endpoint": "https://auth.example.com/token",
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UsesHeadersAuth(tt.service)
			if result != tt.expected {
				t.Errorf("UsesHeadersAuth() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidateRemoteServerAuth(t *testing.T) {
	tests := []struct {
		name        string
		serverName  string
		service     Service
		expectError bool
		errorMsg    string
	}{
		{
			name:       "valid OAuth config",
			serverName: "test-server",
			service: Service{
				Labels: map[string]string{
					"mcp.grant-type":     "client_credentials",
					"mcp.token-endpoint": "https://auth.example.com/token",
					"mcp.client-id":      "client123",
					"mcp.client-secret":  "secret123",
				},
			},
			expectError: false,
		},
		{
			name:       "valid headers config",
			serverName: "test-server",
			service: Service{
				Labels: map[string]string{
					"mcp.header.Authorization": "Bearer token123",
				},
			},
			expectError: false,
		},
		{
			name:       "no auth config",
			serverName: "test-server",
			service: Service{
				Labels: map[string]string{
					"mcp.profile": "default",
				},
			},
			expectError: true,
			errorMsg:    "must have either OAuth labels",
		},
		{
			name:       "both OAuth and headers",
			serverName: "test-server",
			service: Service{
				Labels: map[string]string{
					"mcp.grant-type":           "client_credentials",
					"mcp.header.Authorization": "Bearer token123",
				},
			},
			expectError: true,
			errorMsg:    "cannot have both OAuth labels and headers labels",
		},
		{
			name:       "incomplete OAuth config - missing client-id",
			serverName: "test-server",
			service: Service{
				Labels: map[string]string{
					"mcp.grant-type":     "client_credentials",
					"mcp.token-endpoint": "https://auth.example.com/token",
					"mcp.client-secret":  "secret123",
				},
			},
			expectError: true,
			errorMsg:    "missing required OAuth labels: mcp.client-id",
		},
		{
			name:       "invalid grant type",
			serverName: "test-server",
			service: Service{
				Labels: map[string]string{
					"mcp.grant-type":     "authorization_code",
					"mcp.token-endpoint": "https://auth.example.com/token",
					"mcp.client-id":      "client123",
					"mcp.client-secret":  "secret123",
				},
			},
			expectError: true,
			errorMsg:    "must use 'client_credentials' grant type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRemoteServerAuth(tt.serverName, tt.service)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestValidateRemoteServerOAuth(t *testing.T) {
	// This is a deprecated function that should just call ValidateRemoteServerAuth
	service := Service{
		Labels: map[string]string{
			"mcp.grant-type":     "client_credentials",
			"mcp.token-endpoint": "https://auth.example.com/token",
			"mcp.client-id":      "client123",
			"mcp.client-secret":  "secret123",
		},
	}

	err := ValidateRemoteServerOAuth("test", service)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestExtractHeaders(t *testing.T) {
	envVars := map[string]string{
		"API_TOKEN": "secret123",
		"USER_ID":   "user456",
	}

	tests := []struct {
		name        string
		service     Service
		expectError bool
		expected    map[string]string
	}{
		{
			name: "simple header",
			service: Service{
				Labels: map[string]string{
					"mcp.header.Authorization": "Bearer token123",
				},
			},
			expectError: false,
			expected: map[string]string{
				"Authorization": "Bearer token123",
			},
		},
		{
			name: "multiple headers",
			service: Service{
				Labels: map[string]string{
					"mcp.header.Authorization": "Bearer token123",
					"mcp.header.X-API-Key":     "key456",
				},
			},
			expectError: false,
			expected: map[string]string{
				"Authorization": "Bearer token123",
				"X-API-Key":     "key456",
			},
		},
		{
			name: "headers with env vars",
			service: Service{
				Labels: map[string]string{
					"mcp.header.Authorization": "Bearer ${API_TOKEN}",
					"mcp.header.X-User-ID":     "$USER_ID",
				},
			},
			expectError: false,
			expected: map[string]string{
				"Authorization": "Bearer secret123",
				"X-User-ID":     "user456",
			},
		},
		{
			name: "empty header name",
			service: Service{
				Labels: map[string]string{
					"mcp.header.": "value",
				},
			},
			expectError: false,
			expected:    map[string]string{},
		},
		{
			name: "no headers",
			service: Service{
				Labels: map[string]string{
					"mcp.profile": "default",
				},
			},
			expectError: true,
		},
		{
			name: "unresolved env var",
			service: Service{
				Labels: map[string]string{
					"mcp.header.Authorization": "Bearer ${UNDEFINED_VAR}",
				},
			},
			expectError: true,
		},
		{
			name: "empty placeholder header",
			service: Service{
				Labels: map[string]string{
					"mcp.header.X-Empty": "",
				},
			},
			expectError: false,
			expected:    map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractHeaders(tt.service, envVars)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}

				if len(result) != len(tt.expected) {
					t.Errorf("Expected %d headers, got %d", len(tt.expected), len(result))
				}

				for key, expectedValue := range tt.expected {
					if actualValue, exists := result[key]; !exists {
						t.Errorf("Expected header %q not found", key)
					} else if actualValue != expectedValue {
						t.Errorf("Header %q: expected %q, got %q", key, expectedValue, actualValue)
					}
				}
			}
		})
	}
}

func TestValidateToolSupport(t *testing.T) {
	localServers := map[string]Service{
		"local": {
			Command: "uvx mcp-server-time",
		},
	}

	remoteServers := map[string]Service{
		"remote": {
			Command: "https://api.example.com/mcp",
		},
	}

	mixedServers := map[string]Service{
		"local": {
			Command: "uvx mcp-server-time",
		},
		"remote": {
			Command: "https://api.example.com/mcp",
		},
	}

	tests := []struct {
		name         string
		toolShortcut string
		servers      map[string]Service
		expectError  bool
	}{
		{
			name:         "empty tool with local servers",
			toolShortcut: "",
			servers:      localServers,
			expectError:  false,
		},
		{
			name:         "empty tool with remote servers",
			toolShortcut: "",
			servers:      remoteServers,
			expectError:  false,
		},
		{
			name:         "supported tool with remote servers",
			toolShortcut: "kiro",
			servers:      remoteServers,
			expectError:  false,
		},
		{
			name:         "unsupported tool with remote servers",
			toolShortcut: "claude-desktop",
			servers:      remoteServers,
			expectError:  true,
		},
		{
			name:         "supported tool with mixed servers",
			toolShortcut: "q-cli",
			servers:      mixedServers,
			expectError:  false,
		},
		{
			name:         "unsupported tool with mixed servers",
			toolShortcut: "claude-desktop",
			servers:      mixedServers,
			expectError:  true,
		},
		{
			name:         "any tool with only local servers",
			toolShortcut: "claude-desktop",
			servers:      localServers,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateToolSupport(tt.toolShortcut, tt.servers)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestExtractOAuthConfig(t *testing.T) {
	envVars := map[string]string{
		"CLIENT_ID":     "expanded_client_id",
		"CLIENT_SECRET": "expanded_secret",
		"TOKEN_URL":     "https://expanded.example.com/token",
	}

	tests := []struct {
		name        string
		service     Service
		expectError bool
		expected    OAuthConfig
	}{
		{
			name: "complete OAuth config",
			service: Service{
				Labels: map[string]string{
					"mcp.grant-type":     "client_credentials",
					"mcp.token-endpoint": "https://auth.example.com/token",
					"mcp.client-id":      "client123",
					"mcp.client-secret":  "secret456",
				},
			},
			expectError: false,
			expected: OAuthConfig{
				GrantType:    "client_credentials",
				TokenURL:     "https://auth.example.com/token",
				ClientID:     "client123",
				ClientSecret: "secret456",
			},
		},
		{
			name: "OAuth config with env vars",
			service: Service{
				Labels: map[string]string{
					"mcp.grant-type":     "client_credentials",
					"mcp.token-endpoint": "${TOKEN_URL}",
					"mcp.client-id":      "${CLIENT_ID}",
					"mcp.client-secret":  "$CLIENT_SECRET",
				},
			},
			expectError: false,
			expected: OAuthConfig{
				GrantType:    "client_credentials",
				TokenURL:     "https://expanded.example.com/token",
				ClientID:     "expanded_client_id",
				ClientSecret: "expanded_secret",
			},
		},
		{
			name: "unresolved client ID",
			service: Service{
				Labels: map[string]string{
					"mcp.grant-type":     "client_credentials",
					"mcp.token-endpoint": "https://auth.example.com/token",
					"mcp.client-id":      "${UNDEFINED_CLIENT_ID}",
					"mcp.client-secret":  "secret456",
				},
			},
			expectError: true,
		},
		{
			name: "unresolved client secret",
			service: Service{
				Labels: map[string]string{
					"mcp.grant-type":     "client_credentials",
					"mcp.token-endpoint": "https://auth.example.com/token",
					"mcp.client-id":      "client123",
					"mcp.client-secret":  "${UNDEFINED_SECRET}",
				},
			},
			expectError: true,
		},
		{
			name: "unresolved token URL",
			service: Service{
				Labels: map[string]string{
					"mcp.grant-type":     "client_credentials",
					"mcp.token-endpoint": "${UNDEFINED_URL}",
					"mcp.client-id":      "client123",
					"mcp.client-secret":  "secret456",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractOAuthConfig(tt.service, envVars)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}

				if result.GrantType != tt.expected.GrantType {
					t.Errorf("GrantType: expected %q, got %q", tt.expected.GrantType, result.GrantType)
				}
				if result.TokenURL != tt.expected.TokenURL {
					t.Errorf("TokenURL: expected %q, got %q", tt.expected.TokenURL, result.TokenURL)
				}
				if result.ClientID != tt.expected.ClientID {
					t.Errorf("ClientID: expected %q, got %q", tt.expected.ClientID, result.ClientID)
				}
				if result.ClientSecret != tt.expected.ClientSecret {
					t.Errorf("ClientSecret: expected %q, got %q", tt.expected.ClientSecret, result.ClientSecret)
				}
			}
		})
	}
}
