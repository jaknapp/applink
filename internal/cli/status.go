package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/jaknapp/applink/internal/config"
	"github.com/jaknapp/applink/internal/storage"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status for all services",
	Long:  `Display the authentication status for all supported services.`,
	RunE:  runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	services := config.AllServices()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "SERVICE\tSTATUS\tUSER\tEXPIRES")

	for _, service := range services {
		token, err := storage.GetToken(service.ID)
		if err != nil || token == nil {
			fmt.Fprintf(w, "%s\t✗ not configured\t\t\n", service.ID)
			continue
		}

		// Check if token is expired
		status := "✓ active"
		expires := "never"
		if !token.ExpiresAt.IsZero() {
			if token.ExpiresAt.Before(time.Now()) {
				status = "✗ expired"
			}
			expires = token.ExpiresAt.Format("2006-01-02")
		}

		user := token.User
		if user == "" {
			user = "-"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", service.ID, status, user, expires)
	}

	return w.Flush()
}
