package cli

import (
	"fmt"

	"github.com/jaknapp/applink/internal/auth"
	"github.com/jaknapp/applink/internal/config"
	"github.com/jaknapp/applink/internal/storage"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login <service>",
	Short: "Authenticate with a service",
	Long: `Authenticate with a SaaS service via OAuth or API key.

For OAuth services (slack, notion, linear), this opens your browser
to complete the authentication flow.

For API key services (honeycomb), this prompts you for the key.`,
	Example: `  applink login slack
  applink login notion
  applink login honeycomb`,
	Args: cobra.ExactArgs(1),
	RunE: runLogin,
}

func runLogin(cmd *cobra.Command, args []string) error {
	serviceName := args[0]

	// Get service definition
	service, err := config.GetService(serviceName)
	if err != nil {
		return err
	}

	// Load user config (for client credentials)
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get client credentials for this service
	clientCreds, ok := cfg.Services[serviceName]
	if !ok {
		return fmt.Errorf("no client credentials configured for %s\n\nAdd credentials to ~/.config/applink/config.yaml:\n\nservices:\n  %s:\n    client_id: \"your-client-id\"\n    client_secret: \"your-client-secret\"", serviceName, serviceName)
	}

	fmt.Printf("Authenticating with %s...\n", service.Name)

	var token *auth.Token
	switch service.AuthType {
	case config.AuthTypeOAuth:
		token, err = auth.DoOAuthFlow(service, clientCreds, cfg.Settings.CallbackPort)
	case config.AuthTypeAPIKey:
		token, err = auth.PromptAPIKey(service)
	default:
		return fmt.Errorf("unsupported auth type: %s", service.AuthType)
	}

	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Store token
	if err := storage.StoreToken(serviceName, token); err != nil {
		return fmt.Errorf("failed to store token: %w", err)
	}

	fmt.Printf("✓ Successfully authenticated with %s\n", service.Name)
	return nil
}
