package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all MCP servers from configuration",
	Long:  `Remove all MCP servers from the output MCP JSON configuration file.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load environment variables
		envVars, err := loadEnvVars(composeFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading environment variables: %v\n", err)
			os.Exit(1)
		}

		// Determine the output file path
		outputPath, err := getOutputPath(envVars)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error determining output path: %v\n", err)
			os.Exit(1)
		}

		// Create an empty MCP configuration
		emptyConfig := MCPConfig{
			MCPServers: make(map[string]MCPServer),
		}

		// Write the empty configuration to file
		if err := writeMCPConfig(emptyConfig, outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing MCP config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Cleared all servers from %s\n", outputPath)
	},
}

func init() {
	rootCmd.AddCommand(clearCmd)
	clearCmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to write the MCP JSON configuration file")
	clearCmd.Flags().StringVarP(&toolShortcut, "tool", "t", "", "Tool shortcut (q-cli, claude-desktop, cursor)")
}
