# MCP CLI

MCP CLI is a tool for managing MCP server configuration files.

## Why?

Model Context Protocol (MCP) is a new technology and still evolving.  As I've been using it, I have encountered several pain points:

- Manually editing JSON files
- Dealing with secret envvars
- Experimenting with new MCP servers
- Writing my own MCP servers
- Switching between different configurations based on what I'm doing (e.g., programming, writing, researching)
- Lack of server profiles

I decided to write up some specs for a tool (written in Go) that could help with these pain points and try to "vibe code" it.  This is the result. Please don't judge the code quality. I didn't write or edit a single line :)

## Usage

MCP CLI simplifies managing MCP server configurations through a YAML-based approach.

### Getting Started

1. Create an [mcp-compose.yml](./mcp-compose.yml) file in your home directory with your MCP server configurations.  See example below, or copy the included example file.

```sh
cp ./mcp-compose.yml ~/
```

2. Use the CLI to manage and deploy these configurations to your favorite AI tools

```sh
mcp set -t q-cli # or -t cursor, -t claude-desktop
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

### Setting MCP Configurations

Deploy your MCP server configurations to supported tools:

```sh
# Set default servers for Amazon Q CLI
mcp set -t q-cli

# Set programming profile servers for Cursor
mcp set programming -t cursor

# Set a specific server for Claude Desktop
mcp set -t claude-desktop -s github

# Use a custom output location
mcp set -c /path/to/output/mcp.json
```

### Setting Default Tool

Configure a default tool to avoid specifying `-t` each time:

```sh
# Set Amazon Q CLI as your default tool
mcp config set tool ~/.aws/amazonq/mcp.json

# Now you can simply run:
mcp set programming
# or to switch back to defaults
mcp set
```

### Tool Shortcuts

MCP CLI supports these predefined tool shortcuts:

- `q-cli` - Amazon Q CLI (`$HOME/.aws/amazonq/mcp.json`)
- `claude-desktop` - Claude Desktop (`$HOME/Library/Application Support/Claude/claude_desktop_config.json`)
- `cursor` - Cursor IDE (`$HOME/.cursor/mcp.json`)

### Profiles

Organize your MCP servers with profiles using the `labels` field in your `mcp-compose.yml`:

```yaml
services:
  github:
    command: npx -y @modelcontextprotocol/server-github
    labels:
      mcp.profile: programming
```

Then deploy only those servers:

```sh
mcp set programming -t q-cli
```

Services without the label are considered defaults.


## How?

So turns out that the Docker Compose (`docker-compose.yml`) specification already has good support for MCP stdio configuration where services map to MCP servers with `command`s, `image`s, `environment`s/`env_files`s, and `label`s for profiles. Another added benefit of this is you can run `docker compose pull -f mcp-compose.yml` and it will pre-fetch all the container images.

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
