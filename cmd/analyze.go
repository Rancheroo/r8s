// Package cmd implements the CLI commands for r8s.
// Sprint 9: r8s analyze - Main CLI entry point for bundle analysis (v0.8.0 CLI-First)
// Sprint 11: AI pattern detection integrated
// Sprint 12: Loading UX with personality and progress indicators
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/Rancheroo/r8s/internal/ai"
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
	analyzeFormat   string // Output format: table, json, yaml
	analyzeSeverity string // Filter by severity: critical, warning, info, all
)

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().StringVarP(&analyzeFormat, "format", "f", "table", "Output format: table, json")
	analyzeCmd.Flags().StringVarP(&analyzeSeverity, "severity", "s", "all", "Filter by severity: critical, warning, info, all")
}

// AnalysisResult represents the output of bundle analysis
type AnalysisResult struct {
	BundlePath   string              `json:"bundle_path"`
	BundleType   string              `json:"bundle_type"`
	Completeness float64             `json:"completeness"`
	Issues       []Issue             `json:"issues"`
	Critical     int                 `json:"critical_count"`
	Warning      int                 `json:"warning_count"`
	Info         int                 `json:"info_count"`
	Health       *bundle.HealthCheck `json:"health,omitempty"`
}

// Issue represents a single detected issue
type Issue struct {
	Severity   string `json:"severity"`
	Type       string `json:"type"`
	Resource   string `json:"resource"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

// runAnalyze executes the analyze command with personality and progress
func runAnalyze(cmd *cobra.Command, args []string) error {
	bundlePath := args[0]
	startTime := time.Now()

	// Validate bundle path
	if _, err := os.Stat(bundlePath); err != nil {
		if os.IsNotExist(err) {
			ShowBundleNotFoundError(bundlePath)
			return fmt.Errorf("bundle not found: %s", bundlePath)
		}
		ShowFriendlyError(fmt.Errorf("cannot access bundle path: %w", err))
		return err
	}

	// Initialize loading display for non-JSON output
	var loader *LoadingDisplay
	if analyzeFormat != "json" {
		loader = NewLoadingDisplay(verbose)
		loader.ShowStartMessage(bundlePath)
		loader.ShowRandomLoadingMessage()
		loader.ShowFactAlways()
		fmt.Fprintln(os.Stderr)
	}

	// Show step 1: Checking bundle health
	if loader != nil {
		ShowStep(1, 3, "Checking bundle health...")
	}

	// Check bundle health
	health, err := bundle.CheckHealth(bundlePath)
	if err != nil {
		if loader != nil {
			ShowError("Failed to analyze bundle health")
		}
		ShowFriendlyError(fmt.Errorf("failed to analyze bundle: %w", err))
		return err
	}

	if loader != nil {
		ShowSuccess(fmt.Sprintf("Bundle validated (%s, %.0f%% complete)", health.BundleType, health.Completeness))
		fmt.Fprintln(os.Stderr)
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

	// Show step 2: Running AI pattern analysis
	if loader != nil {
		ShowStep(2, 3, "Scanning for issues with AI pattern detection...")
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
		ShowStep(3, 3, "Analysis complete!")
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
		return outputAnalyzeJSON(result)
	default:
		return outputAnalyzeTable(result)
	}
}

// analyzeBundlePatternsWithProgress scans bundle content with progress indication
func analyzeBundlePatternsWithProgress(bundlePath string, loader *LoadingDisplay) []Issue {
	var issues []Issue
	
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
			ShowWarning("Failed to collect some bundle content")
		}
		return issues
	}
	
	// Show content stats
	if loader != nil && verbose {
		ShowInfo(fmt.Sprintf("Found %d files to analyze", len(bundleContent)))
	}
	
	// Run parallel analysis with progress
	ctx := context.Background()
	
	var result *ai.AnalysisResult
	
	if loader != nil {
		// Show progress with personality
		spinner := NewRancherSpinner("Herding logs into formation...")
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
				ShowFileProgress(filepath.Base(currentFile), completed, total)
				spinner.Start()
			}
		})
		
		spinner.Stop()
		ClearLine()
		
		if err != nil {
			ShowWarning("Analysis encountered some issues")
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
		ShowSuccess(fmt.Sprintf("Found %d pattern matches (%d critical, %d warning)",
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
func analyzeBundlePatterns(bundlePath string) []Issue {
	return analyzeBundlePatternsWithProgress(bundlePath, nil)
}

// hintToIssue converts an AI hint to an Issue for CLI output
func hintToIssue(hint *ai.Hint, source string) Issue {
	severity := string(hint.Severity)
	if severity == "" {
		severity = "warning"
	}

	return Issue{
		Severity:   severity,
		Type:       hint.PatternID,
		Resource:   source,
		Message:    hint.Summary,
		Suggestion: hint.Suggestion,
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
		// Issue #86: Show helpful message when no issues found
		ShowNoIssuesFound(result.BundlePath)
	}

	// Summary line
	fmt.Println(strings.Repeat("─", 60))
	if result.Critical > 0 {
		fmt.Printf("Result: %s (%d critical, %d warning)\n",
			color.RedString("ISSUES FOUND"), result.Critical, result.Warning)
	} else if result.Warning > 0 {
		fmt.Printf("Result: %s (%d warning)\n",
			color.YellowString("WARNINGS"), result.Warning)
	} else {
		fmt.Printf("Result: %s\n", color.GreenString("HEALTHY"))
	}
	fmt.Println()

	// Show random tip at the end
	if rand.Intn(3) == 0 { // 1/3 chance to show tip
		tip := R8sFacts[rand.Intn(len(R8sFacts))]
		tipColor := color.New(color.Italic, color.FgHiBlack)
		tipColor.Fprintln(os.Stderr, "💡 "+tip)
		fmt.Println()
	}

	// Return proper exit code for CI/CD integration
	// Sprint 11: Fix DEFECT #1 - Exit code 1 for issues, not 2
	if result.Critical > 0 {
		return NewExitError(ExitIssuesFound, fmt.Sprintf("analysis found %d critical issues", result.Critical))
	}

	return nil
}

// outputAnalyzeJSON prints analysis as JSON
func outputAnalyzeJSON(result AnalysisResult) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	return encoder.Encode(result)
}

// init seeds the random number generator
func init() {
	rand.Seed(time.Now().UnixNano())
}
