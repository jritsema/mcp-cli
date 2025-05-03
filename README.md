# MCP CLI

MCP CLI is a tool for managing MCP server configuration files.

## Why?

MCP is a new technology and still evolving and as I've been using it, I have a felt several pain points:

- Manually editing JSON files
- Experimenting with new MCP servers
- Writing my own MCP servers
- Switching between different configurations based on what I'm doing (e.g., programming, writing, researching)
- Lack of server profiles

## What?

I decided to write up some specs for a tool that could help with these pain points and "vibe code" it in a couple of hours.

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

```sh
go mod init app
```

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
