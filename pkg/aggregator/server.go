package aggregator

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server wraps multiple stdio MCP servers and exposes them as one.
type Server struct {
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

// New creates a new aggregating server.
func New() *Server {
	impl := &mcp.Implementation{
		Name:    "aggregating-mcp-server",
		Version: "v1.0.0",
	}

	mcpServer := mcp.NewServer(impl, nil)

	return &Server{
		server:  mcpServer,
		clients: make([]*clientConnection, 0),
	}
}

// ConnectToStdioServers establishes connections to all configured stdio servers
// and syncs their capabilities (tools, resources, prompts).
func (s *Server) ConnectToStdioServers(ctx context.Context, configs map[string]Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for name, config := range configs {
		cmd := exec.Command(config.Command, config.Args...)
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
			return fmt.Errorf("failed to connect to %s: %w", name, err)
		}

		conn := &clientConnection{
			name:    name,
			cmd:     cmd,
			client:  client,
			session: session,
		}
		s.clients = append(s.clients, conn)

		// Fetch and register all tools from this server
		if err := s.syncTools(ctx, conn); err != nil {
			slog.Error(fmt.Errorf("failed to sync tools from %s: %w", name, err).Error())
		}

		// Fetch and register all resources from this server
		if err := s.syncResources(ctx, conn); err != nil {
			slog.Error(fmt.Errorf("failed to sync resources from %s: %w", name, err).Error())
		}

		// Fetch and register all prompts from this server
		if err := s.syncPrompts(ctx, conn); err != nil {
			slog.Error(fmt.Errorf("failed to sync prompts from %s: %w", name, err).Error())
		}
	}

	return nil
}

// MCPServer returns the underlying mcp.Server instance.
func (s *Server) MCPServer() *mcp.Server {
	return s.server
}

// Close closes all client connections.
func (s *Server) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, conn := range s.clients {
		if err := conn.session.Close(); err != nil {
			log.Printf("Error closing session for %s: %v", conn.name, err)
		}
	}
	return nil
}

func (s *Server) syncTools(ctx context.Context, conn *clientConnection) error {
	for tool, err := range conn.session.Tools(ctx, nil) {
		if err != nil {
			return err
		}

		// Prefix tool name with server name to avoid conflicts
		prefixedTool := *tool
		if !strings.HasPrefix(tool.Name, conn.name) {
			prefixedTool.Name = conn.name + "." + tool.Name
		}

		// Add tool to our aggregating server
		s.server.AddTool(&prefixedTool, func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return s.routeToolCall(ctx, req, conn, tool.Name)
		})
	}
	return nil
}

func (s *Server) syncResources(ctx context.Context, conn *clientConnection) error {
	for resource, err := range conn.session.Resources(ctx, nil) {
		if err != nil {
			return err
		}

		// Create a prefixed URI to avoid conflicts
		prefixedResource := *resource
		if !strings.HasPrefix(resource.URI, conn.name) {
			prefixedResource.URI = conn.name + "://" + resource.URI
		}

		s.server.AddResource(&prefixedResource, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			return s.routeResourceRead(ctx, req, conn, resource.URI)
		})
	}
	return nil
}

func (s *Server) syncPrompts(ctx context.Context, conn *clientConnection) error {
	for prompt, err := range conn.session.Prompts(ctx, nil) {
		if err != nil {
			return err
		}

		prefixedPrompt := *prompt
		if !strings.HasPrefix(prompt.Name, conn.name) {
			prefixedPrompt.Name = conn.name + "." + prompt.Name
		}

		s.server.AddPrompt(&prefixedPrompt, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			return s.routePromptGet(ctx, req, conn, prompt.Name)
		})
	}
	return nil
}

func (s *Server) routeToolCall(ctx context.Context, req *mcp.CallToolRequest, conn *clientConnection, toolName string) (*mcp.CallToolResult, error) {
	params := &mcp.CallToolParams{
		Name:      toolName,
		Arguments: req.Params.Arguments,
	}

	return conn.session.CallTool(ctx, params)
}

func (s *Server) routeResourceRead(ctx context.Context, _ *mcp.ReadResourceRequest, conn *clientConnection, uri string) (*mcp.ReadResourceResult, error) {
	params := &mcp.ReadResourceParams{
		URI: uri,
	}

	return conn.session.ReadResource(ctx, params)
}

func (s *Server) routePromptGet(ctx context.Context, req *mcp.GetPromptRequest, conn *clientConnection, name string) (*mcp.GetPromptResult, error) {
	params := &mcp.GetPromptParams{
		Name:      name,
		Arguments: req.Params.Arguments,
	}

	return conn.session.GetPrompt(ctx, params)
}
