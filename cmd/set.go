package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	configFile   string
	toolShortcut string
	singleServer string
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set [profile]",
	Short: "Set MCP configuration",
	Long: `Set MCP configuration by writing an MCP JSON file using servers from the specified profile.
If no profile is specified, it uses default servers.`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := loadComposeFile(composeFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading compose file: %v\n", err)
			os.Exit(1)
		}

		// Load environment variables
		envVars, err := loadEnvVars(composeFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading environment variables: %v\n", err)
			os.Exit(1)
		}

		var profile string
		if len(args) > 0 {
			profile = args[0]
		}

		// Determine the output file path
		outputPath, err := getOutputPath(envVars)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error determining output path: %v\n", err)
			os.Exit(1)
		}

		// Filter servers based on profile
		servers := filterServers(config, profile, false)

		// If single server is specified, filter to just that server
		if singleServer != "" {
			if service, exists := servers[singleServer]; exists {
				servers = map[string]Service{singleServer: service}
			} else {
				fmt.Fprintf(os.Stderr, "Server '%s' not found\n", singleServer)
				os.Exit(1)
			}
		}

		// Validate remote servers have required auth configuration (OAuth or headers)
		for name, service := range servers {
			if IsRemoteServerWithEnvExpansion(service, envVars) {
				if err := ValidateRemoteServerAuth(name, service); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
			}
		}

		// Validate tool compatibility with remote servers
		if err := ValidateToolSupportWithEnvExpansion(toolShortcut, servers, envVars); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Convert to MCP JSON format
		mcpConfig := convertToMCPConfig(servers, envVars)

		// Write to file
		if err := writeMCPConfig(mcpConfig, outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing MCP config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Wrote %s\n", outputPath)
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
	setCmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to write the MCP JSON configuration file")
	setCmd.Flags().StringVarP(&toolShortcut, "tool", "t", "", "Tool shortcut (q-cli, claude-desktop, cursor, kiro)")
	setCmd.Flags().StringVarP(&singleServer, "server", "s", "", "Specify a single server to include")
}

func getOutputPath(envVars map[string]string) (string, error) {
	if configFile != "" {
		return expandEnvVars(configFile, envVars), nil
	}

	if toolShortcut != "" {
		path := getPlatformToolPath(toolShortcut)
		if path == "" {
			return "", fmt.Errorf("unknown tool shortcut: %s", toolShortcut)
		}

		// Create directory if it doesn't exist
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", err
		}

		return path, nil
	}

	// Check if there's a default tool configured in the config file
	configDir := getConfigDir()
	configPath := filepath.Join(configDir, "config.json")

	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err == nil {
			var config CLIConfig
			if err := json.Unmarshal(data, &config); err == nil && config.Tool != "" {
				// Create directory if it doesn't exist
				dir := filepath.Dir(config.Tool)
				if err := os.MkdirAll(dir, 0755); err != nil {
					return "", err
				}
				return config.Tool, nil
			}
		}
	}

	return "", fmt.Errorf("either --config or --tool must be specified, or set a default tool with 'mcp config set tool <path>'")
}

func convertToMCPConfig(servers map[string]Service, envVars map[string]string) MCPConfig {
	mcpServers := make(map[string]MCPServer)

	// Get the container tool from config, default to "docker"
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

	for name, service := range servers {
		var mcpServer MCPServer

		if IsRemoteServerWithEnvExpansion(service, envVars) {
			// Remote server - use HTTP-based configuration
			mcpServer.Type = "http"
			mcpServer.URL = expandEnvVars(service.Command, envVars)

			// Merge service environment variables into envVars for expansion
			serviceEnvVars := make(map[string]string)
			for k, v := range envVars {
				serviceEnvVars[k] = v
			}
			// Add service-specific environment variables (with expansion)
			for key, value := range service.Environment {
				expandedValue := expandEnvVars(value, envVars)
				serviceEnvVars[key] = expandedValue
			}

			if UsesHeadersAuth(service) {
				// Headers-based authentication
				headers, err := ExtractHeaders(service, serviceEnvVars)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error extracting headers for '%s': %v\n", name, err)
					os.Exit(1)
				}
				// Always set headers, even if empty (for servers with no auth)
				if headers == nil {
					headers = make(map[string]string)
				}
				mcpServer.Headers = headers
			} else {
				// OAuth-based authentication
				oauthConfig, err := ExtractOAuthConfig(service, serviceEnvVars)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error extracting OAuth config for '%s': %v\n", name, err)
					os.Exit(1)
				}

				accessToken, err := AcquireAccessTokenWithFeedback(name, oauthConfig)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to acquire access token for '%s': %v\n", name, err)
					os.Exit(1)
				}

				// Set Authorization header with Bearer token
				mcpServer.Headers = map[string]string{
					"Authorization": fmt.Sprintf("Bearer %s", accessToken),
				}
			}
		} else if service.Image != "" {
			// Container-based server
			mcpServer.Command = containerTool
			args := []string{"run", "-i", "--rm"}

			// Add environment variables with expanded values
			for key, value := range service.Environment {
				expandedValue := expandEnvVars(value, envVars)
				args = append(args, "-e", fmt.Sprintf("%s=%s", key, expandedValue))
			}

			// Add volume mounts with expanded values
			for _, volume := range service.Volumes {
				expandedVolume := expandEnvVars(volume, envVars)
				args = append(args, "-v", expandedVolume)
			}

			// Expand image name if it contains env vars
			expandedImage := expandEnvVars(service.Image, envVars)
			args = append(args, expandedImage)
			mcpServer.Args = args
		} else {
			// Command-based server
			parts := strings.Fields(service.Command)
			if len(parts) > 0 {
				mcpServer.Command = parts[0]
				if len(parts) > 1 {
					// Expand environment variables in args
					expandedArgs := make([]string, 0, len(parts)-1)
					for _, arg := range parts[1:] {
						expandedArgs = append(expandedArgs, expandEnvVars(arg, envVars))
					}
					mcpServer.Args = expandedArgs
				}
			}
		}

		// Add environment variables with expanded values (only for local servers)
		if !IsRemoteServerWithEnvExpansion(service, envVars) && len(service.Environment) > 0 {
			expandedEnv := make(map[string]string)
			for key, value := range service.Environment {
				// Expand environment variables in the output JSON
				expandedEnv[key] = expandEnvVars(value, envVars)
			}
			mcpServer.Env = expandedEnv
		}

		mcpServers[name] = mcpServer
	}

	return MCPConfig{MCPServers: mcpServers}
}

func writeMCPConfig(config MCPConfig, path string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
