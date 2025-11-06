package proxy

import (
	"context"
	"net/http"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Clients maps server names to their MCP client implementations.
type Clients map[string]Client

// Provider provides MCP clients for each session.
type Provider interface {
	Clients() Clients
}

// manager wraps multiple MCP servers and exposes them as one.
type manager struct {
	provider Provider
}

// Handler returns an HTTP handler that aggregates all clients into one MCP server.
// Each HTTP request creates a new aggregated server instance with prefixed names.
func Handler(provider Provider) *mcp.StreamableHTTPHandler {
	m := &manager{provider: provider}

	// Create HTTP handler that creates a new aggregating server per session
	// This allows different tools to be available for different sessions
	return mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		return m.newProxy(req.Context())
	}, nil)
}

// each newProxy creates a new MCP server instance that aggregates
// all configured backend servers.
func (m *manager) newProxy(ctx context.Context) *mcp.Server {
	p := &proxy{}
	p.server = mcp.NewServer(&mcp.Implementation{
		Name: "chimera",
	}, &mcp.ServerOptions{
		// TODO
		RootsListChangedHandler: nil,
	})

	// Connect to all backend servers async
	wg := sync.WaitGroup{}
	for n, c := range m.provider.Clients() {
		wg.Go(func() {
			p.proxyServer(ctx, n, c)
		})
	}
	wg.Wait()

	return p.server
}
