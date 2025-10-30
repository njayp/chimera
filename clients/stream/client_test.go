package stream

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHTTPClient_Connect(t *testing.T) {
	handler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		impl := &mcp.Implementation{Name: "test-server"}
		return mcp.NewServer(impl, nil)
	}, nil)

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewClient(server.URL, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	session, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if session == nil {
		t.Fatal("expected session, got nil")
	}
}

func TestCustomTransport_RoundTrip(t *testing.T) {
	requestReceived := false
	headers := make(map[string]string)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true
		headers["X-Test"] = r.Header.Get("X-Test")
		headers["Authorization"] = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport := &CustomTransport{
		Transport: http.DefaultTransport,
		Headers: map[string]string{
			"X-Test":        "test-value",
			"Authorization": "Bearer token",
		},
	}

	client := &http.Client{Transport: transport}

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if !requestReceived {
		t.Fatal("request was not received by server")
	}

	if headers["X-Test"] != "test-value" {
		t.Errorf("expected X-Test header 'test-value', got '%s'", headers["X-Test"])
	}

	if headers["Authorization"] != "Bearer token" {
		t.Errorf("expected Authorization header 'Bearer token', got '%s'", headers["Authorization"])
	}
}
