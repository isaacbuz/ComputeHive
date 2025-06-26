package cmd

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
	"github.com/computehive/cli/pkg/client"
	"github.com/computehive/cli/pkg/config"
)

// NewAuthCmd creates the auth command
func NewAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		Long:  "Login, logout, and manage authentication tokens",
	}

	cmd.AddCommand(
		newAuthLoginCmd(),
		newAuthLogoutCmd(),
		newAuthStatusCmd(),
		newAuthTokenCmd(),
	)

	return cmd
}

// newAuthLoginCmd creates the login command
func newAuthLoginCmd() *cobra.Command {
	var (
		email    string
		password string
		token    string
		provider string
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to ComputeHive",
		Long: `Login to ComputeHive using email/password or authentication token.

You can login using:
1. Email and password (interactive)
2. Authentication token (--token flag)
3. OAuth provider (--provider flag)`,
		Example: `  # Interactive login
  computehive auth login

  # Login with email
  computehive auth login --email user@example.com

  # Login with token
  computehive auth login --token YOUR_API_TOKEN

  # Login with OAuth
  computehive auth login --provider github`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _ := config.Load() // Ignore error, we're logging in
			if cfg.APIURL == "" {
				cfg.APIURL = "https://api.computehive.io"
			}

			apiClient := client.New(cfg.APIURL, "")

			// Handle token login
			if token != "" {
				// Verify token
				if err := apiClient.VerifyToken(token); err != nil {
					return fmt.Errorf("invalid token: %w", err)
				}

				// Save token
				cfg.Token = token
				if err := cfg.Save(); err != nil {
					return fmt.Errorf("failed to save config: %w", err)
				}

				fmt.Println("✅ Successfully logged in with token")
				return nil
			}

			// Handle OAuth login
			if provider != "" {
				fmt.Printf("Opening browser for %s authentication...\n", provider)
				authURL, err := apiClient.GetOAuthURL(provider)
				if err != nil {
					return fmt.Errorf("failed to get OAuth URL: %w", err)
				}

				fmt.Printf("Please visit: %s\n", authURL)
				fmt.Print("Enter the authorization code: ")
				
				var code string
				fmt.Scanln(&code)

				token, err := apiClient.ExchangeOAuthCode(provider, code)
				if err != nil {
					return fmt.Errorf("OAuth authentication failed: %w", err)
				}

				cfg.Token = token
				if err := cfg.Save(); err != nil {
					return fmt.Errorf("failed to save config: %w", err)
				}

				fmt.Printf("✅ Successfully logged in with %s\n", provider)
				return nil
			}

			// Handle email/password login
			if email == "" {
				fmt.Print("Email: ")
				fmt.Scanln(&email)
			}

			if password == "" {
				fmt.Print("Password: ")
				passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
				if err != nil {
					return fmt.Errorf("failed to read password: %w", err)
				}
				password = string(passwordBytes)
				fmt.Println() // New line after password
			}

			// Login
			authToken, err := apiClient.Login(email, password)
			if err != nil {
				return fmt.Errorf("login failed: %w", err)
			}

			// Save credentials
			cfg.Token = authToken
			cfg.Email = email
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Println("✅ Successfully logged in!")
			
			// Get user info
			apiClient.SetToken(authToken)
			user, err := apiClient.GetCurrentUser()
			if err == nil {
				fmt.Printf("Welcome back, %s!\n", user.Username)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "email address")
	cmd.Flags().StringVar(&password, "password", "", "password (not recommended, use interactive mode)")
	cmd.Flags().StringVar(&token, "token", "", "authentication token")
	cmd.Flags().StringVar(&provider, "provider", "", "OAuth provider (github, google)")

	return cmd
}

// newAuthLogoutCmd creates the logout command
func newAuthLogoutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout from ComputeHive",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if cfg.Token == "" {
				fmt.Println("Not logged in")
				return nil
			}

			// Clear token
			cfg.Token = ""
			cfg.Email = ""
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Println("✅ Successfully logged out")
			return nil
		},
	}

	return cmd
}

// newAuthStatusCmd creates the status command
func newAuthStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if cfg.Token == "" {
				fmt.Println("Not logged in")
				fmt.Println("\nTo login, run:")
				fmt.Println("  computehive auth login")
				return nil
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)

			// Get current user
			user, err := apiClient.GetCurrentUser()
			if err != nil {
				fmt.Println("Logged in but token may be expired")
				fmt.Println("\nTo re-login, run:")
				fmt.Println("  computehive auth login")
				return nil
			}

			fmt.Println("Authentication Status")
			fmt.Println("===================")
			fmt.Printf("User:         %s\n", user.Username)
			fmt.Printf("Email:        %s\n", user.Email)
			fmt.Printf("User ID:      %s\n", user.ID)
			fmt.Printf("Role:         %s\n", user.Role)
			fmt.Printf("API Endpoint: %s\n", cfg.APIURL)
			
			if user.Organization != "" {
				fmt.Printf("Organization: %s\n", user.Organization)
			}

			// Check token expiry
			if user.TokenExpiresAt != nil {
				fmt.Printf("Token Expires: %s\n", user.TokenExpiresAt.Format("2006-01-02 15:04:05"))
			}

			return nil
		},
	}

	return cmd
}

// newAuthTokenCmd creates the token command
func newAuthTokenCmd() *cobra.Command {
	var (
		create bool
		name   string
		scopes []string
		list   bool
		revoke string
	)

	cmd := &cobra.Command{
		Use:   "token",
		Short: "Manage API tokens",
		Long:  "Create, list, and revoke API tokens for programmatic access",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if cfg.Token == "" {
				return fmt.Errorf("not logged in. Please login first")
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)

			// List tokens
			if list {
				tokens, err := apiClient.ListAPITokens()
				if err != nil {
					return fmt.Errorf("failed to list tokens: %w", err)
				}

				if len(tokens) == 0 {
					fmt.Println("No API tokens found")
					return nil
				}

				fmt.Println("API Tokens")
				fmt.Println("==========")
				for _, token := range tokens {
					fmt.Printf("\nName:    %s\n", token.Name)
					fmt.Printf("ID:      %s\n", token.ID)
					fmt.Printf("Created: %s\n", token.CreatedAt.Format("2006-01-02"))
					fmt.Printf("Scopes:  %s\n", strings.Join(token.Scopes, ", "))
					if token.LastUsed != nil {
						fmt.Printf("Last Used: %s\n", token.LastUsed.Format("2006-01-02"))
					}
				}
				return nil
			}

			// Revoke token
			if revoke != "" {
				fmt.Printf("Revoking token %s...\n", revoke)
				if err := apiClient.RevokeAPIToken(revoke); err != nil {
					return fmt.Errorf("failed to revoke token: %w", err)
				}
				fmt.Println("✅ Token revoked successfully")
				return nil
			}

			// Create token
			if create {
				if name == "" {
					fmt.Print("Token name: ")
					fmt.Scanln(&name)
				}

				fmt.Println("Creating API token...")
				token, err := apiClient.CreateAPIToken(client.CreateTokenRequest{
					Name:   name,
					Scopes: scopes,
				})
				if err != nil {
					return fmt.Errorf("failed to create token: %w", err)
				}

				fmt.Println("✅ API token created successfully!")
				fmt.Println("\n⚠️  Save this token - it won't be shown again:")
				fmt.Printf("\n%s\n\n", token.Token)
				fmt.Println("To use this token:")
				fmt.Printf("  export COMPUTEHIVE_TOKEN=%s\n", token.Token)
				fmt.Println("  computehive auth login --token $COMPUTEHIVE_TOKEN")
				
				return nil
			}

			// Show current token (masked)
			maskedToken := cfg.Token[:10] + "..." + cfg.Token[len(cfg.Token)-10:]
			fmt.Printf("Current token: %s\n", maskedToken)
			fmt.Println("\nTo manage API tokens, use:")
			fmt.Println("  computehive auth token --list")
			fmt.Println("  computehive auth token --create")

			return nil
		},
	}

	cmd.Flags().BoolVar(&create, "create", false, "create a new API token")
	cmd.Flags().StringVar(&name, "name", "", "token name (for --create)")
	cmd.Flags().StringSliceVar(&scopes, "scopes", []string{"read", "write"}, "token scopes (for --create)")
	cmd.Flags().BoolVar(&list, "list", false, "list all API tokens")
	cmd.Flags().StringVar(&revoke, "revoke", "", "revoke a token by ID")

	return cmd
} 