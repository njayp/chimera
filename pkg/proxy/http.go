package proxy

import (
	"context"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// HTTPClient represents the configuration for an HTTP MCP server.
type HTTPClient struct {
	URL     string            `json:"url" yaml:"url"`       // Base URL of the HTTP MCP server
	Headers map[string]string `json:"header" yaml:"header"` // Optional headers to include in requests
}

func (c HTTPClient) connect(ctx context.Context) (*mcp.ClientSession, error) {
	client := mcp.NewClient(&mcp.Implementation{
		Name: "chimera",
	}, nil)

	transport := &mcp.StreamableClientTransport{
		HTTPClient: c.httpClient(),
		Endpoint:   c.URL,
	}

	return client.Connect(ctx, transport, nil)
}

func (c HTTPClient) httpClient() *http.Client {
	return &http.Client{
		Transport: &CustomTransport{
			Transport: http.DefaultTransport,
			Headers:   c.Headers,
		},
	}
}

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
