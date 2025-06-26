package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the CLI configuration
type Config struct {
	APIURL         string            `json:"api_url"`
	Token          string            `json:"token"`
	Email          string            `json:"email"`
	DefaultRegion  string            `json:"default_region"`
	DefaultProject string            `json:"default_project"`
	OutputFormat   string            `json:"output_format"`
	ColorOutput    bool              `json:"color_output"`
	Debug          bool              `json:"debug"`
	ProxyURL       string            `json:"proxy_url,omitempty"`
	Profiles       map[string]Profile `json:"profiles,omitempty"`
	ActiveProfile  string            `json:"active_profile,omitempty"`
}

// Profile represents a named configuration profile
type Profile struct {
	Name          string `json:"name"`
	APIURL        string `json:"api_url"`
	Token         string `json:"token"`
	DefaultRegion string `json:"default_region"`
}

// Default returns a config with default values
func Default() *Config {
	return &Config{
		APIURL:        "https://api.computehive.io",
		DefaultRegion: "us-east-1",
		OutputFormat:  "table",
		ColorOutput:   true,
		Debug:         false,
		Profiles:      make(map[string]Profile),
	}
}

// Load loads the configuration from disk
func Load() (*Config, error) {
	configPath := GetConfigPath()
	
	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}
	
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		cfg := Default()
		if err := cfg.Save(); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
		return cfg, nil
	}
	
	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	// Parse config
	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	
	// Apply defaults for missing values
	if cfg.APIURL == "" {
		cfg.APIURL = "https://api.computehive.io"
	}
	if cfg.DefaultRegion == "" {
		cfg.DefaultRegion = "us-east-1"
	}
	if cfg.OutputFormat == "" {
		cfg.OutputFormat = "table"
	}
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}
	
	// Check environment variables
	if token := os.Getenv("COMPUTEHIVE_TOKEN"); token != "" {
		cfg.Token = token
	}
	if apiURL := os.Getenv("COMPUTEHIVE_API_URL"); apiURL != "" {
		cfg.APIURL = apiURL
	}
	
	return cfg, nil
}

// Save saves the configuration to disk
func (c *Config) Save() error {
	configPath := GetConfigPath()
	
	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	// Marshal config
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	// Write config file
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// GetConfigPath returns the path to the config file
func GetConfigPath() string {
	// Check environment variable first
	if configPath := os.Getenv("COMPUTEHIVE_CONFIG"); configPath != "" {
		return configPath
	}
	
	// Default to ~/.computehive/config.json
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory
		return ".computehive/config.json"
	}
	
	return filepath.Join(home, ".computehive", "config.json")
}

// GetProfile returns a specific profile
func (c *Config) GetProfile(name string) (*Profile, error) {
	if name == "" {
		// Return current config as profile
		return &Profile{
			Name:          "default",
			APIURL:        c.APIURL,
			Token:         c.Token,
			DefaultRegion: c.DefaultRegion,
		}, nil
	}
	
	profile, ok := c.Profiles[name]
	if !ok {
		return nil, fmt.Errorf("profile '%s' not found", name)
	}
	
	return &profile, nil
}

// SetProfile sets the active profile
func (c *Config) SetProfile(name string) error {
	if name == "" || name == "default" {
		c.ActiveProfile = ""
		return nil
	}
	
	if _, ok := c.Profiles[name]; !ok {
		return fmt.Errorf("profile '%s' not found", name)
	}
	
	c.ActiveProfile = name
	
	// Apply profile settings
	profile := c.Profiles[name]
	if profile.APIURL != "" {
		c.APIURL = profile.APIURL
	}
	if profile.Token != "" {
		c.Token = profile.Token
	}
	if profile.DefaultRegion != "" {
		c.DefaultRegion = profile.DefaultRegion
	}
	
	return nil
}

// CreateProfile creates a new profile
func (c *Config) CreateProfile(name string, profile Profile) error {
	if name == "" || name == "default" {
		return fmt.Errorf("invalid profile name")
	}
	
	if c.Profiles == nil {
		c.Profiles = make(map[string]Profile)
	}
	
	profile.Name = name
	c.Profiles[name] = profile
	
	return nil
}

// DeleteProfile deletes a profile
func (c *Config) DeleteProfile(name string) error {
	if name == "" || name == "default" {
		return fmt.Errorf("cannot delete default profile")
	}
	
	if _, ok := c.Profiles[name]; !ok {
		return fmt.Errorf("profile '%s' not found", name)
	}
	
	delete(c.Profiles, name)
	
	// If this was the active profile, clear it
	if c.ActiveProfile == name {
		c.ActiveProfile = ""
	}
	
	return nil
}

// ListProfiles returns a list of profile names
func (c *Config) ListProfiles() []string {
	profiles := []string{"default"}
	
	for name := range c.Profiles {
		profiles = append(profiles, name)
	}
	
	return profiles
} 