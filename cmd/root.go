package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	composeFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mcp",
	Short: "MCP CLI is a tool for managing MCP server configuration files",
	Long: `MCP CLI is a tool for managing MCP server configuration files.
It helps with managing different MCP server configurations based on profiles.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	defaultComposeFile := getDefaultComposeFile()
	rootCmd.PersistentFlags().StringVarP(&composeFile, "file", "f", defaultComposeFile, "Path to the mcp-compose.yml file")
}

// getDefaultComposeFile returns the default compose file path, checking local directory first
func getDefaultComposeFile() string {
	// First check for local mcp-compose.yml in current directory
	localComposeFile := "mcp-compose.yml"
	if _, err := os.Stat(localComposeFile); err == nil {
		return localComposeFile
	}

	// Fall back to the global config directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting user home directory: %v\n", err)
		os.Exit(1)
	}

	return filepath.Join(homeDir, ".config", "mcp", "mcp-compose.yml")
}
