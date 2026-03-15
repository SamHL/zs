package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/SamHL/zs/internal/auth"
	"github.com/SamHL/zs/internal/config"
	"github.com/SamHL/zs/internal/output"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication with Zoho",
	Long:  `Authentication commands for logging in, logging out, and managing OAuth tokens.`,
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Zoho",
	Long: `Initiates the OAuth 2.0 flow to authenticate with Zoho Sprints.

This will:
1. Prompt for your OAuth client credentials (if not already configured)
2. Open your browser to authorize the application
3. Store the access and refresh tokens locally

You need to create an OAuth client in Zoho API Console first:
https://api-console.zoho.com/`,
	Example: `  # Login with browser-based OAuth flow
  zs auth login

  # Check login status
  zs auth status`,
	RunE: runAuthLogin,
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out and clear stored tokens",
	Long:  `Revokes the current access token and clears all stored authentication data.`,
	RunE:  runAuthLogout,
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	Long:  `Displays the current authentication status including token expiry information.`,
	RunE:  runAuthStatus,
}

var authRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh the access token",
	Long:  `Forces a refresh of the access token using the stored refresh token.`,
	RunE:  runAuthRefresh,
}

var authConfigCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure OAuth credentials",
	Long: `Configure your OAuth client credentials.

You need to create an OAuth client in Zoho API Console first:
https://api-console.zoho.com/

When creating the client:
- Client Type: Self Client or Server-based Applications
- Redirect URI: http://localhost:8484/callback`,
	RunE: runAuthConfig,
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authRefreshCmd)
	authCmd.AddCommand(authConfigCmd)

	// Data center flag for login
	authLoginCmd.Flags().StringP("datacenter", "r", "com", "Zoho data center (com, eu, in, au, jp, ca)")
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	cfg := config.Get()

	// Check for existing credentials
	if cfg.Auth.ClientID == "" || cfg.Auth.ClientSecret == "" {
		output.PrintInfo("No OAuth credentials configured. Let's set them up.")
		if err := promptForCredentials(); err != nil {
			return err
		}
		cfg = config.Get()
	}

	// Set data center
	dc, _ := cmd.Flags().GetString("datacenter")
	if dc != "" {
		cfg.Defaults.DataCenter = dc
		config.Save()
	}

	// Start callback server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	codeChan, errChan, err := auth.StartCallbackServer(ctx)
	if err != nil {
		return err
	}

	// Generate and display auth URL
	authURL := auth.GetAuthorizationURL(cfg.Auth.ClientID)
	fmt.Println("\nOpening browser for authentication...")
	fmt.Println("\nIf the browser doesn't open, visit this URL:")
	fmt.Println(authURL)

	// Try to open browser
	openBrowser(authURL)

	fmt.Println("\nWaiting for authorization...")

	// Wait for callback
	select {
	case code := <-codeChan:
		// Exchange code for tokens
		tokenResp, err := auth.ExchangeCodeForToken(code, cfg.Auth.ClientID, cfg.Auth.ClientSecret)
		if err != nil {
			return fmt.Errorf("failed to exchange code: %w", err)
		}

		// Save tokens
		if err := auth.SaveTokens(tokenResp); err != nil {
			return fmt.Errorf("failed to save tokens: %w", err)
		}

		output.PrintSuccess("Successfully authenticated!")
		output.PrintInfo("Access token expires in %d seconds", tokenResp.ExpiresIn)

	case err := <-errChan:
		return fmt.Errorf("authentication failed: %w", err)

	case <-ctx.Done():
		return fmt.Errorf("authentication timed out")
	}

	return nil
}

func runAuthLogout(cmd *cobra.Command, args []string) error {
	if !config.IsAuthenticated() {
		output.PrintInfo("Not currently logged in")
		return nil
	}

	if err := auth.RevokeToken(); err != nil {
		output.PrintWarning("Failed to revoke token: %v", err)
		// Still clear local tokens
	}

	if err := config.ClearAuth(); err != nil {
		return fmt.Errorf("failed to clear auth: %w", err)
	}

	output.PrintSuccess("Successfully logged out")
	return nil
}

func runAuthStatus(cmd *cobra.Command, args []string) error {
	cfg := config.Get()

	fmt.Println("Authentication Status")
	fmt.Println("---------------------")

	if cfg.Auth.ClientID != "" {
		fmt.Printf("Client ID:    %s...%s\n", cfg.Auth.ClientID[:8], cfg.Auth.ClientID[len(cfg.Auth.ClientID)-4:])
	} else {
		fmt.Println("Client ID:    Not configured")
	}

	fmt.Printf("Data Center:  %s\n", cfg.Defaults.DataCenter)

	if !config.IsAuthenticated() {
		fmt.Println("Status:       Not logged in")
		return nil
	}

	fmt.Println("Status:       Logged in")

	if !cfg.Auth.TokenExpiry.IsZero() {
		if time.Now().After(cfg.Auth.TokenExpiry) {
			fmt.Println("Token:        Expired (will auto-refresh)")
		} else {
			remaining := time.Until(cfg.Auth.TokenExpiry)
			fmt.Printf("Token:        Valid (expires in %s)\n", remaining.Round(time.Second))
		}
	}

	if cfg.Auth.RefreshToken != "" {
		fmt.Println("Refresh:      Available")
	}

	return nil
}

func runAuthRefresh(cmd *cobra.Command, args []string) error {
	if !config.IsAuthenticated() {
		return fmt.Errorf("not logged in. Please run 'zs auth login' first")
	}

	output.PrintInfo("Refreshing access token...")

	if err := auth.RefreshAccessToken(); err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	output.PrintSuccess("Access token refreshed successfully")
	return nil
}

func runAuthConfig(cmd *cobra.Command, args []string) error {
	return promptForCredentials()
}

func promptForCredentials() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\nZoho OAuth Configuration")
	fmt.Println("------------------------")
	fmt.Println("Create an OAuth client at: https://api-console.zoho.com/")
	fmt.Println("Use redirect URI: http://localhost:8484/callback")
	fmt.Println()

	fmt.Print("Client ID: ")
	clientID, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	clientID = strings.TrimSpace(clientID)

	fmt.Print("Client Secret: ")
	clientSecret, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	clientSecret = strings.TrimSpace(clientSecret)

	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("client ID and secret are required")
	}

	if err := config.SetCredentials(clientID, clientSecret); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	output.PrintSuccess("Credentials saved")
	return nil
}

func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "darwin":
		err = exec.Command("open", url).Start()
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("cmd", "/c", "start", url).Start()
	}

	if err != nil {
		// Silent fail - user can copy URL manually
	}
}
