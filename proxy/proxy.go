package proxy

import (
	"context"
	"log/slog"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Client connects to an MCP server and returns a session.
type Client interface {
	Transport(ctx context.Context) mcp.Transport
}

type proxy struct {
	server *mcp.Server
}

type cache struct {
	sync.Mutex
	names map[string]bool
}

// proxyServer connects to a backend and registers its tools, resources, and prompts.
func (p *proxy) proxyServer(ctx context.Context, name string, client Client) {
	transport := client.Transport(ctx)
	if transport == nil {
		slog.Error("no transport available for client", "name", name)
		return
	}

	tools := &cache{
		names: make(map[string]bool),
	}
	prompts := &cache{
		names: make(map[string]bool),
	}
	resources := &cache{
		names: make(map[string]bool),
	}

	c := mcp.NewClient(&mcp.Implementation{
		Name: "chimera",
	}, &mcp.ClientOptions{
		ToolListChangedHandler: func(ctx context.Context, req *mcp.ToolListChangedRequest) {
			p.proxyTools(ctx, name, req.Session, tools)
		},
		PromptListChangedHandler: func(ctx context.Context, req *mcp.PromptListChangedRequest) {
			p.proxyPrompts(ctx, name, req.Session, prompts)
		},
		ResourceListChangedHandler: func(ctx context.Context, req *mcp.ResourceListChangedRequest) {
			p.proxyResources(ctx, name, req.Session, resources)
		},

		// TODO
		ResourceUpdatedHandler: nil,
	})

	session, err := c.Connect(ctx, transport, nil)
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

	wg := sync.WaitGroup{}
	wg.Go(func() {
		p.proxyTools(ctx, name, session, tools)
	})
	wg.Go(func() {
		p.proxyPrompts(ctx, name, session, prompts)
	})
	wg.Go(func() {
		p.proxyResources(ctx, name, session, resources)
	})
	wg.Wait()
}

func (p *proxy) proxyTools(ctx context.Context, name string, session *mcp.ClientSession, cache *cache) {
	// Gather prompts
	var tools []*mcp.Tool
	for tool, err := range session.Tools(ctx, nil) {
		if err != nil {
			// Connection failure
			slog.Error("failed to list prompts", "name", name, "err", err)
			return
		}

		tools = append(tools, tool)
	}

	cache.Lock()
	defer cache.Unlock()

	// Register tools
	for _, tool := range tools {
		// Prefix name for uniqueness, but save name for callback
		oldName := tool.Name
		tool.Name = name + "." + tool.Name

		if _, ok := cache.names[tool.Name]; ok {
			// Already registered
			cache.names[tool.Name] = true
			continue
		}
		// mark as registered
		cache.names[tool.Name] = true

		p.server.AddTool(tool, func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			params := &mcp.CallToolParams{
				Name:      oldName,
				Arguments: req.Params.Arguments,
			}

			return session.CallTool(ctx, params)
		})
	}

	// Unregister prompts that are no longer present
	var rm []string
	for name, registered := range cache.names {
		if !registered {
			rm = append(rm, name)
			delete(cache.names, name)
		} else {
			// reset for next iteration
			cache.names[name] = false
		}
	}
	p.server.RemovePrompts(rm...)
}

func (p *proxy) proxyResources(ctx context.Context, name string, session *mcp.ClientSession, cache *cache) {
	for resource, err := range session.Resources(ctx, nil) {
		if err != nil {
			slog.Error("failed to list resources", "name", name, "err", err)
			return
		}

		// Prefix URI unless already prefixed
		prefixedResource := *resource
		prefixedResource.URI = name + "." + resource.URI

		p.server.AddResource(&prefixedResource, func(ctx context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			params := &mcp.ReadResourceParams{
				URI: resource.URI,
			}

			return session.ReadResource(ctx, params)
		})
	}
}

func (p *proxy) proxyPrompts(ctx context.Context, name string, session *mcp.ClientSession, cache *cache) {
	// Gather prompts
	var prompts []*mcp.Prompt
	for prompt, err := range session.Prompts(ctx, nil) {
		if err != nil {
			// Connection failure
			slog.Error("failed to list prompts", "name", name, "err", err)
			return
		}

		prompts = append(prompts, prompt)
	}

	cache.Lock()
	defer cache.Unlock()

	// Register prompts
	for _, prompt := range prompts {
		// Prefix name for uniqueness, but save name for callback
		oldName := prompt.Name
		prompt.Name = name + "." + prompt.Name

		if _, ok := cache.names[prompt.Name]; ok {
			// Already registered
			cache.names[prompt.Name] = true
			continue
		}
		// mark as registered
		cache.names[prompt.Name] = true

		p.server.AddPrompt(prompt, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			params := &mcp.GetPromptParams{
				Name:      oldName,
				Arguments: req.Params.Arguments,
			}

			return session.GetPrompt(ctx, params)
		})
	}

	// Unregister prompts that are no longer present
	var rm []string
	for name, registered := range cache.names {
		if !registered {
			rm = append(rm, name)
			delete(cache.names, name)
		} else {
			// reset for next iteration
			cache.names[name] = false
		}
	}
	p.server.RemovePrompts(rm...)
}
