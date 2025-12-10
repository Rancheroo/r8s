// Package cmd implements the CLI commands and flags for r8s using the Cobra framework.
// It provides the root command, version information, and configuration management.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	cfgFile       string
	contextName   string
	namespace     string
	tuiBundlePath string // Path to bundle for TUI offline mode
	verbose       bool   // Enable verbose error output
	scanDepth     int    // Number of log lines to scan for error/warning detection (default: 200)

	versionInfo struct {
		Version string
		Commit  string
		Date    string
	}
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "r8s [bundle-path]",
	Args:  cobra.MaximumNArgs(1), // Allow 0 or 1 positional argument (bundle path)
	Short: "r8s - The fastest way to understand a broken Kubernetes cluster from a log bundle",
	Long: `r8s — the fastest way to understand a broken Kubernetes cluster from a log bundle.

FEATURES:
  • Attention Dashboard - Instantly see all cluster issues ranked by severity
  • Interactive TUI for browsing pods, deployments, services, and CRDs
  • Color-coded log viewing with search and filtering (errors/warnings highlighted)
  • Smart log analysis - detects crashes, OOM kills, connection failures
  • Bundle-first design - works offline, no API required

QUICKSTART:
  1. Extract your RKE2 support bundle
  2. Run: r8s /path/to/extracted-bundle
  3. Navigate the Attention Dashboard to find issues
  4. Press Enter on any issue to view pod logs

EXAMPLES:
  # Analyze an extracted bundle (instant dashboard)
  r8s ./extracted-bundle-folder/

  # Launch with embedded demo bundle
  r8s

  # Enable verbose error output for debugging
  r8s -v ./bundle/`,
	RunE: runRoot,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.r8s/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose error output for debugging")
	rootCmd.PersistentFlags().StringVar(&contextName, "context", "", "cluster context to start in")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "namespace to start in")

	// Add version command
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(configCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("r8s %s (commit: %s, built: %s)\n",
			versionInfo.Version,
			versionInfo.Commit,
			versionInfo.Date,
		)
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage r8s configuration",
	Long:  "Initialize, view, or edit r8s configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Config management commands:")
		fmt.Println("  init   - Initialize a new config file")
		fmt.Println("  view   - View current configuration")
		fmt.Println("  edit   - Edit configuration in $EDITOR")
		fmt.Println("")
		fmt.Println("Run 'r8s config <command> --help' for more information")
	},
}

// SetVersionInfo sets the version information from main
func SetVersionInfo(version, commit, date string) {
	versionInfo.Version = version
	versionInfo.Commit = commit
	versionInfo.Date = date
}

// runRoot handles execution of the root command with optional bundle path argument
func runRoot(cmd *cobra.Command, args []string) error {
	// Check if a positional argument was provided (bundle path)
	if len(args) > 0 {
		bundlePath := args[0]
		// Auto-detect bundle and launch TUI directly
		tuiBundlePath = bundlePath
	}

	// Delegate to TUI command
	return runTUI(cmd, args)
}
