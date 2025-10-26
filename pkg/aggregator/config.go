package aggregator

// Config represents the configuration for a stdio MCP server.
type Config struct {
	Binary string
	Args   []string
	Env    []string // Additional environment variables for this server (will be appended to inherited env)
}
