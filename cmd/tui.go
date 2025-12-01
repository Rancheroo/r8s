// Package cmd implements the CLI commands and flags for r8s.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/Rancheroo/r8s/internal/config"
	"github.com/Rancheroo/r8s/internal/tui"
)

var (
	mockData bool
)

// tuiCmd represents the tui command
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive terminal UI",
	Long: `Launch the interactive TUI for browsing Rancher clusters or log bundles.

The TUI requires either:
  1. A valid Rancher API connection (RANCHER_URL and RANCHER_TOKEN)
  2. A log bundle via --bundle flag
  3. Demo mode via --mockdata flag

EXAMPLES:
  # Live mode - connect to Rancher API
  r8s tui

  # Demo mode - mock data for testing/screenshots
  r8s tui --mockdata

  # Bundle mode - extract first, then analyze
  tar -xzf support-bundle.tar.gz
  r8s tui --bundle=./w-guard-wg-cp-xyz/

KEYBOARD SHORTCUTS:
  Enter  - Navigate into selected resource
  Esc    - Go back to previous view
  d      - Describe selected resource (JSON)
  l      - View logs for selected pod
  /      - Search in logs
  r      - Refresh current view
  ?      - Show help
  q      - Quit`,
	RunE: runTUI,
}

// runTUI handles launching the TUI application
func runTUI(cmd *cobra.Command, args []string) error {
	// Validate bundle path if provided (before loading config or starting TUI)
	if tuiBundlePath != "" {
		info, err := os.Stat(tuiBundlePath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("bundle path not found: %s", tuiBundlePath)
			}
			return fmt.Errorf("failed to access bundle path: %w", err)
		}

		// Must be a directory (no tarball support)
		if !info.IsDir() {
			return fmt.Errorf("%s is not a directory\n\n"+
				"r8s only supports extracted bundle folders.\n\n"+
				"Extract the bundle first:\n"+
				"  tar -xzf %s\n"+
				"  r8s ./extracted-folder/",
				tuiBundlePath, filepath.Base(tuiBundlePath))
		}
	}

	// Load configuration
	cfg, err := config.Load(cfgFile, profile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override config with CLI flags
	if insecure {
		cfg.Insecure = true
	}
	if contextName != "" {
		cfg.Context = contextName
	}
	if namespace != "" {
		cfg.Namespace = namespace
	}

	// Set mock mode and verbose in config
	cfg.MockMode = mockData
	cfg.Verbose = verbose

	// Create and start TUI with bundle path if provided
	app := tui.NewApp(cfg, tuiBundlePath)
	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(tuiCmd)

	// TUI-specific flags
	tuiCmd.Flags().BoolVar(&mockData, "mockdata", false, "enable demo mode with mock data (no API required)")
	tuiCmd.Flags().StringVar(&tuiBundlePath, "bundle", "", "path to extracted log bundle folder")
}
