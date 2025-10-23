# Requirements Document

## Introduction

This feature adds support for remote MCP servers to the existing MCP CLI tool. Currently, the tool only supports local MCP servers that are executed via commands or containers. This enhancement will allow users to configure and connect to remote MCP servers hosted on external endpoints using OAuth 2.0 client credentials authentication.

The feature will detect remote servers by checking if the command starts with `https://`, automatically handle OAuth token acquisition, and configure the appropriate authentication headers in the generated MCP configuration files.

## Requirements

### Requirement 1

**User Story:** As a developer using MCP CLI, I want to configure remote MCP servers in my YAML configuration file, so that I can connect to MCP servers hosted on external services.

#### Acceptance Criteria

1. WHEN a service command starts with `https://` THEN the system SHALL treat it as a remote MCP server
2. WHEN configuring a remote server THEN the system SHALL support OAuth 2.0 client credentials grant type via labels
3. WHEN a remote server is configured THEN the system SHALL require `mcp.grant-type`, `mcp.token-endpoint`, `mcp.client-id`, and `mcp.client-secret` labels
4. WHEN environment variable expansion is used THEN the system SHALL resolve variables like `${REMOTE_CLIENT_ID}` and `${REMOTE_CLIENT_SECRET}`

### Requirement 2

**User Story:** As a developer using MCP CLI, I want the tool to automatically acquire OAuth access tokens for remote servers, so that I don't have to manually manage authentication.

#### Acceptance Criteria

1. WHEN running `mcp set` with remote servers configured THEN the system SHALL automatically perform OAuth 2.0 client credentials flow
2. WHEN acquiring OAuth tokens THEN the system SHALL output "acquiring access token..." to inform the user
3. WHEN making OAuth requests THEN the system SHALL use POST method with `application/x-www-form-urlencoded` content type
4. WHEN OAuth request succeeds THEN the system SHALL extract the `access_token` from the JSON response
5. WHEN OAuth request fails THEN the system SHALL return an appropriate error message and exit with status 1
6. WHEN OAuth credentials are missing or invalid THEN the system SHALL provide clear error messages

### Requirement 3

**User Story:** As a developer using MCP CLI, I want remote servers to be written to MCP configuration files with proper authentication headers, so that AI tools can connect to them securely.

#### Acceptance Criteria

1. WHEN generating MCP configuration for remote servers THEN the system SHALL use `"type": "http"` instead of command-based configuration
2. WHEN writing remote server configuration THEN the system SHALL set the `url` field to the HTTPS endpoint from the command field
3. WHEN writing remote server configuration THEN the system SHALL include an `Authorization` header with the acquired access token
4. WHEN writing to tool-specific configs THEN the system SHALL support only `kiro` and `q-cli` tools for remote servers
5. WHEN user specifies unsupported tools for remote servers THEN the system SHALL output an error message and exit with status 1
6. WHEN mixing local and remote servers THEN the system SHALL handle both types correctly in the same configuration

### Requirement 4

**User Story:** As a developer using MCP CLI, I want clear error handling and validation for remote server configurations, so that I can quickly identify and fix configuration issues.

#### Acceptance Criteria

1. WHEN a command starts with `https://` but lacks required OAuth labels THEN the system SHALL return a validation error
2. WHEN OAuth token endpoint is unreachable THEN the system SHALL return a network error with helpful context
3. WHEN OAuth credentials are rejected THEN the system SHALL return an authentication error
4. WHEN remote server URL is malformed THEN the system SHALL return a URL validation error
5. WHEN environment variables for OAuth credentials are missing THEN the system SHALL return a clear error message

### Requirement 5

**User Story:** As a developer using MCP CLI, I want the remote server feature to integrate seamlessly with existing functionality, so that my current workflows remain unchanged.

#### Acceptance Criteria

1. WHEN using existing commands like `mcp list` THEN the system SHALL display both local and remote servers
2. WHEN using profile filtering THEN the system SHALL apply profiles to both local and remote servers
3. WHEN using `mcp clear` THEN the system SHALL remove both local and remote server configurations
4. WHEN using existing environment variable expansion THEN the system SHALL work for both local and remote server configurations
5. WHEN backward compatibility is required THEN the system SHALL continue to support all existing local server configurations without changes
