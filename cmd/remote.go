package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// IsRemoteServer detects if a service is a remote MCP server by checking if the command starts with https:// or http://
func IsRemoteServer(service Service) bool {
	return strings.HasPrefix(service.Command, "https://") || strings.HasPrefix(service.Command, "http://")
}

// UsesHeadersAuth checks if a remote server uses headers-based authentication instead of OAuth
func UsesHeadersAuth(service Service) bool {
	// Check if any mcp.header.* labels exist
	for label := range service.Labels {
		if strings.HasPrefix(label, "mcp.header.") {
			return true
		}
	}
	return false
}

// ValidateRemoteServerAuth validates that a remote server has either OAuth or headers-based auth configured
func ValidateRemoteServerAuth(name string, service Service) error {
	usesHeaders := UsesHeadersAuth(service)
	hasOAuthLabels := service.Labels["mcp.grant-type"] != ""

	if !usesHeaders && !hasOAuthLabels {
		return fmt.Errorf("remote server '%s' must have either OAuth labels (mcp.grant-type, mcp.token-endpoint, mcp.client-id, mcp.client-secret) or headers labels (mcp.header.*)", name)
	}

	if usesHeaders && hasOAuthLabels {
		return fmt.Errorf("remote server '%s' cannot have both OAuth labels and headers labels", name)
	}

	if hasOAuthLabels {
		// Validate OAuth configuration
		requiredLabels := []string{
			"mcp.grant-type",
			"mcp.token-endpoint",
			"mcp.client-id",
			"mcp.client-secret",
		}

		var missingLabels []string
		for _, label := range requiredLabels {
			if _, exists := service.Labels[label]; !exists {
				missingLabels = append(missingLabels, label)
			}
		}

		if len(missingLabels) > 0 {
			return fmt.Errorf("remote server '%s' missing required OAuth labels: %s",
				name, strings.Join(missingLabels, ", "))
		}

		// Validate grant type is client_credentials
		if grantType := service.Labels["mcp.grant-type"]; grantType != "client_credentials" {
			return fmt.Errorf("remote server '%s' must use 'client_credentials' grant type, got: %s",
				name, grantType)
		}
	}

	return nil
}

// ValidateRemoteServerOAuth validates that a remote server has all required OAuth labels (deprecated, use ValidateRemoteServerAuth)
func ValidateRemoteServerOAuth(name string, service Service) error {
	return ValidateRemoteServerAuth(name, service)
}

// ExtractHeaders extracts headers from service labels (mcp.header.*) with environment variable expansion
func ExtractHeaders(service Service, envVars map[string]string) (map[string]string, error) {
	headers := make(map[string]string)
	hasHeaders := false

	for label, value := range service.Labels {
		if strings.HasPrefix(label, "mcp.header.") {
			hasHeaders = true
			// Extract header name (everything after "mcp.header.")
			headerName := strings.TrimPrefix(label, "mcp.header.")
			if headerName == "" {
				continue
			}

			// Skip empty placeholder headers (e.g., "X-Empty: "" for servers that need no auth)
			if headerName == "X-Empty" && value == "" {
				continue
			}

			// Expand environment variables in header value
			expandedValue := expandEnvVars(value, envVars)

			// Validate that environment variables were resolved
			if strings.Contains(expandedValue, "${") || (strings.Contains(expandedValue, "$") && !strings.HasPrefix(expandedValue, "$")) {
				return nil, fmt.Errorf("environment variable in header '%s' was not resolved: %s", headerName, expandedValue)
			}

			headers[headerName] = expandedValue
		}
	}

	// Return headers map (can be empty for servers with no authentication)
	// If no mcp.header.* labels exist at all, that's an error (use OAuth or headers)
	if !hasHeaders {
		return nil, fmt.Errorf("no headers found (expected mcp.header.* labels)")
	}

	return headers, nil
}

// remoteSupportedTools defines which tools support remote MCP servers
var remoteSupportedTools = map[string]bool{
	"cursor": true,
	"kiro":   true,
	"q-cli":  true,
}

// ValidateToolSupport validates that the specified tool supports remote servers if any are present
func ValidateToolSupport(toolShortcut string, servers map[string]Service) error {
	hasRemoteServers := false
	for _, service := range servers {
		if IsRemoteServer(service) {
			hasRemoteServers = true
			break
		}
	}

	if hasRemoteServers && toolShortcut != "" {
		if !remoteSupportedTools[toolShortcut] {
			supportedTools := make([]string, 0, len(remoteSupportedTools))
			for tool := range remoteSupportedTools {
				supportedTools = append(supportedTools, tool)
			}
			return fmt.Errorf("tool '%s' does not support remote MCP servers. Supported tools: %s",
				toolShortcut, strings.Join(supportedTools, ", "))
		}
	}

	return nil
}

// ExtractOAuthConfig extracts OAuth configuration from service labels with environment variable expansion
func ExtractOAuthConfig(service Service, envVars map[string]string) (OAuthConfig, error) {
	config := OAuthConfig{
		GrantType:    service.Labels["mcp.grant-type"],
		TokenURL:     service.Labels["mcp.token-endpoint"],
		ClientID:     service.Labels["mcp.client-id"],
		ClientSecret: service.Labels["mcp.client-secret"],
	}

	// Expand environment variables in OAuth configuration
	config.GrantType = expandEnvVars(config.GrantType, envVars)
	config.TokenURL = expandEnvVars(config.TokenURL, envVars)
	config.ClientID = expandEnvVars(config.ClientID, envVars)
	config.ClientSecret = expandEnvVars(config.ClientSecret, envVars)

	// Validate that required environment variables were resolved
	if strings.Contains(config.ClientID, "${") || strings.Contains(config.ClientID, "$") {
		return config, fmt.Errorf("environment variable in OAuth client ID was not resolved: %s", config.ClientID)
	}
	if strings.Contains(config.ClientSecret, "${") || strings.Contains(config.ClientSecret, "$") {
		return config, fmt.Errorf("environment variable in OAuth client secret was not resolved: %s", config.ClientSecret)
	}
	if strings.Contains(config.TokenURL, "${") || strings.Contains(config.TokenURL, "$") {
		return config, fmt.Errorf("environment variable in OAuth token URL was not resolved: %s", config.TokenURL)
	}

	return config, nil
}

// acquireAccessToken performs OAuth 2.0 client credentials flow to acquire an access token
func acquireAccessToken(config OAuthConfig) (string, error) {
	// Prepare form data for client credentials grant
	data := url.Values{}
	data.Set("grant_type", config.GrantType)
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create POST request with application/x-www-form-urlencoded content type
	req, err := http.NewRequest("POST", config.TokenURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create OAuth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	// Handle HTTP error responses
	if resp.StatusCode == 401 {
		return "", fmt.Errorf("authentication failed (401 Unauthorized)")
	}
	if resp.StatusCode == 403 {
		return "", fmt.Errorf("authentication failed (403 Forbidden)")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("OAuth request failed with status %d", resp.StatusCode)
	}

	// Parse JSON response
	var oauthResp OAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&oauthResp); err != nil {
		return "", fmt.Errorf("failed to parse OAuth response: %w", err)
	}

	// Validate that we received an access token
	if oauthResp.AccessToken == "" {
		return "", fmt.Errorf("OAuth response missing access_token field")
	}

	return oauthResp.AccessToken, nil
}

// AcquireAccessTokenWithFeedback acquires an OAuth access token with user feedback
func AcquireAccessTokenWithFeedback(serverName string, config OAuthConfig) (string, error) {
	fmt.Fprintf(os.Stderr, "acquiring access token for '%s'...\n", serverName)
	return acquireAccessToken(config)
}
