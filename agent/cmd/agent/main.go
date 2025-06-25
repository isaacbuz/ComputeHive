package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/computehive/agent/core"
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
