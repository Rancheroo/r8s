// Package cmd implements the CLI commands and flags for r8s using the Cobra framework.
// It provides the root command, version information, and configuration management.
package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/Rancheroo/r8s/internal/config"
	"github.com/Rancheroo/r8s/internal/tui"
)

var (
	cfgFile       string
	profile       string
	insecure      bool
	contextName   string
	namespace     string
	tuiBundlePath string // Path to bundle for TUI offline mode

	versionInfo struct {
		Version string
		Commit  string
		Date    string
	}
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "r8s",
	Short: "r8s - Rancher log viewer and cluster simulator",
	Long: `r8s (Rancheroos) is a terminal UI for managing Rancher-based Kubernetes clusters.
It provides log viewing, filtering, and offline cluster simulation from log bundles,
along with interactive navigation of projects, namespaces, and Rancher resources.`,
	RunE: func(cmd *cobra.Command, args []string) error {
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
	},
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
	rootCmd.PersistentFlags().StringVar(&contextName, "context", "", "cluster context to start in")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "namespace to start in")
	rootCmd.Flags().StringVar(&tuiBundlePath, "bundle", "", "path to bundle for offline mode")

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
	},
}

// SetVersionInfo sets the version information from main
func SetVersionInfo(version, commit, date string) {
	versionInfo.Version = version
	versionInfo.Commit = commit
	versionInfo.Date = date
}
