# Chimera

Aggregates multiple MCP servers (stdio and HTTP) into a single HTTP endpoint.

## Quick Start

## Configuration

Uses VSCode MCP config format:

```json
{
  "servers": {
    "filesystem": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"],
      "env": { "KEY": "value" }
    },
    "api-server": {
      "type": "http",
      "url": "http://localhost:3000/mcp",
      "headers": { "Authorization": "Bearer token" }
    }
  }
}
```

## Features

- **Auto-prefixing**: Prevents name conflicts (`filesystem.read_file`, `api-server.get_user`)
- **Parallel init**: Connects to all backends concurrently
- **Transport agnostic**: Supports stdio and HTTP MCP servers
