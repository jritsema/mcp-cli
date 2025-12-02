# Implementation Plan: Windows Support

- [x] 1. Create platform utilities module





  - Create `cmd/platform.go` with `getPlatformToolPath()` function
  - Function should handle platform-specific tool paths (Claude Desktop differs between Windows/macOS)
  - Hard fail on home directory errors, consistent with `getConfigDir()` in `config.go`
  - _Requirements: 2.4, 2.5, 2.6, 6.1, 6.2, 6.3_

- [x] 2. Update set command to use platform utilities





  - Modify `cmd/set.go` to replace hardcoded `toolShortcuts` map with calls to `getPlatformToolPath()`
  - Update `getOutputPath()` function to use the new platform utilities
  - Verify all path operations use `filepath.Join()` for cross-platform compatibility
  - _Requirements: 1.2, 1.5, 2.4, 2.5, 2.6_

- [x] 3. Add Windows build targets to Makefile









  - Add `build-windows-amd64` target for Windows AMD64 architecture
  - Add `build-windows-arm64` target for Windows ARM64 architecture
  - Add `build-all` target to build binaries for all platforms
  - Ensure `.exe` extension is used for Windows binaries
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

- [x] 4. Update CI/CD pipeline for Windows builds





  - Modify `.github/workflows/release.yml` to build Windows AMD64 binaries
  - Add Windows ARM64 binary build step
  - Package Windows binaries as `.zip` files for release
  - Update release artifact upload to include Windows executables
  - Ensure Windows binaries have clear naming conventions (e.g., `mcp-windows-amd64.exe`)
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

- [ ] 5. Add platform-specific tests
- [ ] 5.1 Create unit tests for `getPlatformToolPath()`
  - Test each tool shortcut returns correct path
  - Verify Claude Desktop path differs between Windows and macOS
  - Test error handling when home directory cannot be determined
  - _Requirements: 2.4, 2.5, 2.6_

- [ ] 5.2 Add cross-platform integration tests
  - Test config directory resolution on Windows
  - Test tool path resolution for all supported tools
  - Verify path separators are correct for the platform
  - _Requirements: 1.1, 1.2, 1.5, 6.1, 6.2, 6.3_

- [x] 6. Update documentation






- [x] 6.1 Add Windows installation instructions to README

  - Document how to download and install Windows binaries
  - Include instructions for both AMD64 and ARM64 architectures
  - Add Windows-specific configuration examples
  - _Requirements: 1.1, 3.1, 3.2_

- [x] 6.2 Create Windows troubleshooting guide


  - Document common Windows path issues (UNC paths, drive letters)
  - Explain Unix-style environment variable syntax usage on Windows
  - Add examples of Windows-specific configurations
  - _Requirements: 5.1, 5.2, 5.3, 5.4_
