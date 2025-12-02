# Design Document: Windows Support for MCP CLI

## Overview

This design document outlines the approach for adding comprehensive Windows platform support to the MCP CLI tool. The implementation will enable Windows users to manage MCP server configurations using platform-appropriate file paths, configuration locations, and build artifacts while maintaining backward compatibility with existing Unix-like system support.

The design follows a platform-agnostic approach using Go's standard library capabilities for cross-platform path handling and runtime OS detection. This ensures the codebase remains maintainable while supporting multiple platforms seamlessly.

## Architecture

### Platform Detection Strategy

The MCP CLI will use Go's `runtime.GOOS` constant to detect the operating system at runtime. This approach eliminates the need for user-specified flags and ensures automatic platform-appropriate behavior.

```go
import "runtime"

func getPlatformConfigDir() string {
    if runtime.GOOS == "windows" {
        return getWindowsConfigDir()
    }
    return getUnixConfigDir()
}
```

### Path Handling Strategy

All path operations will use Go's `filepath` package, which automatically handles platform-specific path separators and conventions. The key principles are:

1. **Use `filepath.Join()` for all path construction** - automatically uses correct separators
2. **Use `os.UserHomeDir()` for home directory** - works across platforms
3. **Consistent configuration directories** - use `.config/mcp` on all platforms
4. **Environment variable expansion** - use Unix-style syntax (`$VAR`, `${VAR}`) on all platforms

### Configuration Directory Structure

**All platforms (macOS, Linux, Windows):**
- Config directory: `$HOME/.config/mcp/` (Unix) or `%USERPROFILE%\.config\mcp\` (Windows)
- Compose file: `$HOME/.config/mcp/mcp-compose.yml` (Unix) or `%USERPROFILE%\.config\mcp\mcp-compose.yml` (Windows)
- CLI config: `$HOME/.config/mcp/config.json` (Unix) or `%USERPROFILE%\.config\mcp\config.json` (Windows)

**Note:** All platforms use the same `.config/mcp` directory structure relative to the user's home directory. This simplifies implementation and provides a consistent experience across platforms.

## Components and Interfaces

### 1. Platform Utilities Module (`cmd/platform.go`)

A new file containing platform-specific utility functions:

```go
package cmd

import (
    "os"
    "path/filepath"
    "runtime"
)

// Note: getConfigDir() already exists in cmd/config.go and works for all platforms
// It returns filepath.Join(homeDir, ".config", "mcp") which is cross-platform compatible

// getPlatformToolPath returns the platform-appropriate path for a tool
// Hard fails on error, consistent with getConfigDir() in config.go
func getPlatformToolPath(tool string) string {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error getting user home directory: %v\n", err)
        os.Exit(1)
    }
    
    switch tool {
    case "q-cli":
        return filepath.Join(homeDir, ".aws", "amazonq", "mcp.json")
    case "claude-desktop":
        if runtime.GOOS == "windows" {
            return filepath.Join(homeDir, "AppData", "Roaming", "Claude", "claude_desktop_config.json")
        }
        return filepath.Join(homeDir, "Library", "Application Support", "Claude", "claude_desktop_config.json")
    case "cursor":
        return filepath.Join(homeDir, ".cursor", "mcp.json")
    case "kiro":
        return filepath.Join(homeDir, ".kiro", "settings", "mcp.json")
    default:
        return ""
    }
}
```

**Note:** The existing `expandEnvVars()` in `cmd/env.go` handles Unix-style syntax (`$VAR`, `${VAR}`) which works on all platforms including Windows. No changes needed for environment variable expansion.

### 2. Modified Existing Files

**`cmd/root.go`:**
- No changes needed - `getDefaultComposeFile()` already uses `filepath.Join()` and works cross-platform

**`cmd/config.go`:**
- No changes needed - `getConfigDir()` already exists and works cross-platform
- Already returns `filepath.Join(homeDir, ".config", "mcp")` which works on all platforms

**`cmd/set.go`:**
- Update `toolShortcuts` map to use `getPlatformToolPath()` function
- Ensure path handling uses `filepath` package consistently (review only)

**`cmd/env.go`:**
- No changes needed - `expandEnvVars()` already works on all platforms with Unix-style syntax

**`cmd/types.go`:**
- No changes needed - review to ensure all path operations use `filepath.Join()`

### 3. Build System Updates

**`Makefile`:**
Add Windows build targets:

```makefile
## build-windows-amd64: build Windows AMD64 binary
.PHONY: build-windows-amd64
build-windows-amd64: test
	GOOS=windows GOARCH=amd64 go build -o ./mcp.exe -v

## build-windows-arm64: build Windows ARM64 binary
.PHONY: build-windows-arm64
build-windows-arm64: test
	GOOS=windows GOARCH=arm64 go build -o ./mcp-arm64.exe -v

## build-all: build binaries for all platforms
.PHONY: build-all
build-all: build build-windows-amd64 build-windows-arm64
```

**`.github/workflows/release.yml`:**
Add Windows build steps:

```yaml
- name: Build for Windows (AMD64)
  run: |
    GOOS=windows GOARCH=amd64 go build -o mcp-windows-amd64.exe -ldflags="-X 'main.Version=${{ env.VERSION }}'" .
    zip -j mcp-windows-amd64.zip mcp-windows-amd64.exe

- name: Build for Windows (ARM64)
  run: |
    GOOS=windows GOARCH=arm64 go build -o mcp-windows-arm64.exe -ldflags="-X 'main.Version=${{ env.VERSION }}'" .
    zip -j mcp-windows-arm64.zip mcp-windows-arm64.exe
```

Update release files section to include Windows artifacts.

## Data Models

No new data models are required. Existing structures (`ComposeConfig`, `Service`, `MCPConfig`, `MCPServer`, `CLIConfig`) remain unchanged.

## Error Handling

### Platform-Specific Error Messages

Error messages will include platform context when relevant:

```go
// Note: getConfigDir() already exists in cmd/config.go
// No changes needed - it already works cross-platform:
//
// func getConfigDir() string {
//     homeDir, err := os.UserHomeDir()
//     if err != nil {
//         fmt.Fprintf(os.Stderr, "Error getting user home directory: %v\n", err)
//         os.Exit(1)
//     }
//     return filepath.Join(homeDir, ".config", "mcp")
// }
```

### Path Validation

Add validation for Windows-specific path edge cases:

```go
func validatePath(path string) error {
    // Check for invalid characters on Windows
    if runtime.GOOS == "windows" {
        invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
        for _, char := range invalidChars {
            if strings.Contains(path, char) {
                return fmt.Errorf("path contains invalid character: %s", char)
            }
        }
    }
    return nil
}
```



## Testing Strategy

### Unit Tests

Create platform-specific unit tests in `cmd/platform_test.go`:

```go
// Note: getConfigDir() already exists and doesn't need testing changes
// It already works cross-platform using filepath.Join()

func TestGetConfigDirCrossPlatform(t *testing.T) {
    dir := getConfigDir()
    
    // All platforms should use .config/mcp
    if !strings.Contains(dir, ".config") || !strings.Contains(dir, "mcp") {
        t.Errorf("Expected .config/mcp path, got: %s", dir)
    }
    
    // Verify filepath.Join() used correct separator automatically
    homeDir, _ := os.UserHomeDir()
    expected := filepath.Join(homeDir, ".config", "mcp")
    if dir != expected {
        t.Errorf("Expected %s, got: %s", expected, dir)
    }
}

func TestGetPlatformToolPath(t *testing.T) {
    tests := []struct {
        tool     string
        contains string
    }{
        {"q-cli", ".aws"},
        {"claude-desktop", "Claude"},
        {"cursor", ".cursor"},
        {"kiro", ".kiro"},
    }
    
    for _, tt := range tests {
        t.Run(tt.tool, func(t *testing.T) {
            path := getPlatformToolPath(tt.tool)
            if !strings.Contains(path, tt.contains) {
                t.Errorf("Expected path to contain %s, got: %s", tt.contains, path)
            }
        })
    }
}

// Note: expandEnvVars() already exists in cmd/env.go and has tests
// Only add this test if implementing expandEnvVarsMultiPlatform()

func TestExpandEnvVarsCrossPlatform(t *testing.T) {
    envVars := map[string]string{
        "TEST_VAR": "test_value",
    }
    
    // Test Unix-style expansion (works on all platforms)
    result := expandEnvVars("${TEST_VAR}/path", envVars)
    if !strings.Contains(result, "test_value") {
        t.Errorf("Failed to expand ${VAR} style variable")
    }
    
    result = expandEnvVars("$TEST_VAR/path", envVars)
    if !strings.Contains(result, "test_value") {
        t.Errorf("Failed to expand $VAR style variable")
    }
}
```

### Integration Tests

Update existing integration tests to be platform-aware:

```go
func TestEndToEndWindows(t *testing.T) {
    if runtime.GOOS != "windows" {
        t.Skip("Skipping Windows-specific test")
    }
    
    // Test Windows-specific functionality
    // - Path handling with backslashes
    // - AppData directory usage
    // - .exe binary execution
}
```

### Manual Testing Checklist

1. **Path Handling:**
   - Verify config files are created in `%USERPROFILE%\.config\mcp\` on Windows
   - Test with paths containing spaces
   - Test with UNC paths (\\server\share)
   - Test with different drive letters
   - Verify path separators are correct (backslashes on Windows)

2. **Tool Shortcuts:**
   - Verify each tool shortcut writes to correct Windows location
   - Test Claude Desktop path in `%USERPROFILE%\AppData\Roaming\Claude\`
   - Test Amazon Q CLI path in `%USERPROFILE%\.aws\amazonq\`
   - Test Cursor path in `%USERPROFILE%\.cursor\`
   - Test Kiro path in `%USERPROFILE%\.kiro\settings\`

3. **Environment Variables:**
   - Test $HOME expansion (Unix-style works on Windows too)
   - Test ${USERPROFILE} expansion
   - Verify Unix-style syntax works correctly on Windows

4. **Binary Execution:**
   - Test .exe binary runs on Windows
   - Verify command-line arguments work correctly
   - Test signal handling (Ctrl+C)

## Implementation Phases

### Phase 1: Core Platform Utilities
- Create `cmd/platform.go` with platform detection and path utilities
- Update existing path handling to use new utilities
- Add unit tests for platform utilities

### Phase 2: Configuration Path Updates
- Update `cmd/root.go` to use platform-aware default paths
- Update `cmd/config.go` to use platform config directory
- Update `cmd/set.go` tool shortcuts for Windows paths
- Add integration tests

### Phase 3: Build System Updates
- Add Windows build targets to Makefile
- Update CI/CD workflow to build Windows binaries
- Test cross-compilation from Unix systems

### Phase 4: Documentation and Testing
- Update README with Windows installation instructions
- Add Windows-specific troubleshooting guide
- Perform comprehensive manual testing on Windows
- Update examples to show Windows paths

## Design Decisions and Rationales

### Decision 1: Use Home Directory (.config/mcp) on All Platforms
**Rationale:** Using the same directory structure across all platforms simplifies implementation, reduces platform-specific code, and provides a consistent user experience. The `.config` directory is a well-established convention that works on all platforms. This eliminates the need for `getPlatformConfigDir()` and makes the codebase more maintainable.

### Decision 2: Maintain Backward Compatibility
**Rationale:** Existing Unix users should not be affected. All changes are additive, using runtime detection to choose appropriate behavior.

### Decision 3: Use Unix-style Environment Variable Syntax on All Platforms
**Rationale:** The existing `expandEnvVars()` function supports Unix-style syntax (`$VAR`, `${VAR}`) which works on all platforms including Windows. Most cross-platform CLI tools use Unix-style syntax universally for consistency. No need to add Windows `%VAR%` support.

### Decision 4: Use Go's filepath Package Exclusively
**Rationale:** Ensures consistent, platform-appropriate path handling without manual string manipulation. Reduces bugs and improves maintainability.

### Decision 5: Build Both AMD64 and ARM64 Windows Binaries
**Rationale:** Windows on ARM is growing (Surface devices, VMs). Providing native ARM64 binaries improves performance and user experience.

## Security Considerations

1. **Path Traversal:** Validate user-provided paths to prevent directory traversal attacks
2. **Environment Variable Injection:** Sanitize environment variables before expansion
3. **File Permissions:** Use appropriate permissions (0644 for files, 0755 for directories) on all platforms
4. **UNC Path Handling:** Be cautious with UNC paths to prevent network-based attacks

## Performance Considerations

- Platform detection using `runtime.GOOS` is a compile-time constant, zero runtime overhead
- `filepath.Join()` is optimized and has minimal performance impact
- No additional dependencies required, keeping binary size small

## Compatibility Matrix

| Platform | Architecture | Status | Binary Name |
|----------|-------------|--------|-------------|
| macOS | AMD64 | Existing | mcp-darwin-amd64 |
| macOS | ARM64 | Existing | mcp-darwin-arm64 |
| Linux | AMD64 | Existing | mcp-linux-amd64 |
| Windows | AMD64 | New | mcp-windows-amd64.exe |
| Windows | ARM64 | New | mcp-windows-arm64.exe |
