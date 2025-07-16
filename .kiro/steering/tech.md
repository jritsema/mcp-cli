# Technology Stack & Build System

## Tech Stack

- **Language**: Go 1.24.2
- **CLI Framework**: Cobra (github.com/spf13/cobra v1.9.1)
- **YAML Processing**: gopkg.in/yaml.v3 v3.0.1
- **Build System**: Make + Go toolchain

## Architecture Patterns

- **Command Pattern**: Each CLI command is implemented as a separate Cobra command in the `cmd/` package
- **Configuration as Code**: Uses Docker Compose YAML format for MCP server definitions
- **Environment Variable Expansion**: Supports `${VAR}` and `$VAR` syntax with .env file loading
- **Profile-based Filtering**: Uses Docker Compose labels for organizing servers by use case

## Key Dependencies

```go
require (
    github.com/spf13/cobra v1.9.1    // CLI framework
    gopkg.in/yaml.v3 v3.0.1          // YAML parsing
)
```

## Common Commands

### Development

```bash
make build      # Build binary to ./app
make test       # Run unit tests with race detection
make vet        # Vet code for issues
make start      # Build and run locally
```

### Auto-development

```bash
make autobuild  # Auto-rebuild on file changes (requires reflex)
```

### Container Operations

```bash
make dockerbuild  # Build Docker image
make deploy       # Deploy to cloud dev environment
```

## File Structure Conventions

- `main.go`: Entry point with signal handling
- `cmd/`: All CLI commands and shared types
- `cmd/root.go`: Root command and global flags
- `cmd/types.go`: Shared data structures and utility functions
- Binary output: `./app`
