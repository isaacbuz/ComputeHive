package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/computehive/cli/pkg/config"
)

var (
	cfgFile string
	debug   bool
	output  string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "computehive",
	Short: "ComputeHive CLI - Distributed compute made simple",
	Long: `ComputeHive CLI is a command-line interface for the ComputeHive platform.

ComputeHive is a distributed compute marketplace that enables you to:
- Submit compute jobs to a global network of providers
- Offer your compute resources to earn rewards
- Manage your agents, jobs, and billing
- Access a marketplace for compute resources

Get started:
  computehive auth login              # Login to your account
  computehive agent start             # Start an agent
  computehive job submit              # Submit a compute job
  computehive marketplace offers      # Browse available resources`,
	Version: "1.0.0",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize config
		if err := initConfig(); err != nil {
			return err
		}
		
		// Set debug mode
		if debug {
			fmt.Fprintln(os.Stderr, "Debug mode enabled")
		}
		
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.computehive/config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug output")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "output format (json, yaml, table)")
	
	// Bind flags to viper
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	
	// Add commands
	rootCmd.AddCommand(
		NewAuthCmd(),
		NewAgentCmd(),
		NewJobCmd(),
		NewMarketplaceCmd(),
		NewBillingCmd(),
		NewStatusCmd(),
		NewConfigCmd(),
		newVersionCmd(),
		newCompletionCmd(),
	)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() error {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		// Search config in home directory with name ".computehive" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(home + "/.computehive")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.SetEnvPrefix("COMPUTEHIVE")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if debug {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}

	return nil
}

// newVersionCmd creates the version command
func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("ComputeHive CLI v%s\n", rootCmd.Version)
			fmt.Println("Built with Go")
			fmt.Println("https://computehive.io")
			
			// Check for updates
			cfg, err := config.Load()
			if err == nil && cfg.Token != "" {
				// In production, would check for updates via API
				fmt.Println("\nChecking for updates...")
			}
		},
	}
	
	return cmd
}

// newCompletionCmd creates the completion command
func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:
  $ source <(computehive completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ computehive completion bash > /etc/bash_completion.d/computehive
  # macOS:
  $ computehive completion bash > /usr/local/etc/bash_completion.d/computehive

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ computehive completion zsh > "${fpath[1]}/_computehive"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ computehive completion fish | source

  # To load completions for each session, execute once:
  $ computehive completion fish > ~/.config/fish/completions/computehive.fish

PowerShell:
  PS> computehive completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> computehive completion powershell > computehive.ps1
  # and source this file from your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				rootCmd.GenZshCompletion(os.Stdout)
			case "fish":
				rootCmd.GenFishCompletion(os.Stdout, true)
			case "powershell":
				rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	}
	
	return cmd
}

// SetVersionInfo sets the version information for the CLI
func SetVersionInfo(version, commit, date, builtBy string) {
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s by %s)", version, commit, date, builtBy)
} 