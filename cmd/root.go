// Package cmd implements the CLI commands and flags for r8s using the Cobra framework.
// It provides the root command, version information, and configuration management.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	cfgFile       string
	profile       string
	insecure      bool
	contextName   string
	namespace     string
	tuiBundlePath string // Path to bundle for TUI offline mode
	verbose       bool   // Enable verbose error output

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
	Short: "r8s - Rancher Cluster Navigator & Log Analyzer",
	Long: `r8s (Rancheroos) - A TUI for browsing Rancher-managed Kubernetes clusters and analyzing log bundles.

FEATURES:
  • Interactive TUI for navigating Rancher clusters, projects, namespaces
  • View pods, deployments, services, and CRDs with live data
  • Analyze RKE2 log bundles offline (no API required)
  • Color-coded log viewing with search and filtering
  • Demo mode with mock data for testing and screenshots

CONFIGURATION:
  r8s uses a config file at ~/.r8s/config.yaml or via environment variables:
    export RANCHER_URL=https://rancher.example.com
    export RANCHER_TOKEN=token-xxxxx:yyyyyyyy

EXAMPLES:
  # Launch TUI with live Rancher connection
  r8s

  # Analyze a bundle (auto-detects and launches TUI)
  r8s ./extracted-bundle-folder/
  r8s ./support-bundle.tar.gz

  # Launch TUI with demo/mock data (no API required)
  r8s --mockdata

  # View bundle metadata only
  r8s bundle info ./extracted-folder/

  # Set up configuration
  r8s config init`,
	RunE: runRoot,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.r8s/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&profile, "profile", "", "Rancher profile to use (default is from config)")
	rootCmd.PersistentFlags().BoolVar(&insecure, "insecure", false, "skip TLS certificate verification")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose error output for debugging")
	rootCmd.PersistentFlags().StringVar(&contextName, "context", "", "cluster context to start in")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "namespace to start in")

	// Root command flags (for direct TUI launch)
	rootCmd.Flags().BoolVar(&mockData, "mockdata", false, "enable demo mode with mock data (no API required)")

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
