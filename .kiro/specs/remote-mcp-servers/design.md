# Design Document

## Overview

This design extends the existing MCP CLI tool to support remote MCP servers alongside the current local command-based and container-based servers. The solution will detect remote servers by checking if the command field starts with `https://` or `http://`, and supports two authentication methods:

1. **Headers-based authentication** - For API keys and custom headers via `mcp.header.*` labels
2. **OAuth 2.0 client credentials** - For OAuth-based authentication via `mcp.grant-type` and related labels

The design maintains backward compatibility with existing functionality while adding new capabilities for remote server authentication and configuration generation.

## Architecture

### Current Architecture Analysis

The existing system follows a clear separation of concerns:

- **Configuration Loading**: `loadComposeFile()` parses YAML configuration
- **Environment Processing**: `loadEnvVars()` and `expandEnvVars()` handle variable expansion
- **Server Filtering**: `filterServers()` applies profile-based filtering
- **Output Generation**: `convertToMCPConfig()` transforms to MCP JSON format
- **File Writing**: `writeMCPConfig()` outputs final configuration

### Enhanced Architecture

The remote server support will be integrated into the existing flow with minimal disruption:

1. **Detection Phase**: During server processing, detect remote servers by URL pattern
2. **Authentication Phase**: For remote servers, perform OAuth 2.0 token acquisition
3. **Configuration Phase**: Generate HTTP-based MCP configuration instead of command-based
4. **Validation Phase**: Ensure tool compatibility and proper error handling

## Components and Interfaces

### File Organization

The remote server functionality will be organized into separate files for better maintainability:

- **`cmd/types.go`**: Core data structures and existing functionality
- **`cmd/remote.go`**: New file containing all remote server-specific functionality

### 1. Remote Server Detection

**Location**: `cmd/remote.go`

The system detects remote servers by examining the command field in service configurations. Remote servers are identified by URLs starting with `https://` or `http://`.

### 2. Authentication Method Detection

**Location**: `cmd/remote.go`

The system determines which authentication method to use:

- **Headers-based**: If any `mcp.header.*` labels exist
- **OAuth-based**: If `mcp.grant-type` label exists

Validation ensures servers use one method or the other, not both.

### 3. Headers Extraction

**Location**: `cmd/remote.go`

For headers-based authentication:

- Extract header names from `mcp.header.*` labels (e.g., `mcp.header.Authorization` → `Authorization`)
- Expand environment variables in header values
- Validate that all environment variables were resolved

### 4. OAuth 2.0 Client

**Location**: `cmd/remote.go`

A complete OAuth 2.0 client credentials implementation will handle:

- Configuration extraction from service labels with environment variable expansion
- HTTP client with proper timeout and content-type handling
- Token acquisition with comprehensive error handling
- User feedback during authentication process
- Response parsing and validation

### 5. Remote Server Validation

**Location**: `cmd/remote.go`

Validation components ensure:

- Remote servers have either headers-based OR OAuth-based auth (not both, not neither)
- All required OAuth labels are present when using OAuth
- Grant type is set to `client_credentials` for OAuth
- Tool compatibility with remote servers (cursor, kiro, and q-cli supported)
- Clear error messages for missing or invalid configuration

### 6. Enhanced MCP Configuration Generation

**Modified Component**: Existing configuration generation logic in `cmd/types.go`

The existing MCP configuration generation will be enhanced to support both local and remote server types. Remote servers will use HTTP transport with authorization headers, while local servers maintain their existing command-based structure.

**Configuration Logic**:

- Local servers: Use existing command/args/env structure
- Remote servers: Use type="http", url, and headers structure
- Mixed configurations: Support both types in same output file

### 7. Tool Compatibility Validation

**New Component**: Tool support validation for remote servers

The system validates tool compatibility by maintaining a list of supported tools for remote servers. Currently, `cursor`, `kiro`, and `q-cli` tools support remote servers, with clear error messages for unsupported combinations. Backward compatibility is maintained for local servers on all tools.

## Data Models

### Headers-Based Authentication Labels

Remote servers can use headers-based authentication via `mcp.header.*` labels:

```yaml
services:
  context7:
    command: https://mcp.context7.com/mcp
    labels:
      mcp.header.Authorization: Bearer ${CONTEXT7_API_KEY}

  custom-server:
    command: https://api.example.com/mcp
    labels:
      mcp.header.X-API-Key: ${API_KEY}
      mcp.header.X-Custom-Header: some-value
```

**Header Label Format**:

- Label name: `mcp.header.<HeaderName>` (e.g., `mcp.header.Authorization`)
- Label value: Header value with optional environment variable expansion

### OAuth Configuration Labels

Remote servers can use OAuth 2.0 client credentials via OAuth labels:

```yaml
services:
  my-remote-server:
    command: https://my-app.gateway.bedrock-agentcore.us-east-1.amazonaws.com/mcp
    labels:
      mcp.grant-type: client_credentials
      mcp.token-endpoint: https://my-app.auth.us-east-1.amazoncognito.com/oauth2/token
      mcp.client-id: ${REMOTE_CLIENT_ID}
      mcp.client-secret: ${REMOTE_CLIENT_SECRET}
```

**Required OAuth Labels**:

- `mcp.grant-type`: Must be "client_credentials"
- `mcp.token-endpoint`: OAuth 2.0 token endpoint URL
- `mcp.client-id`: OAuth client identifier (supports env var expansion)
- `mcp.client-secret`: OAuth client secret (supports env var expansion)

### Authentication Method Selection

- If any `mcp.header.*` labels exist → Headers-based authentication
- If `mcp.grant-type` label exists → OAuth authentication
- Both methods cannot be used on the same server
- One method must be specified for remote servers

### Generated MCP Configuration

**Local Server Output** (unchanged):

```json
{
  "mcpServers": {
    "time": {
      "command": "uvx",
      "args": ["mcp-server-time", "--local-timezone=America/New_York"]
    }
  }
}
```

**Remote Server Output** (headers-based):

```json
{
  "mcpServers": {
    "context7": {
      "type": "http",
      "url": "https://mcp.context7.com/mcp",
      "headers": {
        "Authorization": "Bearer actual-api-key-value"
      }
    }
  }
}
```

**Remote Server Output** (OAuth):

```json
{
  "mcpServers": {
    "my-remote-server": {
      "type": "http",
      "url": "https://my-app.gateway.bedrock-agentcore.us-east-1.amazonaws.com/mcp",
      "headers": {
        "Authorization": "Bearer eyJraWQiOiJsaHNVUndBXC9LMTlvT0FZRmRQUGsrSFNQTzRNT1ZGU0VTekF6NDB6b3hpbz0iLCJhbGciOiJSUzI1NiJ9..."
      }
    }
  }
}
```

## Error Handling

### Validation Errors

1. **Missing Authentication**: When remote server lacks both OAuth and headers configuration

   - Error: "Remote server 'name' must have either OAuth labels (mcp.grant-type, mcp.token-endpoint, mcp.client-id, mcp.client-secret) or headers labels (mcp.header.\*)"
   - Exit code: 1

2. **Conflicting Authentication**: When remote server has both OAuth and headers configuration

   - Error: "Remote server 'name' cannot have both OAuth labels and headers labels"
   - Exit code: 1

3. **Missing OAuth Labels**: When using OAuth but lacking required labels

   - Error: "Remote server 'name' missing required OAuth labels: mcp.grant-type, mcp.token-endpoint, mcp.client-id, mcp.client-secret"
   - Exit code: 1

4. **Unsupported Tool**: When using remote servers with unsupported tools

   - Error: "Tool 'claude-desktop' does not support remote MCP servers. Supported tools: cursor, kiro, q-cli"
   - Exit code: 1

5. **Unresolved Environment Variable**: When header value contains unresolved env var
   - Error: "Environment variable in header 'Authorization' was not resolved: Bearer ${MISSING_VAR}"
   - Exit code: 1

### Runtime Errors

1. **OAuth Network Errors**: When token endpoint is unreachable

   - Error: "Failed to acquire access token for 'server-name': network error: connection timeout"
   - Exit code: 1

2. **OAuth Authentication Errors**: When credentials are rejected

   - Error: "Failed to acquire access token for 'server-name': authentication failed (401 Unauthorized)"
   - Exit code: 1

3. **Missing Environment Variables**: When OAuth credentials are not set
   - Error: "Environment variable 'REMOTE_CLIENT_ID' required for remote server 'server-name' is not set"
   - Exit code: 1

### User Feedback

- **Token Acquisition**: Display "acquiring access token..." message during OAuth flow
- **Success Messages**: Maintain existing success output format
- **Progress Indication**: Show which remote servers are being processed

## Testing Strategy

### Unit Tests

1. **Remote Server Detection**

   - Test remote server detection with various URL formats
   - Verify HTTPS requirement and rejection of HTTP URLs
   - Test edge cases with malformed URLs

2. **OAuth Configuration Parsing**

   - Test extraction of OAuth labels from service configuration
   - Verify environment variable expansion in OAuth credentials
   - Test validation of required OAuth parameters

3. **OAuth Client**

   - Mock HTTP client for token acquisition testing
   - Test successful token response parsing
   - Test error handling for various HTTP status codes
   - Test network timeout scenarios

4. **MCP Configuration Generation**
   - Test remote server configuration output format
   - Verify proper header formatting with Bearer tokens
   - Test mixed local/remote server configurations

### Integration Tests

1. **End-to-End Configuration Generation**

   - Test complete flow from YAML input to MCP JSON output
   - Verify tool-specific output paths and formats
   - Test profile filtering with mixed server types

2. **Tool Compatibility**

   - Test supported tools (kiro, q-cli) with remote servers
   - Verify error handling for unsupported tools
   - Test backward compatibility with existing local server configurations

3. **Environment Variable Expansion**
   - Test OAuth credential expansion from environment
   - Test .env file loading with OAuth variables
   - Test error handling for missing environment variables

### Error Scenario Tests

1. **Network Failures**

   - Test OAuth endpoint unreachable scenarios
   - Test timeout handling and appropriate error messages
   - Test DNS resolution failures

2. **Authentication Failures**

   - Test invalid client credentials
   - Test expired or malformed tokens
   - Test various OAuth error responses

3. **Configuration Validation**
   - Test missing required OAuth labels
   - Test invalid OAuth configuration combinations
   - Test malformed remote server URLs

The testing strategy ensures robust error handling and maintains backward compatibility while validating the new remote server functionality across all supported use cases.
