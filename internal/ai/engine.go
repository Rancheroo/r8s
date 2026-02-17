package ai

import (
	"sort"
	"sync"
	"time"
)

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
				PatternID:   p.ID,
				PatternName: p.Name,
				Severity:    p.Severity,
				Category:    p.Category,
				Message:     result.Message,
				Source:      metadata.SourceType,
				Context:     result.Context,
				Timestamp:   time.Now(),
				Suggestion:  p.Suggestion,
				Confidence:  result.Confidence,
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
		summary.SeverityCount[f.Severity]++
	}

	return summary
}
