# MCP CLI Product Overview

MCP CLI is a command-line tool for managing Model Context Protocol (MCP) server configurations. It simplifies the pain points of working with MCP by providing a YAML-based configuration approach inspired by Docker Compose.

## Core Problems Solved

- Eliminates manual JSON file editing for MCP configurations
- Manages similar config files across different AI tools (Amazon Q CLI, Claude Desktop, Cursor)
- Handles secret environment variables securely
- Enables experimentation with new MCP servers
- Supports switching between different configurations based on context (programming, writing, research)

## Key Features

- **Profile-based organization**: Group MCP servers by use case using labels
- **Multi-tool support**: Deploy configurations to various AI tools via shortcuts
- **Environment variable expansion**: Secure handling of secrets and dynamic values
- **Container support**: Works with both command-based and Docker image-based MCP servers
- **Default configurations**: Sensible defaults with override capabilities

## Target Users

Developers and AI tool users who work with multiple MCP servers and need efficient configuration management across different AI tools and contexts.
