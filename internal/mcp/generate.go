package mcp

import (
	"fmt"
	"sort"

	"github.com/jaknapp/applink/internal/config"
	"github.com/jaknapp/applink/internal/storage"
)

// AddService adds or updates an MCP server configuration for a service
func AddService(serviceName string) error {
	service, err := config.GetService(serviceName)
	if err != nil {
		return err
	}

	if service.MCPPackage == "" {
		return fmt.Errorf("service %s does not have MCP support", serviceName)
	}

	token, err := storage.GetToken(serviceName)
	if err != nil {
		return err
	}
	if token == nil {
		return fmt.Errorf("no token found for %s", serviceName)
	}

	// Build environment variables
	env := make(map[string]string)
	for envVar, tokenField := range service.MCPEnvVars {
		var value string
		switch tokenField {
		case "access_token":
			value = token.AccessToken
		case "team_id":
			value = token.TeamID
		default:
			continue
		}
		if value != "" {
			env[envVar] = value
		}
	}

	// Load current config
	cfg, err := LoadCursorConfig()
	if err != nil {
		return err
	}

	// Add/update server config
	cfg.MCPServers[serviceName] = &ServerConfig{
		Command: "npx",
		Args:    []string{"-y", service.MCPPackage},
		Env:     env,
	}

	// Save config
	return SaveCursorConfig(cfg)
}

// RemoveService removes an MCP server configuration
func RemoveService(serviceName string) error {
	cfg, err := LoadCursorConfig()
	if err != nil {
		return err
	}

	delete(cfg.MCPServers, serviceName)

	return SaveCursorConfig(cfg)
}

// ListServers returns the names of all configured MCP servers
func ListServers() ([]string, error) {
	cfg, err := LoadCursorConfig()
	if err != nil {
		return nil, err
	}

	servers := make([]string, 0, len(cfg.MCPServers))
	for name := range cfg.MCPServers {
		servers = append(servers, name)
	}
	sort.Strings(servers)

	return servers, nil
}
