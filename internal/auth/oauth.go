package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/SamHL/zs/internal/config"
)

const (
	// OAuth scopes required for full Sprints access
	DefaultScopes = "ZOHOSPRINTS.teams.ALL,ZOHOSPRINTS.projects.ALL,ZOHOSPRINTS.sprints.ALL,ZOHOSPRINTS.items.ALL"

	// Local callback server settings
	CallbackPort = 8484
	CallbackPath = "/callback"
)

// TokenResponse represents the OAuth token response from Zoho
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Error        string `json:"error"`
}

// GetAuthorizationURL generates the OAuth authorization URL
func GetAuthorizationURL(clientID string) string {
	cfg := config.Get()
	accountsURL := config.GetAccountsURL()

	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("response_type", "code")
	params.Set("access_type", "offline")
	params.Set("scope", DefaultScopes)
	params.Set("redirect_uri", fmt.Sprintf("http://localhost:%d%s", CallbackPort, CallbackPath))
	params.Set("prompt", "consent")

	// Add data center hint
	if cfg.Defaults.DataCenter != "" && cfg.Defaults.DataCenter != "com" {
		params.Set("access_type", "offline")
	}

	return fmt.Sprintf("%s/oauth/v2/auth?%s", accountsURL, params.Encode())
}

// StartCallbackServer starts a local HTTP server to receive the OAuth callback
// Returns a channel that receives the authorization code
func StartCallbackServer(ctx context.Context) (chan string, chan error, error) {
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", CallbackPort))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start callback server: %w", err)
	}

	mux := http.NewServeMux()
	server := &http.Server{Handler: mux}

	mux.HandleFunc(CallbackPath, func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		errorParam := r.URL.Query().Get("error")

		if errorParam != "" {
			errChan <- fmt.Errorf("OAuth error: %s", errorParam)
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `<html><body><h1>Authentication Failed</h1><p>Error: %s</p><p>You can close this window.</p></body></html>`, errorParam)
			return
		}

		if code == "" {
			errChan <- fmt.Errorf("no authorization code received")
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<html><body><h1>Authentication Failed</h1><p>No authorization code received.</p><p>You can close this window.</p></body></html>`)
			return
		}

		codeChan <- code
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html><body><h1>Authentication Successful!</h1><p>You can close this window and return to the terminal.</p></body></html>`)
	})

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Shutdown server when context is cancelled
	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background())
	}()

	return codeChan, errChan, nil
}

// ExchangeCodeForToken exchanges an authorization code for access and refresh tokens
func ExchangeCodeForToken(code, clientID, clientSecret string) (*TokenResponse, error) {
	accountsURL := config.GetAccountsURL()
	tokenURL := fmt.Sprintf("%s/oauth/v2/token", accountsURL)

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", fmt.Sprintf("http://localhost:%d%s", CallbackPort, CallbackPath))

	resp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("token error: %s", tokenResp.Error)
	}

	return &tokenResp, nil
}

// RefreshAccessToken uses the refresh token to get a new access token
func RefreshAccessToken() error {
	cfg := config.Get()
	if cfg.Auth.RefreshToken == "" {
		return fmt.Errorf("no refresh token available, please run 'zs auth login'")
	}

	accountsURL := config.GetAccountsURL()
	tokenURL := fmt.Sprintf("%s/oauth/v2/token", accountsURL)

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", cfg.Auth.ClientID)
	data.Set("client_secret", cfg.Auth.ClientSecret)
	data.Set("refresh_token", cfg.Auth.RefreshToken)

	resp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.Error != "" {
		return fmt.Errorf("token refresh error: %s", tokenResp.Error)
	}

	// Calculate expiry time
	expiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	// Update config with new access token (refresh token stays the same)
	return config.UpdateAuth(tokenResp.AccessToken, cfg.Auth.RefreshToken, expiry)
}

// EnsureValidToken checks if the token is valid and refreshes if needed
func EnsureValidToken() error {
	if !config.IsAuthenticated() {
		return fmt.Errorf("not authenticated, please run 'zs auth login'")
	}

	if config.IsTokenExpired() {
		return RefreshAccessToken()
	}

	return nil
}

// SaveTokens saves the token response to config
func SaveTokens(tokenResp *TokenResponse) error {
	expiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	return config.UpdateAuth(tokenResp.AccessToken, tokenResp.RefreshToken, expiry)
}

// RevokeToken revokes the current access token
func RevokeToken() error {
	cfg := config.Get()
	if cfg.Auth.AccessToken == "" {
		return nil // Nothing to revoke
	}

	accountsURL := config.GetAccountsURL()
	revokeURL := fmt.Sprintf("%s/oauth/v2/token/revoke?token=%s", accountsURL, cfg.Auth.AccessToken)

	resp, err := http.Post(revokeURL, "application/x-www-form-urlencoded", nil)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	defer resp.Body.Close()

	return config.ClearAuth()
}
