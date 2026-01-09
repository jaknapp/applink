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

	// Setup instructions
	SetupURL          string // URL to create OAuth app
	SetupInstructions string // Step-by-step instructions
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
			"SLACK_BOT_TOKEN": "access_token",
			"SLACK_TEAM_ID":   "team_id",
		},
		SetupURL: "https://api.slack.com/apps",
		SetupInstructions: `1. Go to https://api.slack.com/apps
2. Click "Create New App" → "From scratch"
3. Name your app (e.g., "applink") and select your workspace
4. Go to "OAuth & Permissions" in the sidebar
5. Under "Redirect URLs", add: https://localhost:8888/callback
6. Under "User Token Scopes", add these scopes:
   - channels:read, channels:history
   - groups:read, groups:history
   - chat:write, users:read
7. Go to "Basic Information" to find your Client ID and Client Secret`,
	},
	"notion": {
		ID:         "notion",
		Name:       "Notion",
		AuthType:   AuthTypeOAuth,
		AuthURL:    "https://api.notion.com/v1/oauth/authorize",
		TokenURL:   "https://api.notion.com/v1/oauth/token",
		Scopes:     []string{}, // Notion doesn't use scopes in the same way
		APIURL:     "https://api.notion.com",
		MCPPackage: "@modelcontextprotocol/server-notion",
		MCPEnvVars: map[string]string{
			"NOTION_API_TOKEN": "access_token",
		},
		SetupURL: "https://www.notion.so/my-integrations",
		SetupInstructions: `1. Go to https://www.notion.so/my-integrations
2. Click "New integration"
3. Name your integration (e.g., "applink")
4. Select the workspace to install it in
5. Under "Capabilities", ensure it has the access you need
6. Set the redirect URI to: https://localhost:8888/callback
7. Copy the "OAuth client ID" and "OAuth client secret"

Note: After authenticating, you must share specific pages with the integration.`,
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
		SetupURL: "https://linear.app/settings/api",
		SetupInstructions: `1. Go to https://linear.app/settings/api
2. Under "OAuth applications", click "Create new"
3. Name your application (e.g., "applink")
4. Set the redirect URI to: https://localhost:8888/callback
5. Select the scopes: read, write, issues:create, comments:create
6. Copy the "Client ID" and "Client Secret"`,
	},
	"honeycomb": {
		ID:         "honeycomb",
		Name:       "Honeycomb",
		AuthType:   AuthTypeAPIKey,
		APIURL:     "https://api.honeycomb.io",
		MCPPackage: "", // No MCP server yet
		MCPEnvVars: nil,
		SetupURL:   "https://ui.honeycomb.io/account",
		SetupInstructions: `1. Go to https://ui.honeycomb.io/account
2. Navigate to "Team settings" → "API Keys"
3. Create a new API key with the permissions you need
4. Copy the API key`,
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

// ServiceNames returns the names of all registered services
func ServiceNames() []string {
	names := make([]string, 0, len(serviceRegistry))
	for name := range serviceRegistry {
		names = append(names, name)
	}
	return names
}
