package aggregator

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server wraps multiple MCP servers and exposes them as one.
type Server struct {
	server       *mcp.Server
	stdioClients []*stdioConnection
}

// New creates a new aggregating server.
func New() *Server {
	impl := &mcp.Implementation{
		Name:    "aggregating-mcp-server",
		Version: "v1.0.0",
	}

	mcpServer := mcp.NewServer(impl, nil)

	return &Server{
		server:       mcpServer,
		stdioClients: make([]*stdioConnection, 0),
	}
}

// MCPServer returns the underlying mcp.Server instance.
func (s *Server) MCPServer() *mcp.Server {
	return s.server
}
