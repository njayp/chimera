# Chimera

Aggregates multiple MCP servers (stdio and HTTP) into a single HTTP endpoint.

## Quick Start

## Features

- **Auto-prefixing**: Prevents name conflicts (`filesystem.read_file`, `api-server.get_user`)
- **Parallel init**: Connects to all backends concurrently
- **Transport agnostic**: Supports stdio and HTTP MCP servers
- **Kubernetes-ready**: Helm chart with ConfigMap-based configuration
