package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// CLIConfig represents the structure of the MCP CLI config file
type CLIConfig struct {
	Tool string `json:"tool,omitempty"`
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage MCP CLI configuration",
	Long:  `Manage MCP CLI configuration settings.`,
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Long:  `Set a configuration value in the MCP CLI config file.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		if key != "tool" {
			fmt.Fprintf(os.Stderr, "Error: unsupported configuration key: %s\n", key)
			os.Exit(1)
		}

		// Expand ~ to home directory if present
		if value[:1] == "~" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting user home directory: %v\n", err)
				os.Exit(1)
			}
			value = filepath.Join(homeDir, value[1:])
		}

		// Ensure the config directory exists
		configDir := getConfigDir()
		if err := os.MkdirAll(configDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating config directory: %v\n", err)
			os.Exit(1)
		}

		configPath := filepath.Join(configDir, "config.json")

		// Load existing config if it exists
		config := CLIConfig{}
		if _, err := os.Stat(configPath); err == nil {
			data, err := os.ReadFile(configPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
				os.Exit(1)
			}
			if err := json.Unmarshal(data, &config); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing config file: %v\n", err)
				os.Exit(1)
			}
		}

		// Update the config
		switch key {
		case "tool":
			config.Tool = value
		}

		// Write the updated config
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding config: %v\n", err)
			os.Exit(1)
		}

		if err := os.WriteFile(configPath, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing config file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Set %s to %s in %s\n", key, value, configPath)
	},
}

// getConfigDir returns the path to the MCP CLI config directory
func getConfigDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting user home directory: %v\n", err)
		os.Exit(1)
	}
	return filepath.Join(homeDir, ".config", "mcp")
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
}
