package stream

import (
	"context"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Client represents the configuration for an Client MCP server.
type Client struct {
	url        string
	headers    map[string]string
	mcpClient  *mcp.Client
	httpClient *http.Client
}

// NewClient creates a new Client instance with the specified URL and headers.
func NewClient(url string, headers map[string]string) *Client {
	mcpClient := mcp.NewClient(&mcp.Implementation{
		Name: "chimera",
	}, nil)

	httpClient := &http.Client{
		Transport: &CustomTransport{
			Transport: http.DefaultTransport,
			Headers:   headers,
		},
	}

	return &Client{
		url:        url,
		headers:    headers,
		mcpClient:  mcpClient,
		httpClient: httpClient,
	}
}

// Connect establishes a connection to the HTTP MCP server.
func (c *Client) Connect(ctx context.Context) (*mcp.ClientSession, error) {
	// A new transport must be created for each session
	transport := &mcp.StreamableClientTransport{
		Endpoint:   c.url,
		HTTPClient: c.httpClient,
	}

	return c.mcpClient.Connect(ctx, transport, nil)
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
