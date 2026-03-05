package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Rancheroo/r8s/internal/ai"
	"github.com/Rancheroo/r8s/internal/bundle"
	"github.com/Rancheroo/r8s/internal/rancher"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate [subcommand]",
	Short: "Generate outputs from bundle analysis",
	Long: `Generate AI prompts, scripts, or other outputs from bundle analysis.

SUBCOMMANDS:
  prompt    Generate AI-ready prompts for troubleshooting

EXAMPLES:
  # Generate AI prompt for the entire bundle
  r8s generate prompt ./bundle/

  # Generate terminal-focused prompt (kubectl commands)
  r8s generate prompt ./bundle/ --format=terminal

  # Generate prompt for a specific finding
  r8s generate prompt ./bundle/ --severity=critical

  # Save to file
  r8s generate prompt ./bundle/ --output=analysis.txt`,
}

// promptCmd represents the generate prompt subcommand
var promptCmd = &cobra.Command{
	Use:   "prompt [bundle-path]",
	Short: "Generate AI-ready troubleshooting prompt",
	Long: `Generate a prompt for AI assistants (Claude, GPT, Grok) based on bundle analysis.

This command analyzes the bundle and creates a structured prompt with:
  • Bundle health summary
  • Detected issues and patterns
  • Recommended kubectl commands
  • Context for troubleshooting

FORMATS:
  chatbot   - Analysis-focused for web interfaces (default)
  terminal  - Command-focused for terminal AI (Claude Code, Warp)
  script    - Generates bash remediation script

EXAMPLES:
  # Default chatbot format
  r8s generate prompt ./bundle/

  # Terminal format (kubectl commands)
  r8s generate prompt ./bundle/ --format=terminal

  # Pipe directly to Claude Code
  r8s generate prompt ./bundle/ --format=terminal | claude

  # Generate fix script
  r8s generate prompt ./bundle/ --format=script > fix.sh`,
	Args: cobra.ExactArgs(1),
	RunE: runGeneratePrompt,
}

var (
	promptFormat   string // Output format: chatbot, terminal, script
	promptOutput   string // Output file (default: stdout)
	promptSeverity string // Filter by severity: critical, warning, all
)

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.AddCommand(promptCmd)

	promptCmd.Flags().StringVarP(&promptFormat, "format", "f", "chatbot", "Output format: chatbot, terminal, script")
	promptCmd.Flags().StringVarP(&promptOutput, "output", "o", "", "Output file (default: stdout)")
	promptCmd.Flags().StringVarP(&promptSeverity, "severity", "s", "all", "Filter by severity: critical, warning, all")
}

// runGeneratePrompt executes the generate prompt command
func runGeneratePrompt(cmd *cobra.Command, args []string) error {
	bundlePath := args[0]

	// Validate bundle path
	if _, err := os.Stat(bundlePath); err != nil {
		return fmt.Errorf("cannot access bundle path: %w", err)
	}

	// Load full bundle
	b, err := bundle.LoadFromPath(bundlePath, bundle.ImportOptions{})
	if err != nil {
		return fmt.Errorf("failed to load bundle: %w", err)
	}

	// Manually parse Nodes since they aren't in Bundle struct (yet)
	nodes, _ := bundle.ParseNodes(b.ExtractPath)

	// Run AI Analysis (reuse logic from analyze/export)
	analyzer := ai.NewAnalyzer()
	bundleContent, _ := collectBundleContent(b.ExtractPath, b.Health.BundleType)
	results, _ := analyzer.AnalyzeMultiple(bundleContent, ai.AnalysisOptions{
		MinSeverity: ai.SeverityWarning,
	})
	mergedHints := mergeAnalysisResults(results)

	// Generate prompt based on format
	var prompt string
	switch promptFormat {
	case "terminal":
		prompt = generateTerminalPrompt(b, nodes)
	case "script":
		prompt = generateScriptPrompt(b, nodes)
	default:
		prompt = generateChatbotPrompt(b, nodes, mergedHints)
	}

	// Output
	if promptOutput != "" {
		if err := os.WriteFile(promptOutput, []byte(prompt), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "✓ Prompt saved to %s\n", promptOutput)
	} else {
		fmt.Println(prompt)
	}

	return nil
}

// generateChatbotPrompt creates an analysis-focused prompt
func generateChatbotPrompt(b *bundle.Bundle, nodes []bundle.NodeInfo, hints []*ai.Hint) string {
	var prompt strings.Builder

	prompt.WriteString("# Kubernetes Bundle Analysis Request\n\n")
	prompt.WriteString(fmt.Sprintf("**Generated:** %s\n", time.Now().Format("2006-01-02 15:04:05")))
	prompt.WriteString(fmt.Sprintf("**Bundle:** %s\n", b.Path))
	prompt.WriteString(fmt.Sprintf("**Bundle Type:** %s\n", b.Health.BundleType))
	prompt.WriteString(fmt.Sprintf("**Health:** %d%% complete\n\n", b.Health.Percentage()))

	if b.Health.Percentage() < 50 {
		prompt.WriteString("⚠️ **Warning:** This bundle is missing critical files. Analysis may be limited.\n\n")
	}

	// AI Detected Issues
	if len(hints) > 0 {
		prompt.WriteString("## Detected Issues (AI Analysis)\n\n")
		for _, hint := range hints {
			icon := "🔴"
			if hint.Severity == ai.SeverityWarning {
				icon = "🟡"
			}
			prompt.WriteString(fmt.Sprintf("### %s %s\n", icon, hint.Summary))
			prompt.WriteString(fmt.Sprintf("**Pattern:** %s\n", hint.PatternID))
			if hint.Explanation != "" {
				prompt.WriteString(fmt.Sprintf("%s\n", hint.Explanation))
			}
			prompt.WriteString("\n")
		}
	}

	// Node Status
	if len(nodes) > 0 {
		prompt.WriteString("## Cluster Nodes\n\n")
		prompt.WriteString("| Name | Status |\n|---|---|\n")
		for _, node := range nodes {
			prompt.WriteString(fmt.Sprintf("| %s | %s |\n", node.Name, node.Status))
		}
		prompt.WriteString("\n")
	}

	// Failing Pods
	var failingPods []rancher.Pod
	// Get parsed pods from bundle
	kubectlPods, _ := bundle.ParsePods(b.ExtractPath)
	for _, pod := range kubectlPods {
		if pod.KubectlStatus != "Running" && pod.KubectlStatus != "Completed" {
			failingPods = append(failingPods, pod)
		}
	}

	if len(failingPods) > 0 {
		prompt.WriteString("## Failing Pods\n\n")
		prompt.WriteString("| Namespace | Name | Status | Restarts |\n|---|---|---|---|\n")
		for _, pod := range failingPods {
			prompt.WriteString(fmt.Sprintf("| %s | %s | %s | %d |\n", pod.NamespaceID, pod.Name, pod.KubectlStatus, pod.RestartCount))
		}
		prompt.WriteString("\n")

		// Logs for failing pods (limit to first 3)
		prompt.WriteString("## Relevant Logs (Tail)\n\n")
		count := 0
		for _, pod := range failingPods {
			if count >= 3 {
				break
			}
			// Try to find log file
			logContent := getLogsForPod(b, pod.NamespaceID, pod.Name)
			if logContent != "" {
				prompt.WriteString(fmt.Sprintf("### Logs: %s/%s\n", pod.NamespaceID, pod.Name))
				prompt.WriteString("```\n")
				prompt.WriteString(logContent)
				prompt.WriteString("\n```\n\n")
				count++
			}
		}
	}

	// Warning Events
	prompt.WriteString("## Warning Events\n\n")
	events, _ := bundle.ParseEvents(b.ExtractPath)
	warningCount := 0
	for _, event := range events {
		if strings.Contains(strings.ToLower(event.Type), "warn") {
			if warningCount < 10 { // Limit to 10 warnings
				prompt.WriteString(fmt.Sprintf("- [%s] %s/%s: %s\n", event.LastSeen, event.Namespace, event.Name, event.Message))
				warningCount++
			}
		}
	}
	if warningCount == 0 {
		prompt.WriteString("No warning events found.\n\n")
	} else if warningCount >= 10 {
		prompt.WriteString("... (truncated)\n\n")
	}

	// Missing data
	if len(b.Health.MissingFiles) > 0 {
		prompt.WriteString("## Missing Data\n\n")
		for _, missing := range b.Health.MissingFiles {
			prompt.WriteString(fmt.Sprintf("- **%s**\n", missing))
		}
		prompt.WriteString("\n")
	}

	prompt.WriteString("---\n\n")
	prompt.WriteString("## Request\n\n")
	prompt.WriteString("Please analyze this Kubernetes support bundle and provide:\n\n")
	prompt.WriteString("1. **Root cause analysis** for the failing pods/nodes\n")
	prompt.WriteString("2. **Step-by-step remediation** with specific kubectl commands\n")
	prompt.WriteString("3. **Prevention recommendations**\n\n")

	prompt.WriteString("Assume I have kubectl access to the cluster. Be specific with resource names and commands.")

	return prompt.String()
}

// Helper to get tail of logs for a pod
func getLogsForPod(b *bundle.Bundle, namespace, name string) string {
	// Simple matching: look for log file with namespace and pod name
	// This uses the inventory from b.LogFiles
	for _, logFile := range b.LogFiles {
		// LogFileInfo has Namespace and PodName parsed
		if logFile.Namespace == namespace && strings.Contains(name, logFile.PodName) { // Contains because pod name in log might be prefix
			content, err := b.ReadLogFile(&logFile)
			if err != nil {
				continue
			}
			lines := strings.Split(string(content), "\n")
			if len(lines) > 20 {
				return strings.Join(lines[len(lines)-20:], "\n")
			}
			return string(content)
		}
	}
	return ""
}

// generateTerminalPrompt creates a command-focused prompt for terminal AI
func generateTerminalPrompt(b *bundle.Bundle, nodes []bundle.NodeInfo) string {
	health := b.Health
	var prompt string

	prompt += "# R8S Terminal Analysis\n\n"
	prompt += fmt.Sprintf("Bundle: %s\n", b.Path)
	prompt += fmt.Sprintf("Health: %d%% | Type: %s\n\n", health.Percentage(), health.BundleType)

	// Issues (Missing files for now)
	if len(health.MissingFiles) > 0 {
		prompt += "## Missing Data\n\n"
		for _, missing := range health.MissingFiles {
			prompt += fmt.Sprintf("- %s\n", missing)
		}
		prompt += "\n"
	}
	
	prompt += "---\n\n"
	prompt += "Generate kubectl commands to investigate these issues. For each, provide:\n\n"
	prompt += "1. **Diagnostic command** (what to run to investigate)\n"
	prompt += "2. **Fix command** (kubectl patch, edit, or recreate)\n"
	prompt += "3. **Verify command** (how to confirm the fix)\n\n"
	prompt += "Format as copy-paste ready commands with brief comments.\n\n"

	return prompt
}

// generateScriptPrompt creates a bash script generation prompt
func generateScriptPrompt(b *bundle.Bundle, nodes []bundle.NodeInfo) string {
	health := b.Health
	var prompt string

	prompt += "#!/bin/bash\n"
	prompt += "# R8S Auto-Remediation Script Generator\n"
	prompt += fmt.Sprintf("# Bundle: %s\n", b.Path)
	prompt += fmt.Sprintf("# Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	prompt += "set -euo pipefail\n\n"
	prompt += "echo \"=== R8S Remediation Script ===\"\n"
	prompt += fmt.Sprintf("echo \"Bundle Health: %d%%\"\n\n", health.Percentage())

	// List issues
	if len(health.MissingFiles) > 0 {
		prompt += "# Missing Data in bundle analysis:\n"
		for _, missing := range health.MissingFiles {
			prompt += fmt.Sprintf("# - %s\n", missing)
		}
		prompt += "\n"
	}

	prompt += "---\n\n"
	prompt += "Generate a bash script that:\n"
	prompt += "1. Checks each issue with kubectl diagnostics\n"
	prompt += "2. Applies safe fixes with user confirmation prompts\n"
	prompt += "3. Verifies fixes with health checks\n"
	prompt += "4. Outputs clear status messages\n\n"

	return prompt
}
