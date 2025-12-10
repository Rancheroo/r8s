// Package cmd implements the CLI commands and flags for r8s.
package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/Rancheroo/r8s/internal/config"
	"github.com/Rancheroo/r8s/internal/tui"
)

// tuiCmd represents the tui command
var tuiCmd = &cobra.Command{
	Use:   "tui [bundle-path]",
	Short: "Launch interactive TUI for log bundle analysis",
	Long: `Launch the interactive TUI for analyzing RKE2 log bundles.

QUICKSTART:
  1. Extract your RKE2 support bundle: tar -xzf support-bundle.tar.gz
  2. Launch r8s: r8s ./extracted-bundle/
  3. Use Attention Dashboard to find issues

EXAMPLES:
  # Analyze extracted bundle
  r8s tui ./w-guard-wg-cp-xyz/

  # Launch with embedded demo bundle
  r8s tui

KEYBOARD SHORTCUTS:
  Enter  - Navigate into selected item / view logs
  Esc/b  - Go back to previous view
  d      - Describe selected resource (JSON)
  l      - View logs for selected pod
  /      - Search in logs
  g/G    - Jump to first/last log line
  w      - Toggle word wrap
  Ctrl+E - Filter to errors only
  Ctrl+W - Filter to warnings
  r      - Refresh current view
  ?      - Show help
  q      - Quit`,
	RunE: runTUI,
}

// runTUI handles launching the TUI application
func runTUI(cmd *cobra.Command, args []string) error {
	// Load configuration (simplified for bundle-only mode)
	cfg, err := config.Load(cfgFile, "")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override config with CLI flags
	if contextName != "" {
		cfg.Context = contextName
	}
	if namespace != "" {
		cfg.Namespace = namespace
	}
	cfg.Verbose = verbose

	// Set scan depth (default 200, tunable via --scan flag)
	if scanDepth <= 0 {
		scanDepth = 200 // Ensure positive value
	}
	cfg.ScanDepth = scanDepth

	// Create and start TUI with bundle path
	app := tui.NewApp(cfg, tuiBundlePath)

	// Check if app initialization failed - print error and exit cleanly
	if app.HasError() {
		return fmt.Errorf(app.GetError())
	}

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
	tuiCmd.Flags().StringVar(&tuiBundlePath, "bundle", "", "path to extracted log bundle folder")
	// Note: --scan flag is now a global flag on rootCmd.PersistentFlags
}
