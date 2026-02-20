// Package cmd implements the CLI commands for r8s.
// Sprint 8: r8s validate bundle - Headless bundle health checking
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
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
	bundlePath := args[0]

	// Perform health check
	health, err := bundle.CheckHealth(bundlePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
		return err
	}

	// Standardize format
	format := StandardizeFormat(validateFormat)

	// Output based on format
	switch format {
	case "json":
		outputValidateJSON(health)
	case "summary":
		fmt.Println(health.Summary())
	default:
		outputValidateTable(health)
	}

	// Exit with standardized codes
	if !health.IsValid {
		os.Exit(ExitError)         // 2 - Bundle invalid (critical files missing)
	} else if health.Completeness < 100 {
		os.Exit(ExitIssuesFound)   // 1 - Bundle incomplete but usable
	}
	os.Exit(ExitSuccess)           // 0 - Bundle valid and complete
	return nil
}

// outputValidateTable prints health check results in table format
func outputValidateTable(health *bundle.HealthCheck) {
	fmt.Println()
	fmt.Println(color.New(color.Bold).Sprint("R8S Bundle Health Check"))
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	// Summary
	fmt.Println(health.Summary())
	fmt.Printf("Bundle Type: %s\n", health.BundleType)
	fmt.Printf("Files: %d/%d present (%.0f%%)\n", health.FoundFiles, health.TotalFiles, health.Completeness)
	fmt.Println()

	// Missing files by importance
	if len(health.MissingFiles) > 0 {
		fmt.Println(color.New(color.Bold).Sprint("Missing Files:"))
		fmt.Println()

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
			fmt.Println(color.RedString("🔴 Critical (Bundle unusable without these):"))
			for _, m := range critical {
				fmt.Printf("  • %s\n", m.Path)
				fmt.Printf("    %s\n", m.Impact)
			}
			fmt.Println()
		}

		// Print high
		if len(high) > 0 {
			fmt.Println(color.YellowString("⚠️  High Impact (Major features affected):"))
			for _, m := range high {
				fmt.Printf("  • %s\n", m.Path)
				fmt.Printf("    %s\n", m.Impact)
			}
			fmt.Println()
		}

		// Print medium
		if len(medium) > 0 {
			fmt.Println(color.CyanString("ℹ️  Medium Impact:"))
			for _, m := range medium {
				fmt.Printf("  • %s\n", m.Path)
			}
			fmt.Println()
		}

		// Print low (condensed)
		if len(low) > 0 {
			fmt.Println("📝 Low Impact (Optional data):")
			for _, m := range low {
				fmt.Printf("  • %s\n", m.Path)
			}
			fmt.Println()
		}
	} else {
		fmt.Println(color.GreenString("✓ All expected files present"))
		fmt.Println()
	}

	// Category summary
	if len(health.Categories) > 0 {
		fmt.Println(color.New(color.Bold).Sprint("By Category:"))
		for cat, stats := range health.Categories {
			status := "✓"
			if stats.Missing > 0 {
				if stats.Found == 0 {
					status = "✗"
				} else {
					status = "~"
				}
			}
			fmt.Printf("  %s %s: %d/%d files\n", status, cat, stats.Found, stats.Total)
		}
		fmt.Println()
	}

	fmt.Println(strings.Repeat("-", 60))
}

// outputValidateJSON prints health check results as JSON
func outputValidateJSON(health *bundle.HealthCheck) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.Encode(health)
}
