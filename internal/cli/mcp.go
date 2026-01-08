package cli

import (
	"fmt"

	"github.com/jaknapp/applink/internal/config"
	"github.com/jaknapp/applink/internal/mcp"
	"github.com/jaknapp/applink/internal/storage"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Manage MCP server configuration for Cursor",
	Long:  `Configure MCP servers in Cursor based on your authenticated services.`,
}

var mcpInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Configure MCP servers for all authenticated services",
	Long: `Automatically configure Cursor's MCP servers for all services
you have authenticated with.`,
	RunE: runMCPInstall,
}

var mcpAddCmd = &cobra.Command{
	Use:   "add <service>",
	Short: "Add MCP server for a specific service",
	Args:  cobra.ExactArgs(1),
	RunE:  runMCPAdd,
}

var mcpRemoveCmd = &cobra.Command{
	Use:   "remove <service>",
	Short: "Remove MCP server for a specific service",
	Args:  cobra.ExactArgs(1),
	RunE:  runMCPRemove,
}

var mcpListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured MCP servers",
	RunE:  runMCPList,
}

func init() {
	mcpCmd.AddCommand(mcpInstallCmd)
	mcpCmd.AddCommand(mcpAddCmd)
	mcpCmd.AddCommand(mcpRemoveCmd)
	mcpCmd.AddCommand(mcpListCmd)
}

func runMCPInstall(cmd *cobra.Command, args []string) error {
	services := config.AllServices()
	var authenticated []string

	for _, service := range services {
		if service.MCPPackage == "" {
			continue // No MCP server for this service
		}
		token, err := storage.GetToken(service.ID)
		if err != nil || token == nil {
			continue
		}
		authenticated = append(authenticated, service.ID)
	}

	if len(authenticated) == 0 {
		fmt.Println("No authenticated services found with MCP support.")
		fmt.Println("Run 'applink login <service>' first.")
		return nil
	}

	fmt.Printf("→ Found tokens for: %v\n", authenticated)

	for _, serviceName := range authenticated {
		if err := mcp.AddService(serviceName); err != nil {
			fmt.Printf("  ✗ Failed to add %s: %v\n", serviceName, err)
		} else {
			fmt.Printf("  ✓ Added %s\n", serviceName)
		}
	}

	fmt.Println("→ Updated ~/.cursor/mcp.json")
	fmt.Println("✓ MCP servers configured. Restart Cursor to activate.")
	return nil
}

func runMCPAdd(cmd *cobra.Command, args []string) error {
	serviceName := args[0]

	service, err := config.GetService(serviceName)
	if err != nil {
		return err
	}

	if service.MCPPackage == "" {
		return fmt.Errorf("%s does not have MCP server support", serviceName)
	}

	token, err := storage.GetToken(serviceName)
	if err != nil || token == nil {
		return fmt.Errorf("not authenticated with %s. Run: applink login %s", serviceName, serviceName)
	}

	if err := mcp.AddService(serviceName); err != nil {
		return fmt.Errorf("failed to add MCP server: %w", err)
	}

	fmt.Printf("✓ Added MCP server for %s. Restart Cursor to activate.\n", serviceName)
	return nil
}

func runMCPRemove(cmd *cobra.Command, args []string) error {
	serviceName := args[0]

	if _, err := config.GetService(serviceName); err != nil {
		return err
	}

	if err := mcp.RemoveService(serviceName); err != nil {
		return fmt.Errorf("failed to remove MCP server: %w", err)
	}

	fmt.Printf("✓ Removed MCP server for %s. Restart Cursor to activate.\n", serviceName)
	return nil
}

func runMCPList(cmd *cobra.Command, args []string) error {
	servers, err := mcp.ListServers()
	if err != nil {
		return fmt.Errorf("failed to list MCP servers: %w", err)
	}

	if len(servers) == 0 {
		fmt.Println("No MCP servers configured.")
		return nil
	}

	fmt.Println("Configured MCP servers:")
	for _, name := range servers {
		fmt.Printf("  • %s\n", name)
	}
	return nil
}
