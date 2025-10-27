package aggregator

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type httpConnection struct {
	name    string
	url     string
	client  *mcp.Client
	session *mcp.ClientSession
}

// ConnectToHTTPServers establishes connections to all configured HTTP servers
// and syncs their capabilities (tools, resources, prompts).
func (s *Server) ConnectToHTTPServers(ctx context.Context, configs map[string]HTTPConfig) error {
	for name, config := range configs {
		client := mcp.NewClient(&mcp.Implementation{
			Name:    "aggregating-client",
			Version: "v1.0.0",
		}, nil)

		transport := &mcp.StreamableClientTransport{
			Endpoint: config.URL,
		}
		session, err := client.Connect(ctx, transport, nil)
		if err != nil {
			return fmt.Errorf("failed to connect to %s at %s: %w", name, config.URL, err)
		}

		conn := &httpConnection{
			name:    name,
			url:     config.URL,
			client:  client,
			session: session,
		}
		s.httpClients = append(s.httpClients, conn)

		// Fetch and register all tools from this server
		if err := s.syncHTTPTools(ctx, conn); err != nil {
			slog.Error(fmt.Errorf("failed to sync tools from %s: %w", name, err).Error())
		}

		// Fetch and register all resources from this server
		if err := s.syncHTTPResources(ctx, conn); err != nil {
			slog.Error(fmt.Errorf("failed to sync resources from %s: %w", name, err).Error())
		}

		// Fetch and register all prompts from this server
		if err := s.syncHTTPPrompts(ctx, conn); err != nil {
			slog.Error(fmt.Errorf("failed to sync prompts from %s: %w", name, err).Error())
		}
	}

	return nil
}

func (s *Server) syncHTTPTools(ctx context.Context, conn *httpConnection) error {
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
			return s.routeHTTPToolCall(ctx, req, conn, tool.Name)
		})
	}
	return nil
}

func (s *Server) syncHTTPResources(ctx context.Context, conn *httpConnection) error {
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
			return s.routeHTTPResourceRead(ctx, req, conn, resource.URI)
		})
	}
	return nil
}

func (s *Server) syncHTTPPrompts(ctx context.Context, conn *httpConnection) error {
	for prompt, err := range conn.session.Prompts(ctx, nil) {
		if err != nil {
			return err
		}

		prefixedPrompt := *prompt
		if !strings.HasPrefix(prompt.Name, conn.name) {
			prefixedPrompt.Name = conn.name + "." + prompt.Name
		}

		s.server.AddPrompt(&prefixedPrompt, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			return s.routeHTTPPromptGet(ctx, req, conn, prompt.Name)
		})
	}
	return nil
}

func (s *Server) routeHTTPToolCall(ctx context.Context, req *mcp.CallToolRequest, conn *httpConnection, toolName string) (*mcp.CallToolResult, error) {
	params := &mcp.CallToolParams{
		Name:      toolName,
		Arguments: req.Params.Arguments,
	}

	return conn.session.CallTool(ctx, params)
}

func (s *Server) routeHTTPResourceRead(ctx context.Context, _ *mcp.ReadResourceRequest, conn *httpConnection, uri string) (*mcp.ReadResourceResult, error) {
	params := &mcp.ReadResourceParams{
		URI: uri,
	}

	return conn.session.ReadResource(ctx, params)
}

func (s *Server) routeHTTPPromptGet(ctx context.Context, req *mcp.GetPromptRequest, conn *httpConnection, name string) (*mcp.GetPromptResult, error) {
	params := &mcp.GetPromptParams{
		Name:      name,
		Arguments: req.Params.Arguments,
	}

	return conn.session.GetPrompt(ctx, params)
}
