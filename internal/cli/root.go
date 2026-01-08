package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var debug bool

var rootCmd = &cobra.Command{
	Use:   "applink",
	Short: "Authenticate with SaaS apps and connect them to AI tools",
	Long: `applink is a CLI tool for authenticating with SaaS applications
(Slack, Notion, Linear, etc.) and connecting them to AI tools like Cursor via MCP.

It handles OAuth flows, stores credentials securely in your system keychain,
and automatically configures MCP servers.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug output")

	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(tokenCmd)
	rootCmd.AddCommand(mcpCmd)
	rootCmd.AddCommand(requestCmd)
}

func debugLog(format string, args ...interface{}) {
	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", args...)
	}
}
