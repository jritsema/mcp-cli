# MCP CLI

MCP CLI is a tool for managing MCP server configuration files.

## Why?

Model Context Protocol (MCP) is a new technology and still evolving. As I've been using it, I have encountered several pain points:

- Manually editing JSON files
- Managing similar config files for different AI tools
- Dealing with secret envvars
- Experimenting with new MCP servers
- Switching between different configurations based on what I'm doing (e.g., programming, writing, researching)

I decided to write up some specs for a tool (written in Go) that could help with these pain points and try to "vibe code" it. This is the result. Please don't judge the code quality. I didn't write or edit a single line :)

## Usage

MCP CLI simplifies managing MCP server configurations through a YAML-based approach.

### Getting Started

1. Create an [mcp-compose.yml](./mcp-compose.yml) file with your MCP server configurations. You can place this file in either:
   - Your current working directory (for project-specific configurations)
   - `$HOME/.config/mcp/mcp-compose.yml` (for global configurations)

```sh
# For global configuration
mkdir -p ~/.config/mcp
cp ./mcp-compose.yml $HOME/.config/mcp/
```

2. Use the CLI to manage and deploy these configurations to your favorite AI tools

```sh
mcp set -t q-cli # or -t cursor, -t claude-desktop
```

### Configuration File Resolution

MCP CLI automatically looks for configuration files in the following order:

1. **Local directory**: `./mcp-compose.yml` in your current working directory
2. **Global directory**: `$HOME/.config/mcp/mcp-compose.yml` in your home config directory
3. **Custom path**: Use the `-f` flag to specify a custom location

This allows you to have project-specific MCP server configurations that override your global settings when working in specific directories.

```sh
# Uses local mcp-compose.yml if it exists, otherwise falls back to global
mcp ls

# Explicitly use a custom configuration file
mcp ls -f ./custom-mcp-compose.yml
```

### Listing MCP Servers

View available MCP servers defined in your configuration:

```sh
# List default MCP servers
mcp ls

# List all MCP servers
mcp ls -a

# List servers with specific profile
mcp ls programming

# Use a custom configuration file
mcp ls -f ./custom-mcp-compose.yml
```

The output format shows NAME, PROFILES, COMMAND, and ENVVARS columns.

### Setting MCP Configurations

Deploy your MCP server configurations to supported tools:

```sh
# Set default servers for Amazon Q CLI
mcp set -t q-cli

# Set programming profile servers for Cursor
mcp set programming -t cursor

# Set a specific server for Claude Desktop
mcp set -t claude-desktop -s github

# Set programming profile servers for Kiro IDE
mcp set programming -t kiro

# Use a custom output location
mcp set -c /path/to/output/mcp.json
```

### Clearing MCP Configurations

Remove all MCP servers from a configuration:

```sh
# Clear all servers from Amazon Q CLI configuration
mcp clear -t q-cli

# Clear from a custom output location
mcp clear -c /path/to/output/mcp.json
```

### Tool Shortcuts

MCP CLI supports these predefined tool shortcuts for popular AI tools:

- `q-cli` - Amazon Q CLI (`$HOME/.aws/amazonq/mcp.json`)
- `claude-desktop` - Claude Desktop (`$HOME/Library/Application Support/Claude/claude_desktop_config.json`)
- `cursor` - Cursor IDE (`$HOME/.cursor/mcp.json`)
- `kiro` - Kiro IDE (`$HOME/.kiro/settings/mcp.json`)

### Setting Default AI Tool

Configure a default AI tool to avoid specifying `-t` each time:

```sh
# Set Amazon Q CLI as your default tool
mcp config set tool ~/.aws/amazonq/mcp.json

# Now you can simply run:
mcp set programming

# or to switch back to defaults
mcp set
```

### Setting Container Tool

If you're using containers to run your MCP servers (by setting the `image` property), then MCP CLI will output `docker` run commands by default. If you're using a different container tool such as `finch` or `podman`, etc., then you can use the `set container-tool` command.

```sh
# Set a custom container tool (default is docker)
mcp config set container-tool finch
```

### Profiles

Organize your MCP servers with profiles using the `labels` field in your `mcp-compose.yml`:

```yaml
services:
  brave:
    image: mcp/brave-search
    environment:
      BRAVE_API_KEY: ${BRAVE_API_KEY}
    labels:
      mcp.profile: research

  github:
    command: npx -y @modelcontextprotocol/server-github
    labels:
      mcp.profile: programming
```

Then deploy only those servers:

```sh
mcp set programming -t claude-desktop
```

Services without the label are considered defaults.

## How?

It turns out that the Docker Compose (`docker-compose.yml`) specification already has good support for MCP stdio configuration where services map to MCP servers with `command`s, `image`s, `environment`s/`env_files`s, and `label`s for profiles. Another added benefit of this is you can run `docker compose pull -f mcp-compose.yml` and it will pre-fetch all the container images.

Example:

```yaml
# MCP Servers
services:
  time:
    command: uvx mcp-server-time

  fetch:
    command: uvx mcp-server-fetch

  github:
    command: npx -y @modelcontextprotocol/server-github
    environment:
      GITHUB_PERSONAL_ACCESS_TOKEN: ${GITHUB_PERSONAL_ACCESS_TOKEN}
    labels:
      mcp.profile: programming

  aws-docs:
    command: uvx awslabs.aws-documentation-mcp-server@latest
    environment:
      FASTMCP_LOG_LEVEL: "ERROR"
    labels:
      mcp.profile: programming

  postgres:
    command: npx -y @modelcontextprotocol/server-postgres postgresql://localhost/mydb
    labels:
      mcp.profile: database

  # OR container based MCP servers

  github-docker:
    image: ghcr.io/github/github-mcp-server
    environment:
      GITHUB_PERSONAL_ACCESS_TOKEN: ${GITHUB_PERSONAL_ACCESS_TOKEN}
    labels:
      mcp.profile: programming

  brave:
    image: mcp/brave-search
    environment:
      BRAVE_API_KEY: ${BRAVE_API_KEY}
    labels:
      mcp.profile: programming, research
```

## Development

```
 Choose a make command to run

  vet           vet code
  test          run unit tests
  build         build a binary
  autobuild     auto build when source files change
  dockerbuild   build project into a docker container image
  start         build and run local project
  deploy        build code into a container and deploy it to the cloud dev environment
```
