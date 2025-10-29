package proxy

import (
	"context"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// StdioClient represents the configuration for a stdio MCP server.
type StdioClient struct {
	Command string   `json:"command" yaml:"command"`
	Args    []string `json:"args" yaml:"args"`
	Env     []string `json:"env" yaml:"env"` // Additional environment variables for this server (will be appended to inherited env)
}

func (c StdioClient) connect(ctx context.Context) (*mcp.ClientSession, error) {
	cmd := exec.CommandContext(ctx, c.Command, c.Args...)

	// Append any server-specific environment variables
	if len(c.Env) > 0 {
		cmd.Env = append(cmd.Env, c.Env...)
	}

	transport := &mcp.CommandTransport{Command: cmd}

	client := mcp.NewClient(&mcp.Implementation{
		Name: "chimera",
	}, nil)

	return client.Connect(ctx, transport, nil)
}
