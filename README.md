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

## Installation

### macOS and Linux

Download the appropriate binary for your platform from the [releases page](https://github.com/
jritsema/mcp-cli/releases):

- macOS AMD64: `mcp-darwin-amd64`
- macOS ARM64: `mcp-darwin-arm64`
- Linux AMD64: `mcp-linux-amd64`

Make the binary executable and move it to your PATH:

```sh
chmod +x mcp-darwin-arm64
sudo mv mcp-darwin-arm64 /usr/local/bin/mcp
```

### Windows

Download the appropriate Windows binary for your architecture from the [releases page](https://github.com/
jritsema/mcp-cli/releases):

- **Windows AMD64** (most common): `mcp-windows-amd64.zip`
- **Windows ARM64** (Surface devices, ARM VMs): `mcp-windows-arm64.zip`

#### Installation Steps

1. Download the appropriate `.zip` file for your architecture
2. Extract the `.exe` file from the archive
3. Move `mcp.exe` to a directory in your PATH, or add the directory to your PATH

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

- `q-cli` - Amazon Q CLI
  - macOS/Linux: `$HOME/.aws/amazonq/mcp.json`
  - Windows: `%USERPROFILE%\.aws\amazonq\mcp.json`
- `claude-desktop` - Claude Desktop
  - macOS: `$HOME/Library/Application Support/Claude/claude_desktop_config.json`
  - Windows: `%USERPROFILE%\AppData\Roaming\Claude\claude_desktop_config.json`
- `cursor` - Cursor IDE
  - macOS/Linux: `$HOME/.cursor/mcp.json`
  - Windows: `%USERPROFILE%\.cursor\mcp.json`
- `kiro` - Kiro IDE
  - macOS/Linux: `$HOME/.kiro/settings/mcp.json`
  - Windows: `%USERPROFILE%\.kiro\settings\mcp.json`

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

### Remote MCP Servers

MCP CLI supports remote MCP servers that use `Streamable HTTP` transport with OAuth 2.0 authentication. Remote servers are identified by URLs starting with `https://` in the command field.

#### Configuration

To configure a remote MCP server, use the following format in your `mcp-compose.yml`:

```yaml
services:
  my-remote-server:
    command: https://my-remote-server.gateway.bedrock-agentcore.us-east-1.amazonaws.com/mcp
    labels:
      mcp.grant-type: client_credentials
      mcp.token-endpoint: https://my-app.auth.us-east-1.amazoncognito.com/oauth2/token
      mcp.client-id: ${REMOTE_CLIENT_ID}
      mcp.client-secret: ${REMOTE_CLIENT_SECRET}
```

#### Required Labels for Remote Servers

- `mcp.grant-type`: Must be "client_credentials"
- `mcp.token-endpoint`: OAuth 2.0 token endpoint URL
- `mcp.client-id`: OAuth client identifier (supports environment variable expansion)
- `mcp.client-secret`: OAuth client secret (supports environment variable expansion)

#### Environment Variables

Set your OAuth credentials in your environment or `.env` file:

```bash
REMOTE_CLIENT_ID=your_client_id_here
REMOTE_CLIENT_SECRET=your_client_secret_here
```

#### Tool Support

Remote MCP servers are currently supported by:

- `kiro` - Kiro IDE
- `q-cli` - Amazon Q CLI


#### Authentication Flow

When deploying remote servers, MCP CLI will:

1. Validate the OAuth configuration
2. Acquire an access token using the client credentials flow
3. Generate MCP configuration with HTTP transport and authorization headers

```sh
# Deploy remote servers (will show "acquiring access token..." message)
mcp set
```

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

## Windows Troubleshooting

### Path Issues

**Problem: Configuration files not found**

MCP CLI uses Unix-style paths internally but works with Windows paths. If you're having issues:

```powershell
# Verify your home directory
echo $env:USERPROFILE

# Check if config directory exists
Test-Path "$env:USERPROFILE\.config\mcp"

# Create it if needed
New-Item -ItemType Directory -Path "$env:USERPROFILE\.config\mcp" -Force
```

**Problem: UNC paths not working**

UNC paths (e.g., `\\server\share\path`) are supported but may require proper network permissions:

```powershell
# Test UNC path access
Test-Path "\\server\share\mcp-compose.yml"

# Use mapped drive as alternative
net use Z: \\server\share
mcp ls -f Z:\mcp-compose.yml
```

**Problem: Drive letter issues**

Always use absolute paths with drive letters on Windows:

```powershell
# Good
mcp ls -f C:\projects\mcp-compose.yml

# Also works with forward slashes
mcp ls -f C:/projects/mcp-compose.yml
```

### Environment Variables

**Unix-style syntax on Windows**

MCP CLI uses Unix-style environment variable syntax (`$VAR` or `${VAR}`) on all platforms, including Windows:

```yaml
# In mcp-compose.yml - use Unix-style syntax even on Windows
services:
  github:
    command: npx -y @modelcontextprotocol/server-github
    environment:
      GITHUB_TOKEN: ${GITHUB_TOKEN}  # Correct
      # NOT: %GITHUB_TOKEN%          # Wrong - don't use Windows syntax
```

**Setting environment variables**

Create a `.env` file in `%USERPROFILE%\.config\mcp\`:

```bash
# .env file (use Unix-style syntax)
GITHUB_TOKEN=ghp_your_token_here
BRAVE_API_KEY=your_api_key_here
```

Or set them in PowerShell:

```powershell
# Temporary (current session only)
$env:GITHUB_TOKEN = "ghp_your_token_here"

# Permanent (user-level)
[Environment]::SetEnvironmentVariable("GITHUB_TOKEN", "ghp_your_token_here", "User")
```

### Path Separators

MCP CLI automatically handles path separators. You can use either forward slashes or backslashes:

```powershell
# Both work on Windows
mcp ls -f C:\projects\mcp-compose.yml
mcp ls -f C:/projects/mcp-compose.yml
```

### Common Configuration Examples

**Example 1: Basic Windows setup**

```powershell
# Create config directory
New-Item -ItemType Directory -Path "$env:USERPROFILE\.config\mcp" -Force

# Create mcp-compose.yml
@"
services:
  time:
    command: uvx mcp-server-time
  
  fetch:
    command: uvx mcp-server-fetch
"@ | Out-File -FilePath "$env:USERPROFILE\.config\mcp\mcp-compose.yml" -Encoding UTF8
```

**Example 2: With environment variables**

```powershell
# Create .env file
@"
GITHUB_TOKEN=ghp_your_token_here
BRAVE_API_KEY=your_api_key_here
"@ | Out-File -FilePath "$env:USERPROFILE\.config\mcp\.env" -Encoding UTF8

# Create mcp-compose.yml with env vars
@"
services:
  github:
    command: npx -y @modelcontextprotocol/server-github
    environment:
      GITHUB_PERSONAL_ACCESS_TOKEN: `${GITHUB_TOKEN}
    labels:
      mcp.profile: programming
  
  brave:
    image: mcp/brave-search
    environment:
      BRAVE_API_KEY: `${BRAVE_API_KEY}
    labels:
      mcp.profile: research
"@ | Out-File -FilePath "$env:USERPROFILE\.config\mcp\mcp-compose.yml" -Encoding UTF8
```

**Example 3: Project-specific configuration**

```powershell
# In your project directory
cd C:\projects\myproject

# Create local mcp-compose.yml (overrides global config)
@"
services:
  postgres:
    command: npx -y @modelcontextprotocol/server-postgres postgresql://localhost/mydb
    labels:
      mcp.profile: database
"@ | Out-File -FilePath ".\mcp-compose.yml" -Encoding UTF8

# Use it
mcp ls
```

### Tool-Specific Issues

**Claude Desktop not found**

Verify Claude Desktop is installed and the config directory exists:

```powershell
Test-Path "$env:USERPROFILE\AppData\Roaming\Claude"

# Create directory if needed
New-Item -ItemType Directory -Path "$env:USERPROFILE\AppData\Roaming\Claude" -Force
```

**Amazon Q CLI not found**

Verify AWS directory structure:

```powershell
Test-Path "$env:USERPROFILE\.aws\amazonq"

# Create directory if needed
New-Item -ItemType Directory -Path "$env:USERPROFILE\.aws\amazonq" -Force
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
