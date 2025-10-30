package stdio

import (
	"context"
	"os"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Client manages a stdio-based MCP server connection.
type Client struct {
	command string
	args    []string
	env     []string
	client  *mcp.Client
}

// NewClient creates a stdio client that spawns the given command.
// env is a list of "KEY=value" strings appended to the process environment.
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

// Connect spawns the stdio process and establishes an MCP session.
func (c *Client) Connect(ctx context.Context) (*mcp.ClientSession, error) {
	cmd := exec.CommandContext(ctx, c.command, c.args...)
	// Append any server-specific environment variables
	cmd.Env = append(os.Environ(), c.env...)
	transport := &mcp.CommandTransport{Command: cmd}

	return c.client.Connect(ctx, transport, nil)
}
