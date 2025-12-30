package cmd

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// ComposeConfig represents the structure of a docker-compose.yml file
type ComposeConfig struct {
	Services map[string]Service `yaml:"services"`
}

// loadComposeFile loads and parses the compose file
func loadComposeFile(path string) (*ComposeConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config ComposeConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// filterServers filters servers based on profile
func filterServers(config *ComposeConfig, profile string, all bool) map[string]Service {
	result := make(map[string]Service)

	if all {
		// Return all servers
		return config.Services
	}

	for name, service := range config.Services {
		// Check if this is a default server (no profile or has "default" in profile)
		isDefault := false
		profileStr, hasProfile := service.Labels["mcp.profile"]

		if !hasProfile {
			// No profile specified, consider it default
			isDefault = true
		} else {
			// Check if it has "default" in its profile
			profiles := strings.Split(profileStr, ",")
			for _, p := range profiles {
				if strings.TrimSpace(p) == "default" {
					isDefault = true
					break
				}
			}
		}

		if profile == "" {
			// Only include default servers when no specific profile is requested
			if isDefault {
				result[name] = service
			}
		} else {
			// When a specific profile is requested, include both:
			// 1. Default servers
			// 2. Servers with the requested profile
			if isDefault {
				result[name] = service
				continue
			}

			// Check if server has the requested profile
			if hasProfile {
				profiles := strings.Split(profileStr, ",")
				for _, p := range profiles {
					if strings.TrimSpace(p) == profile {
						result[name] = service
						break
					}
				}
			}
		}
	}

	return result
}

// Service represents a service in the docker-compose.yml file
type Service struct {
	Command     string            `yaml:"command"`
	Image       string            `yaml:"image"`
	Environment map[string]string `yaml:"environment"`
	Labels      map[string]string `yaml:"labels"`
}

// MCPConfig represents the MCP JSON configuration format
type MCPConfig struct {
	MCPServers map[string]MCPServer `json:"mcpServers"`
}

// MCPServer represents a single MCP server in the JSON configuration
type MCPServer struct {
	// Existing fields for local servers
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`

	// New fields for remote servers
	Type    string            `json:"type,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// CLIConfig represents the structure of the MCP CLI config file
type CLIConfig struct {
	Tool          string `json:"tool,omitempty"`
	ContainerTool string `json:"container-tool,omitempty"`
}

// OAuthConfig represents OAuth 2.0 client credentials configuration
type OAuthConfig struct {
	GrantType    string
	TokenURL     string
	ClientID     string
	ClientSecret string
}

// OAuthResponse represents the response from an OAuth 2.0 token endpoint
type OAuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// ServerStatus represents the status of a server in a specific tool
type ServerStatus struct {
	Status      string   // "configured", "not-configured", "different", "unknown"
	Tool        string   // tool shortcut name
	Differences []string // list of differences if status is "different"
	ConfigPath  string   // path to the config file
	Error       string   // error message if status is "unknown"
}

// ToolStatus represents the status of a tool configuration
type ToolStatus struct {
	ToolName    string
	ConfigPath  string
	Exists      bool
	ServerCount int
}

// ToolConfig represents a tool's configuration with metadata
type ToolConfig struct {
	Config  MCPConfig
	Path    string
	Exists  bool
	Error   string
}
