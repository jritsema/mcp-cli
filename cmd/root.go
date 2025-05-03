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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting user home directory: %v\n", err)
		os.Exit(1)
	}

	defaultComposeFile := filepath.Join(homeDir, "mcp-compose.yml")

	rootCmd.PersistentFlags().StringVarP(&composeFile, "file", "f", defaultComposeFile, "Path to the mcp-compose.yml file")
}
