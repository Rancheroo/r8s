// Package ai provides pattern matching and root cause analysis for Kubernetes issues.
// Sprint 11: Root Cause Hint Generation System
package ai

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// Hint represents a generated root cause hint
// Sprint 11: Human-readable explanation with remediation steps
type Hint struct {
	PatternID    string            // Pattern that triggered this hint
	Severity     Severity          // Issue severity
	Confidence   Confidence        // Detection confidence
	Summary      string            // Short description
	Explanation  string            // Detailed explanation
	Suggestion   string            // How to fix
	Command      string            // kubectl command to run
	References   []string          // Links to docs
	Metadata     map[string]string // Additional context from pattern match
}

// HintGenerator produces root cause hints from pattern matches
// Sprint 11: Template-based hint generation with variable substitution
type HintGeneratorV2 struct {
	templates *template.Template
}

// NewHintGenerator creates a new hint generator
func NewHintGenerator() *HintGeneratorV2 {
	return &HintGeneratorV2{
		templates: template.New("hints"),
	}
}

// Generate creates a hint from a pattern match result
// Sprint 11: Applies template with match metadata
func (hg *HintGeneratorV2) Generate(match MatchResultV2, pattern PatternV2) (*Hint, error) {
	if !match.Matched {
		return nil, fmt.Errorf("cannot generate hint for unmatched pattern")
	}

	// Extract values from pattern's HintGenerator
	hgTemplate := pattern.HintGenerator.Template
	suggestion := pattern.HintGenerator.Suggestion
	command := pattern.HintGenerator.Command
	references := pattern.HintGenerator.References

	// Apply template substitution
	summary, err := hg.applyTemplate(hgTemplate, match.Metadata)
	if err != nil {
		// Fallback to generic message if template fails
		summary = fmt.Sprintf("[%s] %s detected", 
			strings.ToUpper(string(match.Severity)), pattern.Name)
	}

	// Apply command template substitution (if variables exist)
	if command != "" {
		command, _ = hg.applyTemplate(command, match.Metadata)
	}

	// Build explanation
	explanation := hg.buildExplanation(pattern, match)

	return &Hint{
		PatternID:   match.PatternID,
		Severity:    match.Severity,
		Confidence:  match.Confidence,
		Summary:     summary,
		Explanation: explanation,
		Suggestion:  suggestion,
		Command:     command,
		References:  references,
		Metadata:    match.Metadata,
	}, nil
}

// applyTemplate applies Go template with metadata variables
func (hg *HintGeneratorV2) applyTemplate(tmpl string, data map[string]string) (string, error) {
	if tmpl == "" {
		return "", nil
	}

	// Check if template has variables
	if !strings.Contains(tmpl, "{{") {
		return tmpl, nil
	}

	// Parse and execute template
	t, err := template.New("hint").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execute error: %w", err)
	}

	result := buf.String()
	// Sprint 11: Detect missing template variables (prevents <no value> in output)
	if strings.Contains(result, "<no value>") {
		return "", fmt.Errorf("template contains unresolved variables: %s", result)
	}

	return result, nil
}

// buildExplanation creates a detailed explanation from match data
func (hg *HintGeneratorV2) buildExplanation(pattern PatternV2, match MatchResultV2) string {
	var parts []string

	// Add pattern description
	parts = append(parts, pattern.Description)

	// Add evidence if available
	if len(match.Evidence) > 0 {
		parts = append(parts, "\nEvidence:")
		for i, evidence := range match.Evidence {
			if i < 3 { // Limit to 3 evidence snippets
				parts = append(parts, fmt.Sprintf("  • %s", evidence))
			}
		}
	}

	// Add correlations if present
	if len(match.Correlated) > 0 {
		parts = append(parts, "\nRelated Issues:")
		for _, corrID := range match.Correlated {
			// Find correlation message
			for _, corr := range pattern.Correlations {
				if corr.PatternID == corrID {
					parts = append(parts, fmt.Sprintf("  • %s", corr.Message))
				}
			}
		}
	}

	// Add metadata context
	if len(match.Metadata) > 0 {
		parts = append(parts, "\nContext:")
		for key, value := range match.Metadata {
			if value != "" {
				parts = append(parts, fmt.Sprintf("  • %s: %s", key, value))
			}
		}
	}

	return strings.Join(parts, "\n")
}

// GenerateAll creates hints for all matched patterns
func (hg *HintGeneratorV2) GenerateAll(matches []MatchResultV2, registry *PatternRegistryV2) []*Hint {
	var hints []*Hint

	for _, match := range matches {
		pattern, found := registry.GetByID(match.PatternID)
		if !found {
			continue
		}

		// Add evidence to metadata for template use
		if match.Metadata == nil {
			match.Metadata = make(map[string]string)
		}
		if len(match.Evidence) > 0 {
			match.Metadata["Evidence"] = match.Evidence[0]
		}

		hint, err := hg.Generate(match, pattern)
		if err != nil {
			continue // Skip failed hints
		}
		hints = append(hints, hint)
	}

	return hints
}

// FormatHint formats a hint for display
func (hg *HintGeneratorV2) FormatHint(hint *Hint) string {
	var parts []string

	// Header with severity
	icon := hg.severityIcon(hint.Severity)
	parts = append(parts, fmt.Sprintf("%s [%s] %s", icon, hint.Confidence, hint.Summary))

	// Explanation
	if hint.Explanation != "" {
		parts = append(parts, "\n"+hint.Explanation)
	}

	// Suggestion
	if hint.Suggestion != "" {
		parts = append(parts, "\n💡 Suggestion:")
		parts = append(parts, "  "+hint.Suggestion)
	}

	// Command
	if hint.Command != "" {
		parts = append(parts, "\n🔧 Command:")
		parts = append(parts, "  $ "+hint.Command)
	}

	// References
	if len(hint.References) > 0 {
		parts = append(parts, "\n📚 References:")
		for _, ref := range hint.References {
			parts = append(parts, "  • "+ref)
		}
	}

	return strings.Join(parts, "\n")
}

// severityIcon returns an emoji for the severity level
func (hg *HintGeneratorV2) severityIcon(s Severity) string {
	switch s {
	case SeverityCritical:
		return "🔴"
	case SeverityWarning:
		return "🟡"
	case SeverityInfo:
		return "🔵"
	default:
		return "⚪"
	}
}

// HintFormatter provides different output formats for hints
type HintFormatter struct{}

// NewHintFormatter creates a new hint formatter
func NewHintFormatter() *HintFormatter {
	return &HintFormatter{}
}

// FormatMarkdown formats hints as Markdown
func (hf *HintFormatter) FormatMarkdown(hints []*Hint) string {
	var parts []string

	parts = append(parts, "# Root Cause Analysis Report")
	parts = append(parts, fmt.Sprintf("\nFound %d issue(s)\n", len(hints)))

	for i, hint := range hints {
		parts = append(parts, fmt.Sprintf("\n## %d. %s [%s]", i+1, hint.Summary, hint.Severity))
		parts = append(parts, fmt.Sprintf("\n**Confidence:** %s\n", hint.Confidence))

		if hint.Explanation != "" {
			parts = append(parts, "### Explanation")
			parts = append(parts, hint.Explanation)
		}

		if hint.Suggestion != "" {
			parts = append(parts, "\n### Suggestion")
			parts = append(parts, hint.Suggestion)
		}

		if hint.Command != "" {
			parts = append(parts, "\n### Command")
			parts = append(parts, "```bash")
			parts = append(parts, hint.Command)
			parts = append(parts, "```")
		}

		if len(hint.References) > 0 {
			parts = append(parts, "\n### References")
			for _, ref := range hint.References {
				parts = append(parts, fmt.Sprintf("- <%s>", ref))
			}
		}
	}

	return strings.Join(parts, "\n")
}

// FormatJSON formats hints as JSON (for programmatic use)
func (hf *HintFormatter) FormatJSON(hints []*Hint) string {
	// Simple JSON formatting (production would use encoding/json)
	var parts []string
	parts = append(parts, "[")
	
	for i, hint := range hints {
		parts = append(parts, "  {")
		parts = append(parts, fmt.Sprintf(`    "pattern_id": %q,`, hint.PatternID))
		parts = append(parts, fmt.Sprintf(`    "severity": %q,`, hint.Severity))
		parts = append(parts, fmt.Sprintf(`    "confidence": %q,`, hint.Confidence))
		parts = append(parts, fmt.Sprintf(`    "summary": %q,`, hint.Summary))
		parts = append(parts, fmt.Sprintf(`    "explanation": %q,`, hint.Explanation))
		parts = append(parts, fmt.Sprintf(`    "suggestion": %q,`, hint.Suggestion))
		parts = append(parts, fmt.Sprintf(`    "command": %q,`, hint.Command))
		parts = append(parts, fmt.Sprintf(`    "references": %q,`, strings.Join(hint.References, ", ")))
		
		if i < len(hints)-1 {
			parts = append(parts, "  },")
		} else {
			parts = append(parts, "  }")
		}
	}
	
	parts = append(parts, "]")
	return strings.Join(parts, "\n")
}

// FilterHints filters hints by severity
func FilterHints(hints []*Hint, severities ...Severity) []*Hint {
	var filtered []*Hint
	severityMap := make(map[Severity]bool)
	for _, s := range severities {
		severityMap[s] = true
	}

	for _, hint := range hints {
		if severityMap[hint.Severity] {
			filtered = append(filtered, hint)
		}
	}
	return filtered
}

// SortHintsBySeverity sorts hints: Critical > Warning > Info
func SortHintsBySeverity(hints []*Hint) []*Hint {
	// Simple bubble sort for severity
	severityOrder := map[Severity]int{
		SeverityCritical: 0,
		SeverityWarning:  1,
		SeverityInfo:     2,
	}

	result := make([]*Hint, len(hints))
	copy(result, hints)

	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if severityOrder[result[i].Severity] > severityOrder[result[j].Severity] {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return result
}