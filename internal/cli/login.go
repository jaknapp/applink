package cli

import (
	"fmt"
	"strings"

	"github.com/jaknapp/applink/internal/auth"
	"github.com/jaknapp/applink/internal/config"
	"github.com/jaknapp/applink/internal/storage"
	"github.com/spf13/cobra"
)

const defaultCallbackPort = 8888

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

	fmt.Printf("Authenticating with %s...\n", service.Name)

	var token *auth.Token
	switch service.AuthType {
	case config.AuthTypeOAuth:
		token, err = doOAuthLogin(service, serviceName)
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

func doOAuthLogin(service *config.Service, serviceName string) (*auth.Token, error) {
	// Get credentials (env vars → keychain)
	creds, err := storage.GetCredentials(serviceName)
	if err != nil {
		if storage.IsKeychainError(err) {
			return nil, err // Error includes env var instructions
		}
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	// No credentials found - guide user through setup
	if creds == nil {
		return nil, &credentialsNotFoundError{service: service, serviceName: serviceName}
	}

	// Convert to config.ClientCredentials for auth package
	clientCreds := config.ClientCredentials{
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
	}

	return auth.DoOAuthFlow(service, clientCreds, defaultCallbackPort)
}

type credentialsNotFoundError struct {
	service     *config.Service
	serviceName string
}

func (e *credentialsNotFoundError) Error() string {
	envPrefix := fmt.Sprintf("APPLINK_%s_", strings.ToUpper(e.serviceName))
	
	return fmt.Sprintf(`No OAuth credentials found for %s.

Option 1: Run setup (stores in keychain)
  applink setup %s

Option 2: Use environment variables
  export %sCLIENT_ID="your-client-id"
  export %sCLIENT_SECRET="your-client-secret"

To create an OAuth app, visit: %s`,
		e.service.Name,
		e.serviceName,
		envPrefix,
		envPrefix,
		e.service.SetupURL,
	)
}
