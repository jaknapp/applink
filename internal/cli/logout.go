package cli

import (
	"fmt"

	"github.com/jaknapp/applink/internal/config"
	"github.com/jaknapp/applink/internal/storage"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout <service>",
	Short: "Remove credentials for a service",
	Long:  `Remove stored credentials for a service from your system keychain.`,
	Example: `  applink logout slack
  applink logout notion`,
	Args: cobra.ExactArgs(1),
	RunE: runLogout,
}

func runLogout(cmd *cobra.Command, args []string) error {
	serviceName := args[0]

	// Verify service exists
	if _, err := config.GetService(serviceName); err != nil {
		return err
	}

	if err := storage.DeleteToken(serviceName); err != nil {
		return fmt.Errorf("failed to remove credentials: %w", err)
	}

	fmt.Printf("✓ Removed credentials for %s\n", serviceName)
	return nil
}
