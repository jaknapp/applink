package cli

import (
	"fmt"

	"github.com/jaknapp/applink/internal/config"
	"github.com/jaknapp/applink/internal/storage"
	"github.com/spf13/cobra"
)

var tokenCmd = &cobra.Command{
	Use:   "token <service>",
	Short: "Print the access token for a service",
	Long: `Print the access token for a service to stdout.
Useful for scripts or debugging.`,
	Example: `  applink token slack
  applink token notion | pbcopy`,
	Args: cobra.ExactArgs(1),
	RunE: runToken,
}

func runToken(cmd *cobra.Command, args []string) error {
	serviceName := args[0]

	// Verify service exists
	if _, err := config.GetService(serviceName); err != nil {
		return err
	}

	token, err := storage.GetToken(serviceName)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}
	if token == nil {
		return fmt.Errorf("not authenticated with %s. Run: applink login %s", serviceName, serviceName)
	}

	fmt.Print(token.AccessToken)
	return nil
}
