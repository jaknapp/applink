package storage

import (
	"encoding/json"
	"time"

	"github.com/zalando/go-keyring"
)

const serviceName = "applink"

// Token represents stored authentication data
type Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	Scope        string    `json:"scope,omitempty"`

	// Service-specific fields
	TeamID string `json:"team_id,omitempty"` // Slack team ID
	User   string `json:"user,omitempty"`    // User email/name
}

// StoreToken saves a token to the system keychain
func StoreToken(service string, token *Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}

	return keyring.Set(serviceName, service, string(data))
}

// GetToken retrieves a token from the system keychain
func GetToken(service string) (*Token, error) {
	data, err := keyring.Get(serviceName, service)
	if err != nil {
		if err == keyring.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	var token Token
	if err := json.Unmarshal([]byte(data), &token); err != nil {
		return nil, err
	}

	return &token, nil
}

// DeleteToken removes a token from the system keychain
func DeleteToken(service string) error {
	err := keyring.Delete(serviceName, service)
	if err == keyring.ErrNotFound {
		return nil // Already deleted
	}
	return err
}
