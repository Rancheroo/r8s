// Package cmd implements the CLI commands for r8s.
// Sprint 11 Day 10: Natural Language Queries v1
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/Rancheroo/r8s/internal/ai"
	"github.com/Rancheroo/r8s/internal/bundle"
)

// askCmd represents the ask command
var askCmd = &cobra.Command{
	Use:   "ask [bundle-path] [question]",
	Short: "Ask natural language questions about your bundle",
	Long: `Ask questions in plain English about issues in your bundle.

This is the natural language interface for r8s. It parses your question,
queries the bundle data, and answers with context-specific responses.

SUPPORTED QUERY TYPES (v1):
  "why is <pod> crashing?"    - Explain crash reasons with logs
  "show me <pattern> issues"   - List all issues of a type
  "which <resource> are <state>?" - Find resources in specific states
  "what is causing <issue>?"   - Root cause analysis

EXAMPLES:
  # Ask about a specific pod
  r8s ask ./bundle/ "why is nginx-pod crashing?"
  
  # Find all image pull issues
  r8s ask ./bundle/ "show me imagepullbackoff issues"
  
  # Check certificate state
  r8s ask ./bundle/ "which certificates are expired?"
  
  # Get help for a specific pod
  r8s ask ./bundle/ "what is wrong with worker-1?"

LIMITATIONS (v1):
- Single-fact questions only (no compound queries)
- No follow-up questions
- Limited to supported query patterns`,
	Args: cobra.MinimumNArgs(2),
	RunE: runAsk,
}

func init() {
	rootCmd.AddCommand(askCmd)
}

// runAsk executes the ask command
func runAsk(cmd *cobra.Command, args []string) error {
	bundlePath := args[0]
	question := strings.Join(args[1:], " ")

	// Validate bundle path
	if _, err := os.Stat(bundlePath); err != nil {
		return fmt.Errorf("cannot access bundle path: %w", err)
	}

	// Parse the question
	intent := parseQueryIntent(question)

	// Load bundle
	health, err := bundle.CheckHealth(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to analyze bundle: %w", err)
	}

	// Run analysis
	analyzer := ai.NewAnalyzer()
	bundleContent, err := collectBundleContent(bundlePath, health.BundleType)
	if err != nil {
		return fmt.Errorf("failed to collect bundle content: %w", err)
	}

	results, err := analyzer.AnalyzeMultiple(bundleContent, ai.AnalysisOptions{
		MinSeverity: ai.SeverityInfo,
		IncludeInfo: true,
	})
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	mergedHints := mergeAnalysisResults(results)

	// Generate response based on intent
	response := generateResponse(intent, mergedHints, question)

	// Output response
	fmt.Println()
	header := color.New(color.Bold, color.FgCyan)
	header.Println("🤖 R8S Natural Language Query")
	header.Println(strings.Repeat("═", 60))
	fmt.Println()

	fmt.Printf("Question: %s\n\n", question)

	fmt.Println(response)

	return nil
}

// QueryIntent represents parsed user query
type QueryIntent struct {
	Type      string // "why", "show", "which", "what"
	Resource  string // "pod", "certificate", "image"
	Name      string // specific resource name if mentioned
	Condition string // "crashing", "expired", "pending"
	Filter    string // additional filters
}

// parseQueryIntent extracts intent from natural language question
func parseQueryIntent(question string) QueryIntent {
	q := strings.ToLower(question)
	intent := QueryIntent{}

	// Determine query type
	switch {
	case strings.HasPrefix(q, "why"):
		intent.Type = "why"
	case strings.HasPrefix(q, "show") || strings.HasPrefix(q, "find") || strings.HasPrefix(q, "list"):
		intent.Type = "show"
	case strings.HasPrefix(q, "which"):
		intent.Type = "which"
	case strings.HasPrefix(q, "what"):
		intent.Type = "what"
	default:
		intent.Type = "unknown"
	}

	// Extract resource type and name
	if strings.Contains(q, "pod") || strings.Contains(q, "container") {
		intent.Resource = "pod"
		// Try to extract pod name
		for _, suffix := range []string{"crashing", "pending", "not ready", "wrong"} {
			if idx := strings.Index(q, "is "); idx != -1 {
				afterIs := q[idx+3:]
				endIdx := strings.Index(afterIs, suffix)
				if endIdx != -1 {
					intent.Name = strings.TrimSpace(afterIs[:endIdx])
					break
				}
			}
		}
	}

	if strings.Contains(q, "certificate") || strings.Contains(q, "cert") {
		intent.Resource = "certificate"
	}

	if strings.Contains(q, "image") || strings.Contains(q, "docker") {
		intent.Resource = "image"
	}

	if strings.Contains(q, "node") {
		intent.Resource = "node"
	}

	// Extract condition
	switch {
	case strings.Contains(q, "crash") || strings.Contains(q, "crashloop"):
		intent.Condition = "crashing"
	case strings.Contains(q, "imagepull") || strings.Contains(q, "can't pull"):
		intent.Condition = "imagepull"
	case strings.Contains(q, "pending"):
		intent.Condition = "pending"
	case strings.Contains(q, "expire"):
		intent.Condition = "expired"
	case strings.Contains(q, "not ready"):
		intent.Condition = "notready"
	}

	return intent
}

// generateResponse creates an answer based on intent and findings
func generateResponse(intent QueryIntent, hints []*ai.Hint, originalQuestion string) string {
	if intent.Type == "unknown" {
		return formatUnknownResponse()
	}

	// Filter hints based on intent
	var relevantHints []*ai.Hint
	for _, hint := range hints {
		if matchesIntent(hint, intent) {
			relevantHints = append(relevantHints, hint)
		}
	}

	if len(relevantHints) == 0 {
		return formatNoResultsResponse(intent)
	}

	// Generate specific response based on query type
	switch intent.Type {
	case "why":
		return formatWhyResponse(intent, relevantHints)
	case "show":
		return formatShowResponse(intent, relevantHints)
	case "which":
		return formatWhichResponse(intent, relevantHints)
	case "what":
		return formatWhatResponse(intent, relevantHints)
	default:
		return formatGeneralResponse(relevantHints)
	}
}

// matchesIntent checks if a hint matches the query intent
func matchesIntent(hint *ai.Hint, intent QueryIntent) bool {
	// Check pattern ID matches resource/condition
	patternIDLower := strings.ToLower(hint.PatternID)
	
	// Check condition match
	if intent.Condition != "" {
		if !strings.Contains(patternIDLower, intent.Condition) {
			// Also check aliases
			if intent.Condition == "crashing" && !strings.Contains(patternIDLower, "crash") {
				return false
			}
			if intent.Condition == "imagepull" && !strings.Contains(patternIDLower, "imagepull") {
				return false
			}
		}
	}

	// Check resource match
	if intent.Resource == "pod" {
		// Many patterns relate to pods
		podPatterns := []string{"crashloop", "imagepull", "oomkill", "pending", "terminating"}
		found := false
		for _, pp := range podPatterns {
			if strings.Contains(patternIDLower, pp) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// Response formatters

func formatWhyResponse(intent QueryIntent, hints []*ai.Hint) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("🔍 Analysis: Why resources are %s\n\n", intent.Condition))
	
	for i, hint := range hints {
		sb.WriteString(fmt.Sprintf("**Finding %d:** %s\n\n", i+1, hint.Summary))
		
		if hint.Explanation != "" {
			sb.WriteString(fmt.Sprintf("**What happened:**\n%s\n\n", hint.Explanation))
		}
		
		if hint.Suggestion != "" {
			sb.WriteString(fmt.Sprintf("**What to do:**\n%s\n\n", hint.Suggestion))
		}
		
		if hint.Command != "" {
			sb.WriteString(fmt.Sprintf("**Try this command:**\n```\n%s\n```\n\n", hint.Command))
		}
		
		if i < len(hints)-1 {
			sb.WriteString("---\n\n")
		}
	}
	
	return sb.String()
}

func formatShowResponse(intent QueryIntent, hints []*ai.Hint) string {
	var sb strings.Builder
	
	title := intent.Condition
	if title == "" {
		title = "matching"
	}
	
	sb.WriteString(fmt.Sprintf("📋 Showing %s %d %s issues:\n\n", title, len(hints), intent.Resource))
	
	for i, hint := range hints {
		icon := "🔵"
		if hint.Severity == ai.SeverityCritical {
			icon = "🔴"
		} else if hint.Severity == ai.SeverityWarning {
			icon = "🟡"
		}
		
		sb.WriteString(fmt.Sprintf("%s **%d.** %s\n", icon, i+1, hint.Summary))
	}
	
	if len(hints) > 5 {
		sb.WriteString(fmt.Sprintf("\n... and %d more. Use 'r8s analyze' for full details.\n", len(hints)-5))
	}
	
	return sb.String()
}

func formatWhichResponse(intent QueryIntent, hints []*ai.Hint) string {
	var sb strings.Builder
	
	resourceName := intent.Resource
	if resourceName == "" {
		resourceName = "resources"
	}
	
	condition := intent.Condition
	if condition == "" {
		condition = "affected"
	}
	
	sb.WriteString(fmt.Sprintf("🎯 Found %d %s that are %s:\n\n", len(hints), resourceName, condition))
	
	for _, hint := range hints {
		// Extract name from hint if possible
		name := hint.PatternID
		if hint.Metadata["PodName"] != "" {
			name = hint.Metadata["PodName"]
		}
		
		sb.WriteString(fmt.Sprintf("• %s\n", name))
	}
	
	return sb.String()
}

func formatWhatResponse(intent QueryIntent, hints []*ai.Hint) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("🤔 Analysis: What is wrong with %s?\n\n", intent.Resource))
	
	if len(hints) == 0 {
		sb.WriteString("No issues detected for this resource.\n")
		return sb.String()
	}
	
	for _, hint := range hints {
		icon := "🔵"
		if hint.Severity == ai.SeverityCritical {
			icon = "🔴"
		} else if hint.Severity == ai.SeverityWarning {
			icon = "🟡"
		}
		
		sb.WriteString(fmt.Sprintf("%s **%s**\n", icon, hint.Summary))
		sb.WriteString(fmt.Sprintf("   %s\n\n", hint.Explanation))
	}
	
	return sb.String()
}

func formatGeneralResponse(hints []*ai.Hint) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("Found %d relevant issues:\n\n", len(hints)))
	
	for _, hint := range hints {
		icon := "🔵"
		if hint.Severity == ai.SeverityCritical {
			icon = "🔴"
		} else if hint.Severity == ai.SeverityWarning {
			icon = "🟡"
		}
		
		sb.WriteString(fmt.Sprintf("%s %s\n", icon, hint.Summary))
	}
	
	return sb.String()
}

func formatNoResultsResponse(intent QueryIntent) string {
	return fmt.Sprintf(`❓ No %s issues found matching your query.

This could mean:
• No issues detected for this resource type
• The issue has been resolved since the bundle was collected
• The query pattern isn't fully matching (Day 10 v1 limitations)

Try:
• 'r8s analyze %s' for a full bundle report
• Different wording for your question
• Checking if the bundle includes recent data`, 
		intent.Condition, "<bundle-path>")
}

func formatUnknownResponse() string {
	return `❓ I didn't understand that question.

Supported query patterns (Day 10 v1):
• "why is <pod> crashing?"
• "show me <pattern> issues"  
• "which <pods> are <state>?"
• "what is wrong with <resource>?"

Try using one of these patterns, or use:
• 'r8s analyze <bundle>' for full analysis
• 'r8s export <bundle> --format=markdown' for a report`
}