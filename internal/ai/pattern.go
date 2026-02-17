// Package ai provides pattern matching for Kubernetes issues.
// Sprint 8: AI Pattern Engine - YAML-driven pattern detection.
package ai

import (
	"fmt"
	"strings"
)

// Pattern represents a detection pattern for a specific Kubernetes issue
type Pattern struct {
	ID          string   `yaml:"id"`          // Unique identifier
	Name        string   `yaml:"name"`        // Human-readable name
	Category    string   `yaml:"category"`    // e.g., "OOM", "ImagePull", "CrashLoop"
	Severity    Severity `yaml:"severity"`    // Critical, Warning, Info
	Keywords    []string `yaml:"keywords"`    // Strings to match (all must match)
	Description string   `yaml:"description"` // What this pattern detects
	Suggestion  string   `yaml:"suggestion"`  // Recommended fix
}

// Severity represents issue severity
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityWarning  Severity = "warning"
	SeverityInfo     Severity = "info"
)

// MatchResult represents the outcome of pattern matching
type MatchResult struct {
	Matched     bool
	PatternID   string
	PatternName string
	Severity    Severity
	Message     string
	Confidence  float64 // 0.0 to 1.0
}

// Matcher provides pattern matching functionality
type Matcher struct {
	pattern Pattern
}

// NewMatcher creates a new pattern matcher
func NewMatcher(p Pattern) *Matcher {
	return &Matcher{pattern: p}
}

// Match checks if the content matches the pattern
// Sprint 8: Simple keyword matching (80/20) - no regex for now
func (m *Matcher) Match(content string) MatchResult {
	content = strings.ToLower(content)
	
	// Count how many keywords matched
	matches := 0
	for _, keyword := range m.pattern.Keywords {
		if strings.Contains(content, strings.ToLower(keyword)) {
			matches++
		}
	}
	
	// All keywords must match
	if matches < len(m.pattern.Keywords) {
		return MatchResult{Matched: false}
	}
	
	// Calculate simple confidence based on match quality
	confidence := 1.0
	if len(content) > 1000 {
		confidence = 0.9 // Lower confidence for very long content
	}
	
	return MatchResult{
		Matched:     true,
		PatternID:   m.pattern.ID,
		PatternName: m.pattern.Name,
		Severity:    m.pattern.Severity,
		Message:     m.detectedMessage(),
		Confidence:  confidence,
	}
}

// detectedMessage returns a human-readable detection message
func (m *Matcher) detectedMessage() string {
	return fmt.Sprintf("[%s] %s: %s", 
		strings.ToUpper(string(m.pattern.Severity)),
		m.pattern.Name,
		m.pattern.Description)
}

// BuiltinPatterns contains the built-in pattern definitions
// Sprint 8: 3 patterns only (80/20) - OOMKill, ImagePullBackOff, CrashLoop
var BuiltinPatterns = []Pattern{
	{
		ID:          "oomkill",
		Name:        "OOMKill Detected",
		Category:    "OOM",
		Severity:    SeverityCritical,
		Keywords:    []string{"out of memory", "oomkill", "oom_kill_process", "killed process"},
		Description: "Container was killed due to memory limits",
		Suggestion:  "Increase memory limit or optimize application memory usage",
	},
	{
		ID:          "imagepullbackoff",
		Name:        "ImagePullBackOff",
		Category:    "Image",
		Severity:    SeverityWarning,
		Keywords:    []string{"imagepullbackoff", "pull access denied", "failed to pull image", "image not found"},
		Description: "Cannot pull container image from registry",
		Suggestion:  "Check image name, registry credentials, and network connectivity",
	},
	{
		ID:          "crashloopbackoff",
		Name:        "CrashLoopBackOff",
		Category:    "Crash",
		Severity:    SeverityCritical,
		Keywords:    []string{"crashloopbackoff", "back-off restarting", "crash loop"},
		Description: "Container repeatedly crashing and restarting",
		Suggestion:  "Check container logs for application errors and exit codes",
	},
}

// PatternRegistry manages pattern definitions
type PatternRegistry struct {
	patterns []Pattern
}

// NewRegistry creates a new pattern registry with built-in patterns
func NewRegistry() *PatternRegistry {
	return &PatternRegistry{
		patterns: BuiltinPatterns,
	}
}

// Register adds a new pattern to the registry
func (r *PatternRegistry) Register(p Pattern) error {
	// Validate pattern
	if p.ID == "" {
		return fmt.Errorf("pattern ID is required")
	}
	if p.Name == "" {
		return fmt.Errorf("pattern name is required")
	}
	if len(p.Keywords) == 0 {
		return fmt.Errorf("at least one keyword is required")
	}
	
	r.patterns = append(r.patterns, p)
	return nil
}

// GetByID retrieves a pattern by ID
func (r *PatternRegistry) GetByID(id string) (Pattern, bool) {
	for _, p := range r.patterns {
		if p.ID == id {
			return p, true
		}
	}
	return Pattern{}, false
}

// GetByCategory retrieves all patterns in a category
func (r *PatternRegistry) GetByCategory(category string) []Pattern {
	var result []Pattern
	for _, p := range r.patterns {
		if p.Category == category {
			result = append(result, p)
		}
	}
	return result
}

// GetAll returns all patterns
func (r *PatternRegistry) GetAll() []Pattern {
	return r.patterns
}

// Analyze scans content against all patterns and returns matches
func (r *PatternRegistry) Analyze(content string) []MatchResult {
	var matches []MatchResult
	
	for _, pattern := range r.patterns {
		matcher := NewMatcher(pattern)
		result := matcher.Match(content)
		if result.Matched {
			matches = append(matches, result)
		}
	}
	
	return matches
}
