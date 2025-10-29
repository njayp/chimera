package client

import (
	"context"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Stdio represents the configuration for a stdio MCP server.
type Stdio struct {
	Command string
	Args    []string
	Env     []string
}

// Connect establishes a connection to the stdio MCP server.
func (c Stdio) Connect(ctx context.Context) (*mcp.ClientSession, error) {
	cmd := exec.CommandContext(ctx, c.Command, c.Args...)
	// Append any server-specific environment variables
	cmd.Env = append(cmd.Env, c.Env...)

	transport := &mcp.CommandTransport{Command: cmd}
	client := mcp.NewClient(&mcp.Implementation{
		Name: "chimera",
	}, nil)

	return client.Connect(ctx, transport, nil)
}
