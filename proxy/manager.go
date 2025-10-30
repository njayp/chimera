package proxy

import (
	"context"
	"net/http"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Clients maps server names to their MCP client implementations.
type Clients map[string]Client

// ClientsFunc provides Clients for each session.
type ClientsFunc func() Clients

// manager wraps multiple MCP servers and exposes them as one.
type manager struct {
	clientsFunc ClientsFunc
}

// Handler returns an HTTP handler that aggregates all clients into one MCP server.
// Each HTTP request creates a new aggregated server instance with prefixed names.
func Handler(clientsFunc ClientsFunc) *mcp.StreamableHTTPHandler {
	m := &manager{clientsFunc: clientsFunc}

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
	for n, c := range m.clientsFunc() {
		wg.Add(1)
		go func(n string, c Client) {
			defer wg.Done()
			proxy.proxyServer(ctx, c, n)
		}(n, c)
	}
	wg.Wait()

	return proxy.server
}
