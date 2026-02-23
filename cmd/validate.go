// Package cmd implements the CLI commands for r8s.
// Sprint 8: r8s validate bundle - Headless bundle health checking
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/Rancheroo/r8s/internal/bundle"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate [bundle-path]",
	Short: "Validate bundle health and completeness",
	Long: `Check if a support bundle is complete and usable.

This command analyzes the bundle structure and reports:
  • Completeness percentage (what's present vs expected)
  • Missing files with impact scoring
  • Bundle type detection (RKE2, K3s, etc.)
  • Critical vs optional data categories

EXAMPLES:
  # Check bundle health with table output (default)
  r8s validate ./extracted-bundle/

  # Check with JSON output for CI pipelines
  r8s validate ./bundle/ --format=json

  # Get short summary only
  r8s validate ./bundle/ --summary

EXIT CODES:
  0 - Bundle is valid (all critical files present)
  1 - Bundle is incomplete but usable (missing non-critical files)
  2 - Invalid bundle or path (missing critical files or not a bundle)`,
	Args: cobra.ExactArgs(1),
	RunE: runValidate,
}

var (
	validateFormat  string // Output format: table, json, summary
	validateSummary bool   // Show only summary line
)

func init() {
	rootCmd.AddCommand(validateCmd)

	validateCmd.Flags().StringVarP(&validateFormat, "format", "f", "table", FormatHelp())
	validateCmd.Flags().BoolVar(&validateSummary, "summary", false, "Show only summary line")
}

// runValidate executes the validate command
func runValidate(cmd *cobra.Command, args []string) error {
	// Validate args length to prevent panic
	if len(args) == 0 {
		return NewExitError(ExitError, "bundle path argument required")
	}

	bundlePath := args[0]

	// Perform health check
	health, err := bundle.CheckHealth(bundlePath)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), fmt.Sprintf("Error: %v", err))
		return NewExitError(ExitError, err.Error())
	}

	// Standardize format
	format := StandardizeFormat(validateFormat)

	// Output based on format
	out := cmd.OutOrStdout()
	switch format {
	case "json":
		outputValidateJSON(out, health)
	case "summary":
		fmt.Fprintln(out, health.Summary())
	default:
		outputValidateTable(out, health)
	}

	// Return appropriate exit code via error
	if !health.IsValid {
		return NewExitError(ExitError, "bundle invalid") // 2 - Bundle invalid
	} else if health.Completeness < 100 {
		return NewExitError(ExitIssuesFound, "bundle incomplete") // 1 - Bundle incomplete
	}
	return nil // 0 - Bundle valid
}

// outputValidateTable prints health check results in table format
func outputValidateTable(w io.Writer, health *bundle.HealthCheck) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, color.New(color.Bold).Sprint("R8S Bundle Health Check"))
	fmt.Fprintln(w, strings.Repeat("=", 60))
	fmt.Fprintln(w)

	// Summary
	fmt.Fprintln(w, health.Summary())
	fmt.Fprintf(w, "Bundle Type: %s\n", health.BundleType)
	fmt.Fprintf(w, "Files: %d/%d present (%.0f%%)\n", health.FoundFiles, health.TotalFiles, health.Completeness)
	fmt.Fprintln(w)

	// Missing files by importance
	if len(health.MissingFiles) > 0 {
		fmt.Fprintln(w, color.New(color.Bold).Sprint("Missing Files:"))
		fmt.Fprintln(w)

		// Group by importance
		critical := []bundle.MissingFile{}
		high := []bundle.MissingFile{}
		medium := []bundle.MissingFile{}
		low := []bundle.MissingFile{}

		for _, m := range health.MissingFiles {
			switch m.Importance {
			case bundle.ImportanceCritical:
				critical = append(critical, m)
			case bundle.ImportanceHigh:
				high = append(high, m)
			case bundle.ImportanceMedium:
				medium = append(medium, m)
			case bundle.ImportanceLow:
				low = append(low, m)
			}
		}

		// Print critical
		if len(critical) > 0 {
			fmt.Fprintln(w, color.RedString("🔴 Critical (Bundle unusable without these):"))
			for _, m := range critical {
				fmt.Fprintf(w, "  • %s\n", m.Path)
				fmt.Fprintf(w, "    %s\n", m.Impact)
			}
			fmt.Fprintln(w)
		}

		// Print high
		if len(high) > 0 {
			fmt.Fprintln(w, color.YellowString("⚠️  High Impact (Major features affected):"))
			for _, m := range high {
				fmt.Fprintf(w, "  • %s\n", m.Path)
				fmt.Fprintf(w, "    %s\n", m.Impact)
			}
			fmt.Fprintln(w)
		}

		// Print medium
		if len(medium) > 0 {
			fmt.Fprintln(w, color.CyanString("ℹ️  Medium Impact:"))
			for _, m := range medium {
				fmt.Fprintf(w, "  • %s\n", m.Path)
			}
			fmt.Fprintln(w)
		}

		// Print low (condensed)
		if len(low) > 0 {
			fmt.Fprintln(w, "📝 Low Impact (Optional data):")
			for _, m := range low {
				fmt.Fprintf(w, "  • %s\n", m.Path)
			}
			fmt.Fprintln(w)
		}
	} else {
		fmt.Fprintln(w, color.GreenString("✓ All expected files present"))
		fmt.Fprintln(w)
	}

	// Category summary
	if len(health.Categories) > 0 {
		fmt.Fprintln(w, color.New(color.Bold).Sprint("By Category:"))
		for cat, stats := range health.Categories {
			status := "✓"
			if stats.Missing > 0 {
				if stats.Found == 0 {
					status = "✗"
				} else {
					status = "~"
				}
			}
			fmt.Fprintf(w, "  %s %s: %d/%d files\n", status, cat, stats.Found, stats.Total)
		}
		fmt.Fprintln(w)
	}

	fmt.Fprintln(w, strings.Repeat("-", 60))
}

// outputValidateJSON prints health check results as JSON
func outputValidateJSON(w io.Writer, health *bundle.HealthCheck) {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(health)
}
