package ui

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

// PrintAnalysisTable prints analysis in table format
func PrintAnalysisTable(result AnalysisResult) error {
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
	rand.Seed(time.Now().UnixNano()) // Ensure randomness (though init() usually does this)
	if rand.Intn(3) == 0 {           // 1/3 chance to show tip
		if len(R8sFacts) > 0 {
			tip := R8sFacts[rand.Intn(len(R8sFacts))]
			tipColor := color.New(color.Italic, color.FgHiBlack)
			tipColor.Fprintln(os.Stderr, "💡 "+tip)
			fmt.Println()
		}
	}

	return nil
}

// PrintAnalysisJSON prints analysis as JSON
func PrintAnalysisJSON(result AnalysisResult) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	return encoder.Encode(result)
}
