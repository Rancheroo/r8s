// Package cmd implements the CLI commands and flags for r8s using the Cobra framework.
// It provides the root command, version information, and configuration management.
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	namespace string // Default namespace for kubectl-style commands
	verbose   bool   // Enable verbose error output

	versionInfo struct {
		Version string
		Commit  string
		Date    string
	}
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:           "r8s [bundle-path]",
	Args:          cobra.MaximumNArgs(1), // Allow 0 or 1 positional argument (bundle path)
	Short:         "r8s - The fastest way to understand a broken Kubernetes cluster from a log bundle",
	SilenceErrors: true,                  // We handle error output ourselves
	SilenceUsage:  true,                  // Don't dump usage on every error
	Long: `r8s — kubectl for Rancher bundles. Analyze clusters offline.

USAGE:
  r8s analyze ./bundle/          # Detect issues
  r8s get pods ./bundle/         # List pods
  r8s logs ./bundle/ <pod>       # Stream logs
  r8s export ./bundle/           # Export findings

EXAMPLES:
  # Quick analysis
  r8s analyze ./bundle/

  # Filter critical issues
  r8s analyze ./bundle/ --format=json | jq '.critical'

  # kubectl-compatible commands
  r8s get pods ./bundle/ -n cattle-system
  r8s logs ./bundle/ nginx-pod

For more: https://github.com/Rancheroo/r8s`,
	RunE: runRoot,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	// Set up custom error handling
	SetUnknownCommandHandler(rootCmd)
	
	// Check for unknown commands before executing
	if len(os.Args) > 1 {
		firstArg := os.Args[1]
		// Skip flags and special commands
		if !strings.HasPrefix(firstArg, "-") && firstArg != "help" && firstArg != "--help" && firstArg != "-h" {
			// Check if it's a known subcommand
			if !isKnownCommand(firstArg) {
				// Check if it's a typo we can suggest
				if suggestion, found := CommandSuggestions[strings.ToLower(firstArg)]; found {
					ShowUnknownCommandError(firstArg)
					// Offer to run the correct command
					fmt.Fprintf(os.Stderr, "Run 'r8s %s' instead? (y/n): ", suggestion.Command)
					os.Exit(ExitError)
				} else if !isValidBundlePath(firstArg) {
					// Could be a bundle path, which is handled by runRoot
					// But if it doesn't exist as a path either, show unknown command
					if _, err := os.Stat(firstArg); os.IsNotExist(err) {
						ShowUnknownCommandError(firstArg)
						os.Exit(ExitError)
					}
				}
			}
		}
	}
	
	if err := rootCmd.Execute(); err != nil {
		// Check for exit code error
		if exitCode := GetExitCode(err); exitCode != ExitSuccess {
			os.Exit(exitCode)
		}
		// Regular error - show friendly version
		ShowFriendlyError(err)
		os.Exit(ExitError)
	}
}

// isKnownCommand checks if a command is known
func isKnownCommand(cmd string) bool {
	knownCommands := []string{
		"analyze", "analyse", "analize",
		"ask",
		"completion",
		"describe", "desc",
		"export",
		"generate", "gen",
		"get",
		"logs", "log",
		"patterns", "pattern",
		"test-cluster", "testcluster",
		"validate", "val", "check",
		"version", "ver", "v",
		"help", "h",
	}

	cmd = strings.ToLower(cmd)
	for _, known := range knownCommands {
		if cmd == known {
			return true
		}
	}
	return false
}

// isValidBundlePath checks if the argument looks like a bundle path
func isValidBundlePath(path string) bool {
	// Check if it looks like a path
	if strings.Contains(path, "/") || strings.Contains(path, ".") {
		return true
	}
	return false
}

func init() {
	// Global flags (Law 3: Simplify - removed cfgFile, contextName, scanDepth)
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "namespace filter")

	// Add version command
	rootCmd.AddCommand(versionCmd)
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
	// Call analyze directly (v0.8.0 CLI-first)
	return runAnalyze(cmd, args)
}
