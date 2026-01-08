package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/jaknapp/applink/internal/config"
	"golang.org/x/term"
)

// PromptAPIKey prompts the user for an API key
func PromptAPIKey(service *config.Service) (*Token, error) {
	fmt.Printf("Enter your %s API key: ", service.Name)

	// Try to read password securely (hidden input)
	var apiKey string
	if term.IsTerminal(int(syscall.Stdin)) {
		keyBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return nil, fmt.Errorf("failed to read API key: %w", err)
		}
		apiKey = string(keyBytes)
		fmt.Println() // New line after hidden input
	} else {
		// Fallback for non-terminal input (e.g., piped)
		reader := bufio.NewReader(os.Stdin)
		key, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read API key: %w", err)
		}
		apiKey = strings.TrimSpace(key)
	}

	if apiKey == "" {
		return nil, fmt.Errorf("API key cannot be empty")
	}

	return &Token{
		AccessToken: apiKey,
		TokenType:   "apikey",
	}, nil
}
