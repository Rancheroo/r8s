// Package cmd implements the CLI commands and flags for r8s using the Cobra framework.
// It provides the root command, version information, and configuration management.
package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Rancheroo/r8s/internal/ui"
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
	SilenceErrors: true, // We handle error output ourselves
	SilenceUsage:  true, // Don't dump usage on every error
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
	ui.SetUnknownCommandHandler(rootCmd)

	// Check for unknown commands before executing
	if len(os.Args) > 1 {
		firstArg := os.Args[1]
		// Skip flags and special commands
		if !strings.HasPrefix(firstArg, "-") && firstArg != "help" && firstArg != "--help" && firstArg != "-h" {
			// Check if it's a known subcommand
			if !isKnownCommand(firstArg) {
				// Check if it's a valid bundle path (if it exists)
				if isValidBundlePath(firstArg) {
					// It might be a bundle path, let runRoot handle it
					// But we should verify it exists before assuming it's a bundle path vs a typo
					if _, err := os.Stat(firstArg); os.IsNotExist(err) {
						// It doesn't exist, so it's likely a typo'd command
						ui.ShowUnknownCommandError(firstArg)
						os.Exit(ExitError)
					}
				} else {
					// Not a path, definitively an unknown command
					ui.ShowUnknownCommandError(firstArg)
					os.Exit(ExitError)
				}
			}
		}
	}

	if err := rootCmd.Execute(); err != nil {
		// Check for exit code error (type assertion, not value check)
		var exitErr *ExitCodeError
		if errors.As(err, &exitErr) {
			// ExitCodeError means the command already displayed the error
			os.Exit(exitErr.Code)
		}
		// Regular error - show friendly version
		ui.ShowFriendlyError(err)
		os.Exit(ExitError)
	}
}

// isKnownCommand checks if a command is known
func isKnownCommand(cmdName string) bool {
	cmdName = strings.ToLower(cmdName)
	for _, c := range rootCmd.Commands() {
		if c.Name() == cmdName || c.HasAlias(cmdName) {
			return true
		}
	}
	// Special case for help which is always available
	if cmdName == "help" {
		return true
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
	
	// Fix: #95 Add completion command explicitly to root
	rootCmd.AddCommand(completionCmd)
	
	// Add custom help command to show tip
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		// Fix: Show Long description if available (restores missing docs)
		if cmd.Long != "" {
			cmd.Printf("%s\n\n", cmd.Long)
		}

		// Default help (using standard Cobra method)
		cmd.Printf(cmd.UsageString())
		
		// Show tip at bottom of help
		fmt.Println()
		ui.ShowRandomTip()
	})
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
		
		// Show a tip on version command
		fmt.Println()
		ui.ShowRandomTip()
	},
}

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `To load completions:

Bash:

  $ source <(r8s completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ r8s completion bash > /etc/bash_completion.d/r8s
  # macOS:
  $ r8s completion bash > /usr/local/etc/bash_completion.d/r8s

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ r8s completion zsh > "${fpath[1]}/_r8s"

  # You will need to start a new shell for this setup to take effect.

Fish:

  $ r8s completion fish | source

  # To load completions for each session, execute once:
  $ r8s completion fish > ~/.config/fish/completions/r8s.fish

PowerShell:

  PS> r8s completion powershell | Out-String | Invoke-Expression

  # To load completions for each session, execute once:
  PS> r8s completion powershell > r8s.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			return cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			return cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
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
	// (Even though we add them explicitly, custom Execute logic can still hide them)
	if args[0] == "completion" || args[0] == "help" {
		// Handled by Cobra's traversal since we added the commands
		return nil
	}

	// If bundle path provided without subcommand, default to analyze
	// Call analyze directly (v0.8.0 CLI-first)
	return runAnalyze(cmd, args)
}
