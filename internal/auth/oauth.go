package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jaknapp/applink/internal/config"
	"github.com/jaknapp/applink/internal/storage"
)

// Token is an alias for storage.Token for convenience
type Token = storage.Token

// DoOAuthFlow performs the OAuth 2.0 authorization code flow
func DoOAuthFlow(service *config.Service, creds config.ClientCredentials, port int) (*Token, error) {
	// Generate state parameter for CSRF protection
	state, err := generateState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	// Build authorization URL
	authURL, err := buildAuthURL(service, creds.ClientID, port, state)
	if err != nil {
		return nil, fmt.Errorf("failed to build auth URL: %w", err)
	}

	// Start callback server
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)
	server := startCallbackServer(port, state, codeChan, errChan)

	// Open browser
	fmt.Printf("Opening browser for authentication...\n")
	fmt.Printf("If the browser doesn't open, visit:\n%s\n\n", authURL)
	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Failed to open browser: %v\n", err)
	}

	// Wait for callback
	var code string
	select {
	case code = <-codeChan:
		// Success
	case err := <-errChan:
		shutdownServer(server)
		return nil, err
	case <-time.After(5 * time.Minute):
		shutdownServer(server)
		return nil, fmt.Errorf("authentication timed out")
	}

	shutdownServer(server)

	// Exchange code for token
	token, err := exchangeCode(service, creds, code, port)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	return token, nil
}

func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func buildAuthURL(service *config.Service, clientID string, port int, state string) (string, error) {
	u, err := url.Parse(service.AuthURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("client_id", clientID)
	q.Set("redirect_uri", fmt.Sprintf("http://localhost:%d/callback", port))
	q.Set("response_type", "code")
	q.Set("state", state)

	if len(service.Scopes) > 0 {
		// Slack uses user_scope for user tokens, not scope
		if service.ID == "slack" {
			q.Set("user_scope", strings.Join(service.Scopes, ","))
		} else {
			q.Set("scope", strings.Join(service.Scopes, " "))
		}
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}

func exchangeCode(service *config.Service, creds config.ClientCredentials, code string, port int) (*Token, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", fmt.Sprintf("http://localhost:%d/callback", port))

	req, err := http.NewRequest("POST", service.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Different services have different auth requirements
	switch service.ID {
	case "notion":
		// Notion uses Basic auth for token exchange
		auth := base64.StdEncoding.EncodeToString([]byte(creds.ClientID + ":" + creds.ClientSecret))
		req.Header.Set("Authorization", "Basic "+auth)
	case "slack":
		// Slack wants credentials in the body
		data.Set("client_id", creds.ClientID)
		data.Set("client_secret", creds.ClientSecret)
		req.Body = io.NopCloser(strings.NewReader(data.Encode()))
	default:
		// Standard OAuth2: credentials in body
		data.Set("client_id", creds.ClientID)
		data.Set("client_secret", creds.ClientSecret)
		req.Body = io.NopCloser(strings.NewReader(data.Encode()))
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	// Parse response based on service
	return parseTokenResponse(service, body)
}

func parseTokenResponse(service *config.Service, body []byte) (*Token, error) {
	var rawResp map[string]interface{}
	if err := json.Unmarshal(body, &rawResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	token := &Token{}

	switch service.ID {
	case "slack":
		// Slack has a different response structure for user tokens
		if ok, exists := rawResp["ok"].(bool); exists && !ok {
			errMsg, _ := rawResp["error"].(string)
			return nil, fmt.Errorf("slack auth failed: %s", errMsg)
		}
		if authedUser, ok := rawResp["authed_user"].(map[string]interface{}); ok {
			token.AccessToken, _ = authedUser["access_token"].(string)
			token.Scope, _ = authedUser["scope"].(string)
		}
		if team, ok := rawResp["team"].(map[string]interface{}); ok {
			token.TeamID, _ = team["id"].(string)
		}
	default:
		// Standard OAuth2 response
		token.AccessToken, _ = rawResp["access_token"].(string)
		token.RefreshToken, _ = rawResp["refresh_token"].(string)
		token.TokenType, _ = rawResp["token_type"].(string)
		token.Scope, _ = rawResp["scope"].(string)

		// Handle expires_in
		if expiresIn, ok := rawResp["expires_in"].(float64); ok && expiresIn > 0 {
			token.ExpiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)
		}
	}

	if token.AccessToken == "" {
		return nil, fmt.Errorf("no access token in response: %s", string(body))
	}

	return token, nil
}

func shutdownServer(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}
