package cli

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/jaknapp/applink/internal/certs"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize applink with trusted certificates",
	Long: `Initialize applink by generating and installing a local Certificate Authority.

This is normally run automatically the first time you login to a service
that requires HTTPS (like Slack). You can also run it manually.

The CA certificate will be installed into your system's trust store:
  macOS:   login keychain
  Linux:   system CA certificates (requires sudo)
  Windows: user certificate store

Use --force to regenerate the CA certificate if needed.`,
	Example: `  applink init
  applink init --force`,
	Args:    cobra.NoArgs,
	RunE:    runInit,
}

var forceInit bool

func init() {
	initCmd.Flags().BoolVarP(&forceInit, "force", "f", false, "Force regeneration of CA certificate")
}

func runInit(cmd *cobra.Command, args []string) error {
	fmt.Println("Initializing applink")
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println()

	// Check if CA already exists
	exists, err := certs.CAExists()
	if err != nil {
		return fmt.Errorf("failed to check CA status: %w", err)
	}

	if exists && !forceInit {
		// Check if it's installed
		if certs.IsCAInstalled() {
			fmt.Println("✓ applink is already initialized with trusted certificates.")
			fmt.Println("  Use --force to regenerate the CA certificate.")
			return nil
		}
		fmt.Println("CA certificate exists but is not installed in the system trust store.")
		fmt.Println()
	}

	// Generate CA if needed
	if !exists || forceInit {
		fmt.Println("Generating local Certificate Authority...")
		if err := certs.GenerateCA(); err != nil {
			return fmt.Errorf("failed to generate CA: %w", err)
		}
		fmt.Println("✓ CA certificate generated")
		fmt.Println()
	}

	// Ask for confirmation before installing
	certPath, _ := certs.GetCACertPath()
	fmt.Println("To enable HTTPS without browser warnings, the CA certificate needs to")
	fmt.Println("be installed in your system's trust store.")
	fmt.Println()
	fmt.Printf("Certificate location: %s\n", certPath)
	fmt.Println()

	if runtime.GOOS == "linux" {
		fmt.Println("Note: On Linux, this requires sudo access.")
		fmt.Println()
	}

	fmt.Print("Install the CA certificate now? [Y/n] ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "" && response != "y" && response != "yes" {
		fmt.Println()
		fmt.Println("CA certificate was generated but not installed.")
		fmt.Println("You can install it manually or run 'applink init' again.")
		fmt.Println()
		fmt.Println("Manual installation:")
		printManualInstructions(certPath)
		return nil
	}

	// Install CA
	fmt.Println()
	fmt.Println("Installing CA certificate...")
	if err := certs.InstallCA(); err != nil {
		fmt.Println()
		fmt.Printf("Failed to install CA: %v\n", err)
		fmt.Println()
		fmt.Println("You can install it manually:")
		printManualInstructions(certPath)
		return nil
	}

	fmt.Println("✓ CA certificate installed")
	fmt.Println()
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println("✓ applink is ready! HTTPS OAuth callbacks will now work without warnings.")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Run 'applink setup <service>' to configure OAuth credentials")
	fmt.Println("  2. Run 'applink login <service>' to authenticate")

	return nil
}

func printManualInstructions(certPath string) {
	switch runtime.GOOS {
	case "darwin":
		fmt.Println("  macOS:")
		fmt.Printf("    security add-trusted-cert -r trustRoot -k login.keychain %s\n", certPath)
	case "linux":
		fmt.Println("  Ubuntu/Debian:")
		fmt.Printf("    sudo cp %s /usr/local/share/ca-certificates/applink-ca.crt\n", certPath)
		fmt.Println("    sudo update-ca-certificates")
		fmt.Println()
		fmt.Println("  RHEL/Fedora:")
		fmt.Printf("    sudo cp %s /etc/pki/ca-trust/source/anchors/applink-ca.crt\n", certPath)
		fmt.Println("    sudo update-ca-trust")
	case "windows":
		fmt.Println("  Windows:")
		fmt.Printf("    certutil -addstore -user Root %s\n", certPath)
	}
	fmt.Println()
}
