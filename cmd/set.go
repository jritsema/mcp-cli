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

// Tool shortcuts mapping
var toolShortcuts = map[string]string{
	"q-cli":          filepath.Join("${HOME}", ".aws", "amazonq", "mcp.json"),
	"claude-desktop": filepath.Join("${HOME}", "Library", "Application Support", "Claude", "claude_desktop_config.json"),
	"cursor":         filepath.Join("${HOME}", ".cursor", "mcp.json"),
}

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

		var profile string
		if len(args) > 0 {
			profile = args[0]
		}

		// Determine the output file path
		outputPath, err := getOutputPath()
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

		// Convert to MCP JSON format
		mcpConfig := convertToMCPConfig(servers)

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
	setCmd.Flags().StringVarP(&toolShortcut, "tool", "t", "", "Tool shortcut (q-cli, claude-desktop, cursor)")
	setCmd.Flags().StringVarP(&singleServer, "server", "s", "", "Specify a single server to include")
}

func getOutputPath() (string, error) {
	if configFile != "" {
		return configFile, nil
	}

	if toolShortcut != "" {
		path, exists := toolShortcuts[toolShortcut]
		if !exists {
			return "", fmt.Errorf("unknown tool shortcut: %s", toolShortcut)
		}

		// Replace ${HOME} with actual home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = strings.Replace(path, "${HOME}", homeDir, 1)

		// Create directory if it doesn't exist
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", err
		}

		return path, nil
	}

	return "", fmt.Errorf("either --config or --tool must be specified")
}

// MCPConfig represents the MCP JSON configuration format
type MCPConfig struct {
	MCPServers map[string]MCPServer `json:"mcpServers"`
}

// MCPServer represents a single MCP server in the JSON configuration
type MCPServer struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

func convertToMCPConfig(servers map[string]Service) MCPConfig {
	mcpServers := make(map[string]MCPServer)

	for name, service := range servers {
		var mcpServer MCPServer

		if service.Image != "" {
			// Docker container
			mcpServer.Command = "docker"
			args := []string{"run", "-i", "--rm"}

			// Add environment variables
			for key, value := range service.Environment {
				args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
			}

			args = append(args, service.Image)
			mcpServer.Args = args
		} else {
			// Command-based server
			parts := strings.Fields(service.Command)
			if len(parts) > 0 {
				mcpServer.Command = parts[0]
				if len(parts) > 1 {
					mcpServer.Args = parts[1:]
				}
			}
		}

		// Add environment variables
		if len(service.Environment) > 0 {
			mcpServer.Env = service.Environment
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
