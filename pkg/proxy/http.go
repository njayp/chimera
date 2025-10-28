package proxy

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// HTTPClient represents the configuration for an HTTP MCP server.
type HTTPClient struct {
	URL string `json:"url" yaml:"url"` // Base URL of the HTTP MCP server
}

func (c HTTPClient) connect(ctx context.Context) (*mcp.ClientSession, error) {
	client := mcp.NewClient(&mcp.Implementation{
		Name: "aggregating-client",
	}, nil)

	transport := &mcp.StreamableClientTransport{
		Endpoint: c.URL,
	}

	return client.Connect(ctx, transport, nil)
}
