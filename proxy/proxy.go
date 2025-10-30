package proxy

import (
	"context"
	"log/slog"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Client connects to an MCP server and returns a session.
type Client interface {
	Connect(ctx context.Context) (*mcp.ClientSession, error)
}

type proxy struct {
	server *mcp.Server
}

// proxyServer connects to a backend and registers its tools, resources, and prompts.
func (p *proxy) proxyServer(ctx context.Context, client Client, name string) {
	session, err := client.Connect(ctx)
	if err != nil {
		slog.Error("failed to connect to server", "name", name, "err", err)
		return
	}

	// Close session when context is cancelled
	go func() {
		<-ctx.Done()
		if err := session.Close(); err != nil {
			slog.Error("failed to close session", "name", name, "err", err)
		}
	}()

	if err := p.proxyTools(ctx, session, name); err != nil {
		slog.Error("failed to sync tools", "name", name, "err", err)
	}

	if err := p.proxyResources(ctx, session, name); err != nil {
		slog.Error("failed to sync resources", "name", name, "err", err)
	}

	if err := p.proxyPrompts(ctx, session, name); err != nil {
		slog.Error("failed to sync prompts", "name", name, "err", err)
	}
}

func (p *proxy) proxyTools(ctx context.Context, session *mcp.ClientSession, name string) error {
	for tool, err := range session.Tools(ctx, nil) {
		if err != nil {
			return err
		}

		// Prefix with server name unless already prefixed
		prefixedTool := *tool
		if !strings.HasPrefix(tool.Name, name) {
			prefixedTool.Name = name + "." + tool.Name
		}

		// Register tool that forwards calls to the backend
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
			return err
		}

		// Prefix URI unless already prefixed
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
