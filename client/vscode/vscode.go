package vscode

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/njayp/chimera/client"
	"github.com/njayp/chimera/proxy"
)

// Config represents the structure of VSCode's MCP configuration file.
type Config struct {
	Servers map[string]Server `json:"servers"`
}

// Server represents an MCP server configuration entry for VSCode.
type Server struct {
	Type    string            `json:"type,omitempty"`
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// Clients loads MCP server configuration from a Clients-style JSON file.
func Clients(path string) (proxy.Clients, error) {
	config, err := readFile(path)
	if err != nil {
		return proxy.Clients{}, fmt.Errorf("failed to load MCP config: %w", err)
	}

	return clients(config), nil
}

func readFile(path string) (Config, error) {
	var config Config

	data, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to parse JSON config: %w", err)
	}

	return config, nil
}

func clients(config Config) proxy.Clients {
	clients := make(proxy.Clients)
	for name, server := range config.Servers {
		switch server.Type {
		case "stdio":
			s := client.Stdio{
				Command: server.Command,
				Args:    server.Args,
			}

			for key, value := range server.Env {
				s.Env = append(s.Env, fmt.Sprintf("%s=%s", key, value))
			}

			clients[name] = s
		case "http":
			s := client.HTTP{
				URL:     server.URL,
				Headers: server.Headers,
			}

			clients[name] = s
		default:
			slog.Error("unsupported server type", "name", name, "type", server.Type)
		}
	}

	return clients
}
