// Package config handles application configuration management, including multi-profile
// support, credential handling, and configuration file persistence. It uses YAML for
// configuration storage and supports both bearer token and API key/secret authentication.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	CurrentProfile  string    `yaml:"currentProfile"`
	Profiles        []Profile `yaml:"profiles"`
	RefreshInterval string    `yaml:"refreshInterval"`
	LogLevel        string    `yaml:"logLevel"`

	// Runtime overrides (not from file)
	Insecure  bool
	Context   string
	Namespace string
	MockMode  bool // Enable demo mode with mock data
}

// Profile represents a Rancher connection profile
type Profile struct {
	Name        string `yaml:"name"`
	URL         string `yaml:"url"`
	BearerToken string `yaml:"bearerToken,omitempty"`
	AccessKey   string `yaml:"accessKey,omitempty"`
	SecretKey   string `yaml:"secretKey,omitempty"`
	Insecure    bool   `yaml:"insecure"`
}

// GetToken returns the bearer token, constructing it from access/secret keys if needed
func (p *Profile) GetToken() string {
	if p.BearerToken != "" {
		return p.BearerToken
	}
	if p.AccessKey != "" && p.SecretKey != "" {
		return p.AccessKey + ":" + p.SecretKey
	}
	return ""
}

// Load loads configuration from file or creates default
func Load(cfgFile, profileName string) (*Config, error) {
	// Determine config file path
	if cfgFile == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		cfgFile = filepath.Join(home, ".r8s", "config.yaml")
	}

	// Check if config file exists
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		return createDefaultConfig(cfgFile)
	}

	// Read config file
	data, err := os.ReadFile(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse config
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Use specified profile or current profile
	if profileName != "" {
		cfg.CurrentProfile = profileName
	}

	// Validate profile exists
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if len(c.Profiles) == 0 {
		return fmt.Errorf("no profiles configured")
	}

	// Check if current profile exists
	found := false
	for _, p := range c.Profiles {
		if p.Name == c.CurrentProfile {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("profile '%s' not found", c.CurrentProfile)
	}

	return nil
}

// GetCurrentProfile returns the current active profile
func (c *Config) GetCurrentProfile() (*Profile, error) {
	for i := range c.Profiles {
		if c.Profiles[i].Name == c.CurrentProfile {
			return &c.Profiles[i], nil
		}
	}
	return nil, fmt.Errorf("profile '%s' not found", c.CurrentProfile)
}

// GetRefreshInterval returns the refresh interval as duration
func (c *Config) GetRefreshInterval() time.Duration {
	if c.RefreshInterval == "" {
		return 5 * time.Second
	}
	d, err := time.ParseDuration(c.RefreshInterval)
	if err != nil {
		return 5 * time.Second
	}
	return d
}

// createDefaultConfig creates a default configuration file
func createDefaultConfig(cfgFile string) (*Config, error) {
	cfg := &Config{
		CurrentProfile: "default",
		Profiles: []Profile{
			{
				Name:     "default",
				URL:      "https://rancher.example.com",
				Insecure: false,
			},
		},
		RefreshInterval: "5s",
		LogLevel:        "info",
	}

	// Create config directory
	dir := filepath.Dir(cfgFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write default config
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(cfgFile, data, 0600); err != nil {
		return nil, fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Created default config file at %s\n", cfgFile)
	fmt.Println("Please edit the file to add your Rancher credentials.")
	os.Exit(0)

	return cfg, nil
}

// Save saves the configuration to file
func (c *Config) Save(cfgFile string) error {
	if cfgFile == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		cfgFile = filepath.Join(home, ".r8s", "config.yaml")
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(cfgFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
