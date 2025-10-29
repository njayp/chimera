package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestManager_NewProxy_MultipleServers(t *testing.T) {
	handler1 := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		impl := &mcp.Implementation{Name: "server1"}
		return mcp.NewServer(impl, nil)
	}, nil)

	handler2 := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		impl := &mcp.Implementation{Name: "server2"}
		return mcp.NewServer(impl, nil)
	}, nil)

	testServer1 := httptest.NewServer(handler1)
	defer testServer1.Close()

	testServer2 := httptest.NewServer(handler2)
	defer testServer2.Close()

	m := &manager{
		servers: Servers{
			StdioServers: make(map[string]StdioClient),
			HTTPServers: map[string]HTTPClient{
				"server1": {URL: testServer1.URL},
				"server2": {URL: testServer2.URL},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	proxy := m.newProxy(ctx)
	if proxy == nil {
		t.Fatal("expected proxy server, got nil")
	}
}
