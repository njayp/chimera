package proxy

import (
	"log"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Servers holds the complete configuration for the aggregating server.
type Servers struct {
	StdioServers map[string]StdioClient `json:"stdioServers" yaml:"stdioServers"`
	HTTPServers  map[string]HTTPClient  `json:"httpServers" yaml:"httpServers"`
}

func Run(servers Servers, addr string) error {
	m := &manager{servers: servers}

	// Create HTTP handler that creates a new aggregating server per session
	handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		return m.newProxy(req.Context())
	}, nil)

	// Start HTTP server
	log.Printf("Starting aggregating MCP HTTP server on %s", addr)
	return http.ListenAndServe(addr, handler)
}
