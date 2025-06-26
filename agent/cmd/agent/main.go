package main

import (
	"context"
<<<<<<< HEAD
	"fmt"
=======
	"flag"
	"fmt"
	"log"
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/computehive/agent/core"
<<<<<<< HEAD
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	cfgFile string
	logger  *zap.Logger
)

var rootCmd = &cobra.Command{
	Use:   "computehive-agent",
	Short: "ComputeHive distributed compute agent",
	Long: `ComputeHive Agent contributes computing resources to the distributed compute network.
	
The agent automatically detects available hardware resources, registers with the control plane,
and executes assigned compute jobs in secure sandboxed environments.`,
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the compute agent",
	Long:  `Start the ComputeHive agent and begin contributing compute resources to the network.`,
	RunE:  runAgent,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("ComputeHive Agent v1.0.0")
		fmt.Println("Build Date: 2024-01-01")
		fmt.Println("Git Commit: unknown")
		fmt.Println("Go Version: go1.21")
		fmt.Println("OS/Arch: " + fmt.Sprintf("%s/%s", os.Getenv("GOOS"), os.Getenv("GOARCH")))
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.computehive/agent.yaml)")
	rootCmd.PersistentFlags().String("log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("log-format", "console", "log format (console, json)")

	// Start command flags
	startCmd.Flags().String("control-plane", "https://api.computehive.io", "Control plane URL")
	startCmd.Flags().Duration("heartbeat-interval", 30*time.Second, "Heartbeat interval")
	startCmd.Flags().Int("max-jobs", 5, "Maximum concurrent jobs")
	startCmd.Flags().Float64("max-cpu", 80.0, "Maximum CPU usage percentage")
	startCmd.Flags().Float64("max-memory", 80.0, "Maximum memory usage percentage")
	startCmd.Flags().Float64("max-disk", 90.0, "Maximum disk usage percentage")
	startCmd.Flags().Bool("enable-gpu", true, "Enable GPU compute if available")
	startCmd.Flags().Bool("enable-tls", true, "Enable TLS for control plane communication")
	startCmd.Flags().String("cert-file", "", "TLS certificate file")
	startCmd.Flags().String("key-file", "", "TLS key file")
	startCmd.Flags().String("ca-file", "", "TLS CA certificate file")

	// Bind flags to viper
	viper.BindPFlag("control_plane_url", startCmd.Flags().Lookup("control-plane"))
	viper.BindPFlag("heartbeat_interval", startCmd.Flags().Lookup("heartbeat-interval"))
	viper.BindPFlag("max_concurrent_jobs", startCmd.Flags().Lookup("max-jobs"))
	viper.BindPFlag("resource_limits.max_cpu_percent", startCmd.Flags().Lookup("max-cpu"))
	viper.BindPFlag("resource_limits.max_memory_percent", startCmd.Flags().Lookup("max-memory"))
	viper.BindPFlag("resource_limits.max_disk_percent", startCmd.Flags().Lookup("max-disk"))
	viper.BindPFlag("enable_gpu", startCmd.Flags().Lookup("enable-gpu"))
	viper.BindPFlag("security.enable_tls", startCmd.Flags().Lookup("enable-tls"))
	viper.BindPFlag("security.cert_file", startCmd.Flags().Lookup("cert-file"))
	viper.BindPFlag("security.key_file", startCmd.Flags().Lookup("key-file"))
	viper.BindPFlag("security.ca_file", startCmd.Flags().Lookup("ca-file"))

	// Add commands
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(versionCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		configDir := filepath.Join(home, ".computehive")
		os.MkdirAll(configDir, 0755)

		viper.AddConfigPath(configDir)
		viper.SetConfigName("agent")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix("COMPUTEHIVE")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("control_plane_url", "https://api.computehive.io")
	viper.SetDefault("heartbeat_interval", 30*time.Second)
	viper.SetDefault("max_concurrent_jobs", 5)
	viper.SetDefault("resource_limits.max_cpu_percent", 80.0)
	viper.SetDefault("resource_limits.max_memory_percent", 80.0)
	viper.SetDefault("resource_limits.max_disk_percent", 90.0)
	viper.SetDefault("security.enable_tls", true)
	viper.SetDefault("log_level", "info")
	viper.SetDefault("log_format", "console")

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	// Initialize logger
	initLogger()
}

func initLogger() {
	logLevel := viper.GetString("log_level")
	logFormat := viper.GetString("log_format")

	// Parse log level
	level := zapcore.InfoLevel
	switch logLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	}

	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create encoder
	var encoder zapcore.Encoder
	if logFormat == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Create core
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		level,
	)

	// Create logger
	logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

func runAgent(cmd *cobra.Command, args []string) error {
	logger.Info("Starting ComputeHive agent",
		zap.String("version", "1.0.0"),
		zap.String("control_plane", viper.GetString("control_plane_url")))

	// Create agent config
	config := &core.Config{
		ControlPlaneURL:   viper.GetString("control_plane_url"),
		HeartbeatInterval: viper.GetDuration("heartbeat_interval"),
		MaxConcurrentJobs: viper.GetInt("max_concurrent_jobs"),
		ResourceLimits: core.ResourceLimits{
			MaxCPUPercent:    viper.GetFloat64("resource_limits.max_cpu_percent"),
			MaxMemoryPercent: viper.GetFloat64("resource_limits.max_memory_percent"),
			MaxDiskPercent:   viper.GetFloat64("resource_limits.max_disk_percent"),
		},
		SecurityConfig: core.SecurityConfig{
			EnableTLS:         viper.GetBool("security.enable_tls"),
			CertFile:          viper.GetString("security.cert_file"),
			KeyFile:           viper.GetString("security.key_file"),
			CAFile:            viper.GetString("security.ca_file"),
			EnableAttestation: viper.GetBool("security.enable_attestation"),
		},
	}

	// Create agent
	agent, err := core.NewAgent(config, logger)
	if err != nil {
		logger.Fatal("Failed to create agent", zap.Error(err))
		return err
	}

	// Start agent
	if err := agent.Start(); err != nil {
		logger.Fatal("Failed to start agent", zap.Error(err))
		return err
	}

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Wait for shutdown signal
	select {
	case sig := <-sigChan:
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
	case <-ctx.Done():
		logger.Info("Context cancelled")
	}

	// Graceful shutdown
	logger.Info("Shutting down agent...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Stop agent
	if err := agent.Stop(); err != nil {
		logger.Error("Error during shutdown", zap.Error(err))
		return err
	}

	logger.Info("Agent stopped successfully")
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
} 
=======
)

func main() {
	// Parse command line flags
	var (
		controlPlaneURL = flag.String("control-plane", "https://api.computehive.io", "Control plane URL")
		token           = flag.String("token", "", "Authentication token")
		workDir         = flag.String("work-dir", getDefaultWorkDir(), "Working directory for jobs")
		maxJobs         = flag.Int("max-jobs", 5, "Maximum concurrent jobs")
		enableGPU       = flag.Bool("enable-gpu", true, "Enable GPU support")
		enableTrusted   = flag.Bool("enable-trusted", false, "Enable trusted execution (TEE)")
		logLevel        = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
		configFile      = flag.String("config", "", "Configuration file path")
		version         = flag.Bool("version", false, "Show version information")
	)
	
	flag.Parse()
	
	if *version {
		fmt.Printf("ComputeHive Agent v%s\n", core.Version)
		os.Exit(0)
	}
	
	// Create configuration
	config := &core.Config{
		ControlPlaneURL:    *controlPlaneURL,
		Token:              *token,
		HeartbeatInterval:  30 * time.Second,
		JobPollingInterval: 10 * time.Second,
		MetricsInterval:    60 * time.Second,
		MaxConcurrentJobs:  *maxJobs,
		WorkDir:            *workDir,
		EnableGPU:          *enableGPU,
		EnableTrustedExec:  *enableTrusted,
		LogLevel:           *logLevel,
	}
	
	// Load config from file if specified
	if *configFile != "" {
		if err := loadConfigFromFile(*configFile, config); err != nil {
			log.Fatalf("Failed to load config file: %v", err)
		}
	}
	
	// Override with environment variables
	loadConfigFromEnv(config)
	
	// Validate configuration
	if err := validateConfig(config); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}
	
	// Create work directory
	if err := os.MkdirAll(config.WorkDir, 0755); err != nil {
		log.Fatalf("Failed to create work directory: %v", err)
	}
	
	// Create agent
	agent, err := core.NewAgent(config)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}
	
	// Start agent
	if err := agent.Start(); err != nil {
		log.Fatalf("Failed to start agent: %v", err)
	}
	
	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	log.Println("Agent is running. Press Ctrl+C to stop.")
	<-sigChan
	
	log.Println("Shutting down agent...")
	
	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	shutdownDone := make(chan error, 1)
	go func() {
		shutdownDone <- agent.Stop()
	}()
	
	select {
	case err := <-shutdownDone:
		if err != nil {
			log.Printf("Error during shutdown: %v", err)
			os.Exit(1)
		}
		log.Println("Agent stopped successfully")
	case <-ctx.Done():
		log.Println("Shutdown timeout exceeded, forcing exit")
		os.Exit(1)
	}
}

// getDefaultWorkDir returns the default working directory
func getDefaultWorkDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/computehive"
	}
	return filepath.Join(homeDir, ".computehive", "work")
}

// loadConfigFromFile loads configuration from a JSON file
func loadConfigFromFile(path string, config *core.Config) error {
	// In a real implementation, this would parse JSON/YAML config
	// For now, we'll skip this
	return nil
}

// loadConfigFromEnv loads configuration from environment variables
func loadConfigFromEnv(config *core.Config) {
	if url := os.Getenv("COMPUTEHIVE_CONTROL_PLANE_URL"); url != "" {
		config.ControlPlaneURL = url
	}
	
	if token := os.Getenv("COMPUTEHIVE_TOKEN"); token != "" {
		config.Token = token
	}
	
	if workDir := os.Getenv("COMPUTEHIVE_WORK_DIR"); workDir != "" {
		config.WorkDir = workDir
	}
	
	if maxJobs := os.Getenv("COMPUTEHIVE_MAX_JOBS"); maxJobs != "" {
		// Parse and set max jobs
	}
}

// validateConfig validates the configuration
func validateConfig(config *core.Config) error {
	if config.ControlPlaneURL == "" {
		return fmt.Errorf("control plane URL is required")
	}
	
	if config.MaxConcurrentJobs <= 0 {
		return fmt.Errorf("max concurrent jobs must be positive")
	}
	
	return nil
}
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
