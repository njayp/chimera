# Chimera - MCP Server Aggregator

Chimera is an HTTP MCP server that aggregates multiple stdio MCP servers into a single unified interface.

## Project Structure

```
chimera/
├── cmd/
│   └── main.go                  # Application entry point
├── pkg/
│   ├── aggregator/              # Core aggregation logic
│   │   ├── config.go            # Configuration types
│   │   ├── config_loader.go     # Config file loading
│   │   └── server.go            # Aggregating server implementation
│   └── server/                  # (Reserved for future server utilities)
├── config.example.yaml          # Example YAML configuration
├── config.example.json          # Example JSON configuration
├── go.mod
├── go.sum
├── makefile
└── README.md
```

## Architecture

### Package: `pkg/aggregator`

The `aggregator` package provides the core functionality for aggregating multiple stdio MCP servers:

- **`Config`**: Configuration for individual stdio MCP servers (binary path, args, env vars)
- **`ServerConfig`**: Top-level configuration including HTTP address and list of servers
- **`LoadConfig`**: Loads configuration from YAML or JSON files
- **`Server`**: The aggregating server that manages multiple stdio server connections
  - Connects to stdio servers via `exec.Command`
  - Syncs capabilities (tools, resources, prompts) from each server
  - Prefixes names/URIs to avoid conflicts between servers
  - Routes requests to the appropriate upstream server

### Package: `cmd`

The main application entry point that:
1. Loads configuration from file (YAML or JSON)
2. Creates an HTTP handler using the MCP SDK's `StreamableHTTPHandler`
3. Instantiates a new aggregating server per HTTP session
4. Starts the HTTP server

## Usage

### Configuration

Create a configuration file (YAML or JSON) to define your stdio MCP servers:

**config.yaml:**
```yaml
address: ":8080"
servers:
  - name: filesystem
    binary: ./filesystem-server
    args: []
    env: []
  
  - name: weather
    binary: ./weather-server
    args: []
    env:
      - WEATHER_API_KEY=your_api_key_here
      - WEATHER_PROVIDER=openweather
  
  - name: database
    binary: /usr/local/bin/db-mcp-server
    args: []
    env:
      - DATABASE_URL=postgresql://localhost/mydb
      - DB_POOL_SIZE=10
```

**config.json:**
```json
{
  "address": ":8080",
  "servers": [
    {
      "name": "filesystem",
      "binary": "./filesystem-server",
      "args": [],
      "env": []
    },
    {
      "name": "weather",
      "binary": "./weather-server",
      "args": [],
      "env": [
        "WEATHER_API_KEY=your_api_key_here",
        "WEATHER_PROVIDER=openweather"
      ]
    }
  ]
}
```

See `config.example.yaml` or `config.example.json` for complete examples.

### Running the Server

```bash
# Using default config.yaml
go run ./cmd/main.go

# Using a specific config file
go run ./cmd/main.go -config=/path/to/config.yaml

# Or with JSON
go run ./cmd/main.go -config=/path/to/config.json
```

The server will start on the address specified in your config file (default `:8080`).

### Configuration Fields

- **`address`**: HTTP server listen address (e.g., `:8080`, `localhost:3000`)
- **`servers`**: Array of stdio MCP server configurations
  - **`name`**: Unique identifier for the server (used as prefix for tools/resources/prompts)
  - **`binary`**: Path to the stdio MCP server executable
  - **`args`**: Command-line arguments to pass to the server
  - **`env`**: Additional environment variables (appended to inherited parent env)

### Environment Variables

- **Inherited**: All stdio servers inherit the parent process's environment variables by default
- **Additional**: Use the `env` field to add server-specific environment variables
- **Override**: Server-specific env vars can override inherited ones (if same key)

## How It Works

1. **HTTP Request** → Client connects to Chimera HTTP server
2. **Session Created** → New `aggregator.Server` instance created per session
3. **Stdio Connections** → Aggregator connects to each configured stdio server
4. **Capability Sync** → Tools, resources, and prompts are discovered and registered
5. **Name Prefixing** → Each capability is prefixed with its server name (e.g., `filesystem.read_file`)
6. **Request Routing** → Incoming requests are routed to the appropriate stdio server based on prefix

## Development

### Build

```bash
make build
```

### Lint

```bash
make lint
```

### Test

```bash
make test
```

### Run All

```bash
make all
```

## Example: Tool Name Mapping

If a stdio server named `"filesystem"` exposes a tool `"read_file"`, it becomes:
- **Aggregated name**: `filesystem.read_file`
- **When called**: The prefix is stripped and routed to the filesystem server as `read_file`

## Example: Resource URI Mapping

If a stdio server named `"database"` exposes resource `"file:///schema"`, it becomes:
- **Aggregated URI**: `database://file:///schema`
- **When read**: The prefix is stripped and routed to the database server as `file:///schema`

## License

See LICENSE file.
