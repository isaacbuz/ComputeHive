package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/computehive/cli/pkg/config"
)

// NewConfigCmd creates the config command
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
		Long:  "View and modify ComputeHive CLI configuration settings",
	}

	cmd.AddCommand(
		newConfigShowCmd(),
		newConfigSetCmd(),
		newConfigGetCmd(),
		newConfigListCmd(),
		newConfigResetCmd(),
	)

	return cmd
}

// newConfigShowCmd creates the show command
func newConfigShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			fmt.Println("ComputeHive CLI Configuration")
			fmt.Println("============================")
			fmt.Printf("Config File:       %s\n", config.GetConfigPath())
			fmt.Printf("API URL:           %s\n", cfg.APIURL)
			fmt.Printf("Default Region:    %s\n", cfg.DefaultRegion)
			fmt.Printf("Output Format:     %s\n", cfg.OutputFormat)
			fmt.Printf("Color Output:      %v\n", cfg.ColorOutput)
			fmt.Printf("Debug Mode:        %v\n", cfg.Debug)
			
			if cfg.Email != "" {
				fmt.Printf("Email:             %s\n", cfg.Email)
			}
			
			if cfg.Token != "" {
				maskedToken := cfg.Token[:10] + "..." + cfg.Token[len(cfg.Token)-10:]
				fmt.Printf("Token:             %s\n", maskedToken)
			} else {
				fmt.Printf("Token:             (not set)\n")
			}

			if cfg.DefaultProject != "" {
				fmt.Printf("Default Project:   %s\n", cfg.DefaultProject)
			}

			if cfg.ProxyURL != "" {
				fmt.Printf("Proxy URL:         %s\n", cfg.ProxyURL)
			}

			fmt.Printf("\nProfiles:          ")
			if len(cfg.Profiles) == 0 {
				fmt.Println("(none)")
			} else {
				fmt.Println(strings.Join(getProfileNames(cfg.Profiles), ", "))
			}

			if cfg.ActiveProfile != "" {
				fmt.Printf("Active Profile:    %s\n", cfg.ActiveProfile)
			}

			return nil
		},
	}

	return cmd
}

// newConfigSetCmd creates the set command
func newConfigSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set KEY VALUE",
		Short: "Set a configuration value",
		Long: `Set a configuration value.

Available keys:
  api-url          - API endpoint URL
  default-region   - Default region for operations
  output-format    - Output format (json, yaml, table)
  color            - Enable/disable colored output (true/false)
  debug            - Enable/disable debug mode (true/false)
  default-project  - Default project ID
  proxy-url        - HTTP proxy URL`,
		Example: `  # Set API URL
  computehive config set api-url https://api.computehive.io

  # Set default region
  computehive config set default-region us-west-2

  # Set output format
  computehive config set output-format json

  # Enable debug mode
  computehive config set debug true`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			switch key {
			case "api-url":
				cfg.APIURL = value
			case "default-region":
				cfg.DefaultRegion = value
			case "output-format":
				if value != "json" && value != "yaml" && value != "table" {
					return fmt.Errorf("invalid output format: %s (must be json, yaml, or table)", value)
				}
				cfg.OutputFormat = value
			case "color":
				cfg.ColorOutput = value == "true"
			case "debug":
				cfg.Debug = value == "true"
			case "default-project":
				cfg.DefaultProject = value
			case "proxy-url":
				cfg.ProxyURL = value
			default:
				return fmt.Errorf("unknown configuration key: %s", key)
			}

			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("✅ Set %s = %s\n", key, value)
			return nil
		},
	}

	return cmd
}

// newConfigGetCmd creates the get command
func newConfigGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get KEY",
		Short: "Get a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			var value string
			switch key {
			case "api-url":
				value = cfg.APIURL
			case "default-region":
				value = cfg.DefaultRegion
			case "output-format":
				value = cfg.OutputFormat
			case "color":
				value = fmt.Sprintf("%v", cfg.ColorOutput)
			case "debug":
				value = fmt.Sprintf("%v", cfg.Debug)
			case "default-project":
				value = cfg.DefaultProject
			case "proxy-url":
				value = cfg.ProxyURL
			case "email":
				value = cfg.Email
			case "token":
				if cfg.Token != "" {
					value = cfg.Token[:10] + "..." + cfg.Token[len(cfg.Token)-10:]
				}
			default:
				return fmt.Errorf("unknown configuration key: %s", key)
			}

			fmt.Println(value)
			return nil
		},
	}

	return cmd
}

// newConfigListCmd creates the list command
func newConfigListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all configuration keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Available configuration keys:")
			fmt.Println()
			fmt.Println("  api-url          - API endpoint URL")
			fmt.Println("  default-region   - Default region for operations")
			fmt.Println("  output-format    - Output format (json, yaml, table)")
			fmt.Println("  color            - Enable/disable colored output (true/false)")
			fmt.Println("  debug            - Enable/disable debug mode (true/false)")
			fmt.Println("  default-project  - Default project ID")
			fmt.Println("  proxy-url        - HTTP proxy URL")
			fmt.Println("  email            - User email (read-only)")
			fmt.Println("  token            - Authentication token (read-only)")
			fmt.Println()
			fmt.Println("Use 'computehive config set KEY VALUE' to modify a value")
			return nil
		},
	}

	return cmd
}

// newConfigResetCmd creates the reset command
func newConfigResetCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset configuration to defaults",
		Long:  "Reset all configuration values to their defaults. This will preserve authentication tokens unless --force is used.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				fmt.Println("This will reset your configuration to defaults (preserving authentication).")
				fmt.Print("Continue? (y/N): ")
				
				var response string
				fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					fmt.Println("Cancelled")
					return nil
				}
			}

			cfg, err := config.Load()
			if err != nil {
				// If we can't load config, create a new one
				cfg = &config.Config{}
			}

			// Preserve auth unless forced
			var token, email string
			if !force {
				token = cfg.Token
				email = cfg.Email
			}

			// Reset to defaults
			*cfg = config.Config{
				APIURL:        "https://api.computehive.io",
				DefaultRegion: "us-east-1",
				OutputFormat:  "table",
				ColorOutput:   true,
				Debug:         false,
				Token:         token,
				Email:         email,
			}

			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Println("✅ Configuration reset to defaults")
			if !force && token != "" {
				fmt.Println("   (authentication preserved)")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "reset everything including authentication")

	return cmd
}

// Helper function to get profile names
func getProfileNames(profiles map[string]config.Profile) []string {
	names := make([]string, 0, len(profiles))
	for name := range profiles {
		names = append(names, name)
	}
	return names
} 