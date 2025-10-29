package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/njayp/chimera/client"
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
		clients: make(Clients),
	}

	m.clients["server1"] = &client.HTTP{URL: testServer1.URL}
	m.clients["server2"] = &client.HTTP{URL: testServer2.URL}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	proxy := m.newProxy(ctx)
	if proxy == nil {
		t.Fatal("expected proxy server, got nil")
	}
}
