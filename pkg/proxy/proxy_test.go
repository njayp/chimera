package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/njayp/chimera/pkg/client"
)

func TestProxy_EmptyServer(t *testing.T) {
	handler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		impl := &mcp.Implementation{Name: "backend"}
		return mcp.NewServer(impl, nil)
	}, nil)

	testServer := httptest.NewServer(handler)
	defer testServer.Close()

	client := client.HTTP{URL: testServer.URL}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	session, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	impl := &mcp.Implementation{Name: "proxy"}
	proxy := &proxy{
		server: mcp.NewServer(impl, nil),
	}

	// Should handle empty server without error
	err = proxy.proxyTools(ctx, session, "backend")
	if err != nil {
		t.Fatalf("failed to proxy tools: %v", err)
	}

	err = proxy.proxyResources(ctx, session, "backend")
	if err != nil {
		t.Fatalf("failed to proxy resources: %v", err)
	}

	err = proxy.proxyPrompts(ctx, session, "backend")
	if err != nil {
		t.Fatalf("failed to proxy prompts: %v", err)
	}
}
