// Package aggregator provides functionality for aggregating multiple MCP stdio servers
// into a single HTTP MCP server.
package aggregator

// StdioConfig represents the configuration for a stdio MCP server.
type StdioConfig struct {
	Command string   `json:"command" yaml:"command"`
	Args    []string `json:"args" yaml:"args"`
	Env     []string `json:"env" yaml:"env"` // Additional environment variables for this server (will be appended to inherited env)
}
