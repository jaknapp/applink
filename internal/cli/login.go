package cli

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/jaknapp/applink/internal/auth"
	"github.com/jaknapp/applink/internal/certs"
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

	// Auto-initialize certificates if needed (for services requiring HTTPS like Slack)
	if err := ensureCertsInitialized(serviceName); err != nil {
		return nil, err
	}

	// Convert to config.ClientCredentials for auth package
	clientCreds := config.ClientCredentials{
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
	}

	return auth.DoOAuthFlow(service, clientCreds, defaultCallbackPort)
}

// ensureCertsInitialized checks if certificates are set up and initializes them if needed
func ensureCertsInitialized(serviceName string) error {
	// Only needed for services that require HTTPS (like Slack)
	if serviceName != "slack" {
		return nil
	}

	// Check if CA exists and is installed
	caExists, _ := certs.CAExists()
	if caExists && certs.IsCAInstalled() {
		return nil // Already set up
	}

	// Need to initialize
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("First-time setup: Installing trusted certificates")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("Slack requires HTTPS for OAuth callbacks. To avoid browser")
	fmt.Println("security warnings, applink will install a local certificate")
	fmt.Println("authority into your system's trust store.")
	fmt.Println()

	if runtime.GOOS == "linux" {
		fmt.Println("Note: This requires sudo access on Linux.")
		fmt.Println()
	}

	fmt.Print("Continue? [Y/n] ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "" && response != "y" && response != "yes" {
		fmt.Println()
		fmt.Println("Certificate setup skipped. Your browser will show a security warning.")
		fmt.Println("Click 'Advanced' → 'Proceed to localhost' to continue.")
		fmt.Println()
		return nil // Continue anyway, will fall back to self-signed
	}

	// Generate CA if it doesn't exist
	if !caExists {
		fmt.Println()
		fmt.Println("Generating certificate authority...")
		if err := certs.GenerateCA(); err != nil {
			fmt.Printf("Warning: Failed to generate CA: %v\n", err)
			fmt.Println("Continuing with self-signed certificate (browser will show warning)")
			fmt.Println()
			return nil
		}
		fmt.Println("✓ Certificate authority generated")
	}

	// Install CA
	fmt.Println("Installing to system trust store...")
	if err := certs.InstallCA(); err != nil {
		certPath, _ := certs.GetCACertPath()
		fmt.Printf("Warning: Failed to install CA: %v\n", err)
		fmt.Println()
		fmt.Println("You can install it manually:")
		switch runtime.GOOS {
		case "darwin":
			fmt.Printf("  security add-trusted-cert -r trustRoot -k login.keychain %s\n", certPath)
		case "linux":
			fmt.Printf("  sudo cp %s /usr/local/share/ca-certificates/applink-ca.crt\n", certPath)
			fmt.Println("  sudo update-ca-certificates")
		case "windows":
			fmt.Printf("  certutil -addstore -user Root %s\n", certPath)
		}
		fmt.Println()
		fmt.Println("Continuing with browser warning...")
		return nil
	}

	fmt.Println("✓ Certificate authority installed")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	return nil
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
