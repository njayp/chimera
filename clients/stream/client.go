package stream

import (
	"context"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Client manages an HTTP-based MCP server connection.
type Client struct {
	url        string
	httpClient *http.Client
}

// NewClient creates an HTTP client with the given URL and headers.
// Headers are added to all requests (useful for authentication).
func NewClient(url string, headers map[string]string) *Client {
	httpClient := &http.Client{
		Transport: &CustomTransport{
			Transport: http.DefaultTransport,
			Headers:   headers,
		},
	}

	return &Client{
		url:        url,
		httpClient: httpClient,
	}
}

// Transport provides a new transport for each session.
func (c *Client) Transport(_ context.Context) mcp.Transport {
	return &mcp.StreamableClientTransport{
		Endpoint:   c.url,
		HTTPClient: c.httpClient,
	}
}

// CustomTransport adds headers to all HTTP requests.
type CustomTransport struct {
	Transport http.RoundTripper
	Headers   map[string]string
}

// RoundTrip adds custom headers and executes the HTTP request.
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
