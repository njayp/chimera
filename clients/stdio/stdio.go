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
}

func New(command string, args []string, env []string) Client {
	return Client{
		command: command,
		args:    args,
		env:     env,
	}
}

// Connect establishes a connection to the stdio MCP server.
func (c Client) Connect(ctx context.Context) (*mcp.ClientSession, error) {
	cmd := exec.CommandContext(ctx, c.command, c.args...)
	// Append any server-specific environment variables
	cmd.Env = append(os.Environ(), c.env...)
	transport := &mcp.CommandTransport{Command: cmd}
	client := mcp.NewClient(&mcp.Implementation{
		Name: "chimera",
	}, nil)

	return client.Connect(ctx, transport, nil)
}
