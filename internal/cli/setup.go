package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"github.com/jaknapp/applink/internal/config"
	"github.com/jaknapp/applink/internal/storage"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var setupCmd = &cobra.Command{
	Use:   "setup <service>",
	Short: "Configure OAuth credentials for a service",
	Long: `Set up OAuth client credentials for a service.

This will guide you through creating an OAuth application and
store the credentials securely in your system keychain.`,
	Example: `  applink setup slack
  applink setup notion`,
	Args: cobra.ExactArgs(1),
	RunE: runSetup,
}

var openBrowserFlag bool

func init() {
	setupCmd.Flags().BoolVarP(&openBrowserFlag, "open", "o", false, "Open the setup URL in your browser")
}

func runSetup(cmd *cobra.Command, args []string) error {
	serviceName := args[0]

	service, err := config.GetService(serviceName)
	if err != nil {
		return err
	}

	// Check if using environment variables
	envPrefix := fmt.Sprintf("APPLINK_%s_", strings.ToUpper(serviceName))
	if os.Getenv(envPrefix+"CLIENT_ID") != "" {
		fmt.Printf("Note: Environment variables are set for %s.\n", service.Name)
		fmt.Printf("Keychain credentials will be used as fallback when env vars are not set.\n\n")
	}

	fmt.Printf("Setting up %s OAuth credentials\n", service.Name)
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println()

	if service.SetupURL != "" {
		fmt.Printf("Create an OAuth app at: %s\n\n", service.SetupURL)
		if openBrowserFlag {
			openSetupURL(service.SetupURL)
		}
	}

	if service.SetupInstructions != "" {
		fmt.Println(service.SetupInstructions)
		fmt.Println()
	}

	fmt.Println(strings.Repeat("─", 50))

	// Prompt for credentials
	creds, err := promptCredentials(service)
	if err != nil {
		return err
	}

	// Store in keychain
	if err := storage.StoreCredentials(serviceName, creds); err != nil {
		if storage.IsKeychainError(err) {
			return err // Error message includes env var instructions
		}
		return fmt.Errorf("failed to store credentials: %w", err)
	}

	fmt.Printf("\n✓ Credentials saved to keychain for %s\n", service.Name)
	fmt.Printf("  Run 'applink login %s' to authenticate.\n", serviceName)
	return nil
}

func promptCredentials(service *config.Service) (*storage.Credentials, error) {
	reader := bufio.NewReader(os.Stdin)

	// Client ID (visible input)
	fmt.Print("Client ID: ")
	clientID, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read client ID: %w", err)
	}
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return nil, fmt.Errorf("client ID cannot be empty")
	}

	// Client Secret (hidden input)
	fmt.Print("Client Secret: ")
	var clientSecret string
	if term.IsTerminal(int(syscall.Stdin)) {
		secretBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return nil, fmt.Errorf("failed to read client secret: %w", err)
		}
		clientSecret = string(secretBytes)
		fmt.Println() // New line after hidden input
	} else {
		secret, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read client secret: %w", err)
		}
		clientSecret = strings.TrimSpace(secret)
	}

	if clientSecret == "" {
		return nil, fmt.Errorf("client secret cannot be empty")
	}

	return &storage.Credentials{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}, nil
}

func openSetupURL(url string) {
	// Reuse the browser opening logic from auth package
	// Import would create circular dependency, so we duplicate
	cmd := exec.Command("open", url)
	if runtime.GOOS == "linux" {
		cmd = exec.Command("xdg-open", url)
	} else if runtime.GOOS == "windows" {
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	}
	cmd.Start()
}
