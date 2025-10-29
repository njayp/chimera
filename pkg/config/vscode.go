// Package config provides configuration file loading and parsing.
package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/njayp/chimera/pkg/proxy"
)

// Config represents the structure of VSCode's MCP configuration file.
type Config struct {
	Inputs  []Input              `json:"inputs,omitempty"`
	Servers map[string]MCPServer `json:"servers"`
}

// Input represents a VSCode input variable configuration.
type Input struct {
	Type        string `json:"type"`
	ID          string `json:"id"`
	Description string `json:"description"`
	Password    bool   `json:"password,omitempty"`
}

// MCPServer represents an MCP server configuration entry for VSCode.
type MCPServer struct {
	Type    string            `json:"type,omitempty"`
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// VSCode loads MCP server configuration from a VSCode-style JSON file.
func VSCode(path string) (proxy.Servers, error) {
	servers := proxy.Servers{
		StdioServers: make(map[string]proxy.StdioClient),
		HTTPServers:  make(map[string]proxy.HTTPClient),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return servers, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return servers, fmt.Errorf("failed to parse JSON config: %w", err)
	}

	for name, server := range config.Servers {
		switch server.Type {
		case "stdio":
			s := proxy.StdioClient{
				Command: server.Command,
				Args:    server.Args,
			}

			for key, value := range server.Env {
				s.Env = append(s.Env, fmt.Sprintf("%s=%s", key, value))
			}

			servers.StdioServers[name] = s

		case "http":
			s := proxy.HTTPClient{
				URL:     server.URL,
				Headers: server.Headers,
			}

			servers.HTTPServers[name] = s

		default:
			return servers, fmt.Errorf("unsupported server type: %s", server.Type)
		}
	}

	return servers, nil
}
