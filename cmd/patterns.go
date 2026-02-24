// Package cmd implements the CLI commands for r8s.
// Sprint 11 Day 11: Pattern Registry Commands
package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/Rancheroo/r8s/internal/ai"
)

// patternsCmd represents the patterns command
var patternsCmd = &cobra.Command{
	Use:   "patterns [subcommand]",
	Short: "Manage and query pattern definitions",
	Long: `View, search, and manage Kubernetes issue detection patterns.

This command provides access to the Sprint 11 pattern registry, allowing
you to browse patterns, understand what they detect, and see examples.

SUBCOMMANDS:
  list      List all available patterns
  show      Show details for a specific pattern
  search    Search patterns by keyword

EXAMPLES:
  # List all patterns
  r8s patterns list

  # Show pattern details
  r8s patterns show crashloopbackoff-v2

  # Search for network-related patterns
  r8s patterns search network`,
}

// listCmd represents the patterns list subcommand
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available patterns",
	Long: `List all Kubernetes issue detection patterns in the registry.

Shows pattern ID, category, severity, and description for each pattern.

EXAMPLES:
  # List all patterns
  r8s patterns list

  # List only critical patterns
  r8s patterns list --severity=critical

  # List patterns by category
  r8s patterns list --category=network`,
	RunE: runPatternsList,
}

// showCmd represents the patterns show subcommand
var showCmd = &cobra.Command{
	Use:   "show [pattern-id]",
	Short: "Show details for a specific pattern",
	Long: `Display detailed information about a specific pattern.

Shows the pattern definition, matchers, examples, and suggestions.

EXAMPLES:
  # Show crashloopbackoff pattern
  r8s patterns show crashloopbackoff-v2

  # Show with examples
  r8s patterns show oomkill-v2 --examples`,
	Args: cobra.ExactArgs(1),
	RunE: runPatternsShow,
}

// searchCmd represents the patterns search subcommand
var searchCmd = &cobra.Command{
	Use:   "search [keyword]",
	Short: "Search patterns by keyword",
	Long: `Search for patterns matching a keyword in their ID, name, or description.

EXAMPLES:
  # Search for network patterns
  r8s patterns search network

  # Search for certificate patterns
  r8s patterns search cert

  # Search for crash patterns
  r8s patterns search crash`,
	Args: cobra.ExactArgs(1),
	RunE: runPatternsSearch,
}

var (
	listSeverity string
	listCategory string
	showExamples bool
)

func init() {
	rootCmd.AddCommand(patternsCmd)
	patternsCmd.AddCommand(listCmd)
	patternsCmd.AddCommand(showCmd)
	patternsCmd.AddCommand(searchCmd)

	listCmd.Flags().StringVar(&listSeverity, "severity", "", "Filter by severity: critical, warning, info")
	listCmd.Flags().StringVar(&listCategory, "category", "", "Filter by category")

	showCmd.Flags().BoolVar(&showExamples, "examples", false, "Show example matches")
}

// runPatternsList executes the list command
func runPatternsList(cmd *cobra.Command, args []string) error {
	registry := ai.NewRegistryV2()
	patterns := registry.GetAll()

	// Filter by severity if specified
	if listSeverity != "" {
		var filtered []ai.PatternV2
		for _, p := range patterns {
			if strings.EqualFold(string(p.Severity), listSeverity) {
				filtered = append(filtered, p)
			}
		}
		patterns = filtered
	}

	// Filter by category if specified
	if listCategory != "" {
		var filtered []ai.PatternV2
		for _, p := range patterns {
			if strings.EqualFold(p.Category, listCategory) {
				filtered = append(filtered, p)
			}
		}
		patterns = filtered
	}

	// Sort by severity (critical first)
	sort.Slice(patterns, func(i, j int) bool {
		severityOrder := map[ai.Severity]int{
			ai.SeverityCritical: 0,
			ai.SeverityWarning:  1,
			ai.SeverityInfo:     2,
		}
		return severityOrder[patterns[i].Severity] < severityOrder[patterns[j].Severity]
	})

	// Print table
	fmt.Println()
	header := color.New(color.Bold, color.FgCyan)
	header.Println("Available Patterns")
	header.Println(strings.Repeat("═", 80))
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tCATEGORY\tSEVERITY\tDESCRIPTION")
	fmt.Fprintln(w, strings.Repeat("-", 80))

	for _, p := range patterns {
		severityColor := color.WhiteString
		switch p.Severity {
		case ai.SeverityCritical:
			severityColor = color.RedString
		case ai.SeverityWarning:
			severityColor = color.YellowString
		case ai.SeverityInfo:
			severityColor = color.CyanString
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			p.ID,
			p.Category,
			severityColor(string(p.Severity)),
			truncateString(p.Description, 40),
		)
	}

	w.Flush()

	fmt.Println()
	fmt.Printf("Total: %d patterns\n", len(patterns))
	fmt.Println()

	return nil
}

// runPatternsShow executes the show command
func runPatternsShow(cmd *cobra.Command, args []string) error {
	patternID := args[0]

	registry := ai.NewRegistryV2()
	pattern, found := registry.GetByID(patternID)
	if !found {
		return fmt.Errorf("pattern not found: %s", patternID)
	}

	fmt.Println()
	header := color.New(color.Bold, color.FgCyan)
	header.Println("Pattern Details")
	header.Println(strings.Repeat("═", 60))
	fmt.Println()

	fmt.Printf("ID:          %s\n", pattern.ID)
	fmt.Printf("Name:        %s\n", pattern.Name)
	fmt.Printf("Category:    %s\n", pattern.Category)
	fmt.Printf("Severity:    %s\n", pattern.Severity)
	fmt.Printf("Confidence:  %s\n", pattern.Confidence)
	fmt.Println()

	fmt.Println("Description:")
	fmt.Printf("  %s\n\n", pattern.Description)

	fmt.Println("Matchers:")
	for i, m := range pattern.Matchers {
		fmt.Printf("  %d. [%s] %s (weight: %.1f)\n", i+1, m.Type, m.Pattern, m.Weight)
	}
	fmt.Println()

	if len(pattern.Correlations) > 0 {
		fmt.Println("Correlations:")
		for _, c := range pattern.Correlations {
			fmt.Printf("  • %s: %s\n", c.PatternID, c.Message)
		}
		fmt.Println()
	}

	fmt.Println("Hint Template:")
	fmt.Printf("  %s\n\n", pattern.HintGenerator.Template)

	fmt.Println("Suggestion:")
	fmt.Printf("  %s\n\n", pattern.HintGenerator.Suggestion)

	if pattern.HintGenerator.Command != "" {
		fmt.Println("Command:")
		fmt.Printf("  %s\n\n", pattern.HintGenerator.Command)
	}

	if len(pattern.HintGenerator.References) > 0 {
		fmt.Println("References:")
		for _, ref := range pattern.HintGenerator.References {
			fmt.Printf("  • %s\n", ref)
		}
		fmt.Println()
	}

	if showExamples {
		fmt.Println("Examples:")
		fmt.Println("  (Use 'r8s analyze' on a bundle to see pattern matches)")
		fmt.Println()
	}

	return nil
}

// runPatternsSearch executes the search command
func runPatternsSearch(cmd *cobra.Command, args []string) error {
	keyword := strings.ToLower(args[0])

	registry := ai.NewRegistryV2()
	patterns := registry.GetAll()

	var matches []ai.PatternV2
	for _, p := range patterns {
		// Search in ID, Name, Description, Category
		searchText := strings.ToLower(p.ID + " " + p.Name + " " + p.Description + " " + p.Category)
		if strings.Contains(searchText, keyword) {
			matches = append(matches, p)
		}
	}

	if len(matches) == 0 {
		fmt.Printf("No patterns found matching '%s'\n", keyword)
		return nil
	}

	fmt.Println()
	header := color.New(color.Bold, color.FgCyan)
	header.Printf("Search Results for '%s'\n", keyword)
	header.Println(strings.Repeat("═", 60))
	fmt.Println()

	for _, p := range matches {
		severityColor := color.WhiteString
		switch p.Severity {
		case ai.SeverityCritical:
			severityColor = color.RedString
		case ai.SeverityWarning:
			severityColor = color.YellowString
		case ai.SeverityInfo:
			severityColor = color.CyanString
		}

		fmt.Printf("%s [%s] %s\n", p.ID, severityColor(string(p.Severity)), p.Name)
		fmt.Printf("  %s\n\n", truncateString(p.Description, 60))
	}

	fmt.Printf("Found %d patterns\n\n", len(matches))

	return nil
}

// truncateString truncates a string to max length with ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}