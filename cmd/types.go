package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
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

// loadEnvVars loads environment variables from the system and .env file
func loadEnvVars(composePath string) (map[string]string, error) {
	envVars := make(map[string]string)

	// First, load all environment variables from the system
	for _, envVar := range os.Environ() {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) == 2 {
			envVars[parts[0]] = parts[1]
		}
	}

	// Then, try to load variables from .env file in the same directory as the compose file
	envFilePath := filepath.Join(filepath.Dir(composePath), ".env")
	file, err := os.Open(envFilePath)
	if err != nil {
		// If the file doesn't exist, that's fine, just return the system env vars
		if os.IsNotExist(err) {
			return envVars, nil
		}
		return nil, fmt.Errorf("error opening .env file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse VAR=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			
			// Remove quotes if present
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
				(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}
			
			// Only set if not already in environment
			if _, exists := envVars[key]; !exists {
				envVars[key] = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading .env file: %w", err)
	}

	return envVars, nil
}

// expandEnvVars replaces ${VAR} or $VAR in the input string with their values from the environment
func expandEnvVars(input string, envVars map[string]string) string {
	result := input

	// Replace ${VAR} format
	for key, value := range envVars {
		result = strings.ReplaceAll(result, "${"+key+"}", value)
	}

	// Replace $VAR format
	for key, value := range envVars {
		result = strings.ReplaceAll(result, "$"+key, value)
	}

	return result
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
