# Implementation Plan: MCP Server Description Display

## Overview

This plan implements the `-d` flag for the `mcp-cli ls` command to display server descriptions from the `mcp.description` label. The implementation follows the existing patterns in the codebase and builds incrementally.

## Tasks

- [x] 1. Add helper functions for description handling
  - [x] 1.1 Add `GetDescription` function to cmd/types.go
    - Extract description from service labels
    - Return empty string if label not present
    - _Requirements: 1.1, 1.2_
  - [x] 1.2 Add `TruncateDescription` function to cmd/types.go
    - Define `MaxDescriptionLength = 60` constant
    - Truncate strings longer than max and append "..."
    - Return original string if within limit
    - _Requirements: 6.3_
  - [~]\* 1.3 Write property test for description extraction
    - **Property 1: Description extraction preserves content**
    - **Validates: Requirements 1.1, 1.3**
  - [~]\* 1.4 Write property test for truncation logic
    - **Property 2: Truncation applies correctly**
    - **Validates: Requirements 6.3**

- [x] 2. Add description flag to ls command
  - [x] 2.1 Add `showDescription` flag variable in cmd/list.go
    - Add boolean variable declaration
    - Register `-d`/`--description` flag in init()
    - _Requirements: 2.1, 2.3_
  - [x] 2.2 Add flag validation function
    - Create `validateDescriptionFlag` function
    - Check for incompatible combinations with -s, -t, --all-tools
    - Return error with descriptive message
    - _Requirements: 5.1, 5.2, 5.3_
  - [x] 2.3 Call validation in command Run function
    - Add validation call at start of Run function
    - Exit with error if validation fails
    - _Requirements: 5.1, 5.2, 5.3_
  - [~]\* 2.4 Write unit tests for flag validation
    - Test -d with -s returns error
    - Test -d with -t returns error
    - Test -d with --all-tools returns error
    - Test -d alone succeeds
    - _Requirements: 5.1, 5.2, 5.3_

- [x] 3. Checkpoint - Ensure helper functions and flag validation work
  - Ensure all tests pass, ask the user if questions arise.

- [x] 4. Update displayServers function for description output
  - [x] 4.1 Update header output in displayServers
    - Add DESCRIPTION header when showDescription is true
    - Append to end of existing headers for each format
    - Add separator line for description column
    - _Requirements: 2.1, 3.1_
  - [x] 4.2 Update printServerRow to include description
    - Get description using GetDescription helper
    - Apply truncation unless commandFormat is true
    - Append description to row output
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 6.1, 6.2, 6.4_
  - [~]\* 4.3 Write property test for command format full description
    - **Property 3: Command format shows full description**
    - **Validates: Requirements 4.3, 6.4**
  - [~]\* 4.4 Write property test for description column position
    - **Property 4: Description column position**
    - **Validates: Requirements 3.1**

- [x] 5. Update command help text
  - [x] 5.1 Update Long description in listCmd
    - Add documentation for -d flag behavior
    - Document incompatible flag combinations
    - _Requirements: 2.1_

- [x] 6. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- The implementation uses existing patterns from the codebase (tabwriter, flag handling)
- Property tests should use Go's `testing/quick` package
- All changes are contained within cmd/list.go and cmd/types.go
