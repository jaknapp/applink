package config

// ClientCredentials holds OAuth client credentials for a service
// Used by the auth package for OAuth flows
type ClientCredentials struct {
	ClientID     string
	ClientSecret string
}
