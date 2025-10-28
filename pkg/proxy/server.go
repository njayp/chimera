package proxy

import (
	"context"
	"log/slog"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type client interface {
	connect(ctx context.Context) (*mcp.ClientSession, error)
}

type proxy struct {
	server *mcp.Server
}

// proxyServer establishes a connection to a backend MCP server
// and syncs its capabilities (tools, resources, prompts).
func (s *proxy) proxyServer(ctx context.Context, client client, name string) {
	// Establish connection to the stdio server
	session, err := client.connect(ctx)
	if err != nil {
		slog.Error("failed to connect to server", "name", name, "err", err)
		// if connection fails, skip this server
		return
	}

	// Fetch and register all tools from this server
	if err := s.proxyTools(ctx, session, name); err != nil {
		slog.Error("failed to sync tools", "name", name, "err", err)
	}

	// Fetch and register all resources from this server
	if err := s.proxyResources(ctx, session, name); err != nil {
		slog.Error("failed to sync resources", "name", name, "err", err)
	}

	// Fetch and register all prompts from this server
	if err := s.proxyPrompts(ctx, session, name); err != nil {
		slog.Error("failed to sync prompts", "name", name, "err", err)
	}
}

//
// Proxy functions to import tools, resources, and prompts from backend servers
//

func (s *proxy) proxyTools(ctx context.Context, session *mcp.ClientSession, name string) error {
	for tool, err := range session.Tools(ctx, nil) {
		if err != nil {
			return err
		}

		// Prefix tool name with server name to avoid conflicts
		prefixedTool := *tool
		if !strings.HasPrefix(tool.Name, name) {
			prefixedTool.Name = name + "." + tool.Name
		}

		// Add tool to our aggregating server
		s.server.AddTool(&prefixedTool, func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return s.routeToolCall(ctx, req, session, name)
		})
	}
	return nil
}

func (s *proxy) proxyResources(ctx context.Context, session *mcp.ClientSession, name string) error {
	for resource, err := range session.Resources(ctx, nil) {
		if err != nil {
			return err
		}

		// Create a prefixed URI to avoid conflicts
		prefixedResource := *resource
		if !strings.HasPrefix(resource.URI, name) {
			prefixedResource.URI = name + "://" + resource.URI
		}

		s.server.AddResource(&prefixedResource, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			return s.routeResourceRead(ctx, req, session, resource.URI)
		})
	}
	return nil
}

func (s *proxy) proxyPrompts(ctx context.Context, session *mcp.ClientSession, name string) error {
	for prompt, err := range session.Prompts(ctx, nil) {
		if err != nil {
			return err
		}

		prefixedPrompt := *prompt
		if !strings.HasPrefix(prompt.Name, name) {
			prefixedPrompt.Name = name + "." + prompt.Name
		}

		s.server.AddPrompt(&prefixedPrompt, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			return s.routePromptGet(ctx, req, session, prompt.Name)
		})
	}
	return nil
}

//
// Routing functions to forward requests to the appropriate backend server
//

func (s *proxy) routeToolCall(ctx context.Context, req *mcp.CallToolRequest, session *mcp.ClientSession, name string) (*mcp.CallToolResult, error) {
	params := &mcp.CallToolParams{
		Name:      name,
		Arguments: req.Params.Arguments,
	}

	return session.CallTool(ctx, params)
}

func (s *proxy) routeResourceRead(ctx context.Context, _ *mcp.ReadResourceRequest, session *mcp.ClientSession, uri string) (*mcp.ReadResourceResult, error) {
	params := &mcp.ReadResourceParams{
		URI: uri,
	}

	return session.ReadResource(ctx, params)
}

func (s *proxy) routePromptGet(ctx context.Context, req *mcp.GetPromptRequest, session *mcp.ClientSession, name string) (*mcp.GetPromptResult, error) {
	params := &mcp.GetPromptParams{
		Name:      name,
		Arguments: req.Params.Arguments,
	}

	return session.GetPrompt(ctx, params)
}
