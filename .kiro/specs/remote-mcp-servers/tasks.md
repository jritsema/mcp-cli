# Implementation Plan

- [x] 1. Add remote server detection and validation

  - Implement `isRemoteServer()` function to detect HTTPS URLs in command field
  - Add validation for required OAuth labels on remote servers
  - Add tool compatibility validation for remote servers
  - _Requirements: 1.1, 4.1, 4.4, 5.2_

- [x] 2. Implement OAuth 2.0 client functionality

  - [x] 2.1 Create OAuth data structures and configuration parsing

    - Define `OAuthConfig` and `OAuthResponse` structs
    - Implement function to extract OAuth configuration from service labels
    - Add environment variable expansion for OAuth credentials
    - _Requirements: 1.3, 1.4, 4.5_

  - [x] 2.2 Implement OAuth token acquisition

    - Create `acquireAccessToken()` function with HTTP client
    - Implement POST request with `application/x-www-form-urlencoded` content type
    - Add JSON response parsing for access token extraction
    - Add user feedback message "acquiring access token..."
    - _Requirements: 2.1, 2.2, 2.3, 2.4_

  - [ ]\* 2.3 Add comprehensive OAuth error handling
    - Handle network errors with descriptive messages
    - Handle authentication failures (401, 403 responses)
    - Handle malformed JSON responses
    - Handle timeout scenarios
    - _Requirements: 2.5, 2.6, 4.2, 4.3_

- [x] 3. Enhance MCP configuration generation

  - [x] 3.1 Extend MCPServer struct for remote servers

    - Add `Type`, `URL`, and `Headers` fields to `MCPServer` struct
    - Ensure backward compatibility with existing command-based fields
    - _Requirements: 3.1, 3.2_

  - [x] 3.2 Update convertToMCPConfig for mixed server types
    - Modify `convertToMCPConfig()` to handle both local and remote servers
    - Generate HTTP-based configuration for remote servers
    - Set Authorization header with Bearer token for remote servers
    - Maintain existing logic for local command-based and container servers
    - _Requirements: 3.3, 3.6_

- [x] 4. Add tool support validation and restrictions

  - Implement tool compatibility checking for remote servers
  - Add `remoteSupportedTools` map with kiro and q-cli support
  - Update `getOutputPath()` to validate tool compatibility
  - Add clear error messages for unsupported tool combinations
  - _Requirements: 3.4, 3.5_

- [x] 5. Update command processing and integration

  - [x] 5.1 Integrate remote server processing in set command

    - Modify set command to detect and process remote servers
    - Add OAuth token acquisition step for remote servers
    - Ensure proper error handling and user feedback
    - _Requirements: 2.1, 2.2, 5.1_

  - [x] 5.2 Update list command for remote server display
    - Modify `displayServers()` to show remote server information
    - Update long format display to indicate remote vs local servers
    - Ensure profile filtering works with remote servers
    - _Requirements: 5.1, 5.3_

- [ ]\* 6. Add comprehensive error handling and validation

  - Add validation for malformed remote server URLs
  - Add checks for missing environment variables in OAuth credentials
  - Add validation for required OAuth label combinations
  - Ensure all error messages include helpful context and exit with status 1
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

- [x] 7. Ensure backward compatibility and integration

  - [x] 7.1 Test mixed configuration scenarios

    - Verify local and remote servers work together in same configuration
    - Test profile filtering with mixed server types
    - Ensure existing environment variable expansion continues to work
    - _Requirements: 5.4, 5.5_

  - [x] 7.2 Validate existing functionality preservation
    - Ensure all existing commands work unchanged with local servers
    - Verify container-based servers continue to work as before
    - Test existing tool shortcuts remain functional for local servers
    - _Requirements: 5.1, 5.2, 5.3, 5.5_
