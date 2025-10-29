package proxy

import (
	"context"
	"net/http"
	"sync"

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

	// Connect to all backend servers async
	wg := sync.WaitGroup{}
	for n, c := range m.clients {
		wg.Add(1)
		go func(n string, c Client) {
			defer wg.Done()
			proxy.proxyServer(ctx, c, n)
		}(n, c)
	}
	wg.Wait()

	return proxy.server
}
