# Instructions

Implement the code for the CLI tool described here and in `README.md` using the https://github.com/spf13/cobra library.

Here's an example of a `mcp-compose.yml` file.

```yaml
# MCP Servers
services:

  time:
    command: uvx mcp-server-time

  fetch:
    command: uvx mcp-server-fetch

  fs:
  command: npx -y @modelcontextprotocol/server-filesystem /Users/username/Desktop
  labels:
    mcp.profile: default, programming

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


The tool should have the following commands:

## CLI Commands

### Listing servers

- `mcp ls` - lists all default mcp servers in `mcp-compose.yml`. Defaults are considered entries with either no profile label or a profile label that contains `default`.

example output

| NAME  | PROFILES             | COMMAND                                                                | ENVVARS |
| ----- | -------------------- | ---------------------------------------------------------------------- | ------- |
| time  | default              | uvx mcp-server-time                                                    |         |
| fetch | default              | uvx mcp-server-fetch                                                   |         |
| fs    | default, programming | npx -y @modelcontextprotocol/server-filesystem /Users/username/Desktop |         |


- `mcp ls -a` - lists all mcp servers in `mcp-compose.yml`

example output

| NAME          | PROFILES              | COMMAND                                                                             | ENVVARS                      |
|---------------|-----------------------|-------------------------------------------------------------------------------------|------------------------------|
| time          | default               | uvx mcp-server-time                                                                 |                              |
| fetch         | default               | uvx mcp-server-fetch                                                                |                              |
| fs            | default, programming  | npx -y @modelcontextprotocol/server-filesystem /Users/username/Desktop              |                              |
| github        | programming           | npx -y @modelcontextprotocol/server-github                                          | GITHUB_PERSONAL_ACCESS_TOKEN |
| aws-docs      | programming           | uvx awslabs.aws-documentation-mcp-server@latest                                     | FASTMCP_LOG_LEVEL            |
| postgres      | database              | npx -y @modelcontextprotocol/server-postgres                                        |                              |
| github-docker | programming           | docker run -i --rm -e GITHUB_PERSONAL_ACCESS_TOKEN ghcr.io/github/github-mcp-server | GITHUB_PERSONAL_ACCESS_TOKEN |
| brave         | programming, research | docker run -i --rm -e BRAVE_API_KEY mcp/brave-search                                | BRAVE_API_KEY                |


- `mcp ls <profile>` - lists all mcp servers in `mcp-compose.yml` with a `label` of `<profile>`

example output `mcp ls programming`

| NAME          | PROFILES              | COMMAND                                                                             | ENVVARS                      |
|---------------|-----------------------|-------------------------------------------------------------------------------------|------------------------------|
| fs            | default, programming  | npx -y @modelcontextprotocol/server-filesystem /Users/username/Desktop              |                              |
| github        | programming           | npx -y @modelcontextprotocol/server-github                                          | GITHUB_PERSONAL_ACCESS_TOKEN |
| aws-docs      | programming           | uvx awslabs.aws-documentation-mcp-server@latest                                     | FASTMCP_LOG_LEVEL            |
| github-docker | programming           | docker run -i --rm -e GITHUB_PERSONAL_ACCESS_TOKEN ghcr.io/github/github-mcp-server | GITHUB_PERSONAL_ACCESS_TOKEN |
| brave         | programming, research | docker run -i --rm -e BRAVE_API_KEY mcp/brave-search                                | BRAVE_API_KEY                |


- By default the tool should look for the `mcp-compose.yml` file in the MCP CLI home directory, which is `$HOME/.config/mcp/`, however the location can be overridden using the `-f` flag. for example:

```
mcp ls -f ./mcp-compose.yml
```

### Setting configurations

- `mcp set <profile> -c <mcp.json>` - writes an MCP JSON file to the specified location using only the servers with a `label` matching `<profile>`. If `<profile>` is not specified, then look only at default servers (servers that either don't have a profile label or one that contains `default`).

Rather than explicitly specifying the `-c` location of the MCP config file you want to write to, the user can use `tool shortcuts` using the `-t` flag for well-known tools that support MCP. Here are the shortcut mappings:

- Amazon Q CLI - `q-cli` - `$HOME/.aws/amazonq/mcp.json`
- Claude Desktop - `claude-desktop` - `$HOME/Library/Application Support/Claude/claude_desktop_config.json`
- Cursor - `cursor` - `$HOME/.cursor/mcp.json`

Example output

write defaults

```
mcp set -t q-cli
Wrote /Users/john/.aws/amazonq/mcp.json
```

write programming

```
mcp set programming -t cursor
Wrote /Users/john/.cursor/mcp.json
```

You can also the `-s` flag to specify setting only a single MCP server.

```
mcp set -t q-cli -s my-specific-tool
Wrote /Users/john/.aws/amazonq/mcp.json
```

And then switch back to the defaults when you're done testing

```
mcp set -t q-cli
Wrote /Users/john/.aws/amazonq/mcp.json
```

### Clearing configurations

- `mcp clear -t <tool>` or `mcp clear -c <mcp.json>` - removes all MCP servers from the specified configuration file.

Example output:

```
mcp clear -t q-cli
Cleared all servers from /Users/john/.aws/amazonq/mcp.json
```

### Setting default tool

- `mcp config set tool ~/.aws/amazonq/mcp.json` - this command sets the `tool` config value in the MCP CLI's config file located in `~/.config/mcp/config.json`.  This file looks like:

```json
{
  "tool": "/Users/<user>/.aws/amazonq/mcp.json"  
}
```

After setting this default tool, you no longer need to specify the `-t` flag.

```
mcp set programming

# then switch back to defaults
mcp set
```


## Output

The output MCP JSON file would look like this.

```json
{
  "mcpServers": {
    "time": {
      "command": "uvx",
      "args": ["mcp-server-time"]
    },
    "fetch": {
      "command": "uvx",
      "args": ["mcp-server-fetch"]
    },
    "fs": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/Users/username/Desktop"]
    },
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "${GITHUB_PERSONAL_ACCESS_TOKEN}"
      }
    },
    "aws-docs": {
      "command": "uvx",
      "args": ["awslabs.aws-documentation-mcp-server@latest"],
      "env": {
        "FASTMCP_LOG_LEVEL": "ERROR"
      }
    },
    "postgres": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-postgres", "postgresql://localhost/mydb"]
    },
    "github-docker": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "-e", "GITHUB_PERSONAL_ACCESS_TOKEN=${GITHUB_PERSONAL_ACCESS_TOKEN}", "ghcr.io/github/github-mcp-server"]
    },
    "brave": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "-e", "BRAVE_API_KEY=${BRAVE_API_KEY}", "mcp/brave-search"]
    }
  }
}
```


# Guidelines

Before making any code changes, first checkout a new git branch.  After updating code, run `make build` to ensure that the code compiles. Then, after testing, git commit your changes with a concise short message on one line along with a longer summary message.
