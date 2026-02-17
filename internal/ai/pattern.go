// Package ai provides intelligent pattern detection for log analysis.
// It uses simple keyword matching (80/20 rule) to detect common Kubernetes issues.
package ai

import (
	"strings"
	"time"
)

// Pattern represents a detectable issue pattern in logs or events
type Pattern struct {
	// ID is the unique identifier for this pattern (e.g., "oom-kill")
	ID string `json:"id" yaml:"id"`

	// Name is the human-readable name (e.g., "Out of Memory Kill")
	Name string `json:"name" yaml:"name"`

	// Description explains what this pattern detects
	Description string `json:"description" yaml:"description"`

	// Severity indicates the impact level: critical, high, medium, low
	Severity string `json:"severity" yaml:"severity"`

	// Category groups related patterns (e.g., "resource", "image", "crash")
	Category string `json:"category" yaml:"category"`

	// Keywords are the search terms that trigger this pattern (case-insensitive)
	Keywords []string `json:"keywords" yaml:"keywords"`

	// Suggestion provides remediation advice
	Suggestion string `json:"suggestion" yaml:"suggestion"`

	// DocumentationURL links to more information
	DocumentationURL string `json:"documentation_url,omitempty" yaml:"documentation_url,omitempty"`
}

// Finding represents a detected pattern instance
type Finding struct {
	// PatternID references the matched pattern
	PatternID string `json:"pattern_id"`

	// PatternName is the human-readable name
	PatternName string `json:"pattern_name"`

	// Severity of the finding (copied from pattern)
	Severity string `json:"severity"`

	// Category of the finding (copied from pattern)
	Category string `json:"category"`

	// Message describes what was found
	Message string `json:"message"`

	// Source indicates where the finding came from (e.g., "dmesg", "pod-logs", "events")
	Source string `json:"source"`

	// Context contains relevant surrounding information
	Context FindingContext `json:"context"`

	// Timestamp when the finding was detected
	Timestamp time.Time `json:"timestamp"`

	// Suggestion provides remediation advice (copied from pattern)
	Suggestion string `json:"suggestion"`

	// Confidence is a simple score: 1.0 = certain, 0.5 = possible
	Confidence float64 `json:"confidence"`
}

// FindingContext provides context about where a pattern was found
type FindingContext struct {
	// PodName identifies the affected pod
	PodName string `json:"pod_name,omitempty"`

	// Namespace of the affected pod
	Namespace string `json:"namespace,omitempty"`

	// ContainerName identifies the specific container
	ContainerName string `json:"container_name,omitempty"`

	// NodeName where the issue occurred
	NodeName string `json:"node_name,omitempty"`

	// ResourceInfo contains resource-related context (limits, requests)
	ResourceInfo map[string]string `json:"resource_info,omitempty"`

	// RawSnippet shows the raw log/event that triggered the match
	RawSnippet string `json:"raw_snippet,omitempty"`
}

// MatchResult contains the outcome of a pattern match attempt
type MatchResult struct {
	// Matched indicates if the pattern was detected
	Matched bool

	// Confidence score (1.0 = high confidence, 0.5 = possible)
	Confidence float64

	// Message describing what was found (if Matched)
	Message string

	// Context contains additional matching details
	Context FindingContext
}

// Matcher is the interface for pattern matching implementations
type Matcher interface {
	// Match attempts to detect the pattern in the given content
	Match(content string, metadata MatchMetadata) MatchResult
}

// MatchMetadata provides context for pattern matching
type MatchMetadata struct {
	// SourceType indicates the type of content ("logs", "events", "dmesg")
	SourceType string

	// PodName if the content is pod-specific
	PodName string

	// Namespace of the pod
	Namespace string

	// ContainerName if container-specific
	ContainerName string

	// NodeName where the content originated
	NodeName string
}

// KeywordMatcher implements simple keyword-based pattern matching
type KeywordMatcher struct {
	Pattern Pattern
}

// Match implements the Matcher interface using keyword matching
func (km *KeywordMatcher) Match(content string, metadata MatchMetadata) MatchResult {
	contentLower := strings.ToLower(content)
	matchCount := 0
	matchedKeywords := []string{}

	for _, keyword := range km.Pattern.Keywords {
		keywordLower := strings.ToLower(keyword)
		if strings.Contains(contentLower, keywordLower) {
			matchCount++
			matchedKeywords = append(matchedKeywords, keyword)
		}
	}

	if matchCount == 0 {
		return MatchResult{Matched: false}
	}

	// Calculate confidence based on keyword matches
	// All keywords = 1.0, half = 0.7, single = 0.5
	confidence := 0.5
	if matchCount >= len(km.Pattern.Keywords) {
		confidence = 1.0
	} else if matchCount >= len(km.Pattern.Keywords)/2 {
		confidence = 0.7
	}

	// Build context from the matched content
	context := FindingContext{
		PodName:       metadata.PodName,
		Namespace:     metadata.Namespace,
		ContainerName: metadata.ContainerName,
		NodeName:      metadata.NodeName,
		RawSnippet:    extractSnippet(content, matchedKeywords),
	}

	return MatchResult{
		Matched:    true,
		Confidence: confidence,
		Message:    buildMessage(km.Pattern, content, matchedKeywords),
		Context:    context,
	}
}

// extractSnippet pulls a relevant portion of the content around matched keywords
func extractSnippet(content string, keywords []string) string {
	// Simple extraction: first 200 chars or around first keyword
	if len(content) <= 200 {
		return content
	}

	// Find first keyword occurrence and extract context around it
	contentLower := strings.ToLower(content)
	for _, keyword := range keywords {
		idx := strings.Index(contentLower, strings.ToLower(keyword))
		if idx != -1 {
			start := idx - 50
			if start < 0 {
				start = 0
			}
			end := idx + 150
			if end > len(content) {
				end = len(content)
			}
			return content[start:end]
		}
	}

	// Fallback: first 200 characters
	return content[:200]
}

// buildMessage creates a human-readable finding message
func buildMessage(pattern Pattern, content string, matchedKeywords []string) string {
	// Use pattern name as base message
	message := pattern.Name

	// Add specificity based on pattern type
	switch pattern.ID {
	case "oom-kill":
		if strings.Contains(strings.ToLower(content), "invoked oom-killer") {
			message = "Node-level OOM killer invoked"
		} else if strings.Contains(strings.ToLower(content), "killed process") {
			message = "Process killed due to memory exhaustion"
		}
	case "image-pull-backoff":
		if strings.Contains(strings.ToLower(content), "imagepullbackoff") ||
			strings.Contains(strings.ToLower(content), "errimagepull") {
			message = "Container image pull failed"
		}
	case "crash-loop-backoff":
		if strings.Contains(strings.ToLower(content), "crashloopbackoff") {
			message = "Container in crash loop"
		}
	}

	return message
}

// NewKeywordMatcher creates a keyword matcher for a pattern
func NewKeywordMatcher(pattern Pattern) *KeywordMatcher {
	return &KeywordMatcher{Pattern: pattern}
}

// Built-in patterns for common Kubernetes issues
var BuiltinPatterns = []Pattern{
	{
		ID:          "oom-kill",
		Name:        "Out of Memory Kill",
		Description: "Process terminated due to memory exhaustion",
		Severity:    "critical",
		Category:    "resource",
		Keywords:    []string{"oom-killer", "killed process", "out of memory", "oom kill"},
		Suggestion:  "Increase memory limit or optimize application memory usage. Check for memory leaks.",
	},
	{
		ID:          "image-pull-backoff",
		Name:        "Image Pull Failure",
		Description: "Container image cannot be pulled from registry",
		Severity:    "high",
		Category:    "image",
		Keywords:    []string{"imagepullbackoff", "errimagepull", "failed to pull image", "image pull"},
		Suggestion:  "Verify image name/tag, check registry credentials, and ensure network connectivity to registry.",
	},
	{
		ID:          "crash-loop-backoff",
		Name:        "Crash Loop Backoff",
		Description: "Container repeatedly crashes and restarts",
		Severity:    "high",
		Category:    "crash",
		Keywords:    []string{"crashloopbackoff", "back-off restarting", "restarting failed container"},
		Suggestion:  "Check container logs for application errors. Verify startup commands and dependencies.",
	},
	{
		ID:          "pod-evicted",
		Name:        "Pod Evicted",
		Description: "Pod was evicted from node due to resource pressure",
		Severity:    "high",
		Category:    "resource",
		Keywords:    []string{"evicted", "pod was evicted", "the node had condition"},
		Suggestion:  "Check node resource pressure. Free disk space or memory on the node.",
	},
	{
		ID:          "disk-pressure",
		Name:        "Disk Pressure",
		Description: "Node is experiencing disk pressure",
		Severity:    "medium",
		Category:    "resource",
		Keywords:    []string{"disk pressure", "diskpressure", "no space left"},
		Suggestion:  "Free up disk space on the node. Clean up unused images and logs.",
	},
}

// SeverityOrder defines the priority order for severities
var SeverityOrder = map[string]int{
	"critical": 4,
	"high":     3,
	"medium":   2,
	"low":      1,
}

// IsHigherSeverity returns true if severity a is higher priority than b
func IsHigherSeverity(a, b string) bool {
	return SeverityOrder[a] > SeverityOrder[b]
}
