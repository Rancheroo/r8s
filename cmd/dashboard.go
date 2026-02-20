// Package cmd implements the CLI commands for r8s.
// Sprint 9 Week 2 Day 8: r8s dashboard - Minimal dashboard TUI
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Rancheroo/r8s/internal/bundle"
	"github.com/Rancheroo/r8s/internal/tui"
)

// dashboardCmd represents the dashboard command
var dashboardCmd = &cobra.Command{
	Use:   "dashboard [bundle-path]",
	Short: "Launch minimal attention dashboard TUI",
	Long: `Launch a lightweight dashboard showing bundle health and issues.

This is the only TUI command in r8s. All other functionality is CLI-based.

EXAMPLES:
  # Launch dashboard with bundle
  r8s dashboard ./extracted-bundle/

  # Launch with demo data (if no bundle provided)
  r8s dashboard

KEYBOARD SHORTCUTS:
  Enter  - View details of selected item
  q      - Quit dashboard
  ?      - Show help

EXIT CODES:
  0 - Dashboard closed normally
  2 - Error (bundle not found, etc.)`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDashboard,
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}

func runDashboard(cmd *cobra.Command, args []string) error {
	bundlePath := ""
	if len(args) > 0 {
		bundlePath = args[0]
	}

	// Validate bundle if provided
	if bundlePath != "" {
		if _, err := os.Stat(bundlePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: bundle path not found: %v\n", err)
			os.Exit(ExitError)
			return nil
		}
	}

	// Launch minimal dashboard
	app, err := tui.NewDashboard(bundlePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create dashboard: %v\n", err)
		os.Exit(ExitError)
		return nil
	}

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: dashboard error: %v\n", err)
		os.Exit(ExitError)
		return nil
	}

	return nil
}
