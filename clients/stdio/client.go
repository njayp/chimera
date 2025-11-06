package stdio

import (
	"context"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Client manages a stdio-based MCP server connection.
type Client struct {
	command string
	args    []string
	env     []string
}

// NewClient creates a stdio client that spawns the given command.
// env is a list of "KEY=value" strings appended to the process environment.
func NewClient(command string, args []string, env []string) *Client {
	return &Client{
		command: command,
		args:    args,
		env:     env,
	}
}

// Transport provides a new transport for each session.
func (c *Client) Transport(ctx context.Context) mcp.Transport {
	cmd := exec.CommandContext(ctx, c.command, c.args...)
	cmd.Env = c.env
	return &mcp.CommandTransport{Command: cmd}
}
