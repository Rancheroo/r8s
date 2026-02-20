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
	Long: `r8s — kubectl for Rancher bundles. Analyze clusters offline, script support workflows.

FEATURES:
  • kubectl-compatible commands (get, logs, describe, analyze)
  • Analyze bundles offline with smart pattern detection
  • Export findings for AI-assisted troubleshooting
  • Bundle-first design - works without cluster access
  • CI/CD ready with JSON output and proper exit codes

QUICKSTART:
  1. Extract your RKE2 support bundle
  2. Run: r8s analyze /path/to/extracted-bundle
  3. Pipe to jq for filtering: r8s analyze ./bundle --format=json | jq '.critical'
  4. Generate AI prompts: r8s generate prompt ./bundle

EXAMPLES:
  # Analyze an extracted bundle
  r8s analyze ./extracted-bundle-folder/

  # Get pods like kubectl
  r8s get pods ./bundle/

  # Stream logs for a pod
  r8s logs ./bundle/ nginx-pod

  # Validate bundle health
  r8s validate ./bundle/

  # Enable verbose error output
  r8s -v analyze ./bundle/`,
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
	rootCmd.PersistentFlags().IntVar(&scanDepth, "scan", 200, "number of log lines to scan for error/warning detection")

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
	// v0.8.0: CLI-first - show help if no subcommand
	if len(args) == 0 {
		return cmd.Help()
	}

	// If bundle path provided without subcommand, default to analyze
	bundlePath := args[0]
	// Store for analyze command
	tuiBundlePath = bundlePath
	
	// Call analyze directly (v0.8.0 CLI-first)
	return runAnalyze(cmd, args)
}
