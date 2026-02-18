// Package ai provides pattern matching for Kubernetes issues.
// Sprint 8: AI Pattern Engine - YAML-driven pattern detection.
// This file replaces the previous engine.go with a simpler implementation.
package ai

import (
	"sort"
	"sync"
	"time"
)

// MatchMetadata provides context about the content being analyzed
type MatchMetadata struct {
	SourceType    string    // e.g., "log", "event", "dmesg"
	SourceName    string    // e.g., pod name, file name
	Timestamp     time.Time
	NodeName      string    // Optional: for node-level sources
	PodName       string    // Optional: for pod-level sources
	Namespace     string    // Optional: Kubernetes namespace
	ContainerName string    // Optional: container name
}

// Finding represents a detected issue
type Finding struct {
	PatternID     string    `json:"pattern_id"`
	PatternName   string    `json:"pattern_name"`
	Severity      Severity  `json:"severity"`
	Category      string    `json:"category"`
	Message       string    `json:"message"`
	Source        string    `json:"source"`
	Namespace     string    `json:"namespace"`
	PodName       string    `json:"pod_name"`
	ContainerName string    `json:"container_name"`
	Timestamp     time.Time `json:"timestamp"`
	Suggestion    string    `json:"suggestion"`
	Confidence    float64   `json:"confidence"`
}

// KeywordMatcher wraps a Pattern for matching
type KeywordMatcher struct {
	pattern Pattern
}

// NewKeywordMatcher creates a new keyword matcher for a pattern
func NewKeywordMatcher(p Pattern) *KeywordMatcher {
	return &KeywordMatcher{pattern: p}
}

// Match checks if content matches the pattern using metadata
func (km *KeywordMatcher) Match(content string, metadata MatchMetadata) MatchResult {
	return NewMatcher(km.pattern).Match(content)
}

// Engine orchestrates log analysis and pattern detection
type Engine struct {
	patterns []Pattern
	mu       sync.RWMutex
}

// NewEngine creates a new analysis engine with builtin patterns
func NewEngine() *Engine {
	return &Engine{
		patterns: BuiltinPatterns,
	}
}

// Analyze processes content and returns a list of findings
func (e *Engine) Analyze(content string, metadata MatchMetadata) []Finding {
	e.mu.RLock()
	defer e.mu.RUnlock()

	findings := []Finding{}
	
	// Create matchers and run analysis
	for _, p := range e.patterns {
		matcher := NewKeywordMatcher(p)
		result := matcher.Match(content, metadata)
		
		if result.Matched {
			finding := Finding{
				PatternID:     p.ID,
				PatternName:   p.Name,
				Severity:      p.Severity,
				Category:      p.Category,
				Message:       result.Message,
				Source:        metadata.SourceType,
				Namespace:     metadata.Namespace,
				PodName:       metadata.PodName,
				ContainerName: metadata.ContainerName,
				Timestamp:     time.Now(),
				Suggestion:    p.Suggestion,
				Confidence:    result.Confidence,
			}
			findings = append(findings, finding)
		}
	}

	// Sort findings by severity (highest first)
	sort.Slice(findings, func(i, j int) bool {
		return IsHigherSeverity(findings[i].Severity, findings[j].Severity)
	})

	return findings
}

// AddPattern adds a custom pattern to the engine
func (e *Engine) AddPattern(p Pattern) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.patterns = append(e.patterns, p)
}

// GetPatterns returns the current list of patterns
func (e *Engine) GetPatterns() []Pattern {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	// Return a copy
	p := make([]Pattern, len(e.patterns))
	copy(p, e.patterns)
	return p
}

// IsHigherSeverity returns true if s1 is higher severity than s2
func IsHigherSeverity(s1, s2 Severity) bool {
	severityOrder := map[Severity]int{
		SeverityCritical: 3,
		SeverityWarning:  2,
		SeverityInfo:     1,
	}
	return severityOrder[s1] > severityOrder[s2]
}

// AnalysisSummary provides a high-level view of findings
type AnalysisSummary struct {
	TotalFindings int            `json:"total_findings"`
	SeverityCount map[string]int `json:"severity_count"`
	TopFinding    *Finding       `json:"top_finding,omitempty"`
}

// GetSummary generates a summary from a list of findings
func GetSummary(findings []Finding) AnalysisSummary {
	summary := AnalysisSummary{
		TotalFindings: len(findings),
		SeverityCount: make(map[string]int),
	}

	if len(findings) == 0 {
		return summary
	}

	summary.TopFinding = &findings[0]
	for _, f := range findings {
		summary.SeverityCount[string(f.Severity)]++
	}

	return summary
}
