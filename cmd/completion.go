// Package cmd implements the CLI commands for r8s.
// Sprint 9 Day 1: r8s completion - Shell completion support
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion scripts for r8s.

To load completions:

Bash:
  $ source <(r8s completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ r8s completion bash > /etc/bash_completion.d/r8s
  # macOS:
  $ r8s completion bash > $(brew --prefix)/etc/bash_completion.d/r8s

Zsh:
  $ source <(r8s completion zsh)
  # To load completions for each session, execute once:
  $ r8s completion zsh > "${fpath[1]}/_r8s"

Fish:
  $ r8s completion fish | source
  # To load completions for each session, execute once:
  $ r8s completion fish > ~/.config/fish/completions/r8s.fish

PowerShell:
  $ r8s completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, run:
  $ r8s completion powershell > r8s.ps1
  # and source this file from your PowerShell profile.
`,
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	RunE:      runCompletion,
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

func runCompletion(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return NewExitError(ExitError, "shell argument required")
	}

	shell := args[0]

	switch shell {
	case "bash":
		return rootCmd.GenBashCompletion(os.Stdout)
	case "zsh":
		return rootCmd.GenZshCompletion(os.Stdout)
	case "fish":
		return rootCmd.GenFishCompletion(os.Stdout, true)
	case "powershell":
		return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
	default:
		return NewExitError(ExitError, fmt.Sprintf("invalid shell: %s (valid: bash, zsh, fish, powershell)", shell))
	}
}
