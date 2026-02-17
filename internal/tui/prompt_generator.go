package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/Rancheroo/r8s/internal/bundle"
)

// PromptGenerator creates AI-ready prompts from bundle context
type PromptGenerator struct {
	bundleHealth *bundle.BundleHealth
	findings     []AttentionItem
	bundlePath   string
}

// NewPromptGenerator creates a prompt generator
func NewPromptGenerator(health *bundle.BundleHealth, findings []AttentionItem, path string) *PromptGenerator {
	return &PromptGenerator{
		bundleHealth: health,
		findings:     findings,
		bundlePath:   path,
	}
}

// GenerateSupportPrompt creates a comprehensive prompt for troubleshooting
func (pg *PromptGenerator) GenerateSupportPrompt() string {
	var b strings.Builder

	// System context
	b.WriteString("# R8S Support Bundle Analysis\n\n")
	b.WriteString(fmt.Sprintf("**Generated:** %s\n", time.Now().Format("2006-01-02 15:04:05")))
	b.WriteString(fmt.Sprintf("**Bundle:** %s\n", pg.bundlePath))
	
	if pg.bundleHealth != nil {
		b.WriteString(fmt.Sprintf("**Bundle Health:** %s (%d%% complete)\n", 
			pg.bundleHealth.Color(), pg.bundleHealth.Percentage()))
		if pg.bundleHealth.BundleType != "" {
			b.WriteString(fmt.Sprintf("**Bundle Type:** %s\n", pg.bundleHealth.BundleType))
		}
	}
	
	b.WriteString("\n---\n\n")

	// Critical findings summary
	criticalCount := 0
	warningCount := 0
	for _, f := range pg.findings {
		if f.Severity == SeverityCritical {
			criticalCount++
		} else if f.Severity == SeverityWarning {
			warningCount++
		}
	}

	b.WriteString(fmt.Sprintf("## Summary: %d Critical, %d Warning Issues Detected\n\n", 
		criticalCount, warningCount))

	// Top issues
	b.WriteString("### Priority Issues:\n\n")
	shown := 0
	for _, f := range pg.findings {
		if f.Severity == SeverityCritical && shown < 5 {
			b.WriteString(fmt.Sprintf("%d. **%s** (%s)\n", shown+1, f.Title, f.Description))
			if f.Namespace != "" && f.Namespace != "cluster" {
				b.WriteString(fmt.Sprintf("   - Namespace: `%s`\n", f.Namespace))
			}
			if f.PodName != "" {
				b.WriteString(fmt.Sprintf("   - Pod: `%s`\n", f.PodName))
			}
			b.WriteString("\n")
			shown++
		}
	}

	b.WriteString("---\n\n")

	// Context for AI
	b.WriteString("## Context\n\n")
	b.WriteString("This is a Kubernetes support bundle analyzed with r8s (Rancher Support Tool). ")
	b.WriteString("The tool has detected patterns in logs, events, and cluster state that indicate issues.\n\n")

	if pg.bundleHealth != nil && len(pg.bundleHealth.Warnings) > 0 {
		b.WriteString("### Bundle Quality Notes:\n")
		for _, w := range pg.bundleHealth.Warnings {
			b.WriteString(fmt.Sprintf("- %s\n", w))
		}
		b.WriteString("\n")
	}

	b.WriteString("---\n\n")

	// The ask
	b.WriteString("## Request\n\n")
	b.WriteString("Please analyze these Kubernetes issues and provide:\n\n")
	b.WriteString("1. **Root Cause Analysis** for the critical issues above\n")
	b.WriteString("2. **Step-by-Step Remediation** with specific kubectl commands\n")
	b.WriteString("3. **Prevention Recommendations** to avoid recurrence\n")
	b.WriteString("4. **Priority Order** for fixes (what to do first)\n\n")
	
	b.WriteString("Assume I'm a Kubernetes administrator with cluster access. ")
	b.WriteString("Be specific with commands and configuration changes.\n")

	return b.String()
}

// GenerateFindingPrompt creates a focused prompt for a single finding
func (pg *PromptGenerator) GenerateFindingPrompt(item AttentionItem) string {
	var b strings.Builder

	b.WriteString("# R8S Finding Analysis\n\n")
	b.WriteString(fmt.Sprintf("**Issue:** %s\n", item.Title))
	b.WriteString(fmt.Sprintf("**Description:** %s\n", item.Description))
	b.WriteString(fmt.Sprintf("**Severity:** %s\n", severityString(item.Severity)))
	
	if item.Namespace != "" {
		b.WriteString(fmt.Sprintf("**Namespace:** `%s`\n", item.Namespace))
	}
	if item.PodName != "" {
		b.WriteString(fmt.Sprintf("**Pod:** `%s`\n", item.PodName))
	}
	if item.ContainerName != "" {
		b.WriteString(fmt.Sprintf("**Container:** `%s`\n", item.ContainerName))
	}

	b.WriteString("\n---\n\n")

	// Bundle context
	if pg.bundleHealth != nil {
		b.WriteString(fmt.Sprintf("**Bundle Health:** %d%% complete\n", pg.bundleHealth.Percentage()))
		b.WriteString(fmt.Sprintf("**Bundle Type:** %s\n\n", pg.bundleHealth.BundleType))
	}

	b.WriteString("## Request\n\n")
	b.WriteString(fmt.Sprintf("This Kubernetes %s issue was detected in cluster logs/events. ", item.ResourceType))
	b.WriteString("Please provide:\n\n")
	b.WriteString("1. **Likely Root Cause** based on the issue type\n")
	b.WriteString("2. **Immediate Fix** - specific commands to resolve\n")
	b.WriteString("3. **Verification** - how to confirm the fix worked\n")
	
	if item.Severity == SeverityCritical {
		b.WriteString("4. **Urgency Assessment** - is this causing downtime?\n")
	}

	return b.String()
}

// GenerateComparisonPrompt creates a prompt comparing current vs previous state
func (pg *PromptGenerator) GenerateComparisonPrompt(previousFindings []AttentionItem) string {
	var b strings.Builder

	b.WriteString("# R8S Cluster Health Comparison\n\n")
	b.WriteString(fmt.Sprintf("**Analysis Time:** %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	// Current state
	currentCritical := 0
	currentWarning := 0
	for _, f := range pg.findings {
		if f.Severity == SeverityCritical {
			currentCritical++
		} else if f.Severity == SeverityWarning {
			currentWarning++
		}
	}

	// Previous state
	prevCritical := 0
	prevWarning := 0
	for _, f := range previousFindings {
		if f.Severity == SeverityCritical {
			prevCritical++
		} else if f.Severity == SeverityWarning {
			prevWarning++
		}
	}

	b.WriteString("## State Comparison\n\n")
	b.WriteString(fmt.Sprintf("| Metric | Previous | Current | Change |\n"))
	b.WriteString(fmt.Sprintf("|--------|----------|---------|--------|\n"))
	b.WriteString(fmt.Sprintf("| Critical | %d | %d | %+d |\n", prevCritical, currentCritical, currentCritical-prevCritical))
	b.WriteString(fmt.Sprintf("| Warning | %d | %d | %+d |\n\n", prevWarning, currentWarning, currentWarning-prevWarning))

	b.WriteString("## Request\n\n")
	b.WriteString("Please analyze this cluster health trend:\n\n")
	
	if currentCritical > prevCritical {
		b.WriteString("⚠️ **Critical issues have increased.**\n\n")
		b.WriteString("- What could cause this degradation?\n")
		b.WriteString("- What should be investigated first?\n")
	} else if currentCritical < prevCritical {
		b.WriteString("✅ **Critical issues have decreased.**\n\n")
		b.WriteString("- What fixes likely helped?\n")
		b.WriteString("- How to maintain this improvement?\n")
	}
	
	b.WriteString("\nProvide specific diagnostic steps and preventive measures.\n")

	return b.String()
}

// severityString converts severity to string
func severityString(s AttentionSeverity) string {
	switch s {
	case SeverityCritical:
		return "Critical"
	case SeverityWarning:
		return "Warning"
	case SeverityInfo:
		return "Info"
	default:
		return "Unknown"
	}
}
