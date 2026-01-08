package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// CursorConfig represents the structure of ~/.cursor/mcp.json
type CursorConfig struct {
	MCPServers map[string]*ServerConfig `json:"mcpServers"`
}

// ServerConfig represents an MCP server configuration
type ServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

// cursorConfigPath returns the path to the Cursor MCP config file
func cursorConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cursor", "mcp.json"), nil
}

// LoadCursorConfig loads the Cursor MCP configuration
func LoadCursorConfig() (*CursorConfig, error) {
	path, err := cursorConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &CursorConfig{
				MCPServers: make(map[string]*ServerConfig),
			}, nil
		}
		return nil, err
	}

	var cfg CursorConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.MCPServers == nil {
		cfg.MCPServers = make(map[string]*ServerConfig)
	}

	return &cfg, nil
}

// SaveCursorConfig saves the Cursor MCP configuration
func SaveCursorConfig(cfg *CursorConfig) error {
	path, err := cursorConfigPath()
	if err != nil {
		return err
	}

	// Ensure .cursor directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
