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
		cfg, err := config.Load(cfgFile, profile)
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
