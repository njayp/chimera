package aggregator

// StdioConfig represents the configuration for a stdio MCP server.
type StdioConfig struct {
	Command string   `json:"command" yaml:"command"`
	Args    []string `json:"args" yaml:"args"`
	Env     []string `json:"env" yaml:"env"` // Additional environment variables for this server (will be appended to inherited env)
}

// HTTPConfig represents the configuration for an HTTP MCP server.
type HTTPConfig struct {
	URL string `json:"url" yaml:"url"` // Base URL of the HTTP MCP server
}
