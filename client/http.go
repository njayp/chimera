package client

import (
	"context"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// HTTP represents the configuration for an HTTP MCP server.
type HTTP struct {
	URL     string
	Headers map[string]string
}

// Connect establishes a connection to the HTTP MCP server.
func (c HTTP) Connect(ctx context.Context) (*mcp.ClientSession, error) {
	client := mcp.NewClient(&mcp.Implementation{
		Name: "chimera",
	}, nil)

	transport := &mcp.StreamableClientTransport{
		HTTPClient: c.client(),
		Endpoint:   c.URL,
	}

	return client.Connect(ctx, transport, nil)
}

func (c HTTP) client() *http.Client {
	return &http.Client{
		Transport: &CustomTransport{
			Transport: http.DefaultTransport,
			Headers:   c.Headers,
		},
	}
}

// CustomTransport wraps an HTTP transport to add custom headers to all requests.
type CustomTransport struct {
	Transport http.RoundTripper
	Headers   map[string]string
}

// RoundTrip executes a single HTTP transaction, adding custom headers
func (t *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	reqClone := req.Clone(req.Context())

	// Add custom headers
	for key, value := range t.Headers {
		reqClone.Header.Set(key, value)
	}

	// Use the underlying transport to execute the request
	return t.Transport.RoundTrip(reqClone)
}
