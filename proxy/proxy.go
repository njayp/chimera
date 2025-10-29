package proxy

import (
	"context"
	"log/slog"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Client represents a generic MCP client that can establish a connection.
type Client interface {
	Connect(ctx context.Context) (*mcp.ClientSession, error)
}

type proxy struct {
	server *mcp.Server
}

// proxyServer establishes a connection to a backend MCP server
// and syncs its capabilities (tools, resources, prompts).
func (p *proxy) proxyServer(ctx context.Context, client Client, name string) {
	// Establish connection to the server
	session, err := client.Connect(ctx)
	if err != nil {
		slog.Error("failed to connect to server", "name", name, "err", err)
		// if connection fails, skip this server
		return
	}

	// Fetch and register all tools from this server
	if err := p.proxyTools(ctx, session, name); err != nil {
		slog.Error("failed to sync tools", "name", name, "err", err)
	}

	// Fetch and register all resources from this server
	if err := p.proxyResources(ctx, session, name); err != nil {
		slog.Error("failed to sync resources", "name", name, "err", err)
	}

	// Fetch and register all prompts from this server
	if err := p.proxyPrompts(ctx, session, name); err != nil {
		slog.Error("failed to sync prompts", "name", name, "err", err)
	}
}

//
// Proxy functions to import tools, resources, and prompts from backend servers
//

func (p *proxy) proxyTools(ctx context.Context, session *mcp.ClientSession, name string) error {
	for tool, err := range session.Tools(ctx, nil) {
		if err != nil {
			// iteration stops at first error
			return err
		}

		// Prefix tool name with server name to avoid conflicts
		prefixedTool := *tool
		if !strings.HasPrefix(tool.Name, name) {
			prefixedTool.Name = name + "." + tool.Name
		}

		// Add tool to our aggregating server
		p.server.AddTool(&prefixedTool, func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			params := &mcp.CallToolParams{
				Name:      tool.Name,
				Arguments: req.Params.Arguments,
			}

			return session.CallTool(ctx, params)
		})
	}

	return nil
}

func (p *proxy) proxyResources(ctx context.Context, session *mcp.ClientSession, name string) error {
	for resource, err := range session.Resources(ctx, nil) {
		if err != nil {
			// iteration stops at first error
			return err
		}

		// Create a prefixed URI to avoid conflicts
		prefixedResource := *resource
		if !strings.HasPrefix(resource.URI, name) {
			prefixedResource.URI = name + "." + resource.URI
		}

		p.server.AddResource(&prefixedResource, func(ctx context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			params := &mcp.ReadResourceParams{
				URI: resource.URI,
			}

			return session.ReadResource(ctx, params)
		})
	}

	return nil
}

func (p *proxy) proxyPrompts(ctx context.Context, session *mcp.ClientSession, name string) error {
	for prompt, err := range session.Prompts(ctx, nil) {
		if err != nil {
			// iteration stops at first error
			return err
		}

		prefixedPrompt := *prompt
		if !strings.HasPrefix(prompt.Name, name) {
			prefixedPrompt.Name = name + "." + prompt.Name
		}

		p.server.AddPrompt(&prefixedPrompt, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			params := &mcp.GetPromptParams{
				Name:      prompt.Name,
				Arguments: req.Params.Arguments,
			}

			return session.GetPrompt(ctx, params)
		})
	}

	return nil
}
