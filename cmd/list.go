package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var (
	allServers bool
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "ls [profile]",
	Aliases: []string{"list"},
	Short:   "List MCP servers",
	Long: `List MCP servers from the mcp-compose.yml file.
Without arguments, it lists all default servers.
With a profile argument, it lists all servers with that profile.
With the -a flag, it lists all servers.`,
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
		displayServers(servers)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVarP(&allServers, "all", "a", false, "List all servers")
}

func displayServers(servers map[string]Service) {
	if len(servers) == 0 {
		fmt.Println("No servers found")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Name\tCommand\tImage\tEnvvars\tProfiles")
	fmt.Fprintln(w, "----\t-------\t-----\t-------\t--------")

	for name, service := range servers {
		command := service.Command
		image := service.Image
		
		// Get environment variables
		var envVars []string
		for key := range service.Environment {
			envVars = append(envVars, key)
		}
		envVarsStr := strings.Join(envVars, ", ")
		
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
		
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", name, command, image, envVarsStr, profilesStr)
	}
	w.Flush()
}
