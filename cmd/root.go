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
	Use:   "r8s",
	Short: "r8s - Rancher Cluster Navigator & Log Analyzer",
	Long: `r8s (Rancheroos) - A TUI for browsing Rancher-managed Kubernetes clusters and analyzing log bundles.

FEATURES:
  • Interactive TUI for navigating Rancher clusters, projects, namespaces
  • View pods, deployments, services, and CRDs with live data
  • Analyze RKE2 log bundles offline (no API required)
  • Color-coded log viewing with search and filtering
  • Demo mode with mock data for testing and screenshots

CONFIGURATION:
  r8s uses a config file at ~/.config/r8s/config.yaml or via environment variables:
    export RANCHER_URL=https://rancher.example.com
    export RANCHER_TOKEN=token-xxxxx:yyyyyyyy

EXAMPLES:
  # Launch TUI with live Rancher connection
  r8s tui

  # Launch TUI with demo/mock data (no API required)
  r8s tui --mockdata

  # Analyze a log bundle offline
  r8s bundle import --path=w-guard-wg-cp-svtk6-lqtxw.tar.gz

  # Show bundle summary without launching TUI
  r8s bundle info --path=logs.tar.gz

  # Set up configuration
  r8s config init`,
	// No RunE - shows help by default when run without subcommands
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
