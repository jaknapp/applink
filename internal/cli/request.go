package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/jaknapp/applink/internal/config"
	"github.com/jaknapp/applink/internal/storage"
	"github.com/spf13/cobra"
)

var requestData string

var requestCmd = &cobra.Command{
	Use:   "request <service> <method> <path>",
	Short: "Send an authenticated API request",
	Long: `Send an authenticated HTTP request to a service's API.
The request will automatically include the appropriate authentication headers.`,
	Example: `  applink request slack GET /api/conversations.list
  applink request notion POST /v1/search --data '{"query": "meeting notes"}'
  applink request linear POST /graphql --data '{"query": "{ viewer { id } }"}'`,
	Args: cobra.ExactArgs(3),
	RunE: runRequest,
}

func init() {
	requestCmd.Flags().StringVarP(&requestData, "data", "d", "", "Request body (JSON)")
}

func runRequest(cmd *cobra.Command, args []string) error {
	serviceName := args[0]
	method := strings.ToUpper(args[1])
	path := args[2]

	service, err := config.GetService(serviceName)
	if err != nil {
		return err
	}

	token, err := storage.GetToken(serviceName)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}
	if token == nil {
		return fmt.Errorf("not authenticated with %s. Run: applink login %s", serviceName, serviceName)
	}

	// Build URL
	url := service.APIURL + path

	// Create request
	var body io.Reader
	if requestData != "" {
		body = bytes.NewBufferString(requestData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add auth header based on service type
	switch serviceName {
	case "notion":
		req.Header.Set("Authorization", "Bearer "+token.AccessToken)
		req.Header.Set("Notion-Version", "2022-06-28")
	case "linear":
		req.Header.Set("Authorization", token.AccessToken)
	default:
		req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	}

	if requestData != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	debugLog("Request: %s %s", method, url)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read and format response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Pretty print JSON
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, respBody, "", "  "); err != nil {
		// Not JSON, print raw
		fmt.Println(string(respBody))
	} else {
		fmt.Println(prettyJSON.String())
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("request returned status %d", resp.StatusCode)
	}

	return nil
}
