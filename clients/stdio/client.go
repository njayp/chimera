package stdio

import (
	"context"
	"os"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Client represents the configuration for a stdio MCP server.
type Client struct {
	command string
	args    []string
	env     []string
	client  *mcp.Client
}

// NewClient creates a new stdio Client instance with the specified command, arguments, and environment variables.
func NewClient(command string, args []string, env []string) *Client {
	client := mcp.NewClient(&mcp.Implementation{
		Name: "chimera",
	}, nil)

	return &Client{
		command: command,
		args:    args,
		env:     env,
		client:  client,
	}
}

// Connect establishes a connection to the stdio MCP server.
func (c *Client) Connect(ctx context.Context) (*mcp.ClientSession, error) {
	cmd := exec.CommandContext(ctx, c.command, c.args...)
	// Append any server-specific environment variables
	cmd.Env = append(os.Environ(), c.env...)
	transport := &mcp.CommandTransport{Command: cmd}

	return c.client.Connect(ctx, transport, nil)
}
