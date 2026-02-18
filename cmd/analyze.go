// Package cmd implements the CLI commands for r8s.
// Sprint 9: r8s analyze - Main CLI entry point for bundle analysis (v0.8.0 CLI-First)
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

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze [bundle-path]",
	Short: "Analyze bundle and detect issues (default command)",
	Long: `Analyze a Rancher support bundle and detect issues.

This is the primary command for r8s v0.8.0+. It analyzes the bundle
and reports critical issues, warnings, and informational findings.

EXAMPLES:
  # Analyze a bundle with table output (default)
  r8s analyze ./extracted-bundle/

  # Analyze with JSON output for piping
  r8s analyze ./bundle/ --format=json | jq '.critical'

EXIT CODES:
  0 - No issues found or analyzed successfully
  1 - Issues detected (warnings or critical)
  2 - Invalid bundle or path`,
	Args: cobra.ExactArgs(1),
	RunE: runAnalyze,
}

var (
	analyzeFormat string // Output format: table, json, yaml
)

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().StringVarP(&analyzeFormat, "format", "f", "table", "Output format: table, json, yaml")
}

// AnalysisResult represents the output of bundle analysis
type AnalysisResult struct {
	BundlePath   string           `json:"bundle_path"`
	BundleType   string           `json:"bundle_type"`
	Completeness float64          `json:"completeness"`
	Issues       []Issue          `json:"issues"`
	Critical     int              `json:"critical_count"`
	Warning      int              `json:"warning_count"`
	Info         int              `json:"info_count"`
	Health       *bundle.HealthCheck `json:"health,omitempty"`
}

// Issue represents a single detected issue
type Issue struct {
	Severity    string `json:"severity"`
	Type        string `json:"type"`
	Resource    string `json:"resource"`
	Message     string `json:"message"`
	Suggestion  string `json:"suggestion,omitempty"`
}

// runAnalyze executes the analyze command
func runAnalyze(cmd *cobra.Command, args []string) error {
	bundlePath := args[0]

	// Validate bundle path
	if _, err := os.Stat(bundlePath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot access bundle path: %v\n", err)
		os.Exit(2)
		return err
	}

	// Check bundle health
	health, err := bundle.CheckHealth(bundlePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to analyze bundle: %v\n", err)
		os.Exit(2)
		return err
	}

	// Build analysis result
	result := AnalysisResult{
		BundlePath:   bundlePath,
		BundleType:   health.BundleType,
		Completeness: health.Completeness,
		Issues:       []Issue{},
		Health:       health,
	}

	// Add health-based issues
	if health.Completeness < 100 {
		for _, missing := range health.MissingFiles {
			severity := "info"
			switch missing.Importance {
			case bundle.ImportanceCritical:
				severity = "critical"
				result.Critical++
			case bundle.ImportanceHigh:
				severity = "warning"
				result.Warning++
			default:
				result.Info++
			}
			
			result.Issues = append(result.Issues, Issue{
				Severity:   severity,
				Type:       "missing_file",
				Resource:   missing.Path,
				Message:    fmt.Sprintf("Missing %s file", missing.Path),
				Suggestion: missing.Impact,
			})
		}
	}

	// Output based on format
	switch analyzeFormat {
	case "json":
		return outputAnalyzeJSON(result)
	default:
		return outputAnalyzeTable(result)
	}
}

// outputAnalyzeTable prints analysis in table format
func outputAnalyzeTable(result AnalysisResult) error {
	fmt.Println()
	header := color.New(color.Bold, color.FgCyan)
	header.Println("R8S Bundle Analysis")
	header.Println(strings.Repeat("═", 60))
	fmt.Println()

	// Bundle summary
	fmt.Printf("Bundle: %s\n", result.BundlePath)
	fmt.Printf("Type:   %s\n", result.BundleType)
	
	// Completeness indicator
	completenessColor := color.GreenString
	if result.Completeness < 70 {
		completenessColor = color.RedString
	} else if result.Completeness < 90 {
		completenessColor = color.YellowString
	}
	fmt.Printf("Health: %s (%.0f%% complete)\n", completenessColor("●"), result.Completeness)
	fmt.Println()

	// Issues summary
	if result.Critical > 0 || result.Warning > 0 {
		issueHeader := color.New(color.Bold)
		issueHeader.Println("Issues Found:")
		fmt.Println()

		// Group by severity
		criticalIssues := []Issue{}
		warningIssues := []Issue{}
		infoIssues := []Issue{}

		for _, issue := range result.Issues {
			switch issue.Severity {
			case "critical":
				criticalIssues = append(criticalIssues, issue)
			case "warning":
				warningIssues = append(warningIssues, issue)
			default:
				infoIssues = append(infoIssues, issue)
			}
		}

		// Print critical
		for _, issue := range criticalIssues {
			fmt.Println(color.RedString("🔴 CRITICAL"))
			fmt.Printf("   %s: %s\n", issue.Type, issue.Resource)
			fmt.Printf("   %s\n", issue.Message)
			if issue.Suggestion != "" {
				fmt.Printf("   → %s\n", issue.Suggestion)
			}
			fmt.Println()
		}

		// Print warnings
		for _, issue := range warningIssues {
			fmt.Println(color.YellowString("⚠️  WARNING"))
			fmt.Printf("   %s: %s\n", issue.Type, issue.Resource)
			fmt.Printf("   %s\n", issue.Message)
			fmt.Println()
		}

		// Print info (condensed)
		if len(infoIssues) > 0 {
			fmt.Println(color.CyanString("ℹ️  INFO"))
			for _, issue := range infoIssues {
				fmt.Printf("   • %s: %s\n", issue.Type, issue.Resource)
			}
			fmt.Println()
		}
	} else {
		fmt.Println(color.GreenString("✓ No issues detected"))
		fmt.Println()
	}

	// Summary line
	fmt.Println(strings.Repeat("─", 60))
	if result.Critical > 0 {
		fmt.Printf("Result: %s (%d critical, %d warning)\n", 
			color.RedString("ISSUES FOUND"), result.Critical, result.Warning)
		os.Exit(1)
	} else if result.Warning > 0 {
		fmt.Printf("Result: %s (%d warning)\n", 
			color.YellowString("WARNINGS"), result.Warning)
	} else {
		fmt.Printf("Result: %s\n", color.GreenString("HEALTHY"))
	}
	fmt.Println()

	return nil
}

// outputAnalyzeJSON prints analysis as JSON
func outputAnalyzeJSON(result AnalysisResult) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	// Exit with appropriate code
	if result.Critical > 0 {
		os.Exit(1)
	}
	return nil
}
