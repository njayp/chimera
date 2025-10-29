package proxy

import (
	"context"
	"log"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Clients is a map of named MCP clients.
type Clients map[string]client

// manager wraps multiple MCP servers and exposes them as one.
type manager struct {
	clients Clients
}

// Run starts the aggregating MCP HTTP server on the specified address.
func Run(clients Clients, addr string) error {
	m := &manager{clients: clients}

	// Create HTTP handler that creates a new aggregating server per session
	handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		return m.newProxy(req.Context())
	}, nil)

	// Start HTTP server
	log.Printf("Starting aggregating MCP HTTP server on %s", addr)
	return http.ListenAndServe(addr, handler)
}

// each newProxy creates a new MCP server instance that aggregates
// all configured backend servers.
func (s *manager) newProxy(ctx context.Context) *mcp.Server {
	impl := &mcp.Implementation{
		Name: "aggregate-proxy",
	}

	proxy := &proxy{
		server: mcp.NewServer(impl, nil),
	}

	for n, c := range s.clients {
		proxy.proxyServer(ctx, c, n)
	}

	return proxy.server
}
