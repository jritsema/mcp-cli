# Project Structure & Organization

## Directory Layout

```
mcp/
├── main.go                 # Application entry point with signal handling
├── cmd/                    # CLI command implementations
│   ├── root.go            # Root command and global flags
│   ├── types.go           # Shared data structures and utilities
│   ├── list.go            # List/ls command implementation
│   ├── set.go             # Set command implementation
│   ├── clear.go           # Clear command implementation
│   └── config.go          # Config command implementation
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── Makefile               # Build automation
├── mcp-compose.yml        # Example MCP server configuration
├── .env                   # Environment variables (gitignored)
└── app                    # Built binary output
```

## Command Structure

All CLI commands follow the Cobra pattern:

- Each command is a separate file in `cmd/`
- Commands are registered in their `init()` functions
- Shared functionality lives in `cmd/types.go`

## Configuration Files

### User Configuration Locations

- **Compose file**: `$HOME/.config/mcp/mcp-compose.yml` (default)
- **CLI config**: `$HOME/.config/mcp/config.json`
- **Environment**: `.env` file in same directory as compose file

### Tool-specific Output Locations

- **Amazon Q CLI**: `$HOME/.aws/amazonq/mcp.json`
- **Claude Desktop**: `$HOME/Library/Application Support/Claude/claude_desktop_config.json`
- **Cursor**: `$HOME/.cursor/mcp.json`

## Data Flow

1. Load `mcp-compose.yml` (Docker Compose format)
2. Load environment variables from system + `.env` file
3. Filter services by profile using `mcp.profile` labels
4. Transform to MCP JSON format with environment expansion
5. Write to tool-specific configuration location

## Key Conventions

- **Profile organization**: Use `mcp.profile` labels for grouping servers
- **Environment expansion**: Support `${VAR}` and `$VAR` syntax
- **Default behavior**: Services without profiles are considered "default"
- **Container support**: Services can use `command` or `image` fields
- **Error handling**: Exit with status 1 on errors, write to stderr
