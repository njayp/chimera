package proxy

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// manager wraps multiple MCP servers and exposes them as one.
type manager struct {
	servers Servers
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

	for n, c := range s.servers.StdioServers {
		proxy.proxyServer(ctx, c, n)
	}

	for n, c := range s.servers.HTTPServers {
		proxy.proxyServer(ctx, c, n)
	}

	return proxy.server
}
