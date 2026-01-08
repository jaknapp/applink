package config

import "fmt"

// AuthType represents the authentication method for a service
type AuthType string

const (
	AuthTypeOAuth  AuthType = "oauth"
	AuthTypeAPIKey AuthType = "apikey"
)

// Service defines a SaaS service that applink can authenticate with
type Service struct {
	ID       string   // Unique identifier (e.g., "slack")
	Name     string   // Display name (e.g., "Slack")
	AuthType AuthType // oauth or apikey

	// OAuth configuration
	AuthURL  string   // OAuth authorization URL
	TokenURL string   // OAuth token exchange URL
	Scopes   []string // OAuth scopes to request

	// API configuration
	APIURL string // Base URL for API requests

	// MCP configuration
	MCPPackage string            // npm package name for MCP server
	MCPEnvVars map[string]string // Environment variable mappings
}

// serviceRegistry holds all supported services
var serviceRegistry = map[string]*Service{
	"slack": {
		ID:       "slack",
		Name:     "Slack",
		AuthType: AuthTypeOAuth,
		AuthURL:  "https://slack.com/oauth/v2/authorize",
		TokenURL: "https://slack.com/api/oauth.v2.access",
		Scopes: []string{
			"channels:read",
			"channels:history",
			"groups:read",
			"groups:history",
			"chat:write",
			"users:read",
		},
		APIURL:     "https://slack.com",
		MCPPackage: "@modelcontextprotocol/server-slack",
		MCPEnvVars: map[string]string{
			"SLACK_USER_TOKEN": "access_token",
			"SLACK_TEAM_ID":    "team_id",
		},
	},
	"notion": {
		ID:       "notion",
		Name:     "Notion",
		AuthType: AuthTypeOAuth,
		AuthURL:  "https://api.notion.com/v1/oauth/authorize",
		TokenURL: "https://api.notion.com/v1/oauth/token",
		Scopes:   []string{}, // Notion doesn't use scopes in the same way
		APIURL:   "https://api.notion.com",
		MCPPackage: "@modelcontextprotocol/server-notion",
		MCPEnvVars: map[string]string{
			"NOTION_API_TOKEN": "access_token",
		},
	},
	"linear": {
		ID:       "linear",
		Name:     "Linear",
		AuthType: AuthTypeOAuth,
		AuthURL:  "https://linear.app/oauth/authorize",
		TokenURL: "https://api.linear.app/oauth/token",
		Scopes: []string{
			"read",
			"write",
			"issues:create",
			"comments:create",
		},
		APIURL:     "https://api.linear.app",
		MCPPackage: "@linear/mcp-server",
		MCPEnvVars: map[string]string{
			"LINEAR_API_KEY": "access_token",
		},
	},
	"honeycomb": {
		ID:         "honeycomb",
		Name:       "Honeycomb",
		AuthType:   AuthTypeAPIKey,
		APIURL:     "https://api.honeycomb.io",
		MCPPackage: "", // No MCP server yet
		MCPEnvVars: nil,
	},
}

// GetService returns the service definition for a given service name
func GetService(name string) (*Service, error) {
	service, ok := serviceRegistry[name]
	if !ok {
		return nil, fmt.Errorf("unknown service: %s\n\nSupported services: slack, notion, linear, honeycomb", name)
	}
	return service, nil
}

// AllServices returns all registered services
func AllServices() []*Service {
	services := make([]*Service, 0, len(serviceRegistry))
	for _, s := range serviceRegistry {
		services = append(services, s)
	}
	return services
}
