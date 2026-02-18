// Package cmd implements the CLI commands for r8s.
// Sprint 8: r8s generate prompt - AI prompt generation from bundle analysis.
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/Rancheroo/r8s/internal/bundle"
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

	// Check bundle health
	health, err := bundle.CheckHealth(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to analyze bundle: %w", err)
	}

	// Generate prompt based on format
	var prompt string
	switch promptFormat {
	case "terminal":
		prompt = generateTerminalPrompt(bundlePath, health)
	case "script":
		prompt = generateScriptPrompt(bundlePath, health)
	default:
		prompt = generateChatbotPrompt(bundlePath, health)
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
func generateChatbotPrompt(bundlePath string, health *bundle.HealthCheck) string {
	var prompt string

	prompt += "# Kubernetes Bundle Analysis Request\n\n"
	prompt += fmt.Sprintf("**Generated:** %s\n", time.Now().Format("2006-01-02 15:04:05"))
	prompt += fmt.Sprintf("**Bundle:** %s\n", bundlePath)
	prompt += fmt.Sprintf("**Bundle Type:** %s\n", health.BundleType)
	prompt += fmt.Sprintf("**Health:** %.0f%% complete\n\n", health.Completeness)

	if !health.IsValid {
		prompt += "⚠️ **Warning:** This bundle is missing critical files. Analysis may be limited.\n\n"
	}

	// High impact issues
	highImpact := health.GetHighImpactMissing()
	if len(highImpact) > 0 {
		prompt += "## Missing Critical Data\n\n"
		for _, missing := range highImpact {
			prompt += fmt.Sprintf("- **%s** (%s): %s\n", missing.Path, missing.Category, missing.Impact)
		}
		prompt += "\n"
	}

	if len(health.MissingFiles) > 0 {
		prompt += "## Bundle Completeness\n\n"
		prompt += fmt.Sprintf("Present: %d/%d files (%.0f%%)\n\n", health.FoundFiles, health.TotalFiles, health.Completeness)
		
		prompt += "### Missing Files by Category\n\n"
		for cat, stats := range health.Categories {
			if stats.Missing > 0 {
				prompt += fmt.Sprintf("- %s: %d/%d files missing\n", cat, stats.Missing, stats.Total)
			}
		}
		prompt += "\n"
	} else {
		prompt += "✓ **Bundle is complete** — all expected files present.\n\n"
	}

	prompt += "---\n\n"
	prompt += "## Request\n\n"
	prompt += "Please analyze this Kubernetes support bundle and provide:\n\n"
	prompt += "1. **Root cause analysis** for any detected issues\n"
	prompt += "2. **Step-by-step remediation** with specific kubectl commands\n"
	prompt += "3. **Prevention recommendations** to avoid recurrence\n"
	prompt += "4. **Priority order** for any fixes needed\n\n"
	
	if health.Completeness < 100 {
		prompt += "Note: This bundle may be incomplete. Recommendations should note where missing data limits analysis.\n\n"
	}

	prompt += "Assume I have kubectl access to the cluster. Be specific with resource names and commands."

	return prompt
}

// generateTerminalPrompt creates a command-focused prompt for terminal AI
func generateTerminalPrompt(bundlePath string, health *bundle.HealthCheck) string {
	var prompt string

	prompt += "# R8S Terminal Analysis\n\n"
	prompt += fmt.Sprintf("Bundle: %s\n", bundlePath)
	prompt += fmt.Sprintf("Health: %.0f%% | Type: %s\n\n", health.Completeness, health.BundleType)

	// High impact issues only
	highImpact := health.GetHighImpactMissing()
	if len(highImpact) > 0 {
		prompt += "## Issues Found\n\n"
		for i, missing := range highImpact {
			importanceStr := importanceToString(missing.Importance)
			prompt += fmt.Sprintf("%d. [%s] %s: %s\n",
				i+1,
				importanceStr,
				missing.Path,
				missing.Impact)
		}
		prompt += "\n"
	}

	prompt += "---\n\n"
	prompt += "Generate kubectl commands to investigate these issues. For each, provide:\n\n"
	prompt += "1. **Diagnostic command** (what to run to investigate)\n"
	prompt += "2. **Fix command** (kubectl patch, edit, or recreate)\n"
	prompt += "3. **Verify command** (how to confirm the fix)\n\n"
	prompt += "Format as copy-paste ready commands with brief comments.\n\n"
	prompt += "```bash\n# Example format:\n"
	prompt += "# Check status\n"
	prompt += "kubectl get pods -n <namespace>\n\n"
	prompt += "# Apply fix\n"
	prompt += "kubectl patch ...\n\n"
	prompt += "# Verify\n"
	prompt += "kubectl get pods -n <namespace> | grep <pod>\n"
	prompt += "```\n"

	return prompt
}

// generateScriptPrompt creates a bash script generation prompt
func generateScriptPrompt(bundlePath string, health *bundle.HealthCheck) string {
	var prompt string

	prompt += "#!/bin/bash\n"
	prompt += "# R8S Auto-Remediation Script Generator\n"
	prompt += fmt.Sprintf("# Bundle: %s\n", bundlePath)
	prompt += fmt.Sprintf("# Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	prompt += "set -euo pipefail\n\n"
	prompt += "echo \"=== R8S Remediation Script ===\"\n"
	prompt += fmt.Sprintf("echo \"Bundle Health: %.0f%%\"\n\n", health.Completeness)

	// List issues
	if len(health.MissingFiles) > 0 {
		prompt += "# Issues detected in bundle analysis:\n"
		for _, missing := range health.GetHighImpactMissing() {
			prompt += fmt.Sprintf("# - %s: %s\n", missing.Path, missing.Impact)
		}
		prompt += "\n"
	}

	prompt += "---\n\n"
	prompt += "Generate a bash script that:\n"
	prompt += "1. Checks each issue with kubectl diagnostics\n"
	prompt += "2. Applies safe fixes with user confirmation prompts\n"
	prompt += "3. Verifies fixes with health checks\n"
	prompt += "4. Outputs clear status messages\n\n"
	prompt += "Requirements:\n"
	prompt += "- Add 'read -p' prompts before destructive operations\n"
	prompt += "- Include error handling and rollback options\n"
	prompt += "- Make it idempotent (safe to run multiple times)\n"
	prompt += "- End with verification that issues are resolved\n"

	return prompt
}

// importanceToString converts FileImportance to human-readable string
func importanceToString(imp bundle.FileImportance) string {
	switch imp {
	case bundle.ImportanceCritical:
		return "critical"
	case bundle.ImportanceHigh:
		return "high"
	case bundle.ImportanceMedium:
		return "medium"
	case bundle.ImportanceLow:
		return "low"
	default:
		return "unknown"
	}
}
