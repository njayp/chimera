package vscode

import (
	"fmt"
	"log/slog"

	"github.com/njayp/chimera/clients/stdio"
	"github.com/njayp/chimera/clients/stream"
	"github.com/njayp/chimera/proxy"
)

// Config is the structure of a VSCode MCP configuration file.
type Config struct {
	Servers map[string]Server `json:"servers"`
}

// Server defines a single MCP server (stdio or HTTP).
type Server struct {
	Type    string            `json:"type,omitempty"`
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

func (c *Config) Clients() proxy.Clients {
	clients := make(proxy.Clients)
	for name, server := range c.Servers {
		switch server.Type {
		case "stdio":
			env := make([]string, 0, len(server.Env))
			for key, value := range server.Env {
				env = append(env, fmt.Sprintf("%s=%s", key, value))
			}
			clients[name] = stdio.NewClient(server.Command, server.Args, env)
		case "http":
			clients[name] = stream.NewClient(server.URL, server.Headers)
		default:
			slog.Error("unsupported server type", "name", name, "type", server.Type)
		}
	}

	return clients
}
