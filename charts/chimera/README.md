# Chimera Helm Chart

Deploys the Chimera MCP server aggregator proxy.

## Installation

```bash
helm install chimera ./charts/chimera -f my-values.yaml
```

## Configuration

Configure MCP servers in `values.yaml`:

```yaml
config:
  servers:
    filesystem:
      type: stdio
      command: /usr/local/bin/mcp-filesystem
      args: [start]
    api-server:
      type: http
      url: http://api-server:8080/mcp
```

The config is mounted as a ConfigMap at `/etc/chimera/config.json`.
