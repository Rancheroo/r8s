// Package cmd implements the CLI commands for r8s.
// Sprint 11 Day 8: Export Formats - SARIF, JUnit, Markdown
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Rancheroo/r8s/internal/ai"
	"github.com/Rancheroo/r8s/internal/bundle"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export [bundle-path]",
	Short: "Export bundle analysis to various formats",
	Long: `Export bundle analysis results to SARIF, JUnit XML, or Markdown formats.

This command analyzes the bundle and exports findings for integration with
CI/CD pipelines, security scanners, and documentation.

FORMATS:
  sarif     - Static Analysis Results Interchange Format (GitHub, Azure DevOps)
  junit     - JUnit XML for test result reporting (Jenkins, GitHub Actions)
  markdown  - Human-readable report for documentation

EXAMPLES:
  # Export to SARIF for GitHub Advanced Security
  r8s export ./bundle/ --format=sarif --output=results.sarif

  # Export to JUnit for CI/CD integration
  r8s export ./bundle/ --format=junit --output=test-results.xml

  # Export to Markdown for documentation
  r8s export ./bundle/ --format=markdown --output=report.md

  # Export only critical issues
  r8s export ./bundle/ --format=sarif --min-severity=critical`,
	Args: cobra.ExactArgs(1),
	RunE: runExport,
}

var (
	exportFormat   string // Export format: sarif, junit, markdown
	exportOutput   string // Output file (default: stdout)
	exportMinSev   string // Minimum severity: critical, warning, info
	exportWithLogs bool   // Include log snippets in output
)

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "sarif", "Export format: sarif, junit, markdown")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file (default: stdout)")
	exportCmd.Flags().StringVar(&exportMinSev, "min-severity", "info", "Minimum severity to include")
	exportCmd.Flags().BoolVar(&exportWithLogs, "with-logs", false, "Include log snippets in output")
}

// runExport executes the export command
func runExport(cmd *cobra.Command, args []string) error {
	bundlePath := args[0]

	// Validate bundle path
	if _, err := os.Stat(bundlePath); err != nil {
		return fmt.Errorf("cannot access bundle path: %w", err)
	}

	// Check bundle health
	health, err := bundle.CheckHealth(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to analyze bundle: %w", err)
	}

	// Run Sprint 11 AI pattern analysis
	analyzer := ai.NewAnalyzer()
	opts := ai.AnalysisOptions{
		MinSeverity:  parseSeverity(exportMinSev),
		IncludeInfo:  exportMinSev == "info",
		MaxHints:     0, // No limit
	}

	// Scan bundle content
	bundleContent, err := collectBundleContent(bundlePath, health.BundleType)
	if err != nil {
		return fmt.Errorf("failed to collect bundle content: %w", err)
	}

	// Analyze
	results, err := analyzer.AnalyzeMultiple(bundleContent, opts)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Merge all results
	mergedHints := mergeAnalysisResults(results)

	// Export based on format
	var output []byte
	switch exportFormat {
	case "sarif":
		output, err = exportSARIF(bundlePath, health, mergedHints)
	case "junit":
		output, err = exportJUnit(bundlePath, health, mergedHints)
	case "markdown":
		output, err = exportMarkdown(bundlePath, health, mergedHints)
	default:
		return fmt.Errorf("unknown export format: %s (use: sarif, junit, markdown)", exportFormat)
	}

	if err != nil {
		return fmt.Errorf("export failed: %w", err)
	}

	// Output to file or stdout
	if exportOutput != "" {
		if err := os.WriteFile(exportOutput, output, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Exported %d findings to %s\n", len(mergedHints), exportOutput)
	} else {
		fmt.Print(string(output))
	}

	return nil
}

// collectBundleContent gathers content from bundle files
func collectBundleContent(bundlePath, bundleType string) (map[string]string, error) {
	content := make(map[string]string)

	// File patterns to scan based on bundle type
	// Force lowercase for directory names (RKE2 -> rke2)
	lowerType := strings.ToLower(bundleType)
	
	patterns := []string{
		// Type-specific paths
		fmt.Sprintf("%s/kubectl/pods", lowerType),
		fmt.Sprintf("%s/podlogs/*", lowerType),
		fmt.Sprintf("%s/agent/logs/*.log", lowerType),
		
		// Generic paths
		"kubectl/pods",
		"pod-logs/*.log",
		"journald/*.log",
		"cluster/events.json",
	}

	const maxFileSize = 2 * 1024 * 1024 // 2MB per file

	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(bundlePath, pattern))
		if err != nil {
			continue
		}

		for _, match := range matches {
			// Check file size before reading to avoid OOM
			info, err := os.Stat(match)
			if err != nil {
				continue
			}
			if info.IsDir() || info.Size() > maxFileSize {
				continue // Skip directories and large files
			}

			data, err := os.ReadFile(match)
			if err != nil {
				continue
			}
			content[match] = string(data)
		}
	}

	return content, nil
}

// mergeAnalysisResults combines hints from multiple files
func mergeAnalysisResults(results map[string]*ai.AnalysisResult) []*ai.Hint {
	var allHints []*ai.Hint
	seen := make(map[string]bool)

	for _, result := range results {
		for _, hint := range result.Hints {
			key := hint.PatternID + hint.Summary
			if !seen[key] {
				allHints = append(allHints, hint)
				seen[key] = true
			}
		}
	}

	// Sort for deterministic output (by severity, then pattern ID)
	sort.Slice(allHints, func(i, j int) bool {
		severityOrder := map[ai.Severity]int{
			ai.SeverityCritical: 0,
			ai.SeverityWarning:  1,
			ai.SeverityInfo:     2,
		}
		if severityOrder[allHints[i].Severity] != severityOrder[allHints[j].Severity] {
			return severityOrder[allHints[i].Severity] < severityOrder[allHints[j].Severity]
		}
		return allHints[i].PatternID < allHints[j].PatternID
	})

	return allHints
}

// parseSeverity converts string to Severity type
func parseSeverity(s string) ai.Severity {
	switch s {
	case "critical":
		return ai.SeverityCritical
	case "warning":
		return ai.SeverityWarning
	default:
		return ai.SeverityInfo
	}
}