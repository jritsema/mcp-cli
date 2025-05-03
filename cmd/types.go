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

// Service represents a service in the docker-compose.yml file
type Service struct {
	Command     string            `yaml:"command"`
	Image       string            `yaml:"image"`
	Environment map[string]string `yaml:"environment"`
	Labels      map[string]string `yaml:"labels"`
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
		if profile == "" {
			// Default profile - include servers with no profile or with "default" in their profile
			profileStr, hasProfile := service.Labels["mcp.profile"]
			if !hasProfile {
				// No profile specified, consider it default
				result[name] = service
				continue
			}

			profiles := strings.Split(profileStr, ",")
			for _, p := range profiles {
				if strings.TrimSpace(p) == "default" {
					result[name] = service
					break
				}
			}
		} else {
			// Specific profile
			profileStr, hasProfile := service.Labels["mcp.profile"]
			if !hasProfile {
				continue
			}

			profiles := strings.Split(profileStr, ",")
			for _, p := range profiles {
				if strings.TrimSpace(p) == profile {
					result[name] = service
					break
				}
			}
		}
	}

	return result
}
