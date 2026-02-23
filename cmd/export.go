// Package cmd implements the CLI commands for r8s.
// Sprint 9 Day 4: r8s export - JSON/YAML findings export for CI integration
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/Rancheroo/r8s/internal/bundle"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export [bundle-path]",
	Short: "Export bundle findings as JSON/YAML",
	Long: `Export health, patterns, and analysis findings for CI/CD integration.

Perfect for automated workflows, monitoring pipelines, or processing with jq.

EXAMPLES:
  # Export as JSON (default)
  r8s export ./bundle/

  # Export as YAML
  r8s export ./bundle/ --format=yaml

  # Export only critical issues
  r8s export ./bundle/ --severity=critical

  # Pipe to jq for filtering
  r8s export ./bundle/ | jq '.health.completeness'

  # Save to file
  r8s export ./bundle/ --output=findings.json

INTEGRATION:
  # CI/CD pipeline
  r8s export ./bundle/ --format=json | \
    jq -e '.health.is_valid' || exit 1

  # Monitoring integration
  r8s export ./bundle/ --severity=critical | \
    curl -X POST -d @- https://monitoring.example.com/alerts`,
	Args: cobra.ExactArgs(1),
	RunE: runExport,
}

var (
	exportFormat   string // json, yaml
	exportSeverity string // critical, warning, all
	exportOutput   string // output file
	exportPattern  string // pattern filter
)

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "json", FormatHelp())
	exportCmd.Flags().StringVarP(&exportSeverity, "severity", "s", "all", "Filter by severity: critical, warning, all")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file (default: stdout)")
	exportCmd.Flags().StringVar(&exportPattern, "pattern", "", "Filter by pattern ID (e.g., oomkill)")
}

// ExportReport represents the full export structure
type ExportReport struct {
	Meta     ExportMeta          `json:"meta" yaml:"meta"`
	Health   *bundle.HealthCheck `json:"health" yaml:"health"`
	Findings []ExportFinding     `json:"findings" yaml:"findings"`
	Summary  ExportSummary       `json:"summary" yaml:"summary"`
}

// ExportMeta contains report metadata
type ExportMeta struct {
	GeneratedAt string `json:"generated_at" yaml:"generated_at"`
	BundlePath  string `json:"bundle_path" yaml:"bundle_path"`
	BundleType  string `json:"bundle_type" yaml:"bundle_type"`
	R8SVersion  string `json:"r8s_version" yaml:"r8s_version"`
}

// ExportFinding represents a single finding
type ExportFinding struct {
	ID          string                 `json:"id" yaml:"id"`
	Severity    string                 `json:"severity" yaml:"severity"`
	Category    string                 `json:"category" yaml:"category"`
	Title       string                 `json:"title" yaml:"title"`
	Description string                 `json:"description" yaml:"description"`
	Namespace   string                 `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Resource    string                 `json:"resource,omitempty" yaml:"resource,omitempty"`
	Suggestion  string                 `json:"suggestion,omitempty" yaml:"suggestion,omitempty"`
	Raw         map[string]interface{} `json:"raw,omitempty" yaml:"raw,omitempty"`
}

// ExportSummary contains high-level stats
type ExportSummary struct {
	TotalFindings    int     `json:"total_findings" yaml:"total_findings"`
	CriticalCount    int     `json:"critical_count" yaml:"critical_count"`
	WarningCount     int     `json:"warning_count" yaml:"warning_count"`
	InfoCount        int     `json:"info_count" yaml:"info_count"`
	HealthPercentage float64 `json:"health_percentage" yaml:"health_percentage"`
	IsValid          bool    `json:"is_valid" yaml:"is_valid"`
}

func runExport(cmd *cobra.Command, args []string) error {
	// Validate args length to prevent panic
	if len(args) == 0 {
		return NewExitError(ExitError, "bundle path argument required")
	}

	bundlePath := args[0]

	// Validate bundle
	if _, err := os.Stat(bundlePath); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), fmt.Sprintf("Error: bundle path not found: %v", err))
		return NewExitError(ExitError, fmt.Sprintf("bundle path not found: %v", err))
	}

	// Generate report
	report, err := generateExportReport(bundlePath)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), fmt.Sprintf("Error: failed to generate report: %v", err))
		return NewExitError(ExitError, fmt.Sprintf("failed to generate report: %v", err))
	}

	// Apply filters
	if exportSeverity != "all" {
		report.Findings = filterFindingsBySeverity(report.Findings, exportSeverity)
	}
	if exportPattern != "" {
		report.Findings = filterFindingsByPattern(report.Findings, exportPattern)
	}

	// Update summary after filtering
	report.Summary = calculateSummary(report)

	// Standardize format and output
	format := StandardizeFormat(exportFormat)

	var output []byte
	var marshalErr error

	switch format {
	case "yaml":
		output, marshalErr = yaml.Marshal(report)
	default:
		output, marshalErr = json.MarshalIndent(report, "", "  ")
	}

	if marshalErr != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), fmt.Sprintf("Error: failed to marshal output: %v", marshalErr))
		return NewExitError(ExitError, fmt.Sprintf("failed to marshal output: %v", marshalErr))
	}

	if exportOutput != "" {
		if err := os.WriteFile(exportOutput, output, 0644); err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), fmt.Sprintf("Error: failed to write output file: %v", err))
			return NewExitError(ExitError, fmt.Sprintf("failed to write output file: %v", err))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), fmt.Sprintf("✓ Report exported to %s", exportOutput))
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), string(output))
	}

	// Return appropriate exit code via error
	if !report.Summary.IsValid {
		return NewExitError(ExitIssuesFound, "bundle has issues")
	}
	return nil
}

// generateExportReport creates the full export report
func generateExportReport(bundlePath string) (*ExportReport, error) {
	report := &ExportReport{
		Meta: ExportMeta{
			GeneratedAt: time.Now().Format(time.RFC3339),
			BundlePath:  bundlePath,
			R8SVersion:  "v0.8.0-alpha",
		},
	}

	// Get health
	health, err := bundle.CheckHealth(bundlePath)
	if err != nil {
		return nil, err
	}
	report.Health = health
	report.Meta.BundleType = health.BundleType

	// Generate findings from health and patterns
	report.Findings = generateFindingsFromHealth(health)

	// Calculate summary
	report.Summary = calculateSummary(report)

	return report, nil
}

// generateFindingsFromHealth creates findings from health check
func generateFindingsFromHealth(health *bundle.HealthCheck) []ExportFinding {
	var findings []ExportFinding

	// Critical missing files
	for _, missing := range health.MissingFiles {
		severity := "warning"
		if missing.Importance == bundle.ImportanceCritical {
			severity = "critical"
		} else if missing.Importance == bundle.ImportanceHigh {
			severity = "warning"
		}

		finding := ExportFinding{
			ID:          fmt.Sprintf("missing-%s", missing.Path),
			Severity:    severity,
			Category:    "completeness",
			Title:       fmt.Sprintf("Missing %s", missing.Category),
			Description: missing.Impact,
			Resource:    missing.Path,
			Suggestion:  fmt.Sprintf("Verify collector script ran correctly for %s", missing.Category),
		}
		findings = append(findings, finding)
	}

	// Health status finding
	if !health.IsValid {
		findings = append(findings, ExportFinding{
			ID:          "bundle-invalid",
			Severity:    "critical",
			Category:    "health",
			Title:       "Bundle Incomplete",
			Description: fmt.Sprintf("Bundle is %.0f%% complete, missing critical files", health.Completeness),
			Suggestion:  "Re-run log collector or obtain complete bundle",
		})
	}

	return findings
}

// filterFindingsBySeverity filters findings by minimum severity
func filterFindingsBySeverity(findings []ExportFinding, minSeverity string) []ExportFinding {
	severityOrder := map[string]int{
		"info":     0,
		"warning":  1,
		"critical": 2,
	}

	minLevel := severityOrder[minSeverity]
	var filtered []ExportFinding

	for _, f := range findings {
		if severityOrder[f.Severity] >= minLevel {
			filtered = append(filtered, f)
		}
	}

	return filtered
}

// filterFindingsByPattern filters findings by pattern ID
func filterFindingsByPattern(findings []ExportFinding, patternID string) []ExportFinding {
	var filtered []ExportFinding

	for _, f := range findings {
		if f.ID == patternID || containsSubstring(f.ID, patternID) {
			filtered = append(filtered, f)
		}
	}

	return filtered
}

// calculateSummary calculates the summary statistics
func calculateSummary(report *ExportReport) ExportSummary {
	summary := ExportSummary{
		TotalFindings:    len(report.Findings),
		IsValid:          report.Health != nil && report.Health.IsValid,
		HealthPercentage: 0,
	}

	if report.Health != nil {
		summary.HealthPercentage = report.Health.Completeness
	}

	for _, f := range report.Findings {
		switch f.Severity {
		case "critical":
			summary.CriticalCount++
		case "warning":
			summary.WarningCount++
		default:
			summary.InfoCount++
		}
	}

	return summary
}

// containsSubstring checks if str contains substr
func containsSubstring(str, substr string) bool {
	return strings.Contains(str, substr)
}
