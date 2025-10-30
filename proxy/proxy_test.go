package proxy

import (
	"context"
	"fmt"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type testClient struct {
	server *mcp.Server
}

func (c *testClient) Connect(ctx context.Context) (*mcp.ClientSession, error) {
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.1.0"}, nil)

	if _, err := c.server.Connect(ctx, serverTransport, nil); err != nil {
		return nil, fmt.Errorf("failed to connect server: %w", err)
	}

	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect client: %w", err)
	}

	return session, nil
}

func createTestServer(name string) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: name, Version: "0.1.0"}, nil)

	type EchoArgs struct {
		Message string `json:"message"`
	}
	mcp.AddTool(
		server,
		&mcp.Tool{Name: "echo", Description: "Echo back a message"},
		func(_ context.Context, _ *mcp.CallToolRequest, args EchoArgs) (*mcp.CallToolResult, struct{}, error) {
			msg := args.Message
			if msg == "" {
				msg = "no message"
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Echo: %s", msg)}},
			}, struct{}{}, nil
		},
	)

	server.AddResource(
		&mcp.Resource{Name: "data", URI: "test://data", MIMEType: "text/plain"},
		func(_ context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{{URI: "test://data", Text: fmt.Sprintf("Data from %s", name), MIMEType: "text/plain"}},
			}, nil
		},
	)

	server.AddPrompt(
		&mcp.Prompt{Name: "greet", Arguments: []*mcp.PromptArgument{{Name: "name", Required: true}}},
		func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			name := req.Params.Arguments["name"]
			return &mcp.GetPromptResult{
				Messages: []*mcp.PromptMessage{{Role: "user", Content: &mcp.TextContent{Text: fmt.Sprintf("Please greet %s", name)}}},
			}, nil
		},
	)

	return server
}

func connectProxyClient(ctx context.Context, t *testing.T, clients Clients) *mcp.ClientSession {
	t.Helper()

	m := &manager{clients: clients}
	proxyServer := m.newProxy(ctx)

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	if _, err := proxyServer.Connect(ctx, serverTransport, nil); err != nil {
		t.Fatalf("Failed to connect proxy server: %v", err)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "proxy-client", Version: "0.1.0"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("Failed to connect proxy client: %v", err)
	}

	t.Cleanup(func() {
		if err := session.Close(); err != nil {
			t.Errorf("Failed to close client session: %v", err)
		}
	})

	return session
}

func TestProxySingleBackend(t *testing.T) {
	ctx := context.Background()
	clients := Clients{"backend": &testClient{server: createTestServer("test-server")}}
	session := connectProxyClient(ctx, t, clients)

	t.Run("Tools", func(t *testing.T) {
		var tools []*mcp.Tool
		for tool, err := range session.Tools(ctx, nil) {
			if err != nil {
				t.Fatalf("Failed to list tools: %v", err)
			}
			tools = append(tools, tool)
		}

		if len(tools) != 1 || tools[0].Name != "backend.echo" {
			t.Errorf("Expected tool 'backend.echo', got %d tools", len(tools))
		}

		result, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "backend.echo",
			Arguments: map[string]any{"message": "hello"},
		})
		if err != nil {
			t.Fatalf("Failed to call tool: %v", err)
		}

		if text := result.Content[0].(*mcp.TextContent).Text; text != "Echo: hello" {
			t.Errorf("Expected 'Echo: hello', got %q", text)
		}
	})

	t.Run("Resources", func(t *testing.T) {
		var resources []*mcp.Resource
		for resource, err := range session.Resources(ctx, nil) {
			if err != nil {
				t.Fatalf("Failed to list resources: %v", err)
			}
			resources = append(resources, resource)
		}

		if len(resources) != 1 || resources[0].URI != "backend.test://data" {
			t.Errorf("Expected resource 'backend.test://data', got %d resources", len(resources))
		}

		result, err := session.ReadResource(ctx, &mcp.ReadResourceParams{URI: "backend.test://data"})
		if err != nil {
			t.Fatalf("Failed to read resource: %v", err)
		}

		if text := result.Contents[0].Text; text != "Data from test-server" {
			t.Errorf("Expected 'Data from test-server', got %q", text)
		}
	})

	t.Run("Prompts", func(t *testing.T) {
		var prompts []*mcp.Prompt
		for prompt, err := range session.Prompts(ctx, nil) {
			if err != nil {
				t.Fatalf("Failed to list prompts: %v", err)
			}
			prompts = append(prompts, prompt)
		}

		if len(prompts) != 1 || prompts[0].Name != "backend.greet" {
			t.Errorf("Expected prompt 'backend.greet', got %d prompts", len(prompts))
		}

		result, err := session.GetPrompt(ctx, &mcp.GetPromptParams{
			Name:      "backend.greet",
			Arguments: map[string]string{"name": "Alice"},
		})
		if err != nil {
			t.Fatalf("Failed to get prompt: %v", err)
		}

		if text := result.Messages[0].Content.(*mcp.TextContent).Text; text != "Please greet Alice" {
			t.Errorf("Expected 'Please greet Alice', got %q", text)
		}
	})
}

func TestProxyMultipleBackends(t *testing.T) {
	ctx := context.Background()
	clients := Clients{
		"backend1": &testClient{server: createTestServer("server1")},
		"backend2": &testClient{server: createTestServer("server2")},
	}
	session := connectProxyClient(ctx, t, clients)

	var tools []*mcp.Tool
	for tool, err := range session.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("Failed to list tools: %v", err)
		}
		tools = append(tools, tool)
	}

	if len(tools) != 2 {
		t.Fatalf("Expected 2 tools, got %d", len(tools))
	}

	toolNames := map[string]bool{}
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	if !toolNames["backend1.echo"] || !toolNames["backend2.echo"] {
		t.Errorf("Expected tools 'backend1.echo' and 'backend2.echo', got %v", toolNames)
	}

	result1, _ := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "backend1.echo",
		Arguments: map[string]any{"message": "from backend1"},
	})
	if text := result1.Content[0].(*mcp.TextContent).Text; text != "Echo: from backend1" {
		t.Errorf("Expected 'Echo: from backend1', got %q", text)
	}

	result2, _ := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "backend2.echo",
		Arguments: map[string]any{"message": "from backend2"},
	})
	if text := result2.Content[0].(*mcp.TextContent).Text; text != "Echo: from backend2" {
		t.Errorf("Expected 'Echo: from backend2', got %q", text)
	}
}

func TestProxyEmpty(t *testing.T) {
	ctx := context.Background()
	session := connectProxyClient(ctx, t, Clients{})

	var tools []*mcp.Tool
	for tool, err := range session.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("Failed to list tools: %v", err)
		}
		tools = append(tools, tool)
	}

	if len(tools) != 0 {
		t.Errorf("Expected 0 tools, got %d", len(tools))
	}
}

func TestProxyPrefixing(t *testing.T) {
	ctx := context.Background()

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.1.0"}, nil)
	mcp.AddTool(
		server,
		&mcp.Tool{Name: "backend.prefixed"},
		func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, struct{}, error) {
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "ok"}}}, struct{}{}, nil
		},
	)

	clients := Clients{"backend": &testClient{server: server}}
	session := connectProxyClient(ctx, t, clients)

	var tools []*mcp.Tool
	for tool, err := range session.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("Failed to list tools: %v", err)
		}
		tools = append(tools, tool)
	}

	if len(tools) != 1 || tools[0].Name != "backend.prefixed" {
		t.Errorf("Expected tool 'backend.prefixed' (no double-prefix), got %q", tools[0].Name)
	}
}
