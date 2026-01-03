package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var (
	allServers    bool
	longFormat    bool
	showStatus    bool
	toolFilter    string
	allTools      bool
	commandFormat bool
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "ls [profile]",
	Aliases: []string{"list"},
	Short:   "List MCP servers",
	Long: `List MCP servers from the mcp-compose.yml file.
Without arguments, it lists all default servers.
With a profile argument, it lists all servers with that profile.
With the -a flag, it lists all servers.
With the -l flag, it shows detailed information including command and environment variables.
With the -s flag, it shows deployment status across configured tools.
With the -c flag, it shows the executable command with environment variables expanded and inline.`,
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

		// Filter servers based on profile or show all
		servers := filterServers(config, profile, allServers)

		// Display the servers
		if showStatus {
			displayServersWithStatus(servers)
		} else {
			displayServers(servers)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVarP(&allServers, "all", "a", false, "List all servers")
	listCmd.Flags().BoolVarP(&longFormat, "long", "l", false, "Show detailed information including command and environment variables")
	listCmd.Flags().BoolVarP(&showStatus, "status", "s", false, "Show deployment status across configured tools")
	listCmd.Flags().StringVarP(&toolFilter, "tool", "t", "", "Show status for specific tool only (q-cli, claude-desktop, cursor, kiro)")
	listCmd.Flags().BoolVar(&allTools, "all-tools", false, "Show status across all supported tools")
	listCmd.Flags().BoolVarP(&commandFormat, "command", "c", false, "Show executable command with environment variables expanded inline. WARNING: may expose sensitive data such as API keys and secrets")
}

func displayServers(servers map[string]Service) {
	if len(servers) == 0 {
		fmt.Println("No servers found")
		return
	}

	// Load environment variables if command format flag is set
	var envVars map[string]string
	if commandFormat {
		var err error
		envVars, err = loadEnvVars(composeFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: error loading environment variables: %v\n", err)
			envVars = make(map[string]string)
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Display headers based on format
	if commandFormat {
		fmt.Fprintln(w, "NAME\tCOMMAND")
		fmt.Fprintln(w, "----\t-------")
	} else if longFormat {
		fmt.Fprintln(w, "NAME\tPROFILES\tCOMMAND\tENVVARS")
		fmt.Fprintln(w, "----\t--------\t-------\t-------")
	} else {
		fmt.Fprintln(w, "NAME\tPROFILES")
		fmt.Fprintln(w, "----\t--------")
	}

	// Get the original order from the compose file
	config, err := loadComposeFile(composeFile)
	if err != nil {
		// If we can't load the file again, just use the map order
		for name, service := range servers {
			printServerRow(w, name, service, envVars)
		}
	} else {
		// Create two lists: one for default servers and one for non-default servers
		var defaultServers []string
		var otherServers []string

		// Categorize servers
		for name := range config.Services {
			if _, exists := servers[name]; exists {
				service := servers[name]
				isDefault := false

				// Check if this is a default server (no profile or has "default" in profile)
				profileStr, hasProfile := service.Labels["mcp.profile"]
				if !hasProfile {
					isDefault = true
				} else {
					profiles := strings.Split(profileStr, ",")
					for _, p := range profiles {
						if strings.TrimSpace(p) == "default" {
							isDefault = true
							break
						}
					}
				}

				if isDefault {
					defaultServers = append(defaultServers, name)
				} else {
					otherServers = append(otherServers, name)
				}
			}
		}

		// Sort both lists alphabetically
		sort.Strings(defaultServers)
		sort.Strings(otherServers)

		// Print default servers first (alphabetically sorted)
		for _, name := range defaultServers {
			printServerRow(w, name, servers[name], envVars)
		}

		// Then print other servers (alphabetically sorted)
		for _, name := range otherServers {
			printServerRow(w, name, servers[name], envVars)
		}
	}

	w.Flush()
}

// shellQuote quotes a string for safe use in shell commands
func shellQuote(s string) string {
	// If the string contains no special characters, return as-is
	if !strings.ContainsAny(s, " \t\n\"'\\`!") {
		return s
	}
	// Use double quotes and escape special characters (but not $, to preserve unexpanded vars)
	escaped := strings.ReplaceAll(s, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
	escaped = strings.ReplaceAll(escaped, "`", "\\`")
	return "\"" + escaped + "\""
}

// Helper function to print a single server row
func printServerRow(w *tabwriter.Writer, name string, service Service, envVars map[string]string) {
	// Get profiles
	var profiles []string
	if profilesStr, ok := service.Labels["mcp.profile"]; ok {
		profiles = strings.Split(profilesStr, ",")
		for i, p := range profiles {
			profiles[i] = strings.TrimSpace(p)
		}
	}
	if len(profiles) == 0 {
		profiles = append(profiles, "default")
	}
	profilesStr := strings.Join(profiles, ", ")

	if commandFormat {
		// Command format: NAME + executable command with env vars inline
		var commandStr string

		// Check if this is a remote server
		if IsRemoteServer(service) {
			// For remote servers, just show the URL
			commandStr = service.Command
		} else {
			// Build env var prefix for the command
			var envPrefix string
			if !IsRemoteServer(service) && len(service.Environment) > 0 {
				var envParts []string
				// Sort keys for consistent output
				var keys []string
				for key := range service.Environment {
					keys = append(keys, key)
				}
				sort.Strings(keys)
				for _, key := range keys {
					value := service.Environment[key]
					expandedValue := expandEnvVars(value, envVars)
					envParts = append(envParts, fmt.Sprintf("%s=%s", key, shellQuote(expandedValue)))
				}
				envPrefix = strings.Join(envParts, " ") + " "
			}

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

			if service.Image != "" {
				// For image-based servers, show the container run command format
				commandStr = fmt.Sprintf("%s run -i --rm", containerTool)

				// Add environment variables as -e flags
				var keys []string
				for key := range service.Environment {
					keys = append(keys, key)
				}
				sort.Strings(keys)
				for _, key := range keys {
					value := service.Environment[key]
					expandedValue := expandEnvVars(value, envVars)
					commandStr += fmt.Sprintf(" -e %s=%s", key, shellQuote(expandedValue))
				}

				// Add volume mounts as -v flags
				for _, volume := range service.Volumes {
					expandedVolume := expandEnvVars(volume, envVars)
					commandStr += fmt.Sprintf(" -v %s", shellQuote(expandedVolume))
				}

				// Add the image name
				commandStr += fmt.Sprintf(" %s", service.Image)
			} else {
				// For command-based servers, prepend env vars and expand command
				expandedCommand := expandEnvVars(service.Command, envVars)
				commandStr = envPrefix + expandedCommand
			}
		}

		fmt.Fprintf(w, "%s\t%s\n", name, commandStr)
	} else if longFormat {
		var commandStr string

		// Check if this is a remote server
		if IsRemoteServer(service) {
			// For remote servers, show the URL
			commandStr = service.Command
		} else {
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

			if service.Image != "" {
				// For image-based servers, show the container run command format
				commandStr = fmt.Sprintf("%s run -i --rm", containerTool)

				// Add environment variables to the command
				for key := range service.Environment {
					commandStr += fmt.Sprintf(" -e %s", key)
				}

				// Add volume mounts to the command
				for _, volume := range service.Volumes {
					commandStr += fmt.Sprintf(" -v %s", volume)
				}

				// Add the image name
				commandStr += fmt.Sprintf(" %s", service.Image)
			} else {
				// For command-based servers, show the command
				commandStr = service.Command
			}
		}

		// Get environment variables (only for local servers, remote servers use OAuth)
		var envVarsDisplay []string
		if !IsRemoteServer(service) {
			for key := range service.Environment {
				envVarsDisplay = append(envVarsDisplay, key)
			}
		}
		sort.Strings(envVarsDisplay)
		envVarsStr := strings.Join(envVarsDisplay, ", ")

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", name, profilesStr, commandStr, envVarsStr)
	} else {
		// Simple format with just name and profiles
		fmt.Fprintf(w, "%s\t%s\n", name, profilesStr)
	}
}

// displayServersWithStatus displays servers with their deployment status across tools
func displayServersWithStatus(servers map[string]Service) {
	if len(servers) == 0 {
		fmt.Println("No servers found")
		return
	}

	// Determine which tools to check
	var tools []string
	if toolFilter != "" {
		// Check if tool shortcut exists
		if getPlatformToolPath(toolFilter) == "" {
			fmt.Fprintf(os.Stderr, "Error: unknown tool shortcut: %s\n", toolFilter)
			os.Exit(1)
		}
		tools = []string{toolFilter}
	} else if allTools {
		// Get all tool shortcuts
		tools = supportedTools
	} else {
		// Default: show all tools
		tools = supportedTools
	}

	// Load environment variables for comparison
	envVars, err := loadEnvVars(composeFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: error loading environment variables: %v\n", err)
		envVars = make(map[string]string)
	}

	// Load tool configs
	toolConfigs := getToolConfigs(tools)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Print headers
	if longFormat {
		// Long format with status
		header := "NAME\tPROFILES\tTYPE"
		for _, tool := range tools {
			header += "\t" + normalizeToolName(tool) + " STATUS"
		}
		fmt.Fprintln(w, header)
		separator := "----\t--------\t----"
		for range tools {
			separator += "\t" + strings.Repeat("-", 15)
		}
		fmt.Fprintln(w, separator)
	} else {
		// Simple format with status
		header := "NAME\tPROFILES"
		for _, tool := range tools {
			header += "\t" + normalizeToolName(tool)
		}
		fmt.Fprintln(w, header)
		separator := "----\t--------"
		for range tools {
			separator += "\t" + strings.Repeat("-", 6)
		}
		fmt.Fprintln(w, separator)
	}

	// Get the original order from the compose file
	config, err := loadComposeFile(composeFile)
	if err != nil {
		// If we can't load the file again, just use the map order
		for name, service := range servers {
			printServerRowWithStatus(w, name, service, tools, toolConfigs, envVars)
		}
	} else {
		// Create two lists: one for default servers and one for non-default servers
		var defaultServers []string
		var otherServers []string

		// Categorize servers
		for name := range config.Services {
			if _, exists := servers[name]; exists {
				service := servers[name]
				isDefault := false

				// Check if this is a default server (no profile or has "default" in profile)
				profileStr, hasProfile := service.Labels["mcp.profile"]
				if !hasProfile {
					isDefault = true
				} else {
					profiles := strings.Split(profileStr, ",")
					for _, p := range profiles {
						if strings.TrimSpace(p) == "default" {
							isDefault = true
							break
						}
					}
				}

				if isDefault {
					defaultServers = append(defaultServers, name)
				} else {
					otherServers = append(otherServers, name)
				}
			}
		}

		// Sort both lists alphabetically
		sort.Strings(defaultServers)
		sort.Strings(otherServers)

		// Print default servers first (alphabetically sorted)
		for _, name := range defaultServers {
			printServerRowWithStatus(w, name, servers[name], tools, toolConfigs, envVars)
		}

		// Then print other servers (alphabetically sorted)
		for _, name := range otherServers {
			printServerRowWithStatus(w, name, servers[name], tools, toolConfigs, envVars)
		}
	}

	w.Flush()
}

// printServerRowWithStatus prints a server row with status information
func printServerRowWithStatus(w *tabwriter.Writer, name string, service Service, tools []string, toolConfigs map[string]ToolConfig, envVars map[string]string) {
	// Get profiles
	var profiles []string
	if profilesStr, ok := service.Labels["mcp.profile"]; ok {
		profiles = strings.Split(profilesStr, ",")
		for i, p := range profiles {
			profiles[i] = strings.TrimSpace(p)
		}
	}
	if len(profiles) == 0 {
		profiles = append(profiles, "default")
	}
	profilesStr := strings.Join(profiles, ", ")

	// Get server status for each tool
	serverStatuses := getServerStatus(name, service, toolConfigs, envVars)

	// Build status indicators
	var statusIndicators []string
	for _, tool := range tools {
		status, exists := serverStatuses[tool]
		if !exists {
			statusIndicators = append(statusIndicators, "?")
			continue
		}

		var indicator string
		switch status.Status {
		case "configured":
			indicator = "✓"
		case "not-configured":
			indicator = "✗"
		case "different":
			indicator = "~"
		case "unknown":
			indicator = "?"
		default:
			indicator = "?"
		}

		if longFormat {
			// Long format: show status text
			switch status.Status {
			case "configured":
				indicator = "✓ configured"
			case "not-configured":
				indicator = "✗ not configured"
			case "different":
				indicator = "~ different"
			case "unknown":
				indicator = "? unknown"
			default:
				indicator = "? unknown"
			}
		}

		statusIndicators = append(statusIndicators, indicator)
	}

	if longFormat {
		// Determine server type
		var serverType string
		if IsRemoteServer(service) {
			serverType = "remote"
		} else if service.Image != "" {
			serverType = "container"
		} else {
			serverType = "local"
		}

		row := fmt.Sprintf("%s\t%s\t%s", name, profilesStr, serverType)
		for _, indicator := range statusIndicators {
			row += "\t" + indicator
		}
		fmt.Fprintln(w, row)

		// For remote servers, show URL in long format
		if IsRemoteServer(service) {
			indent := strings.Repeat("\t", 2+len(tools))
			fmt.Fprintf(w, "%sURL: %s\n", indent, service.Command)
		}
	} else {
		// Simple format
		row := fmt.Sprintf("%s\t%s", name, profilesStr)
		for _, indicator := range statusIndicators {
			row += "\t" + indicator
		}
		fmt.Fprintln(w, row)
	}
}
