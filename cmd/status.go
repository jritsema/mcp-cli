package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// loadToolConfig reads the MCP config file for a given tool shortcut
// Returns parsed MCPConfig or error if file doesn't exist
// Handles missing files gracefully (returns empty config)
func loadToolConfig(toolShortcut string) (MCPConfig, string, error) {
	path := getPlatformToolPath(toolShortcut)
	if path == "" {
		return MCPConfig{}, "", fmt.Errorf("unknown tool shortcut: %s", toolShortcut)
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return MCPConfig{}, path, nil // Return empty config for missing file
	}

	// Read and parse the config file
	data, err := os.ReadFile(path)
	if err != nil {
		return MCPConfig{}, path, fmt.Errorf("error reading config file: %w", err)
	}

	var config MCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return MCPConfig{}, path, fmt.Errorf("error parsing config file: %w", err)
	}

	return config, path, nil
}

// getToolConfigs loads MCP configs for all specified tools
// Returns map of tool -> MCPConfig
// Handles missing files gracefully
func getToolConfigs(tools []string) map[string]ToolConfig {
	result := make(map[string]ToolConfig)

	for _, tool := range tools {
		config, path, err := loadToolConfig(tool)
		if err != nil {
			result[tool] = ToolConfig{
				Config: MCPConfig{},
				Path:   path,
				Exists: false,
				Error:  err.Error(),
			}
			continue
		}

		exists := path != "" && fileExists(path)
		result[tool] = ToolConfig{
			Config: config,
			Path:   path,
			Exists: exists,
			Error:  "",
		}
	}

	return result
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// compareServerConfig compares a service from compose file with deployed server config
// Returns status: "configured", "not-configured", "different", "unknown"
// Returns list of differences (command mismatch, missing env vars, etc.)
// Handles both local and remote servers
func compareServerConfig(serverName string, composeService Service, deployedServer MCPServer, envVars map[string]string) (string, []string) {
	// If deployed server doesn't exist (empty struct), it's not configured
	if deployedServer.Command == "" && deployedServer.URL == "" {
		return "not-configured", nil
	}

	// Check if this is a remote server
	if IsRemoteServer(composeService) {
		status, differences := compareRemoteServers(composeService, deployedServer, envVars)
		return status, differences
	}

	// Compare local servers
	return compareLocalServers(serverName, composeService, deployedServer, envVars)
}

// compareRemoteServers compares remote server configs
// Checks URL, headers, type
// Returns match status and differences
func compareRemoteServers(composeService Service, deployedServer MCPServer, envVars map[string]string) (string, []string) {
	var differences []string

	// Check type
	if deployedServer.Type != "http" {
		differences = append(differences, fmt.Sprintf("type mismatch: expected 'http', got '%s'", deployedServer.Type))
	}

	// Check URL
	expectedURL := composeService.Command
	if deployedServer.URL != expectedURL {
		differences = append(differences, fmt.Sprintf("URL mismatch: expected '%s', got '%s'", expectedURL, deployedServer.URL))
	}

	// Check headers (if using headers auth)
	if UsesHeadersAuth(composeService) {
		// Merge service environment variables for expansion
		serviceEnvVars := make(map[string]string)
		// Copy envVars first
		for k, v := range envVars {
			serviceEnvVars[k] = v
		}
		// Add service-specific environment variables
		for key, value := range composeService.Environment {
			expandedValue := expandEnvVars(value, envVars)
			serviceEnvVars[key] = expandedValue
		}

		expectedHeaders, err := ExtractHeaders(composeService, serviceEnvVars)
		if err == nil {
			if !compareHeaders(expectedHeaders, deployedServer.Headers) {
				differences = append(differences, "headers mismatch")
			}
		}
	} else {
		// For OAuth, we can't easily compare tokens, so we check if headers exist
		// OAuth tokens are acquired at deployment time, so we just check if Authorization header exists
		if deployedServer.Headers == nil || len(deployedServer.Headers) == 0 {
			differences = append(differences, "missing OAuth headers")
		} else if authHeader, exists := deployedServer.Headers["Authorization"]; !exists || !strings.HasPrefix(authHeader, "Bearer ") {
			differences = append(differences, "invalid OAuth headers")
		}
	}

	if len(differences) > 0 {
		return "different", differences
	}

	return "configured", nil
}

// compareLocalServers compares local server configs
// Checks command, args, env vars
// Handles container vs command differences
func compareLocalServers(serverName string, composeService Service, deployedServer MCPServer, envVars map[string]string) (string, []string) {
	var differences []string

	// Get container tool from config
	containerTool := "docker"
	configDir := getConfigDir()
	configPath := filepath.Join(configDir, "config.json")
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err == nil {
			var config CLIConfig
			if err := json.Unmarshal(data, &config); err == nil && config.ContainerTool != "" {
				containerTool = config.ContainerTool
			}
		}
	}

	// Handle container-based servers
	if composeService.Image != "" {
		expectedCommand := containerTool
		if deployedServer.Command != expectedCommand {
			differences = append(differences, fmt.Sprintf("command mismatch: expected '%s', got '%s'", expectedCommand, deployedServer.Command))
		}

		// Check args - should start with "run", "-i", "--rm"
		expectedArgsPrefix := []string{"run", "-i", "--rm"}
		if len(deployedServer.Args) < len(expectedArgsPrefix) {
			differences = append(differences, "missing container run arguments")
		} else {
			for i, expectedArg := range expectedArgsPrefix {
				if deployedServer.Args[i] != expectedArg {
					differences = append(differences, fmt.Sprintf("arg mismatch at position %d: expected '%s', got '%s'", i, expectedArg, deployedServer.Args[i]))
				}
			}
		}

		// Check environment variables
		expectedEnv := make(map[string]string)
		for key, value := range composeService.Environment {
			expectedEnv[key] = expandEnvVars(value, envVars)
		}
		if !compareEnvVars(expectedEnv, deployedServer.Env) {
			differences = append(differences, "environment variables mismatch")
		}

		// Check image name (should be last arg)
		expandedImage := expandEnvVars(composeService.Image, envVars)
		if len(deployedServer.Args) > 0 {
			lastArg := deployedServer.Args[len(deployedServer.Args)-1]
			if lastArg != expandedImage {
				differences = append(differences, fmt.Sprintf("image mismatch: expected '%s', got '%s'", expandedImage, lastArg))
			}
		}
	} else {
		// Command-based server
		parts := strings.Fields(composeService.Command)
		if len(parts) > 0 {
			expectedCommand := parts[0]
			if deployedServer.Command != expectedCommand {
				differences = append(differences, fmt.Sprintf("command mismatch: expected '%s', got '%s'", expectedCommand, deployedServer.Command))
			}

			// Check args
			expectedArgs := make([]string, 0)
			if len(parts) > 1 {
				for _, arg := range parts[1:] {
					expectedArgs = append(expectedArgs, expandEnvVars(arg, envVars))
				}
			}

			if !compareStringSlices(expectedArgs, deployedServer.Args) {
				differences = append(differences, "arguments mismatch")
			}
		}

		// Check environment variables
		expectedEnv := make(map[string]string)
		for key, value := range composeService.Environment {
			expectedEnv[key] = expandEnvVars(value, envVars)
		}
		if !compareEnvVars(expectedEnv, deployedServer.Env) {
			differences = append(differences, "environment variables mismatch")
		}
	}

	if len(differences) > 0 {
		return "different", differences
	}

	return "configured", nil
}

// compareHeaders compares two header maps
func compareHeaders(expected, actual map[string]string) bool {
	if len(expected) != len(actual) {
		return false
	}

	for key, expectedValue := range expected {
		actualValue, exists := actual[key]
		if !exists || actualValue != expectedValue {
			return false
		}
	}

	return true
}

// compareEnvVars compares two environment variable maps
func compareEnvVars(expected, actual map[string]string) bool {
	if len(expected) != len(actual) {
		return false
	}

	for key, expectedValue := range expected {
		actualValue, exists := actual[key]
		if !exists || actualValue != expectedValue {
			return false
		}
	}

	return true
}

// compareStringSlices compares two string slices
func compareStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// getServerStatus gets the status of a server across all tools
// Returns map of tool -> ServerStatus
func getServerStatus(serverName string, composeService Service, toolConfigs map[string]ToolConfig, envVars map[string]string) map[string]ServerStatus {
	result := make(map[string]ServerStatus)

	for tool, toolConfig := range toolConfigs {
		if !toolConfig.Exists {
			result[tool] = ServerStatus{
				Status:     "not-configured",
				Tool:       tool,
				ConfigPath: toolConfig.Path,
			}
			continue
		}

		if toolConfig.Error != "" {
			result[tool] = ServerStatus{
				Status:     "unknown",
				Tool:       tool,
				ConfigPath: toolConfig.Path,
				Error:      toolConfig.Error,
			}
			continue
		}

		// Find the server in the deployed config
		deployedServer, exists := toolConfig.Config.MCPServers[serverName]
		if !exists {
			result[tool] = ServerStatus{
				Status:     "not-configured",
				Tool:       tool,
				ConfigPath: toolConfig.Path,
			}
			continue
		}

		// Compare the server configs
		status, differences := compareServerConfig(serverName, composeService, deployedServer, envVars)
		result[tool] = ServerStatus{
			Status:      status,
			Tool:        tool,
			Differences: differences,
			ConfigPath:  toolConfig.Path,
		}
	}

	return result
}

// normalizeToolName normalizes tool names for display
func normalizeToolName(tool string) string {
	switch tool {
	case "q-cli":
		return "Q-CLI"
	case "claude-desktop":
		return "CLAUDE"
	case "cursor":
		return "CURSOR"
	case "kiro":
		return "KIRO"
	default:
		return strings.ToUpper(tool)
	}
}
