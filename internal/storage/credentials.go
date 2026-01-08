package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/zalando/go-keyring"
)

const credentialsPrefix = "applink_creds"

// Credentials holds OAuth client credentials for a service
type Credentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// GetCredentials retrieves client credentials for a service.
// Priority: Environment variables → Keychain
func GetCredentials(service string) (*Credentials, error) {
	// 1. Check environment variables first
	envPrefix := fmt.Sprintf("APPLINK_%s_", strings.ToUpper(service))
	clientID := os.Getenv(envPrefix + "CLIENT_ID")
	clientSecret := os.Getenv(envPrefix + "CLIENT_SECRET")

	if clientID != "" && clientSecret != "" {
		return &Credentials{
			ClientID:     clientID,
			ClientSecret: clientSecret,
		}, nil
	}

	// 2. Try keychain
	creds, err := getCredentialsFromKeychain(service)
	if err != nil {
		return nil, err
	}

	return creds, nil
}

// StoreCredentials saves client credentials to the system keychain
func StoreCredentials(service string, creds *Credentials) error {
	data, err := json.Marshal(creds)
	if err != nil {
		return err
	}

	err = keyring.Set(credentialsPrefix, service, string(data))
	if err != nil {
		return fmt.Errorf("failed to store credentials in keychain: %w\n\n%s", err, keychainHelpMessage(service))
	}

	return nil
}

// DeleteCredentials removes client credentials from the keychain
func DeleteCredentials(service string) error {
	err := keyring.Delete(credentialsPrefix, service)
	if err == keyring.ErrNotFound {
		return nil
	}
	return err
}

// HasCredentials checks if credentials exist (env vars or keychain)
func HasCredentials(service string) bool {
	creds, _ := GetCredentials(service)
	return creds != nil
}

// getCredentialsFromKeychain retrieves credentials from the system keychain
func getCredentialsFromKeychain(service string) (*Credentials, error) {
	data, err := keyring.Get(credentialsPrefix, service)
	if err != nil {
		if err == keyring.ErrNotFound {
			return nil, nil
		}
		// Keychain not available or other error
		return nil, &KeychainError{Err: err, Service: service}
	}

	var creds Credentials
	if err := json.Unmarshal([]byte(data), &creds); err != nil {
		return nil, err
	}

	return &creds, nil
}

// KeychainError indicates the system keychain is not available
type KeychainError struct {
	Err     error
	Service string
}

func (e *KeychainError) Error() string {
	return fmt.Sprintf("keychain error: %v", e.Err)
}

func (e *KeychainError) Unwrap() error {
	return e.Err
}

// IsKeychainError checks if an error is a keychain availability error
func IsKeychainError(err error) bool {
	_, ok := err.(*KeychainError)
	return ok
}

func keychainHelpMessage(service string) string {
	envPrefix := fmt.Sprintf("APPLINK_%s_", strings.ToUpper(service))
	return fmt.Sprintf(`System keychain is not available.

Use environment variables instead:
  export %sCLIENT_ID="your-client-id"
  export %sCLIENT_SECRET="your-client-secret"

Then run: applink login %s`, envPrefix, envPrefix, service)
}
