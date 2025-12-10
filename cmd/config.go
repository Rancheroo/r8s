// Package cmd implements the CLI commands and flags for r8s.
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Rancheroo/r8s/internal/config"
)

func init() {
	// Add subcommands to config
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configViewCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configSetCmd)
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new config file",
	Long: `Initialize a new r8s configuration file with a template.

The config file will be created at ~/.r8s/config.yaml with helpful
comments and examples. You can then edit it to add your Rancher
credentials.

EXAMPLES:
  # Create config file at default location
  r8s config init

  # Create config file at custom location
  r8s config init --config=/path/to/config.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create the config file
		if err := config.InitConfig(cfgFile); err != nil {
			return err
		}

		// Get the actual path that was created
		configPath := config.GetConfigPath(cfgFile)

		fmt.Printf("✓ Created config file at %s\n\n", configPath)
		fmt.Println("Next steps:")
		fmt.Println("  1. Edit the config file and add your Rancher credentials:")
		fmt.Printf("     %s\n\n", configPath)
		fmt.Println("  2. Option A - Edit the YAML file directly:")
		fmt.Println("     - Set url: https://your-rancher-url.com")
		fmt.Println("     - Set bearerToken: token-xxxxx:yyyyyyyy")
		fmt.Println("")
		fmt.Println("  2. Option B - Use environment variables:")
		fmt.Println("     export RANCHER_URL=https://your-rancher-url.com")
		fmt.Println("     export RANCHER_TOKEN=token-xxxxx:yyyyyyyy")
		fmt.Println("")
		fmt.Println("  3. Launch r8s:")
		fmt.Println("     r8s tui")
		fmt.Println("")
		fmt.Println("  Or try demo mode without configuration:")
		fmt.Println("     r8s tui --mockdata")

		return nil
	},
}

var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "View current configuration",
	Long: `Display the current r8s configuration.

Shows the active profile, configured profiles, and settings.
Tokens are masked for security.

EXAMPLES:
  # View current configuration
  r8s config view

  # View specific profile
  r8s config view --profile=production`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get config path
		configPath := config.GetConfigPath(cfgFile)

		// Check if config exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			fmt.Printf("No config file found at %s\n", configPath)
			fmt.Println("\nRun 'r8s config init' to create one.")
			return nil
		}

		// Load config
		cfg, err := config.Load(cfgFile, "")
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Display configuration
		fmt.Printf("# r8s Configuration\n")
		fmt.Printf("# File: %s\n\n", configPath)

		fmt.Printf("Current Profile: %s\n", cfg.CurrentProfile)
		fmt.Printf("Refresh Interval: %s\n", cfg.RefreshInterval)
		fmt.Printf("Log Level: %s\n\n", cfg.LogLevel)

		fmt.Printf("Profiles (%d):\n", len(cfg.Profiles))
		for _, p := range cfg.Profiles {
			fmt.Printf("\n  %s:\n", p.Name)
			fmt.Printf("    URL: %s\n", p.URL)

			// Mask token for security
			if p.BearerToken != "" {
				fmt.Printf("    Token: ******** (configured)\n")
			} else if p.AccessKey != "" && p.SecretKey != "" {
				fmt.Printf("    Access Key: %s\n", p.AccessKey)
				fmt.Printf("    Secret Key: ******** (configured)\n")
			} else {
				fmt.Printf("    Token: (not configured)\n")
			}

			fmt.Printf("    Insecure: %v\n", p.Insecure)
		}

		return nil
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit configuration in $EDITOR",
	Long: `Open the configuration file in your default editor.

Uses the EDITOR environment variable to determine which editor to use.
Falls back to 'vi' if EDITOR is not set.

EXAMPLES:
  # Edit config with default editor
  r8s config edit

  # Edit specific config file
  r8s config edit --config=/path/to/config.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get config path
		configPath := config.GetConfigPath(cfgFile)

		// Check if config exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			fmt.Printf("No config file found at %s\n", configPath)
			fmt.Println("\nRun 'r8s config init' to create one first.")
			return nil
		}

		// Determine editor
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}

		// Split editor command (might have args like "code --wait")
		parts := strings.Fields(editor)
		editorCmd := parts[0]
		editorArgs := parts[1:]
		editorArgs = append(editorArgs, configPath)

		// Open editor
		editCmd := exec.Command(editorCmd, editorArgs...)
		editCmd.Stdin = os.Stdin
		editCmd.Stdout = os.Stdout
		editCmd.Stderr = os.Stderr

		fmt.Printf("Opening %s in %s...\n", configPath, editorCmd)
		if err := editCmd.Run(); err != nil {
			return fmt.Errorf("failed to open editor: %w", err)
		}

		fmt.Println("\n✓ Config file saved")
		return nil
	},
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration file",
	Long: `Validate the configuration file syntax and structure.

Checks for:
- Valid YAML syntax
- Required fields present
- Valid profile references
- Proper URL formatting

EXAMPLES:
  # Validate current config
  r8s config validate

  # Validate specific config file
  r8s config validate --config=/path/to/config.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get config path
		configPath := config.GetConfigPath(cfgFile)

		// Check if config exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return fmt.Errorf("no config file found at %s", configPath)
		}

		// Try to load config
		cfg, err := config.Load(cfgFile, "")
		if err != nil {
			fmt.Printf("✗ Config validation failed\n\n")
			return fmt.Errorf("validation error: %w", err)
		}

		// Validate config
		if err := cfg.Validate(); err != nil {
			fmt.Printf("✗ Config validation failed\n\n")
			return fmt.Errorf("validation error: %w", err)
		}

		// Success
		fmt.Printf("✓ Config file is valid!\n\n")
		fmt.Printf("File: %s\n", configPath)
		fmt.Printf("Current Profile: %s\n", cfg.CurrentProfile)
		fmt.Printf("Profiles: %d\n", len(cfg.Profiles))

		// Check for profiles without tokens
		missingTokens := []string{}
		for _, p := range cfg.Profiles {
			if p.GetToken() == "" {
				missingTokens = append(missingTokens, p.Name)
			}
		}

		if len(missingTokens) > 0 {
			fmt.Printf("\n⚠️  Warning: The following profiles are missing credentials:\n")
			for _, name := range missingTokens {
				fmt.Printf("  - %s\n", name)
			}
			fmt.Println("\nThese profiles won't work until credentials are added.")
		}

		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value for the current profile.

Supported keys:
  url             - Rancher server URL
  token           - Bearer token (format: token-xxxxx:yyyyyyyy)
  insecure        - Skip TLS verification (true/false)
  currentProfile  - Default profile name

EXAMPLES:
  # Set URL for current profile
  r8s config set url https://rancher.example.com

  # Set bearer token
  r8s config set token token-xxxxx:yyyyyyyy

  # Enable insecure mode (skip TLS verification)
  r8s config set insecure true

  # Change default profile
  r8s config set currentProfile production

  # Set value for specific profile
  r8s config set url https://staging.example.com --profile=staging`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		// Get config path
		configPath := config.GetConfigPath(cfgFile)

		// Check if config exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return fmt.Errorf("no config file found at %s\nRun 'r8s config init' first", configPath)
		}

		// Load config
		cfg, err := config.Load(cfgFile, "")
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Handle global settings
		if key == "currentProfile" {
			// Verify profile exists
			found := false
			for _, p := range cfg.Profiles {
				if p.Name == value {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("profile '%s' not found", value)
			}
			cfg.CurrentProfile = value
			if err := cfg.Save(cfgFile); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			fmt.Printf("✓ Set current profile to: %s\n", value)
			return nil
		}

		// Determine which profile to modify
		profileName := ""
		if profileName == "" {
			profileName = cfg.CurrentProfile
		}

		// Find the profile
		var targetProfile *config.Profile
		for i := range cfg.Profiles {
			if cfg.Profiles[i].Name == profileName {
				targetProfile = &cfg.Profiles[i]
				break
			}
		}

		if targetProfile == nil {
			return fmt.Errorf("profile '%s' not found", profileName)
		}

		// Set the value
		switch key {
		case "url":
			targetProfile.URL = value
			fmt.Printf("✓ Set URL to: %s (profile: %s)\n", value, profileName)

		case "token", "bearerToken":
			targetProfile.BearerToken = value
			// Clear access/secret keys if setting bearer token
			targetProfile.AccessKey = ""
			targetProfile.SecretKey = ""
			fmt.Printf("✓ Set bearer token (profile: %s)\n", profileName)

		case "insecure":
			if value == "true" {
				targetProfile.Insecure = true
				fmt.Printf("✓ Enabled insecure mode (profile: %s)\n", profileName)
				fmt.Println("⚠️  WARNING: TLS verification is now disabled. Use only for dev/testing!")
			} else if value == "false" {
				targetProfile.Insecure = false
				fmt.Printf("✓ Disabled insecure mode (profile: %s)\n", profileName)
			} else {
				return fmt.Errorf("invalid value for insecure: %s (use 'true' or 'false')", value)
			}

		default:
			return fmt.Errorf("unknown config key: %s\nSupported keys: url, token, insecure, currentProfile", key)
		}

		// Save config
		if err := cfg.Save(cfgFile); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		return nil
	},
}
