package proxy

import (
	"context"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Clients is a map of named MCP clients.
type Clients map[string]Client

// manager wraps multiple MCP servers and exposes them as one.
type manager struct {
	clients Clients
}

// Handler starts the aggregating MCP HTTP server on the specified address.
func Handler(clients Clients) *mcp.StreamableHTTPHandler {
	m := &manager{clients: clients}

	// Create HTTP handler that creates a new aggregating server per session
	return mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		return m.newProxy(req.Context())
	}, nil)
}

// each newProxy creates a new MCP server instance that aggregates
// all configured backend servers.
func (m *manager) newProxy(ctx context.Context) *mcp.Server {
	proxy := &proxy{
		server: mcp.NewServer(&mcp.Implementation{
			Name: "chimera",
		}, nil),
	}

	for n, c := range m.clients {
		proxy.proxyServer(ctx, c, n)
	}

	return proxy.server
}
