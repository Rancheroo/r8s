// Package cmd implements the CLI commands for r8s.
// Sprint 11 Day 10: Natural Language Queries v1
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"regexp"

	"github.com/Rancheroo/r8s/internal/ai"
	"github.com/Rancheroo/r8s/internal/bundle"
	"github.com/Rancheroo/r8s/internal/rancher"
	"github.com/Rancheroo/r8s/internal/ui"
)

// askCmd represents the ask command
var askCmd = &cobra.Command{
	Use:   "ask [bundle-path] [question]",
	Short: "Ask natural language questions about your bundle",
	Long: `Ask questions in plain English about issues in your bundle.

This is the natural language interface for r8s. It parses your question,
queries the bundle data, and answers with context-specific responses.

🧱 BUILDING A QUERY (THE LEGO BLOCKS):
  Entities:  pod, node, certificate, image, pvc, service
  States:    crashing, pending, expired, not ready, failed, running, ready
  Issues:    oomkill, imagepullbackoff, crashloopbackoff, etcd-latency

EXAMPLES:
  # Root Cause Analysis
  r8s ask ./bundle/ "what caused the outage?"
  r8s ask ./bundle/ "why is nginx-pod crashing?"

  # Issue Discovery
  r8s ask ./bundle/ "show me imagepullbackoff issues"
  r8s ask ./bundle/ "which certificates are expired?"

  # Resource Status
  r8s ask ./bundle/ "which nodes are not ready?"
  r8s ask ./bundle/ "which pods are running?"
  r8s ask ./bundle/ "what is wrong with worker-1?"

SUPPORTED PATTERNS:
  <issue> types:
    - oomkill
    - crashloopbackoff
    - imagepullbackoff
    - etcd-latency
    - etcd-corruption
    - certificate-expired
    - dns-failure
    - cni-error

LIMITATIONS (v1):
- Single-fact questions only (no compound queries)
- No follow-up questions
- Context is limited to bundle contents`,
	RunE: runAsk,
}

func init() {
	rootCmd.AddCommand(askCmd)
}

// runAsk executes the ask command
func runAsk(cmd *cobra.Command, args []string) error {
	// If no args provided, show guide instead of error
	if len(args) == 0 {
		ui.ShowCmdUsage("ask", "r8s ask [bundle-path] [question]", cmd.Long)
		return nil
	}

	// Validate we have the right number of args
	if len(args) < 2 {
		ui.ShowUsageError("ask", "r8s ask <bundle> <question>")
		return ui.NewUsageError("ask", "r8s ask <bundle> <question>")
	}

	bundlePath := args[0]
	question := strings.Join(args[1:], " ")

	// Check if first arg looks like a question (starts with quote) and second is a path
	// This helps detect wrong argument order - but only for very clear cases
	// to avoid rejecting valid bundle paths
	if len(args) > 1 && isLikelyQuestion(bundlePath) && !isLikelyPath(bundlePath) && isLikelyPath(args[1]) {
		// Double-check: first arg is clearly a question, second is clearly a path
		// and the first arg contains actual question markers (?, why, what, how)
		questionLower := strings.ToLower(bundlePath)
		if strings.Contains(questionLower, "?") ||
			strings.HasPrefix(questionLower, "why ") ||
			strings.HasPrefix(questionLower, "what ") ||
			strings.HasPrefix(questionLower, "how ") {
			// First arg looks like a question, not a path
			ui.ShowUsageError("ask", "r8s ask <bundle> <question>")
			fmt.Fprintf(os.Stderr, "\nIt looks like you might have the arguments in the wrong order.\n")
			fmt.Fprintf(os.Stderr, "Expected:    r8s ask ./bundle/ \"your question\"\n")
			fmt.Fprintf(os.Stderr, "You entered: r8s ask \"%s\" %s\n\n", bundlePath, args[1])
			return ui.NewUsageError("ask", "r8s ask <bundle> <question>")
		}
	}

	// Validate bundle path
	if _, err := os.Stat(bundlePath); err != nil {
		if os.IsNotExist(err) {
			ui.ShowBundleNotFoundError(bundlePath)
			return &ExitCodeError{Code: ExitError, Message: fmt.Sprintf("bundle not found: %s", bundlePath)}
		}
		return fmt.Errorf("cannot access bundle path: %w", err)
	}

	// Parse the question
	intent := parseQueryIntent(question)

	// STRICT STRUCTURE OPTIMIZATION (Elon's Laws):
	// Handle direct state queries (running/ready) using raw data parsing
	// instead of inferring from AI hints. This is less dynamic but more reliable.
	if response, handled, err := handleStateQuery(bundlePath, intent); handled {
		if err != nil {
			return fmt.Errorf("failed to query state: %w", err)
		}
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
	
	// Check for "image" resource using word boundaries to avoid matching "imagepull"
	// This matches "image", "images", "docker" but not "imagepull"
	imageRegex := regexp.MustCompile(`\b(image|images|docker)\b`)
	if imageRegex.MatchString(q) {
		intent.Resource = "image"
	}

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

	if strings.Contains(q, "node") {
		intent.Resource = "node"
	}

	// Extract condition
	switch {
	case strings.Contains(q, "crash") || strings.Contains(q, "crashloop") || strings.Contains(q, "restarting"):
		intent.Condition = "crashing"
	case strings.Contains(q, "imagepull") || strings.Contains(q, "can't pull") || strings.Contains(q, "pull error"):
		intent.Condition = "imagepull"
	case strings.Contains(q, "pending") || strings.Contains(q, "stuck") || strings.Contains(q, "scheduling"):
		intent.Condition = "pending"
	case strings.Contains(q, "expire"):
		intent.Condition = "expired"
	case strings.Contains(q, "not ready") || strings.Contains(q, "unhealthy") || strings.Contains(q, "dead"):
		intent.Condition = "notready"
	case strings.Contains(q, "oom") || strings.Contains(q, "memory") || strings.Contains(q, "killed"):
		intent.Condition = "oom"
	case strings.Contains(q, "fail") || strings.Contains(q, "broken") || strings.Contains(q, "error"):
		intent.Condition = "failed"
	case strings.Contains(q, "ready") || strings.Contains(q, "running") || strings.Contains(q, "healthy"):
		intent.Condition = "ready"
	case strings.Contains(q, "slow") || strings.Contains(q, "latency") || strings.Contains(q, "lag") || strings.Contains(q, "timeout"):
		intent.Condition = "latency"
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
	// If resource matches but condition doesn't, filter carefully
	// Sprint 11: Fix imprecise matching logic

	// 1. Filter by resource type
	if intent.Resource != "" {
		isPodPattern := isPatternForResource(hint.PatternID, "pod")
		isNodePattern := isPatternForResource(hint.PatternID, "node")
		
		if intent.Resource == "pod" && !isPodPattern {
			return false
		}
		if intent.Resource == "node" && !isNodePattern {
			return false
		}
	}

	// 2. Filter by condition
	if intent.Condition != "" {
		// Use strict condition matching
		patternLower := strings.ToLower(hint.PatternID)
		
		switch intent.Condition {
		case "crashing":
			return strings.Contains(patternLower, "crash")
		case "oom":
			return strings.Contains(patternLower, "oom")
		case "pending":
			return strings.Contains(patternLower, "pending")
		case "terminating":
			return strings.Contains(patternLower, "terminating")
		case "imagepull":
			return strings.Contains(patternLower, "imagepull")
		case "notready":
			return strings.Contains(patternLower, "not-ready")
		case "failed", "failing":
			// Failing should match any non-info issue
			return hint.Severity == ai.SeverityCritical || hint.Severity == ai.SeverityWarning
		case "ready", "running":
			// We now handle positive state queries ("which pods are running") in handleStateQuery.
			// If we are here, it means we are looking for issues.
			// Issues are by definition NOT ready/running.
			return false
		default:
			// Generic match
			return strings.Contains(patternLower, intent.Condition)
		}
	}

	return true
}

// isPatternForResource checks if a pattern ID matches a specific resource type.
// It uses keyword matching to associate patterns with resource types like "pod" or "node".
func isPatternForResource(patternID, resourceType string) bool {
	pid := strings.ToLower(patternID)
	if resourceType == "pod" {
		return strings.Contains(pid, "crash") || 
			strings.Contains(pid, "oom") || 
			strings.Contains(pid, "image") || 
			strings.Contains(pid, "pending") || 
			strings.Contains(pid, "terminating")
	}
	if resourceType == "node" {
		return strings.Contains(pid, "node") || 
			strings.Contains(pid, "disk") || 
			strings.Contains(pid, "memory") || 
			strings.Contains(pid, "pid") ||
			strings.Contains(pid, "pressure")
	}
	return true
}

// Response formatters

// formatWhyResponse formats the response for "why" type queries, focusing on explanations and suggestions.
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

// formatShowResponse formats the response for "show" type queries, listing matching issues.
func formatShowResponse(intent QueryIntent, hints []*ai.Hint) string {
	var sb strings.Builder

	title := intent.Condition
	if title == "" {
		title = "matching"
	}

	// Use "resources" for positive states, "issues" for negative ones
	term := "issues"
	if intent.Condition == "running" || intent.Condition == "ready" {
		term = "resources"
	}

	sb.WriteString(fmt.Sprintf("📋 Showing %s %d %s %s:\n\n", title, len(hints), intent.Resource, term))

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

// formatWhichResponse formats the response for "which" type queries, listing affected resource names.
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

	// Terminology fix: Don't imply "issues" for neutral queries
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

// formatWhatResponse formats the response for "what" type queries, providing a general analysis of a resource.
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

// formatGeneralResponse formats a generic response when the query type is not specific.
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

// formatNoResultsResponse creates a helpful message when no issues are found matching the query.
func formatNoResultsResponse(intent QueryIntent) string {
	// Customize message based on query type
	term := "issues"
	if intent.Condition == "running" || intent.Condition == "ready" {
		term = "resources"
	}

	msg := fmt.Sprintf("❓ No %s found matching your query for '%s'.\n", term, intent.Condition)
	if intent.Condition == "" {
		msg = fmt.Sprintf("❓ No relevant %s found matching your query.\n", term)
	}
	if intent.Resource != "" {
		if intent.Condition != "" {
			msg = fmt.Sprintf("❓ No %s %s found matching '%s'.\n", intent.Resource, term, intent.Condition)
		} else {
			msg = fmt.Sprintf("❓ No %s %s found matching your query.\n", intent.Resource, term)
		}
	}

	return fmt.Sprintf(`%s
This could mean:
• No issues detected for this resource/condition
• The issue has been resolved since the bundle was collected
• The query pattern isn't fully matching (Day 10 v1 limitations)

Try:
• 'r8s analyze %s' for a full bundle report
• Different wording for your question
• Checking if the bundle includes recent data

SUGGESTED PATTERNS:
• Entities: pod, node, certificate, image, pvc, service
• Issues: oomkill, crashloop, imagepull, etcd, cni, dns`,
		msg, "<bundle-path>")
}

// formatUnknownResponse returns a help message for queries that could not be parsed.
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

// isLikelyQuestion checks if text looks like a question
func isLikelyQuestion(text string) bool {
	lower := strings.ToLower(text)
	questionWords := []string{"why", "what", "which", "show", "find", "list", "how", "is", "are", "can", "will", "?"}
	for _, word := range questionWords {
		if strings.HasPrefix(lower, word+" ") || strings.HasPrefix(lower, word+"?") || strings.HasSuffix(lower, word) {
			return true
		}
	}
	return false
}

// isLikelyPath checks if text looks like a file path
func isLikelyPath(text string) bool {
	// Common path indicators
	if strings.HasPrefix(text, "./") || strings.HasPrefix(text, "/") || strings.HasPrefix(text, "..") {
		return true
	}
	// Check for file separators
	if strings.Contains(text, "/") || strings.Contains(text, "\\") {
		return true
	}
	return false
}

// handleStateQuery processes requests for current state (running/ready) using raw parsers.
// It returns the response string, a boolean indicating if the query was handled, and any error.
func handleStateQuery(bundlePath string, intent QueryIntent) (string, bool, error) {
	// We only handle "which", "show", or "what" queries
	if intent.Type != "which" && intent.Type != "show" && intent.Type != "what" {
		return "", false, nil
	}

	// We only handle positive state assertions
	if intent.Condition != "ready" && intent.Condition != "running" {
		return "", false, nil
	}

	var sb strings.Builder

	if intent.Resource == "pod" {
		pods, err := bundle.ParsePods(bundlePath)
		if err != nil {
			// If file missing or parse error, fall back to analysis
			return "", false, nil
		}

		var matching []string
		for _, pod := range pods {
			if isPodReady(pod) {
				matching = append(matching, pod.Name)
			}
		}

		sb.WriteString(fmt.Sprintf("🎯 Found %d pod%s that are ready:\n\n", len(matching), pluralize(len(matching))))
		for _, name := range matching {
			sb.WriteString(fmt.Sprintf("• %s\n", name))
		}
		return sb.String(), true, nil
	}

	if intent.Resource == "node" {
		nodes, err := bundle.ParseNodes(bundlePath)
		if err != nil {
			return "", false, nil
		}

		var matching []string
		for _, node := range nodes {
			if strings.ToLower(node.Status) == "ready" {
				matching = append(matching, node.Name)
			}
		}

		sb.WriteString(fmt.Sprintf("🎯 Found %d node%s that are ready:\n\n", len(matching), pluralize(len(matching))))
		for _, name := range matching {
			sb.WriteString(fmt.Sprintf("• %s\n", name))
		}
		return sb.String(), true, nil
	}

	return "", false, nil
}

// isPodReady checks if a pod is considered ready based on its status and ready condition.
func isPodReady(pod rancher.Pod) bool {
	// 1. Check Status
	if pod.KubectlStatus != "Running" && pod.KubectlStatus != "Completed" {
		return false
	}

	// 2. Check Ready fraction (e.g. "1/1", "2/2")
	parts := strings.Split(pod.KubectlReady, "/")
	if len(parts) == 2 && parts[0] == parts[1] {
		return true
	}
	
	// Completed pods are "ready" in the sense that they succeeded
	if pod.KubectlStatus == "Completed" {
		return true
	}

	return false
}

// pluralize returns "s" if count is not 1, for simple pluralization.
func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
