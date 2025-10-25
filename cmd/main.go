// Package main provides an HTTP MCP server that aggregates multiple stdio MCP servers.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// AggregatingServer wraps multiple stdio MCP servers and exposes them as one
type AggregatingServer struct {
	server  *mcp.Server
	clients []*clientConnection
	mu      sync.RWMutex
}

type clientConnection struct {
	name    string
	cmd     *exec.Cmd
	client  *mcp.Client
	session *mcp.ClientSession
}

func NewAggregatingServer() *AggregatingServer {
	impl := &mcp.Implementation{
		Name:    "aggregating-mcp-server",
		Version: "v1.0.0",
	}

	server := mcp.NewServer(impl, &mcp.ServerOptions{
		HasTools:     true,
		HasResources: true,
		HasPrompts:   true,
	})

	agg := &AggregatingServer{
		server:  server,
		clients: make([]*clientConnection, 0),
	}

	return agg
}

func (a *AggregatingServer) ConnectToStdioServers(ctx context.Context, configs []StdioServerConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, config := range configs {
		cmd := exec.Command(config.Binary, config.Args...)
		// Inherit environment variables from parent process
		cmd.Env = os.Environ()
		// Append any server-specific environment variables
		if len(config.Env) > 0 {
			cmd.Env = append(cmd.Env, config.Env...)
		}

		client := mcp.NewClient(&mcp.Implementation{
			Name:    "aggregating-client",
			Version: "v1.0.0",
		}, nil)

		transport := &mcp.CommandTransport{Command: cmd}
		session, err := client.Connect(ctx, transport, nil)
		if err != nil {
			return fmt.Errorf("failed to connect to %s: %w", config.Name, err)
		}

		conn := &clientConnection{
			name:    config.Name,
			cmd:     cmd,
			client:  client,
			session: session,
		}
		a.clients = append(a.clients, conn)

		// Fetch and register all tools from this server
		if err := a.syncTools(ctx, conn); err != nil {
			return fmt.Errorf("failed to sync tools from %s: %w", config.Name, err)
		}

		// Fetch and register all resources from this server
		if err := a.syncResources(ctx, conn); err != nil {
			return fmt.Errorf("failed to sync resources from %s: %w", config.Name, err)
		}

		// Fetch and register all prompts from this server
		if err := a.syncPrompts(ctx, conn); err != nil {
			return fmt.Errorf("failed to sync prompts from %s: %w", config.Name, err)
		}
	}

	return nil
}

func (a *AggregatingServer) syncTools(ctx context.Context, conn *clientConnection) error {
	for tool, err := range conn.session.Tools(ctx, nil) {
		if err != nil {
			return err
		}

		// Prefix tool name with server name to avoid conflicts
		prefixedTool := *tool
		prefixedTool.Name = conn.name + "." + tool.Name

		// Add tool to our aggregating server
		a.server.AddTool(&prefixedTool, func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return a.routeToolCall(ctx, req, conn)
		})
	}
	return nil
}

func (a *AggregatingServer) syncResources(ctx context.Context, conn *clientConnection) error {
	for resource, err := range conn.session.Resources(ctx, nil) {
		if err != nil {
			return err
		}

		// Create a prefixed URI to avoid conflicts
		prefixedResource := *resource
		prefixedResource.URI = conn.name + "://" + resource.URI

		a.server.AddResource(&prefixedResource, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			return a.routeResourceRead(ctx, req, conn)
		})
	}
	return nil
}

func (a *AggregatingServer) syncPrompts(ctx context.Context, conn *clientConnection) error {
	for prompt, err := range conn.session.Prompts(ctx, nil) {
		if err != nil {
			return err
		}

		prefixedPrompt := *prompt
		prefixedPrompt.Name = conn.name + "." + prompt.Name

		a.server.AddPrompt(&prefixedPrompt, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			return a.routePromptGet(ctx, req, conn)
		})
	}
	return nil
}

func (a *AggregatingServer) routeToolCall(ctx context.Context, req *mcp.CallToolRequest, conn *clientConnection) (*mcp.CallToolResult, error) {
	// Remove the prefix to get the original tool name
	originalName := req.Params.Name[len(conn.name)+1:]

	params := &mcp.CallToolParams{
		Name:      originalName,
		Arguments: req.Params.Arguments,
	}

	return conn.session.CallTool(ctx, params)
}

func (a *AggregatingServer) routeResourceRead(ctx context.Context, req *mcp.ReadResourceRequest, conn *clientConnection) (*mcp.ReadResourceResult, error) {
	// Remove the prefix to get the original URI
	originalURI := req.Params.URI[len(conn.name)+3:] // +3 for "://"

	params := &mcp.ReadResourceParams{
		URI: originalURI,
	}

	return conn.session.ReadResource(ctx, params)
}

func (a *AggregatingServer) routePromptGet(ctx context.Context, req *mcp.GetPromptRequest, conn *clientConnection) (*mcp.GetPromptResult, error) {
	originalName := req.Params.Name[len(conn.name)+1:]

	params := &mcp.GetPromptParams{
		Name:      originalName,
		Arguments: req.Params.Arguments,
	}

	return conn.session.GetPrompt(ctx, params)
}

func (a *AggregatingServer) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, conn := range a.clients {
		if err := conn.session.Close(); err != nil {
			log.Printf("Error closing session for %s: %v", conn.name, err)
		}
	}
	return nil
}

type StdioServerConfig struct {
	Name   string
	Binary string
	Args   []string
	Env    []string // Additional environment variables for this server (will be appended to inherited env)
}

func main() {
	ctx := context.Background()

	// Define your stdio MCP servers
	stdioServers := []StdioServerConfig{
		{
			Name:   "filesystem",
			Binary: "./filesystem-server",
			Args:   []string{},
			Env:    []string{}, // Inherits all parent env vars
		},
		{
			Name:   "weather",
			Binary: "./weather-server",
			Args:   []string{},
			Env: []string{
				"WEATHER_API_KEY=your_api_key_here",
				"WEATHER_PROVIDER=openweather",
			},
		},
		{
			Name:   "database",
			Binary: "/usr/local/bin/db-mcp-server",
			Args:   []string{},
			Env: []string{
				"DATABASE_URL=postgresql://localhost/mydb",
				"DB_POOL_SIZE=10",
			},
		},
	}

	// Create HTTP handler that creates a new aggregating server per session
	handler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		agg := NewAggregatingServer()

		// Connect to all stdio servers for this HTTP session
		if err := agg.ConnectToStdioServers(ctx, stdioServers); err != nil {
			log.Printf("Failed to connect to stdio servers: %v", err)
			return nil
		}

		return agg.server
	}, nil)

	// Start HTTP server
	addr := ":8080"
	log.Printf("Starting aggregating MCP HTTP server on %s", addr)
	log.Printf("Aggregating servers: %v", stdioServers)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
