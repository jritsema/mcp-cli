# Requirements Document

## Introduction

This feature adds support for displaying MCP server descriptions in the `mcp-cli ls` command output. Developers can add descriptions to their MCP servers using a new `mcp.description` label in the YAML configuration, and view these descriptions using a new `-d` flag on the ls command.

## Glossary

- **MCP_CLI**: The command-line interface tool for managing MCP server configurations
- **Service**: An MCP server definition in the mcp-compose.yml file
- **Label**: A key-value metadata field attached to a service (e.g., `mcp.profile`, `mcp.description`)
- **Description**: A human-readable text explaining what an MCP server does
- **Tabwriter**: Go's text/tabwriter package used for aligned tabular output

## Requirements

### Requirement 1: Description Label Support

**User Story:** As a developer, I want to add descriptions to my MCP servers in the YAML configuration, so that I can document what each server does.

#### Acceptance Criteria

1. WHEN a service has a `mcp.description` label THEN THE MCP_CLI SHALL parse and store the description value
2. WHEN a service does not have a `mcp.description` label THEN THE MCP_CLI SHALL treat the description as empty
3. THE MCP_CLI SHALL support description values containing spaces, punctuation, and special characters

### Requirement 2: Description Flag for List Command

**User Story:** As a developer, I want to use a `-d` flag with the ls command, so that I can see server descriptions in the output.

#### Acceptance Criteria

1. WHEN the user runs `mcp-cli ls -d` THEN THE MCP_CLI SHALL display a DESCRIPTION column at the end of the output
2. WHEN the user runs `mcp-cli ls` without the `-d` flag THEN THE MCP_CLI SHALL NOT display the description column
3. THE MCP_CLI SHALL support both `-d` and `--description` as flag variants

### Requirement 3: Description Column Positioning

**User Story:** As a developer, I want the description column to appear at the end of the output, so that it doesn't disrupt the existing column layout.

#### Acceptance Criteria

1. WHEN displaying output with `-d` flag THEN THE MCP_CLI SHALL append the DESCRIPTION column after all other columns

### Requirement 4: Flag Combination Support

**User Story:** As a developer, I want to combine the `-d` flag with compatible ls flags, so that I can view descriptions alongside other information.

#### Acceptance Criteria

1. WHEN the user combines `-d` with `-a` THEN THE MCP_CLI SHALL show truncated descriptions for all servers
2. WHEN the user combines `-d` with `-l` THEN THE MCP_CLI SHALL show truncated descriptions in long format
3. WHEN the user combines `-d` with `-c` THEN THE MCP_CLI SHALL show full descriptions with command format
4. WHEN the user combines `-d` with a profile argument THEN THE MCP_CLI SHALL show truncated descriptions for filtered servers

### Requirement 5: Incompatible Flag Validation

**User Story:** As a developer, I want clear error messages when I use incompatible flag combinations, so that I understand how to use the CLI correctly.

#### Acceptance Criteria

1. WHEN the user combines `-d` with `-s` THEN THE MCP_CLI SHALL display an error message indicating incompatible flags
2. WHEN the user combines `-d` with `-t` THEN THE MCP_CLI SHALL display an error message indicating incompatible flags
3. WHEN the user combines `-d` with `--all-tools` THEN THE MCP_CLI SHALL display an error message indicating incompatible flags

### Requirement 6: Description Display Formatting

**User Story:** As a developer, I want descriptions to be displayed in a readable format, so that I can easily understand what each server does.

#### Acceptance Criteria

1. WHEN a server has a description THEN THE MCP_CLI SHALL display the description text in the DESCRIPTION column
2. WHEN a server has no description THEN THE MCP_CLI SHALL display an empty value in the DESCRIPTION column
3. WHEN using `-d` without `-c` and a description exceeds 60 characters THEN THE MCP_CLI SHALL truncate it and append "..." to indicate truncation
4. WHEN using `-d` with `-c` THEN THE MCP_CLI SHALL display the full description without truncation
5. THE MCP_CLI SHALL preserve the tabular alignment when displaying descriptions of varying lengths
