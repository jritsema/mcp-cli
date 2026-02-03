package cmd

import (
	"testing"
)

// TestRemoteServerWithEnvExpansion tests the new environment-aware remote server detection
func TestRemoteServerWithEnvExpansion(t *testing.T) {
	envVars := map[string]string{
		"REMOTE_URL": "https://api.example.com/mcp",
		"LOCAL_CMD":  "uvx mcp-server-time",
	}

	testCases := []struct {
		name           string
		service        Service
		shouldBeRemote bool
	}{
		{
			name: "direct_https_url",
			service: Service{
				Command: "https://api.example.com/mcp",
			},
			shouldBeRemote: true,
		},
		{
			name: "env_var_with_https_url",
			service: Service{
				Command: "${REMOTE_URL}",
			},
			shouldBeRemote: true,
		},
		{
			name: "env_var_with_local_command",
			service: Service{
				Command: "${LOCAL_CMD}",
			},
			shouldBeRemote: false,
		},
		{
			name: "local_command",
			service: Service{
				Command: "uvx mcp-server-time",
			},
			shouldBeRemote: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test original function
			originalResult := IsRemoteServer(tc.service)

			// Test new environment-aware function
			newResult := IsRemoteServerWithEnvExpansion(tc.service, envVars)

			if tc.shouldBeRemote {
				if !newResult {
					t.Errorf("IsRemoteServerWithEnvExpansion should detect service as remote")
				}
				// For direct URLs, original function should also work
				if tc.service.Command == "https://api.example.com/mcp" {
					if !originalResult {
						t.Errorf("IsRemoteServer should detect direct HTTPS URL as remote")
					}
				}
			} else {
				if newResult {
					t.Errorf("IsRemoteServerWithEnvExpansion should not detect service as remote")
				}
				if originalResult {
					t.Errorf("IsRemoteServer should not detect service as remote")
				}
			}
		})
	}
}

// TestValidateToolSupportWithEnvExpansion tests the new environment-aware tool validation
func TestValidateToolSupportWithEnvExpansion(t *testing.T) {
	envVars := map[string]string{
		"REMOTE_URL": "https://api.example.com/mcp",
	}

	servers := map[string]Service{
		"remote-server": {
			Command: "${REMOTE_URL}",
		},
		"local-server": {
			Command: "uvx mcp-server-time",
		},
	}

	// Test with supported tool (cursor)
	err := ValidateToolSupportWithEnvExpansion("cursor", servers, envVars)
	if err != nil {
		t.Errorf("cursor should support remote servers, got error: %v", err)
	}

	// Test with unsupported tool (claude-desktop)
	err = ValidateToolSupportWithEnvExpansion("claude-desktop", servers, envVars)
	if err == nil {
		t.Error("claude-desktop should not support remote servers, but got no error")
	}

	// Test with only local servers (should pass for any tool)
	localServers := map[string]Service{
		"local-server": {
			Command: "uvx mcp-server-time",
		},
	}
	err = ValidateToolSupportWithEnvExpansion("claude-desktop", localServers, envVars)
	if err != nil {
		t.Errorf("claude-desktop should work with local servers, got error: %v", err)
	}
}
