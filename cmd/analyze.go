// Package cmd implements the CLI commands for r8s.
// Sprint 9: r8s analyze - Main CLI entry point for bundle analysis (v0.8.0 CLI-First)
// Sprint 11: AI pattern detection integrated
// Sprint 12: Loading UX with personality and progress indicators
package cmd

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/Rancheroo/r8s/internal/ai"
	"github.com/Rancheroo/r8s/internal/bundle"
	"github.com/Rancheroo/r8s/internal/ui"
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
	RunE: runAnalyze,
}

var (
	analyzeFormat   string // Output format: table, json, yaml
	analyzeSeverity string // Filter by severity: critical, warning, info, all
)

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().StringVarP(&analyzeFormat, "format", "f", "table", "Output format: table, json")
	analyzeCmd.Flags().StringVarP(&analyzeSeverity, "severity", "s", "all", "Filter by severity: critical, warning, info, all")
}

// runAnalyze executes the analyze command with personality and progress
func runAnalyze(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		ui.ShowCmdUsage("analyze", "r8s analyze [bundle-path]", cmd.Long)
		return nil
	}
	bundlePath := args[0]
	startTime := time.Now()

	// Validate bundle path
	if _, err := os.Stat(bundlePath); err != nil {
		if os.IsNotExist(err) {
			ui.ShowBundleNotFoundError(bundlePath)
			return &ExitCodeError{Code: ExitError, Message: fmt.Sprintf("bundle not found: %s", bundlePath)}
		}
		return fmt.Errorf("cannot access bundle path: %w", err)
	}

	// Initialize loading display for non-JSON output
	var loader *ui.LoadingDisplay
	if analyzeFormat != "json" {
		loader = ui.NewLoadingDisplay(verbose)
		loader.ShowStartMessage(bundlePath)
		loader.ShowRandomLoadingMessage()
		loader.ShowFactAlways()
		fmt.Fprintln(os.Stderr)
	}

	// Show step 1: Checking bundle health
	if loader != nil {
		ui.ShowStep(1, 3, "Checking bundle health...")
	}

	// Check bundle health
	health, err := bundle.CheckHealth(bundlePath)
	if err != nil {
		if loader != nil {
			ui.ShowError("Failed to analyze bundle health")
		}
		ui.ShowFriendlyError(fmt.Errorf("failed to analyze bundle: %w", err))
		return err
	}

	if loader != nil {
		ui.ShowSuccess(fmt.Sprintf("Bundle validated (%s, %.0f%% complete)", health.BundleType, health.Completeness))
		fmt.Fprintln(os.Stderr)
	}

	// Build analysis result
	result := ui.AnalysisResult{
		BundlePath:   bundlePath,
		BundleType:   health.BundleType,
		Completeness: health.Completeness,
		Issues:       []ui.Issue{},
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

			result.Issues = append(result.Issues, ui.Issue{
				Severity:   severity,
				Type:       "missing_file",
				Resource:   missing.Path,
				Message:    fmt.Sprintf("Missing %s file", missing.Path),
				Suggestion: missing.Impact,
			})
		}
	}

	// Show step 2: Running AI pattern analysis
	if loader != nil {
		ui.ShowStep(2, 3, "Scanning for issues with AI pattern detection...")
		// Show another random loading message
		loader.ShowRandomLoadingMessage()
	}

	// Sprint 12: Run AI pattern analysis with progress
	aiIssues := analyzeBundlePatternsWithProgress(bundlePath, loader)
	for _, issue := range aiIssues {
		result.Issues = append(result.Issues, issue)
		switch issue.Severity {
		case "critical":
			result.Critical++
		case "warning":
			result.Warning++
		default:
			result.Info++
		}
	}

	// Show step 3: Complete
	if loader != nil {
		fmt.Fprintln(os.Stderr)
		ui.ShowStep(3, 3, "Analysis complete!")
		loader.ShowCompletionMessage()
		loader.ShowElapsedTime()
		fmt.Fprintln(os.Stderr)
	}

	// Record total analysis time in result for JSON output
	analysisDuration := time.Since(startTime)
	_ = analysisDuration // Available for future use in JSON output

	// Output based on format
	switch analyzeFormat {
	case "json":
		return ui.PrintAnalysisJSON(result)
	default:
		err := ui.PrintAnalysisTable(result)
		if err != nil {
			return err
		}
		// Return proper exit code for CI/CD integration
		// Sprint 11: Fix DEFECT #1 - Exit code 1 for issues, not 2
		if result.Critical > 0 {
			return NewExitError(ExitIssuesFound, fmt.Sprintf("analysis found %d critical issues", result.Critical))
		}
		if result.Warning > 0 {
			return NewExitError(ExitIssuesFound, fmt.Sprintf("analysis found %d warnings", result.Warning))
		}
		return nil
	}
}

// analyzeBundlePatternsWithProgress scans bundle content with progress indication
func analyzeBundlePatternsWithProgress(bundlePath string, loader *ui.LoadingDisplay) []ui.Issue {
	var issues []ui.Issue

	// Use parallel analyzer
	analyzer := ai.NewParallelAnalyzer()

	// Determine minimum severity from flag
	minSeverity := ai.SeverityInfo
	switch analyzeSeverity {
	case "critical":
		minSeverity = ai.SeverityCritical
	case "warning":
		minSeverity = ai.SeverityWarning
	}

	opts := ai.AnalysisOptions{
		MinSeverity: minSeverity,
	}

	// Collect bundle content
	bundleContent, err := collectBundleContent(bundlePath, "")
	if err != nil {
		if loader != nil {
			ui.ShowWarning("Failed to collect some bundle content")
		}
		return issues
	}

	// Show content stats
	if loader != nil && verbose {
		ui.ShowInfo(fmt.Sprintf("Found %d files to analyze", len(bundleContent)))
	}

	// Run parallel analysis with progress
	ctx := context.Background()

	var result *ai.AnalysisResult

	if loader != nil {
		// Show progress with personality
		spinner := ui.NewRancherSpinner("Herding logs into formation...")
		spinner.Start()

		result, err = analyzer.AnalyzeParallelWithProgress(ctx, bundleContent, opts, func(completed, total int, currentFile string) {
			// Update spinner message occasionally
			if completed%10 == 0 {
				messages := []string{
					"Moo-ving through your bundle...",
					"Rancher wrangling those logs...",
					"Herding log entries into formation...",
					"Scanning for CrashLoops...",
					"Checking etcd health...",
				}
				spinner.UpdateMessage(messages[rand.Intn(len(messages))])
			}

			// Show progress bar for verbose mode
			if verbose && completed%5 == 0 {
				spinner.Stop()
				ui.ShowFileProgress(filepath.Base(currentFile), completed, total)
				spinner.Start()
			}
		})

		spinner.Stop()
		ui.ClearLine()

		if err != nil {
			ui.ShowWarning("Analysis encountered some issues")
		}

		// Show fact during analysis
		loader.ShowFact()

	} else {
		// No loader - just run analysis silently
		result, err = analyzer.AnalyzeParallel(ctx, bundleContent, opts)
	}

	if err != nil {
		return issues
	}

	// Show summary of what was found
	if loader != nil && result.Summary.MatchesFound > 0 {
		ui.ShowSuccess(fmt.Sprintf("Found %d pattern matches (%d critical, %d warning)",
			result.Summary.MatchesFound,
			result.Summary.CriticalIssues,
			result.Summary.WarningIssues))
	}

	// Convert hints to issues
	for _, hint := range result.Hints {
		issue := hintToIssue(hint, hint.PatternID)
		if issue.Severity != "" {
			issues = append(issues, issue)
		}
	}

	return issues
}

// analyzeBundlePatternsLegacy scans bundle content using legacy analyzer
// Kept for backward compatibility
func analyzeBundlePatterns(bundlePath string) []ui.Issue {
	return analyzeBundlePatternsWithProgress(bundlePath, nil)
}

// hintToIssue converts an AI hint to an Issue for CLI output
func hintToIssue(hint *ai.Hint, source string) ui.Issue {
	severity := string(hint.Severity)
	if severity == "" {
		severity = "warning"
	}

	return ui.Issue{
		Severity:   severity,
		Type:       hint.PatternID,
		Resource:   source,
		Message:    hint.Summary,
		Suggestion: hint.Suggestion,
	}
}
